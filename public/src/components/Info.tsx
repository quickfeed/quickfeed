import * as React from 'react'
import { useOvermind } from "../overmind";
import NavBar from './NavBar';


const Info = () => {
    const { state } = useOvermind()

    return (
        <div>
        <h4>Test to see</h4>
        </div>
    )
}

export default Info