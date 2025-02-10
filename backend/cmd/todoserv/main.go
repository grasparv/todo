package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"todolist/internal/api"
	"todolist/internal/db"

	"github.com/gofrs/uuid"
	"github.com/phsym/console-slog"
)

func main() {
	path := flag.String("db", "/tmp/db.bin", "path to database")
	flag.Parse()

	logger := slog.New(console.NewHandler(os.Stdout, &console.HandlerOptions{Level: slog.LevelDebug}))
	ctx := context.Background()

	store, err := db.NewDB(ctx, db.Options{DSN: *path})
	if err != nil {
		logger.Error("failed to create database", "error", err)
		return
	}
	defer store.Close(ctx)

	lists, err := store.GetTodoLists(ctx)
	if err != nil {
		logger.Error("failed to query database", "error", err)
		return
	}

	if len(lists) == 0 {
		// Add some sample data for the project
		id, err := uuid.NewV4()
		if err != nil {
			logger.Error("failed to create sample data", "error", err)
			return
		}

		ida, err := uuid.NewV4()
		if err != nil {
			logger.Error("failed to create sample data", "error", err)
			return
		}

		idb, err := uuid.NewV4()
		if err != nil {
			logger.Error("failed to create sample data", "error", err)
			return
		}

		stringp := func(s string) *string {
			return &s
		}

		boolp := func(b bool) *bool {
			return &b
		}

		sample := db.TodoList{
			ID:    &id,
			Owner: stringp("Jonas"),
			Name:  stringp("Shopping list, Sunday"),
			Items: []db.TodoItem{
				db.TodoItem{
					ID:     &ida,
					Text:   stringp("Salad"),
					Marked: boolp(false),
				},
				db.TodoItem{
					ID:     &idb,
					Text:   stringp("Potatoes"),
					Marked: boolp(true),
				},
			},
		}

		err = store.AddTodoList(ctx, sample)
		if err != nil {
			logger.Error("failed to create sample data", "error", err)
			return
		}
	}

	lists, err = store.GetTodoLists(ctx)
	if err != nil {
		logger.Error("failed to get todo lists", "error", err)
		return
	}

	for _, list := range lists {
		logger.Info("todo list", list.ID.String(), *list.Name)
	}

	service := api.New(ctx, logger, store)
	service.Run()
}
