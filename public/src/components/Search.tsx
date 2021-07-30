import React, { useEffect } from "react"
import { useActions } from "../overmind"

export const Search = ({placeholder}: {placeholder?: string}) => {
    const actions = useActions()

    useEffect(() => {
        return actions.setQuery("")
    }, [])

    return (
        <input type={"text"} onFocus={() => actions.enableRedirect(false)} onBlur={() => actions.enableRedirect(true)} placeholder={placeholder ? placeholder : "Search"} onKeyUp={(e) => actions.setQuery(e.currentTarget.value.toLowerCase()) }></input>
    )
}

export default Search