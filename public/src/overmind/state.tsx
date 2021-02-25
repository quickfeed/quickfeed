import { userInfo } from "os";
import { derived } from "overmind";
import { isMetaProperty } from "typescript";
import { User } from "../proto/ag_pb";


export interface Todo {
   id: number;
   title: string;
   completed: boolean;
}

export interface IUser {
    remoteID: number;
    avatarurl: string;
    email: string;
    id: number;
    isadmin: boolean;
    name: string;
    studentid: number;
}




export type State = {
    user: IUser,
    todos: Todo[],
    num: number,
    isEditing: number,
    numShow: number,
    todoSlice: Todo[],
    Metadata: {user: string},
    users: User[]
}

const getUserID = (currentState: State) => {
    return {'user': currentState.user.id.toString()}
}

export const state: State = {
    user: {avatarurl: '', email: '', id: -1, isadmin: false, name: '', remoteID: -1, studentid: -1},
    todos: [],
    num: derived((state: State) => state.todos.length),
    isEditing: -1,
    numShow: 200,
    todoSlice: derived((state: State) => state.todos.slice(0, state.numShow)),
    Metadata: derived((state: State) =>  getUserID(state)),
    users: []
};