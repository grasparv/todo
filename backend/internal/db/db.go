package db

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/gofrs/uuid"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

type Options struct {
	DSN string
}

func NewDB(ctx context.Context, opt Options) (*DB, error) {
	var create bool
	_, err := os.Stat(opt.DSN)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		create = true
	}

	db, err := sql.Open("sqlite", opt.DSN)
	if err != nil {
		return nil, err
	}

	if create {
		err := createTables(ctx, db)
		if err != nil {
			return nil, err
		}
	}

	return &DB{
		db: db,
	}, nil
}

func createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE list (
   id UUID PRIMARY KEY NOT NULL,
   owner TEXT NOT NULL,
   name TEXT NOT NULL
);`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE list_item (
   id UUID PRIMARY KEY NOT NULL,
   list_id UUID NOT NULL,
   text TEXT NOT NULL,
   marked BOOLEAN NOT NULL,
   FOREIGN KEY (list_id) REFERENCES lists(id)
);`)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) AddTodoList(ctx context.Context, todo TodoList) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT INTO list VALUES (?, ?, ?)", todo.ID, todo.Owner, todo.Name)
	if err != nil {
		return err
	}
	for _, item := range todo.Items {
		_, err = tx.ExecContext(ctx, "INSERT INTO list_item VALUES (?, ?, ?, ?)", item.ID, todo.ID, item.Text, item.Marked)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetTodoLists(ctx context.Context) ([]*TodoList, error) {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT l.id, l.owner, l.name, i.id, i.text, i.marked FROM list AS l LEFT JOIN list_item AS i ON l.id = i.list_id")
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]*TodoList)
	for rows.Next() {
		var list TodoList
		var item TodoItem
		err = rows.Scan(&list.ID, &list.Owner, &list.Name, &item.ID, &item.Text, &item.Marked)
		if err != nil {
			return nil, err
		}
		_, ok := m[*list.ID]
		if !ok {
			list.Items = make([]TodoItem, 0)
			m[*list.ID] = &list
		}
		if item.ID != nil {
			m[*list.ID].Items = append(m[*list.ID].Items, item)
		}
	}
	var lists []*TodoList
	for _, list := range m {
		lists = append(lists, list)
	}
	return lists, nil
}

func (d *DB) RemoveTodoList(ctx context.Context, id uuid.UUID) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "DELETE FROM list WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM list_item WHERE list_id = ?", id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) AddTodoItem(ctx context.Context, listId uuid.UUID, todo TodoItem) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT INTO list_item (id, list_id, text, marked) VALUES (?, ?, ?, ?)", *todo.ID, listId, *todo.Text, *todo.Marked)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) UpdateTodoItem(ctx context.Context, todo TodoItem) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "UPDATE list_item SET text=?, marked=? WHERE id=?", *todo.Text, *todo.Marked, *todo.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteTodoItem(ctx context.Context, itemId uuid.UUID) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "DELETE FROM list_item WHERE id = ?", itemId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Close(ctx context.Context) error {
	return d.db.Close()
}
