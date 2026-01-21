#!/usr/bin/env bash
#
# Go Backend 交互式部署脚本
# 通过创建 tag 触发 GitHub Actions 工作流进行部署
#
# 使用方式: bash scripts/deploy.sh
#
set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
TAG_PREFIX="go-"
REMOTE_SERVER="154.53.61.227"
DEPLOY_PATH="/www/web_project/server/kxl_backend_go"

# 获取脚本所在目录的父目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 打印函数
print_header() {
    echo -e "${CYAN}=========================================="
    echo -e "$1"
    echo -e "==========================================${NC}"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在正确的目录
check_directory() {
    cd "$PROJECT_ROOT"

    if [[ ! -f "go.mod" ]] || [[ ! -f "cmd/api/main.go" ]]; then
        print_error "无法找到 Go 后端项目文件"
        print_info "请确保脚本位于 kxl_backend_go/scripts/ 目录下"
        exit 1
    fi
}

# 检查 git 状态
check_git_status() {
    print_header "检查 Git 状态"

    cd "$PROJECT_ROOT"

    if [[ -n $(git status --porcelain) ]]; then
        print_warning "有未提交的更改:"
        git status --short
        echo
        echo "请选择处理方式:"
        echo "  1) 自动提交这些更改并继续部署 (推荐)"
        echo "  2) 忽略这些更改继续部署 (本次部署不会包含这些改动)"
        echo "  3) 取消部署"
        echo

        read -p "请输入选项 [1-3] (默认 1): " dirty_choice
        dirty_choice="${dirty_choice:-1}"

        case "$dirty_choice" in
            1)
                auto_commit_changes
                ;;
            2)
                print_warning "将忽略未提交更改继续部署：本次 tag 指向的仍是当前 HEAD(未包含工作区改动)"
                ;;
            3)
                print_info "部署已取消"
                exit 0
                ;;
            *)
                print_error "无效选项: $dirty_choice"
                print_info "部署已取消"
                exit 1
                ;;
        esac
    else
        print_success "工作目录干净"
    fi
}

# 自动提交工作区更改（用于确保 tag/部署包含最新代码）
auto_commit_changes() {
    print_header "自动提交未提交更改"

    cd "$PROJECT_ROOT"

    # 基础校验：git 用户信息缺失时 commit 会失败，这里提前提示但不阻止继续尝试。
    local git_name git_email
    git_name="$(git config user.name || true)"
    git_email="$(git config user.email || true)"
    if [[ -z "$git_name" ]] || [[ -z "$git_email" ]]; then
        print_warning "未配置 git user.name/user.email，自动提交可能失败"
        print_info "可执行: git config user.name \"Your Name\" && git config user.email \"you@example.com\""
        echo
    fi

    # 允许在 detached HEAD 上提交（tag 仍可指向该提交），但给出提示。
    local current_branch
    current_branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "HEAD")"
    if [[ "$current_branch" == "HEAD" ]]; then
        print_warning "当前处于 detached HEAD，提交将不在任何分支上（但部署 tag 仍会指向该提交）"
        echo
    fi

    local default_msg="chore: deploy go backend"
    read -p "请输入提交说明(commit message) (默认: $default_msg): " commit_message
    commit_message="${commit_message:-$default_msg}"

    print_info "暂存所有更改..."
    git add -A

    if git diff --cached --quiet; then
        print_info "没有可提交的内容（可能是仅有未暂存/已忽略文件）"
        return 0
    fi

    print_info "创建提交..."
    git commit -m "$commit_message"

    print_success "已提交本地更改：$commit_message"
    echo
}

# 显示最近的提交
show_recent_commits() {
    print_header "最近的提交记录 (Go Backend)"

    cd "$PROJECT_ROOT"
    echo
    git log --oneline -10 --decorate
    echo
}

