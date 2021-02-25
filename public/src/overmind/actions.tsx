import { Context, Action } from "overmind";
import { Todo } from './state'
import { User } from "../proto/ag_pb";
import { useEffects } from ".";

export const addTodo: Action<string> = ({state}, title) => {
    state.todos.push({
        id: state.todos.length ,
        title: title,
        completed: false,
    });
};

export const onLoad: ({state, effects, actions}: { state: any; effects: any; actions: any }) => Promise<void> = async ({state, effects, actions}) => {
    await effects.api.getTodos().then((response: any) => {
        let todos = Object(response)
        todos.forEach(async function (todo: {title: string, completed: boolean}) {
            await actions.addTodo(todo.title)
        })
    })
}

export const toggleTodo: Action<number> = ({state}, id) => {
    state.todos[id].completed = !state.todos[id].completed
}

export const editTodo: Action<number> = ({state}, id) => {
    state.isEditing = id
}

export const saveEdit: Action<Todo> = ({state}, todo) => {
    state.isEditing = -1
    console.log(todo)
    if (todo.title.length > 0) {state.todos[todo.id].title = todo.title}
}

export const changeShowCount: Action<string> = ({state}, num) => {
    state.numShow = +num
}

export const getUser: Action<void, Promise<boolean>> = ({state, effects}) => {
    return effects.api.getUser()
    .then((user) => {
        if (user.id == undefined) {
            return false
        }
        state.user = user;
        return true
    })
    
}
export const getUsers: Action<void> = ({state, effects}) => {
    state.users = []
    effects.api.getUsers(state).then(users => {
        users.forEach(user => {
            if (user.getStudentid() != "") {
                state.users.push(user)
            }
        });
    })
}

export const getCourses: Action<void> = ({state, effects}) => {
    effects.api.getCourses(state).then(courses => {
        console.log("getting courses")
    })
}
