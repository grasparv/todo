import React from 'react';
import '../../styles/TodoList.css';
import { ITodoList } from '../../types';
import TodoItems from '../TodoItems/TodoItems';
import Button from '@mui/material/Button';
import DeleteIcon from '@mui/icons-material/Delete';

interface ITodoListProps {
    user: string | undefined;
    list: ITodoList;
    itemTextFinish: (item: ITodoItem) => void;
    itemDelete: (item: ITodoItem) => void;
    itemAdd: (id: string) => void;
    listDelete: (id: string) => void;
}

const TodoList: React.FC<ITodoListProps> = ({ user, list, itemTextFinish, itemAdd, itemDelete, listDelete }) => {
    if (list === undefined || list.items === undefined) {
        return (
            <div>
                <p>Empty list.</p>
            </div>
        );
    }

    const handleDelete = () => {
        listDelete(list.id);
    };

    
    return (
        <div className="list">
		  <div className="grid-container">
	  	    <div className="left-panel">
              <div className="button-container">
			    <Button className="delete-button" variant="contained" endIcon={<DeleteIcon />} onClick={handleDelete}>Delete</Button>
		      </div>
              <p className="listInfo">Owner: {list.owner}</p>
		    </div>
		    <div className="right-panel">
		  	<div className="complex-component">
              <p className="listHeader">{list.name}</p>
		  	  <TodoItems list={list} itemTextFinish={itemTextFinish} itemAdd={itemAdd} itemDelete={itemDelete} />
		  	</div>
		    </div>
		  </div>
        </div>
    );
}

export default TodoList;
