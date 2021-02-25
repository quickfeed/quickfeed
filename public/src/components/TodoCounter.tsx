import * as React from 'react'
import { useOvermind } from "../overmind";
import NavBar from './NavBar';


const TodoCounter = () => {
    const { state } = useOvermind()

    return (
        <div>
        <NavBar></NavBar>
        <h4>{state.num}</h4>
        </div>
    )
}

export default TodoCounter