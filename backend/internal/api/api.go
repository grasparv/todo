package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid"

	"todolist/internal/db"
	"todolist/internal/sse"
)

const (
	tokenList = "listID"
	tokenItem = "itemID"
)

type api struct {
	store  *db.DB
	logger *slog.Logger
	server *sse.Server
}

func New(ctx context.Context, logger *slog.Logger, store *db.DB) *api {
	return &api{
		store:  store,
		logger: logger,
		server: sse.New(ctx, logger),
	}
}

func (a *api) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.logger.Info("http enter", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		a.logger.Info("http leave", r.Method, r.URL.Path)
	})
}

func (a *api) Run() {
	// Create a new Chi router
	r := chi.NewRouter()
	r.Use(a.logRequest)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Handle the HTTP route for SSE
	r.Get("/events", a.handleEvents)

	// Lists management
	r.Post("/list", a.handleNewList)
	r.Route("/list/{"+tokenList+"}", func(r chi.Router) {
		r.Use(a.listContext)
		r.Delete("/", a.handleDeleteList)
		r.Put("/add", a.handleAddItem)
		r.Route("/item/{"+tokenItem+"}", func(r chi.Router) {
			r.Use(a.itemContext)
			r.Put("/", a.handleUpdateItem)
			r.Delete("/", a.handleDeleteItem)
		})
	})

	http.ListenAndServe(":3000", r)
}

func (a *api) listContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, tokenList)
		ctx := context.WithValue(r.Context(), tokenList, listID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *api) itemContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		itemID := chi.URLParam(r, tokenItem)
		ctx := context.WithValue(r.Context(), tokenItem, itemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *api) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	todos, err := a.store.GetTodoLists(r.Context())
	if err != nil {
		a.logger.Error("failed to get todo lists", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var events []ListEvent
	for _, todo := range todos {
		nlist := NewTodoList(todo)
		event := ListEvent{
			Type:     UpdateList,
			TodoList: &nlist,
		}
		events = append(events, event)
	}

	session, err := a.server.NewSession(w, r)
	if err != nil {
		a.logger.Error("failed to make SSE session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send all existing todo lists for a new client
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			a.logger.Error("failed to marshal event", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session.Send(data)
	}

	session.Wait()
}

func (a *api) handleNewList(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var t TodoList
	err = json.Unmarshal(data, &t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.ID = id
	for i, _ := range t.Items {
		id, err = uuid.NewV4()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t.Items[i].ID = id
	}

	if t.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if t.Owner == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = a.store.AddTodoList(r.Context(), t.Record())
	if err != nil {
		a.logger.Error("failed to add to store", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := ListEvent{UpdateList, &t}
	data, err = json.Marshal(event)
	if err == nil {
		a.server.Broadcast(data)
	} else {
		a.logger.Error("unable to broadcast event", "error", err)
	}
}

func (a *api) handleDeleteList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listID, ok := ctx.Value(tokenList).(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := uuid.FromString(listID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = a.store.RemoveTodoList(ctx, id)
	if err != nil {
		a.logger.Error("failed to remove from store", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	event := ListEvent{RemoveList, &TodoList{ID: id}}
	data, err := json.Marshal(event)
	if err == nil {
		a.server.Broadcast(data)
	} else {
		a.logger.Error("unable to broadcast event", "error", err)
	}
}

func (a *api) handleAddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listID, ok := ctx.Value(tokenList).(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ulistID, err := uuid.FromString(listID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	todo := TodoItem{
		ID:     id,
		List:   ulistID,
		Text:   "My new item",
		Marked: false,
	}
	err = a.store.AddTodoItem(ctx, ulistID, todo.Record())
	if err != nil {
		a.logger.Error("failed to add todo item", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := ItemEvent{AddItem, &todo}
	data, err := json.Marshal(event)
	if err == nil {
		a.server.Broadcast(data)
	} else {
		a.logger.Error("unable to broadcast event", "error", err)
	}
}

func (a *api) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var t TodoItem
	err = json.Unmarshal(data, &t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = a.store.UpdateTodoItem(ctx, t.Record())
	if err != nil {
		a.logger.Error("failed to update todo item", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := ItemEvent{UpdateItem, &t}
	data, err = json.Marshal(event)
	if err == nil {
		a.server.Broadcast(data)
	} else {
		a.logger.Error("unable to broadcast event", "error", err)
	}
}

func (a *api) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listID, ok := ctx.Value(tokenList).(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	itemID, ok := ctx.Value(tokenItem).(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ulistID, err := uuid.FromString(listID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uitemID, err := uuid.FromString(itemID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = a.store.DeleteTodoItem(ctx, uitemID)
	if err != nil {
		a.logger.Error("failed to delete todo item", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := ItemEvent{RemoveItem, &TodoItem{ID: uitemID, List: ulistID}}
	data, err := json.Marshal(event)
	if err == nil {
		a.server.Broadcast(data)
	} else {
		a.logger.Error("unable to broadcast event", "error", err)
	}
}
