import React, { useEffect } from "react"
import { Void } from "../../proto/ag/ag_pb"
import { useActions } from "../overmind"

export const Search = ({placeholder, setQuery}: {placeholder?: string, setQuery?: Function}) => {
    const actions = useActions()

    useEffect(() => {
        return actions.setQuery("")
    }, [])

    return (
        <input  
                type={"text"} 
                onFocus={() => actions.enableRedirect(false)} 
                onBlur={() => actions.enableRedirect(true)} 
                placeholder={placeholder ? placeholder : "Search"} 
                onKeyUp={(e) => setQuery ? setQuery(e.currentTarget.value.toLowerCase()) : actions.setQuery(e.currentTarget.value.toLowerCase()) }>
        </input>
    )
}

export default Search