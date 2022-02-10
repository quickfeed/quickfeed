import React, { useEffect } from "react"
import { SubmissionSort } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"


/**
 *  TableSort displays a widget that aids in sorting and filtering
 *  table contents.
 *  The widget modifies contents of the state on user interaction.
 *  It is up to each component to use the modified state with
 *  sorting and filtering functions based on the modified values.
 *  TODO: We could modify the state to react to changes coming from this component.
 */
const TableSort = () => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        return () => {
            // Reset sort state to default when component is unmounted
            actions.setSubmissionSort(SubmissionSort.Approved)
            actions.setAscending(true)
        }
    }, [])

    const handleChange = (sort: SubmissionSort) => {
        actions.setSubmissionSort(sort)
    }

    return (
        <div className="p-3 mb-2 bg-dark text-white d-flex flex-row">
            <div className="p-2">
                <span>Sort by:</span>
            </div>
            <div className={`${state.sortSubmissionsBy === SubmissionSort.Approved ? "font-weight-bold" : ""} p-2`} role={"button"} onClick={() => handleChange(SubmissionSort.Approved)}>
                Approved
            </div>
            <div className={`${state.sortSubmissionsBy === SubmissionSort.Score ? "font-weight-bold" : ""} p-2`} role={"button"} onClick={() => handleChange(SubmissionSort.Score)}>
                Score
            </div>
            <div className="p-2" role={"button"} onClick={() => actions.setAscending(!state.sortAscending)}>
                <i className={state.sortAscending ? "icon fa fa-caret-down" : "icon fa fa-caret-down fa-rotate-180"} />
            </div>
            <div className="p-2">
                Show:
            </div>
            <div className="p-2" role={"button"} onClick={() => actions.setSubmissionFilter("teachers")}>
                {state.submissionFilters.includes("teachers") ? <del>Teachers</del> : "Teachers"}
            </div>
            <div className="p-2" role={"button"} onClick={() => actions.setSubmissionFilter("approved")}>
                {state.submissionFilters.includes("approved") ? <del>Graded</del> : "Graded"}
            </div>
        </div>
    )
}

export default TableSort
