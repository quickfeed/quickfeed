import { userInfo } from "os";
import { derived } from "overmind";
import { isMetaProperty } from "typescript";
import { Course, User } from "../proto/ag_pb";


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
    Metadata: {user: string},
    users: User[],
    courses: Course[],
    currentPage: string,
    theme: string
}

const getUserID = (currentState: State) => {
    return {'user': currentState.user.id.toString()}
}

export const state: State = {
    user: {avatarurl: '', email: '', id: -1, isadmin: false, name: '', remoteID: -1, studentid: -1},
    Metadata: derived((state: State) =>  getUserID(state)),
    users: [],
    courses: [],
    currentPage: "home",
    theme: "light"
};