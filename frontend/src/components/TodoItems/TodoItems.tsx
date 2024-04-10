import React from 'react';
import '../../styles/TodoItems.css';
import { ITodoList } from '../../types';
import List from '@mui/material/List';
import TodoItem from '../TodoItem/TodoItem';
import IconButton from '@mui/material/IconButton';
import AddIcon from '@mui/icons-material/Add';

interface ITodoItemsProps {
    list: ITodoList;
    itemTextFinish: (item: ITodoItem) => void;
    itemDelete: (item: ITodoItem) => void;
    itemAdd: (id: string) => void;
}

const TodoItems: React.FC<ITodoItemsProps> = ({ list, itemTextFinish, itemDelete, itemAdd }) => {
    if (list === undefined || list.items === undefined) {
        return (
            <div>
                <p>No items loaded yet.</p>
            </div>
        );
    }
    
    const handleAdd = (e: React.FormEvent<HTMLInputElement>) => {
        itemAdd(list.id);
    }

    return (
        <div>
            <List>
                {list.items.map((item, index) => (
                <TodoItem key={item.id} item={item} itemTextFinish={itemTextFinish} itemDelete={itemDelete} />
                ))}
            </List>
            <IconButton onClick={handleAdd} size="small" aria-label="add">
                <AddIcon fontSize="small" />
            </IconButton>
        </div>
    );
};

export default TodoItems;
