import React, { useState, useEffect} from 'react';
import isEqual from 'lodash/isEqual';
import { ITodoList } from '../../types';
import TodoLists from '../components/TodoLists/TodoLists';
import ReconnectingEventSource from "reconnecting-eventsource";
import { enableMapSet, produce } from 'immer';

enableMapSet();

const serviceUri = 'http://localhost:2000'

interface IHomeProps {
    user: string | undefined;
}

const Home: React.FC<IHomeProps> = ({ user }) => {
    const [lists, setLists] = useState<Map<string,ITodoList>>(new Map<string,ITodoList>());
    const [newListName, setNewListName] = useState<string>();
    const [isConnected, setIsConnected] = useState<boolean>(false);

    useEffect(() => {
        const eventSource = new ReconnectingEventSource(serviceUri+'/events');

        eventSource.onopen = () => {
            setIsConnected(true);
        };

        eventSource.onmessage = (event) => {
            if (!isConnected) {
                return;
            }
            try {
                const msg = JSON.parse(event.data);
                console.log("received", event.data);
                const todoList = msg.todolist;
                const todoItem = msg.todoitem;
                switch (msg.type) {
                    case "update-list":
                        const opUpdate = produce(lists, draft => {
                            const oldList = draft.get(todoList.id);
                            if (!isEqual(oldList, todoList)) {
                                draft.set(todoList.id, todoList);
                            }
                        });
                        setLists(opUpdate);
                        break;
                    case "remove-list":
                        const opDelete = produce(lists, draft => {
                            draft.delete(todoList.id);
                        });
                        setLists(opDelete);
                        break;
                    case "update-item":
                        setLists(draftLists =>
                            produce(draftLists, draft => {
                                const tlist = draft.get(todoItem.list);
                                if (tlist) {
                                    const titem = tlist.items.find((item: ITodoItem) => item.id === todoItem.id);
                                    if (titem) {
                                        const updatedItem = { ...titem, text: todoItem.text };
                                        const itemIndex = tlist.items.findIndex((item: ITodoItem) => item.id === todoItem.id);
                                        if (itemIndex !== -1) {
                                            tlist.items[itemIndex] = updatedItem;
                                        }
                                    }
                                }
                            })
                        );
                        break;
                    case "add-item":
                        setLists(draftLists =>
                            produce(draftLists, draft => {
                                const tlist = draft.get(todoItem.list);
                                if (tlist) {
                                    tlist.items.push(todoItem);
                                }
                            })
                        );
                        break;
                    case "remove-item":
                        setLists(draftLists =>
                            produce(draftLists, draft => {
                                const tlist = draft.get(todoItem.list);
                                if (tlist) {
                                    const itemIndex = tlist.items.findIndex((item: ITodoItem) => item.id === todoItem.id);
                                    if (itemIndex !== -1) {
                                        tlist.items.splice(itemIndex, 1);
                                    }
                                }
                            })
                        );
                }
            } catch (error) {
                console.error('error handling message:', error);
            }
        };

        eventSource.onerror = (error) => {
            setIsConnected(false);
        };

        eventSource.onReconnected = (event) => {
            setIsConnected(true);
        };

        return () => {
            eventSource.close();
        };
    }, [lists, isConnected]);

    const doNewTextEdit = (id: string, text: string) => {
        setNewListName(text);
    };

    const doCreateNewList = async () => {
        if (newListName === undefined || newListName.length < 1) {
            alert("Please fill in the name of the list");
            return;
        }

        const url = serviceUri+'/list';
        const newList: ITodoList = {
            id: undefined,
            owner: user,
            name: newListName,
            items: [],
        };

        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(newList)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }

            await response.json();
        } catch (error) {
            console.error('Error creating new list:', error);
        }
    };

    const doDeleteListClick = async (id: string) => {
        const url = serviceUri+'/list/'+id;

        try {
            const response = await fetch(url, {
                method: 'DELETE',
            });

            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }

            await response.json();
        } catch (error) {
            console.error('Error creating new list:', error);
        }
    }

    const doItemAdd = async (id: string) => {
        try {
            const response = await fetch(serviceUri+'/list/'+id+'/add', {
                method: 'PUT',
            });

            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }

            await response.json();
        } catch (error) {
            console.error('Error creating new item:', error);
        }
    };

    const doItemTextFinish = async (item: ITodoItem) => {
        try {
            const response = await fetch(serviceUri+'/list/'+item.list+'/item/'+item.id, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(item)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }

            await response.json();
        } catch (error) {
            console.error('Error updating item:', error);
        }
    };

    const doItemDelete = async (item: ITodoItem) => {
        try {
            const response = await fetch(serviceUri+'/list/'+item.list+'/item/'+item.id, {
                method: 'DELETE',
            });

            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }

            await response.json();
        } catch (error) {
            console.error('Error deleting item:', error);
        }
    };

    return (
        <TodoLists user={user} lists={lists}
            onDeleteListClick={doDeleteListClick} onNewListClick={doCreateNewList}
            itemTextFinish={doItemTextFinish} itemAdd={doItemAdd} itemDelete={doItemDelete}
            newTextEdit={doNewTextEdit} newTextFinish={doCreateNewList} />
    );
}

export default Home;
