import React, { useState} from 'react';
import '../../styles/TodoItem.css';
import ListItem from '@mui/material/ListItem';
import { ITodoItem } from '../../types';
import Input from '@mui/material/Input';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';

interface ITodoItemProps {
	item: ITodoItem;
    itemTextFinish: (item: ITodoItem) => void;
    itemDelete: (item: ITodoItem) => void;
}

const TodoItem: React.FC<ITodoItemProps> = ({ item, editText, itemTextFinish, itemDelete }) => {
    const [text, setText] = useState<string>(item.text);
    const [isEditing, setIsEditing] = useState<boolean>(false);

    const handleChange = (e: React.FormEvent<HTMLInputElement>) => {
        setText(e.currentTarget.value);
    };

    const finishText = () => {
        const newItem : ITodoItem = {
            id: item.id,
            list: item.list,
            text: text,
            marked: item.marked,
        };
        itemTextFinish(newItem);
        setIsEditing(false);
    }

    const handleDelete = (e: React.FormEvent<HTMLInputElement>) => {
        itemDelete(item);
    }

    const handleEdit = (e: React.FormEvent<HTMLInputElement>) => {
        setText(item.text);
        setIsEditing(true);
    }

    if (!isEditing) {
        return (
            <ListItem key={item.id}>
                <div className="item-grid-container">
                    <div className="item-left-panel">
                        <IconButton onClick={handleDelete} size="small" aria-label="delete">
                            <DeleteIcon fontSize="small" />
                        </IconButton>
                    </div>
                    <div className="item-right-panel" onClick={handleEdit}>
                        {item.text}
                    </div>
                </div>
            </ListItem>
        );
    }

	return (
        <ListItem key={item.id}>
            <Input
                value={text}
                onChange={handleChange}
                onBlur={finishText}
                className="editText"
            />
		</ListItem>
	);
};

export default TodoItem;
