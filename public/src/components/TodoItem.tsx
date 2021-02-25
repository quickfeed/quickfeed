import * as React from 'react'
import { useOvermind } from "../overmind";
import {state, Todo} from '../overmind/state'
import {useCallback} from "react";



type Props = {
    todo: Todo;

}



const TodoItem = ({ todo }: Props) => {

    const { state, actions } = useOvermind()

    const handleToggleChange = () => {
        actions.toggleTodo(todo.id)
    }

    const editHandler = () => {
        actions.editTodo(todo.id);
    }

    const saveEditHandler = useCallback((event) => {
        if (event.keyCode != 13) { return }
        actions.saveEdit({id: todo.id, title: event.target.value, completed: todo.completed})
    }, [actions.saveEdit])

    let completed = { }

    if(todo.completed){
        completed = {
            color: 'green',
            'font-style': 'italic',
            'text-decoration': 'line-through'

        }
    }

    if (state.isEditing == todo.id) {
        return (
            <div className='todo-item'>
                <h4><input type='checkbox' onChange={handleToggleChange} checked={todo.completed}/><input placeholder={todo.title} onKeyDown={saveEditHandler}/></h4>
            </div>
        )
    }   else {
        return (
            <div className='todo-item' style={completed}>

                <h4><input type='checkbox' onChange={handleToggleChange} checked={todo.completed}/> {todo.title}<span onClick={editHandler}> *</span></h4>
            </div>
        )
    }
}

export default TodoItem