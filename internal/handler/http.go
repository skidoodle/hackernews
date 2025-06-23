package handler

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"

	"hackernews/internal/config"
	"hackernews/internal/hn"
	"hackernews/internal/view"
)

type App struct {
	Logger        *slog.Logger
	Config        *config.Config
	HackerNews    *hn.Client
	TemplateCache map[string]*template.Template
	StaticFS      fs.FS
}

func (a *App) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(a.StaticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("GET /new", a.storiesHandler("new"))
	mux.HandleFunc("GET /ask", a.storiesHandler("ask"))
	mux.HandleFunc("GET /show", a.storiesHandler("show"))
	mux.HandleFunc("GET /job", a.storiesHandler("job"))
	mux.HandleFunc("GET /item", a.itemHandler)

	mux.HandleFunc("GET /", a.catchAllHandler)

	return mux
}

func (a *App) catchAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		a.storiesHandler("top")(w, r)
		return
	}
	a.notFoundHandler(w, r)
}

func (a *App) notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Unknown."))
}

func (a *App) storiesHandler(storyType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		stories, err := a.HackerNews.GetStoriesForPage(r.Context(), storyType, page)
		if err != nil {
			a.Logger.Error("failed to get stories", "type", storyType, "page", page, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		tmpl, ok := a.TemplateCache["index.page.tmpl"]
		if !ok {
			a.Logger.Error("template not found: index.page.tmpl")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := &view.TemplateData{
			Stories:      stories,
			ActiveNav:    storyType,
			CurrentPage:  page,
			NextPage:     page + 1,
			ItemsPerPage: a.Config.HackerNewsAPI.ItemsPerPage,
		}
		view.Render(w, r, tmpl, data)
	}
}

func (a *App) itemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := a.HackerNews.GetItem(r.Context(), itemID)
	if err != nil {
		a.Logger.Error("failed to get item", "id", itemID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, ok := a.TemplateCache["item.page.tmpl"]
	if !ok {
		a.Logger.Error("template not found: item.page.tmpl")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := &view.TemplateData{
		Item:      item,
		ActiveNav: "",
	}
	view.Render(w, r, tmpl, data)
}
