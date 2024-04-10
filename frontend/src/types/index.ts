export interface ITodoList {
    id: string;
    owner: string;
    name: string;
    items: Map<string,ITodoItem>; // item id -> ITodoItem
}

export interface ITodoItem {
    id: string;
    list: string;
    text: string;
    marked: boolean;
}
