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
const TableSort = ({ review }: { review: boolean }) => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        return () => {
            // Reset sort state to default when component is unmounted
            actions.setSubmissionSort(SubmissionSort.Approved)
            actions.setAscending(true)
            actions.clearSubmissionFilter()
        }
    }, [actions])

    const handleChange = (sort: SubmissionSort) => {
        actions.setSubmissionSort(sort)
    }

    const toggleIndividualSubmissions = () => {
        actions.setIndividualSubmissionsView(!state.individualSubmissionView)
    }

    return (
        <div className="p-1 mb-2 bg-dark text-white d-flex flex-row">
            <div className="d-inline-flex flex-row justify-content-center">
                <div className="p-2">
                    <span>Sort by:</span>
                </div>
                <div className={`${state.sortSubmissionsBy === SubmissionSort.Approved ? "font-weight-bold" : ""} p-2`} role="button" aria-hidden="true" onClick={() => handleChange(SubmissionSort.Approved)}>
                    Approved
                </div>
                <div className={`${state.sortSubmissionsBy === SubmissionSort.Score ? "font-weight-bold" : ""} p-2`} role="button" aria-hidden="true" onClick={() => handleChange(SubmissionSort.Score)}>
                    Score
                </div>
                <div className="p-2" role="button" aria-hidden="true" onClick={() => actions.setAscending(!state.sortAscending)}>
                    <i className={state.sortAscending ? "icon fa fa-caret-down" : "icon fa fa-caret-down fa-rotate-180"} />
                </div>
            </div>
            <div className="d-inline-flex flex-row">
                <div className="p-2">
                    Show:
                </div>
                <div className="p-2" role="button" aria-hidden="true" onClick={() => actions.setSubmissionFilter("teachers")}>
                    {state.submissionFilters.includes("teachers") ? <del>Teachers</del> : "Teachers"}
                </div>
                <div className="p-2" role="button" aria-hidden="true" onClick={() => actions.setSubmissionFilter("approved")}>
                    {state.submissionFilters.includes("approved") ? <del>Graded</del> : "Graded"}
                </div>
                {review ?
                    <div className="p-2" role="button" aria-hidden="true" onClick={() => actions.setSubmissionFilter("released")}>
                        {state.submissionFilters.includes("released") ? <del>Released</del> : "Released"}
                    </div>
                    : null
                }
            </div>
            <div className="d-inline-flex flex-row">
                <div className="p-2" role="button" aria-hidden="true" onClick={toggleIndividualSubmissions}>
                    {state.individualSubmissionView ? "Individual" : "Group"}
                </div>
            </div>
        </div>
    )
}

export default TableSort
