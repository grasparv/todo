package db_test

import (
	"context"
	"os"
	"testing"
	"todolist/internal/db"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	const path = "/tmp/test.db"
	t.Cleanup(func() {
		os.Remove(path)
	})
	ctx := context.Background()
	d, err := db.NewDB(ctx, db.Options{DSN: path})
	require.NoError(t, err)
	defer d.Close(ctx)

	stringp := func(s string) *string {
		return &s
	}

	boolp := func(b bool) *bool {
		return &b
	}

	ids := make([]*uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		id := uuid.Must(uuid.NewV4())
		ids[i] = &id
	}

	a := db.TodoList{
		ID:    ids[0],
		Owner: stringp("Jonas"),
		Name:  stringp("Shopping list, Sunday"),
		Items: []db.TodoItem{
			db.TodoItem{
				ID:     ids[1],
				Text:   stringp("Salad"),
				Marked: boolp(false),
			},
			db.TodoItem{
				ID:     ids[2],
				Text:   stringp("Potatoes"),
				Marked: boolp(true),
			},
		},
	}

	b := db.TodoList{
		ID:    ids[3],
		Owner: stringp("Jonas"),
		Name:  stringp("X"),
		Items: []db.TodoItem{},
	}

	c := db.TodoList{
		ID:    ids[4],
		Owner: stringp("Jonas"),
		Name:  stringp("Y"),
		Items: []db.TodoItem{},
	}

	err = d.AddTodoList(ctx, a)
	require.NoError(t, err)

	err = d.AddTodoList(ctx, b)
	require.NoError(t, err)

	err = d.AddTodoList(ctx, c)
	require.NoError(t, err)

	lists, err := d.GetTodoLists(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, len(lists))

	err = d.RemoveTodoList(ctx, *b.ID)
	lists, err = d.GetTodoLists(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(lists))

	ie := a.Items[0]
	ie.Text = stringp("testie")
	err = d.UpdateTodoItem(ctx, ie)
	require.NoError(t, err)
	lists, err = d.GetTodoLists(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(lists))
	var found bool
	for _, list := range lists {
		for _, item := range list.Items {
			if *item.Text == "testie" {
				found = true
			}
		}
	}
	require.True(t, found)
}
