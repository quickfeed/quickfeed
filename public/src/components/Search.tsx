import React, { useEffect } from "react"
import { useActions } from "../overmind"


/** Search is used to update the query in state when the user types in the search bar.
 *  If setQuery is passed, it will modify the local state of a component instead of the global state. */
const Search = ({ placeholder, setQuery, className, children }: { placeholder?: string, setQuery?: (e: unknown) => void, className?: string, children?: React.ReactNode }): JSX.Element => {
    const actions = useActions()

    useEffect(() => {
        // Reset query in state when component unmounts
        return () => { actions.setQuery("") }
    }, [])

    return (
        <div className={`input-group ${className}`}>
            <input
                type={"text"}
                className="form-control"
                placeholder={placeholder ? placeholder : "Search"}
                onKeyUp={(e) => setQuery ? setQuery(e.currentTarget.value.toLowerCase()) : actions.setQuery(e.currentTarget.value.toLowerCase())} 
            />
            {children}
        </div>
    )
}

export default Search
