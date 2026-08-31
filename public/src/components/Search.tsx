import { useEffect } from "react"
import type { KeyboardEvent, ReactNode } from "react"
import { useActions } from "../overmind"


type SearchProps = {
    placeholder?: string
    setQuery?: (query: string) => void
    className?: string
    children?: ReactNode
}

/** Search is used to update the query in state when the user types in the search bar.
 *  If setQuery is passed, it will modify the local state of a component instead of the global state. */
const Search = ({ placeholder, setQuery, className, children }: SearchProps) => {
    const actions = useActions().global

    const handleKeyUp = (e: KeyboardEvent<HTMLInputElement>) => {
        const value = e.currentTarget.value.toLowerCase()
        setQuery ? setQuery(value) : actions.setQuery(value)
    }

    useEffect(() => {
        // Reset query in state when component unmounts
        return () => { actions.setQuery("") }
    }, [actions])

    return (
        <div className={`flex items-center gap-2 ${className ?? ""}`}>
            <input
                type="text"
                className="input input-bordered flex-1 focus:input-primary"
                placeholder={placeholder ?? "Search"}
                onKeyUp={handleKeyUp}
            />
            {children}
        </div>
    )
}

export default Search
