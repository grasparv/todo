package api

import (
	"todolist/internal/db"

	"github.com/gofrs/uuid"
)

const (
	UpdateList = "update-list"
	RemoveList = "remove-list"
	AddItem    = "add-item"
	UpdateItem = "update-item"
	RemoveItem = "remove-item"
)

type ListEvent struct {
	Type     string    `json:"type,omitempty"`
	TodoList *TodoList `json:"todolist,omitempty"`
}

type ItemEvent struct {
	Type     string    `json:"type,omitempty"`
	TodoItem *TodoItem `json:"todoitem,omitempty"`
}

type TodoList struct {
	ID    uuid.UUID  `json:"id,omitempty"`
	Owner string     `json:"owner,omitempty"`
	Name  string     `json:"name,omitempty"`
	Items []TodoItem `json:"items"`
}

type TodoItem struct {
	ID     uuid.UUID `json:"id,omitempty"`
	List   uuid.UUID `json:"list,omitempty"`
	Text   string    `json:"text,omitempty"`
	Marked bool      `json:"marked,omitempty"`
}

func newTodoItem(list uuid.UUID, in db.TodoItem) (out TodoItem) {
	out.ID = *in.ID
	out.List = list
	out.Text = *in.Text
	out.Marked = *in.Marked
	return
}

func NewTodoList(in *db.TodoList) (out TodoList) {
	out.ID = *in.ID
	out.Owner = *in.Owner
	out.Name = *in.Name
	out.Items = make([]TodoItem, len(in.Items))
	for i := range in.Items {
		out.Items[i] = newTodoItem(out.ID, in.Items[i])
	}
	return
}

func (in TodoItem) Record() (out db.TodoItem) {
	out.ID = &in.ID
	out.Text = &in.Text
	out.Marked = &in.Marked
	return
}

func (in TodoList) Record() (out db.TodoList) {
	out.ID = &in.ID
	out.Owner = &in.Owner
	out.Name = &in.Name
	out.Items = make([]db.TodoItem, len(in.Items))
	for i := range in.Items {
		out.Items[i] = in.Items[i].Record()
	}
	return
}
