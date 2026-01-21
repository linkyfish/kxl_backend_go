package web

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/labstack/echo/v4"
)

var registerOnce sync.Once

func registerPongo2() {
	registerOnce.Do(func() {
		// Match the existing Rust Tera filter names used by templates.
		_ = pongo2.RegisterFilter("truncate_text", truncateTextFilter)
		_ = pongo2.RegisterFilter("format_date", formatDateFilter)
		_ = pongo2.RegisterFilter("highlight", highlightFilter)

		// Global helper used in templates.
		pongo2.Globals["current_year"] = func() int {
			return time.Now().UTC().Year()
		}
		pongo2.Globals["range"] = func(start, end int) []int {
			if end <= start {
				return []int{}
			}
			out := make([]int, 0, end-start)
			for i := start; i < end; i++ {
				out = append(out, i)
			}
			return out
		}
	})
}

type Renderer struct {
	set   *pongo2.TemplateSet
	mu    sync.RWMutex
	cache map[string]*pongo2.Template
}

func NewRenderer(templateDir string) (*Renderer, error) {
	registerPongo2()

	loader, err := pongo2.NewLocalFileSystemLoader(templateDir)
	if err != nil {
		return nil, err
	}
	set := pongo2.NewSet("kxl_backend_go", loader)
	return &Renderer{
		set:   set,
		cache: make(map[string]*pongo2.Template),
	}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tpl, err := r.get(name)
	if err != nil {
		return err
	}

	ctx := pongo2.Context{}
	switch v := data.(type) {
	case pongo2.Context:
		ctx = v
	case map[string]interface{}:
		for k, vv := range v {
			ctx[k] = vv
		}
	default:
		// Echo sometimes passes `map[string]any` or a struct; allow struct by storing it under "data".
		ctx["data"] = data
	}

	// Common values available to all templates.
	if c != nil && c.Request() != nil && c.Request().URL != nil {
		if _, ok := ctx["current_path"]; !ok {
			ctx["current_path"] = c.Request().URL.Path
		}
	}

	return tpl.ExecuteWriter(ctx, w)
}

func (r *Renderer) get(name string) (*pongo2.Template, error) {
	name = strings.TrimPrefix(name, "/")

	r.mu.RLock()
	if tpl, ok := r.cache[name]; ok {
		r.mu.RUnlock()
		return tpl, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if tpl, ok := r.cache[name]; ok {
		return tpl, nil
	}

	tpl, err := r.set.FromFile(name)
	if err != nil {
		return nil, err
	}
	r.cache[name] = tpl
	return tpl, nil
}

func truncateTextFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	s := in.String()
	n := int64(100)
	if param != nil && !param.IsNil() {
		if param.IsInteger() {
			n = int64(param.Integer())
		} else if param.IsFloat() {
			n = int64(param.Float())
		} else {
			// Accept numeric strings.
			if parsed, err := parseInt64(param.String()); err == nil {
				n = parsed
			}
		}
	}
	if n <= 0 {
		return pongo2.AsValue(""), nil
	}

	runes := []rune(s)
	if int64(len(runes)) <= n {
		return pongo2.AsValue(s), nil
	}
	out := string(runes[:n]) + "..."
	return pongo2.AsValue(out), nil
}

func formatDateFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	format := "%Y-%m-%d"
	if param != nil && !param.IsNil() {
		format = param.String()
	}
	layout := strftimeToGoLayout(format)

	// Pongo2 can pass time.Time through Value.Time().
	if in.IsTime() {
		return pongo2.AsValue(in.Time().Format(layout)), nil
	}

	// Attempt to parse RFC3339 or common date formats.
	raw := strings.TrimSpace(in.String())
	if raw == "" {
		return pongo2.AsValue(""), nil
	}
	if tm, err := time.Parse(time.RFC3339, raw); err == nil {
		return pongo2.AsValue(tm.Format(layout)), nil
	}
	if tm, err := time.Parse("2006-01-02", raw); err == nil {
		return pongo2.AsValue(tm.Format(layout)), nil
	}
	return pongo2.AsValue(raw), nil
}

func highlightFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	text := in.String()
	keyword := ""
	if param != nil && !param.IsNil() {
		keyword = strings.TrimSpace(param.String())
	}
	if keyword == "" || text == "" {
		return pongo2.AsValue(text), nil
	}
	highlighted := strings.ReplaceAll(text, keyword,
		fmt.Sprintf("<mark class=\"bg-primary-100 text-primary-800 px-1 rounded\">%s</mark>", keyword),
	)
	return pongo2.AsValue(highlighted), nil
}

func strftimeToGoLayout(format string) string {
	// Minimal subset used in our templates.
	replacer := strings.NewReplacer(
		"%Y", "2006",
		"%m", "01",
		"%d", "02",
		"%H", "15",
		"%M", "04",
		"%S", "05",
	)
	return replacer.Replace(format)
}

func parseInt64(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty")
	}
	var n int64
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("not int")
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}
