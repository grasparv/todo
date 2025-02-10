package db_test

import (
	"context"
	"os"
	"testing"
	"todolist/internal/conv"
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

	ids := make([]*uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		id := uuid.Must(uuid.NewV4())
		ids[i] = &id
	}

	a := db.TodoList{
		ID:    ids[0],
		Owner: conv.Pointer("Jonas"),
		Name:  conv.Pointer("Shopping list, Sunday"),
		Items: []db.TodoItem{
			{
				ID:     ids[1],
				Text:   conv.Pointer("Salad"),
				Marked: conv.Pointer(false),
			},
			{
				ID:     ids[2],
				Text:   conv.Pointer("Potatoes"),
				Marked: conv.Pointer(true),
			},
		},
	}

	b := db.TodoList{
		ID:    ids[3],
		Owner: conv.Pointer("Jonas"),
		Name:  conv.Pointer("X"),
		Items: []db.TodoItem{},
	}

	c := db.TodoList{
		ID:    ids[4],
		Owner: conv.Pointer("Jonas"),
		Name:  conv.Pointer("Y"),
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
	ie.Text = conv.Pointer("testie")
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