# 显示现有的 tags
show_existing_tags() {
    print_header "现有的 Go 部署 Tags"

    cd "$PROJECT_ROOT"

    local tags
    tags="$(git tag -l "${TAG_PREFIX}*" --sort=-version:refname 2>/dev/null | head -10 || true)"

    if [[ -n "$tags" ]]; then
        echo "$tags"
    else
        print_info "暂无 Go 部署 tags"
    fi
    echo
}

# 获取下一个版本号建议
suggest_next_version() {
    cd "$PROJECT_ROOT"

    local latest_tag
    latest_tag="$(git tag -l "${TAG_PREFIX}*" --sort=-version:refname 2>/dev/null | head -1 || true)"

    if [[ -n "$latest_tag" ]]; then
        # 提取版本号 (例如 go-1.0.0 -> 1.0.0)
        local version="${latest_tag#${TAG_PREFIX}}"

        # 尝试递增补丁版本
        if [[ "$version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
            local major="${BASH_REMATCH[1]}"
            local minor="${BASH_REMATCH[2]}"
            local patch="${BASH_REMATCH[3]}"
            echo "${TAG_PREFIX}${major}.${minor}.$((patch + 1))"
        elif [[ "$version" =~ ^([0-9]+)\.([0-9]+)$ ]]; then
            local major="${BASH_REMATCH[1]}"
            local minor="${BASH_REMATCH[2]}"
            echo "${TAG_PREFIX}${major}.$((minor + 1))"
        else
            echo "${TAG_PREFIX}1.0.0"
        fi
    else
        echo "${TAG_PREFIX}1.0.0"
    fi
}

# 验证 tag 格式
validate_tag() {
    local tag="$1"

    if [[ ! "$tag" =~ ^${TAG_PREFIX}[0-9]+(\.[0-9]+)*$ ]]; then
        print_error "无效的 tag 格式: $tag"
        print_info "Tag 必须以 '${TAG_PREFIX}' 开头，后跟版本号 (例如: ${TAG_PREFIX}1.0.0)"
        return 1
    fi

    # 检查 tag 是否已存在
    cd "$PROJECT_ROOT"
    if git rev-parse "$tag" >/dev/null 2>&1; then
        print_error "Tag '$tag' 已存在"
        return 1
    fi

    return 0
}

# 创建并推送 tag
create_and_push_tag() {
    local tag="$1"
    local message="$2"

    print_header "创建并推送 Tag"

    cd "$PROJECT_ROOT"

    # 创建带注释的 tag
    print_info "创建 tag: $tag"
    git tag -a "$tag" -m "$message"

    # 推送 tag
    print_info "推送 tag 到远程仓库..."
    git push origin "$tag"

    print_success "Tag '$tag' 已创建并推送"
}

# 显示部署信息
show_deployment_info() {
    local tag="$1"

    print_header "部署信息"

    echo -e "Tag:            ${GREEN}$tag${NC}"
    echo -e "目标服务器:      ${GREEN}$REMOTE_SERVER${NC}"
    echo -e "部署路径:        ${GREEN}$DEPLOY_PATH${NC}"
    echo
    print_info "GitHub Actions 工作流将自动触发部署"
    print_info "请前往 GitHub 仓库的 Actions 页面查看部署进度"
    echo
}

# 检查 GitHub Secrets 提示
check_secrets_reminder() {
    print_header "重要提示: GitHub Secrets 配置"

    echo "请确保在 Go 后端仓库的 GitHub Settings -> Secrets and variables -> Actions 中配置以下 secrets:"
    echo
    echo -e "  ${YELLOW}GO_DEPLOY_HOST${NC}     = $REMOTE_SERVER"
    echo -e "  ${YELLOW}GO_DEPLOY_PORT${NC}     = 22 (或您的 SSH 端口)"
    echo -e "  ${YELLOW}GO_DEPLOY_USER${NC}     = (SSH 用户名)"
    echo -e "  ${YELLOW}GO_DEPLOY_PASSWORD${NC} = (SSH 密码)"
    echo
    print_info "提示：如需与 PHP/Rust 后端并行部署，可在服务器的 .env 中设置不同的 SERVER_PORT"
    echo
}

# 主菜单
main_menu() {
    print_header "Go Backend 部署工具"

    echo "服务器: $REMOTE_SERVER"
    echo "部署路径: $DEPLOY_PATH"
    echo

    echo "请选择操作:"
    echo "  1) 创建新的部署 tag 并触发部署"
    echo "  2) 查看现有的 tags"
    echo "  3) 查看最近的提交记录"
    echo "  4) 手动触发部署 (使用现有 tag)"
    echo "  5) 查看 GitHub Secrets 配置说明"
    echo "  q) 退出"
    echo

    read -p "请输入选项 [1-5/q]: " choice

    case $choice in
        1)
            deploy_new_version
            ;;
        2)
            show_existing_tags
            read -p "按回车键继续..."
            main_menu
            ;;
        3)
            show_recent_commits
            read -p "按回车键继续..."
            main_menu
            ;;
        4)
            trigger_existing_tag
            ;;
        5)
            check_secrets_reminder
            read -p "按回车键继续..."
            main_menu
            ;;
        q|Q)
            print_info "再见!"
            exit 0
            ;;
        *)
            print_error "无效选项"
            main_menu
            ;;
    esac
}

