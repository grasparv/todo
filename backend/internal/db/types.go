package db

import (
	"github.com/gofrs/uuid"
)

type TodoList struct {
	ID    *uuid.UUID
	Owner *string
	Name  *string
	Items []TodoItem
}

type TodoItem struct {
	ID     *uuid.UUID
	Text   *string
	Marked *bool
}
