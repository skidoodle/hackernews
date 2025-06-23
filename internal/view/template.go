package view

import (
	"bytes"
	"fmt"

	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"hackernews/internal/hn"
)

type TemplateData struct {
	Stories        []*hn.Item
	Item           *hn.Item
	User           *hn.User
	Submissions    []*hn.Item
	Comments       []*hn.Item
	ActiveNav      string
	ActiveUserView string
	CurrentPage    int
	NextPage       int
	ItemsPerPage   int
}

func formatDate(t int64) string {
	return time.Unix(t, 0).Format("January 2, 2006")
}

var functions = template.FuncMap{
	"host": func(s string) string {
		item := hn.Item{URL: s}
		return item.Host()
	},
	"timeAgo": func(t int64) string {
		item := hn.Item{Time: t}
		return item.TimeAgo()
	},
	"add": func(a, b int) int {
		return a + b
	},
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"rank": func(idx, page, itemsPerPage int) int {
		return idx + ((page - 1) * itemsPerPage) + 1
	},
	"formatText": FormatText,
	"formatDate": formatDate,
}

func NewTemplateCache(dir fs.FS) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(dir, "*.page.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFS(dir, page)
		if err != nil {
			return nil, err
		}

		layouts, err := fs.Glob(dir, "*.layout.tmpl")
		if err != nil {
			return nil, err
		}

		if len(layouts) > 0 {
			ts, err = ts.ParseFS(dir, layouts...)
			if err != nil {
				return nil, err
			}
		}

		cache[name] = ts
	}

	return cache, nil
}

func Render(w http.ResponseWriter, r *http.Request, t *template.Template, data *TemplateData) {
	buf := new(bytes.Buffer)
	err := t.ExecuteTemplate(buf, "base", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("error executing template: %v", err), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}
