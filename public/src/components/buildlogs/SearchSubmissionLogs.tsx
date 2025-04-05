import React, { useEffect } from "react"
import { useActions } from "../../overmind"
import SubmissionSearchResults from "./SubmissionSearchResults"
import { Submission } from "../../../proto/qf/types_pb"


// SearchSubmissionLogs is a component that displays the results of a search for submission build logs.
// It is used by the teacher to search for build logs from all submissions.
const SearchSubmissionLogs = () => {
    const actions = useActions()

    const [searchResult, setSearchResult] = React.useState<Submission[]>([])
    const [searching, setSearching] = React.useState<boolean>(false)
    const [query, setQuery] = React.useState<string>("")


    useEffect(() => {
        // Every time the query changes, we start a new search.
        // The search is debounced, such that the search is only performed after no new input has been received for 1000ms.
        const timeout = setTimeout(async () => {
            const result = await actions.searchLogs(query)
            setSearchResult(result)
            setSearching(false)
        }, 1000)

        // If the query changes before the timeout has expired, we cancel the previous search.
        return () => clearTimeout(timeout)
    }, [query])

    function handleSearchChange(event: React.ChangeEvent<HTMLInputElement>) {
        setQuery(event.target.value)
        setSearching(true)
    }

    return (
        <div>
            <input type="text" onChange={handleSearchChange} />
            {searching ? <p>Searching...</p> : null}
            <SubmissionSearchResults submissions={searchResult} />
        </div>
    )
}

export default SearchSubmissionLogs
