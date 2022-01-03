import React, { useEffect } from "react"
import { useActions } from "../overmind"

//** Search */
/**
 *  This component updates either the supplied query state, or the query in Overmind state
 *  Used to determine if elements in the component it is in should be hidden or not.
 */
export const Search = ({placeholder, setQuery}: {placeholder?: string, setQuery?: (e: unknown) => void}): JSX.Element => {
    const actions = useActions()

    useEffect(() => {
        // Reset query in state when component loads
        return actions.setQuery("")
    }, [])

    return (
        <div className="input-group">
            <input  
                    type={"text"} 
                    className="form-control"
                    onFocus={() => actions.enableRedirect(false)} 
                    onBlur={() => actions.enableRedirect(true)} 
                    placeholder={placeholder ? placeholder : "Search"} 
                    onKeyUp={(e) => setQuery ? setQuery(e.currentTarget.value.toLowerCase()) : actions.setQuery(e.currentTarget.value.toLowerCase()) }>
            </input>
        </div>
    )
}

export default Search

