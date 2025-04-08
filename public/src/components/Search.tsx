import React, { useCallback, useEffect } from "react"
import { useActions } from "../overmind"


/** Search is used to update the query in state when the user types in the search bar.
 *  If setQuery is passed, it will modify the local state of a component instead of the global state. */
const Search = ({ placeholder, setQuery, className, children }: { placeholder?: string, setQuery?: (e: unknown) => void, className?: string, children?: React.ReactNode }) => {
    const actions = useActions()

    useEffect(() => {
        // Reset query in state when component unmounts
        return () => { actions.setQuery("") }
    }, [actions])

    const handleKeyUp = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
        if (setQuery) {
            setQuery(e.currentTarget.value.toLowerCase())
        }
        else {
            actions.setQuery(e.currentTarget.value.toLowerCase())
        }
    }, [setQuery, actions])

    return (
        <div className={`input-group ${className}`}>
            <input
                type={"text"}
                className="form-control"
                placeholder={placeholder ? placeholder : "Search"}
                onKeyUp={handleKeyUp}
            />
            {children}
        </div>
    )
}

export default Search