# 部署新版本
deploy_new_version() {
    check_git_status
    show_recent_commits
    show_existing_tags

    local suggested_tag
    suggested_tag="$(suggest_next_version)"

    print_info "建议的下一个版本: $suggested_tag"
    echo

    read -p "请输入 tag 名称 (直接回车使用建议版本): " input_tag

    local tag="${input_tag:-$suggested_tag}"

    if ! validate_tag "$tag"; then
        read -p "按回车键返回..."
        main_menu
        return
    fi

    echo
    read -p "请输入部署说明 (可选): " deploy_message

    local message="${deploy_message:-Deploy Go backend $tag}"

    echo
    print_header "确认部署"
    echo -e "Tag:     ${GREEN}$tag${NC}"
    echo -e "说明:    ${GREEN}$message${NC}"
    echo -e "服务器:  ${GREEN}$REMOTE_SERVER${NC}"
    echo

    read -p "确认创建 tag 并触发部署？(y/N) " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        create_and_push_tag "$tag" "$message"
        show_deployment_info "$tag"
    else
        print_info "部署已取消"
    fi

    read -p "按回车键继续..."
    main_menu
}

# 使用现有 tag 触发部署
trigger_existing_tag() {
    show_existing_tags

    read -p "请输入要重新部署的 tag 名称: " tag

    if [[ -z "$tag" ]]; then
        print_error "未输入 tag"
        read -p "按回车键返回..."
        main_menu
        return
    fi

    cd "$PROJECT_ROOT"

    if ! git rev-parse "$tag" >/dev/null 2>&1; then
        print_error "Tag '$tag' 不存在"
        read -p "按回车键返回..."
        main_menu
        return
    fi

    echo
    print_info "要重新触发已存在的 tag，您可以:"
    echo "  1. 删除远程 tag 并重新推送"
    echo "  2. 在 GitHub Actions 页面手动触发 workflow_dispatch"
    echo

    read -p "是否删除并重新推送 tag '$tag'？(y/N) " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "删除远程 tag..."
        git push origin --delete "$tag" 2>/dev/null || true

        print_info "重新推送 tag..."
        git push origin "$tag"

        print_success "Tag '$tag' 已重新推送，部署工作流将触发"
    else
        print_info "操作已取消"
        print_info "您可以访问 GitHub Actions 页面手动触发部署"
    fi

    read -p "按回车键继续..."
    main_menu
}

# 脚本入口
main() {
    check_directory
    main_menu
}

# 运行
main "$@"
