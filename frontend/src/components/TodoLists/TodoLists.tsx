import React from 'react';
import '../../styles/TodoLists.css';
import Button from '@mui/material/Button';
import { ITodoList } from '../../types';
import TodoList from '../TodoList/TodoList';
import TextField from '@mui/material/TextField';

interface ITodoListsProps {
    user: string | undefined;
    lists: Map<string,ITodoList>;
    onDeleteListClick: (id: string) => void;
    onNewListClick: () => void;
    itemTextFinish: (item: ITodoItem) => void;
    itemDelete: (item: ITodoItem) => void;
    itemAdd: (id: string) => void;
    newTextEdit: (text: string) => void;
    newTextFinish: () => void;
}

const TodoLists: React.FC<ITodoListsProps> = ({ user, lists, onDeleteListClick, onNewListClick, newTextEdit, newTextFinish, itemTextFinish, itemAdd, itemDelete }) => {
    const handleEditNewText = (event: React.ChangeEvent<HTMLInputElement>) => {
        newTextEdit("", event.target.value);
    }

    const finishNewText = (event: React.KeyboardEvent<HTMLInputElement>) => {
        if (event.key === "Enter") {
            newTextFinish();
        }
    }

    return (
        <>
        <div className="userInfo">
            <p>Logged in as {user}</p>
        </div>
        <div className="listsView">
            <div className="todoLists">
                {lists && [...lists].map(([key, list]) => (
                    <TodoList key={list.id} user={user} list={list} itemTextFinish={itemTextFinish} itemAdd={itemAdd} itemDelete={itemDelete} listDelete={onDeleteListClick} />
                ))}
            </div>
            <div className="newLists">
                <TextField autoFocus variant="outlined" label="New list" onChange={(event) => handleEditNewText(event)} onKeyDown={finishNewText} />
                <Button variant="contained" onClick={onNewListClick} >Create</Button>
            </div>
        </div>
        </>
    );
}

export default TodoLists;
