import * as React from 'react'
import { useOvermind } from "../overmind";



const FilterTodo = () => {
    const { actions } = useOvermind()

    const changeShowCount = (event: React.ChangeEvent<HTMLSelectElement>) => {
        actions.changeShowCount(event.target.value);
    }

    return (
        <div>
            Filter Todos:{' '}
            <select onChange={(event => changeShowCount(event))}>
            <option value="200">200</option>
            <option value="100">100</option>
            <option value="50">50</option>
            <option value="25">25</option>
            <option value="10">10</option>
            <option value="5">5</option>
        </select>
</div>
    )
}

export default FilterTodo;