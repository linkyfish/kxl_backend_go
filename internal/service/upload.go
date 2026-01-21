package service

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	kxlcfg "github.com/linkyfish/kxl_backend_go/internal/config"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/util"
)

const (
	UploadKindImage = "image"
	UploadKindVideo = "video"
)

type UploadService struct {
	cfg *kxlcfg.Config
}

func NewUploadService(cfg *kxlcfg.Config) *UploadService {
	return &UploadService{cfg: cfg}
}

func (s *UploadService) Upload(ctx context.Context, file *multipart.FileHeader, kind string) (string, error) {
	if file == nil {
		return "", kxlerrors.Validation("validation error: missing file")
	}

	maxBytes := s.maxSizeBytes(kind)
	if file.Size > maxBytes {
		return "", kxlerrors.Validation("validation error: file too large")
	}

	mime, err := detectMime(file)
	if err != nil {
		return "", kxlerrors.Validation("validation error: invalid upload")
	}

	ext := extensionForMime(kind, mime)
	if ext == "" {
		return "", kxlerrors.Validation("validation error: unsupported file type")
	}

	baseDir := s.uploadsDir()
	subdir := "images"
	if kind == UploadKindVideo {
		subdir = "videos"
	}
	now := time.Now().UTC()
	year := now.Format("2006")
	month := now.Format("01")

	uuid := util.NewUUID()
	rel := filepath.ToSlash(filepath.Join(subdir, year, month, uuid+"."+ext))
	dest := filepath.Join(baseDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", kxlerrors.Internal("io error: failed to create directory")
	}

	src, err := file.Open()
	if err != nil {
		return "", kxlerrors.Validation("validation error: invalid upload")
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return "", kxlerrors.Internal("io error: failed to write file")
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		_ = os.Remove(dest)
		return "", kxlerrors.Internal("io error: failed to write file")
	}

	return "/uploads/" + rel, nil
}

func (s *UploadService) DeleteRelativePath(ctx context.Context, relPath string) error {
	relPath = strings.TrimSpace(relPath)
	relPath = strings.TrimPrefix(relPath, "/")
	if relPath == "" || strings.Contains(relPath, "\x00") {
		return kxlerrors.Validation("validation error: invalid path")
	}

	parts := splitPath(relPath)
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return kxlerrors.Validation("validation error: invalid path")
		}
	}

	baseDir := s.uploadsDir()
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return kxlerrors.Internal("io error")
	}

	full := filepath.Join(baseAbs, filepath.FromSlash(relPath))
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return kxlerrors.Internal("io error")
	}

	// Disallow directory traversal.
	if !strings.HasPrefix(fullAbs, baseAbs+string(os.PathSeparator)) && fullAbs != baseAbs {
		return kxlerrors.NotFound("not found: file not found")
	}

	info, err := os.Stat(fullAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return kxlerrors.NotFound("not found: file not found")
		}
		return kxlerrors.Internal("io error")
	}
	if info.IsDir() {
		return kxlerrors.Validation("validation error: path is a directory")
	}

	if err := os.Remove(fullAbs); err != nil {
		return kxlerrors.New(50003, "io error: failed to delete file", http.StatusInternalServerError, nil)
	}
	return nil
}

func (s *UploadService) uploadsDir() string {
	if s != nil && s.cfg != nil && s.cfg.Uploads.Dir != "" {
		return strings.TrimSpace(s.cfg.Uploads.Dir)
	}
	return "uploads"
}

func (s *UploadService) maxSizeBytes(kind string) int64 {
	if s == nil || s.cfg == nil {
		if kind == UploadKindVideo {
			return 500 * 1024 * 1024
		}
		return 10 * 1024 * 1024
	}
	if kind == UploadKindVideo {
		if s.cfg.Uploads.VideoMaxBytes > 0 {
			return s.cfg.Uploads.VideoMaxBytes
		}
		return 500 * 1024 * 1024
	}
	if s.cfg.Uploads.ImageMaxBytes > 0 {
		return s.cfg.Uploads.ImageMaxBytes
	}
	return 10 * 1024 * 1024
}

func detectMime(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	m, err := mimetype.DetectReader(f)
	if err != nil {
		return "", err
	}
	return m.String(), nil
}

func extensionForMime(kind, mime string) string {
	if kind == UploadKindVideo {
		switch mime {
		case "video/mp4":
			return "mp4"
		case "video/webm":
			return "webm"
		default:
			return ""
		}
	}
	switch mime {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	default:
		return ""
	}
}

func splitPath(p string) []string {
	p = strings.ReplaceAll(p, "\\", "/")
	parts := strings.Split(p, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, part)
	}
	return out
}

