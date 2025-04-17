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

    const boldText = (sort: SubmissionSort) => {
        return state.sortSubmissionsBy === sort ? "font-weight-bold" : ""
    }
    const pointer = state.sortAscending ? "fa fa-caret-down" : "fa fa-caret-down fa-rotate-180"
    const textForToggleIndividualViewButton = state.individualSubmissionView ? "Individual" : "Group"

    const submissionFilters = [
        { name: "teachers", text: "Teachers", show: true },
        { name: "approved", text: "Graded", show: true },
        { name: "released", text: "Released", show: review }
    ]

    const filterElements = submissionFilters.map((filter) => {
        const displayText = state.submissionFilters.includes(filter.name)
            ? <del>{filter.text}</del>
            : filter.text
        return filter.show
            ? <DivButton key={filter.name} text={displayText} onclick={() => actions.setSubmissionFilter(filter.name)} />
            : null
    })

    const sortByButtons = [
        { key: "approved", text: "Approved", className: boldText(SubmissionSort.Approved), onclick: () => handleChange(SubmissionSort.Approved) },
        { key: "score", text: "Score", className: boldText(SubmissionSort.Score), onclick: () => handleChange(SubmissionSort.Score) },
        { key: "pointer", text: <i className={pointer} />, onclick: () => actions.setAscending(!state.sortAscending) }
    ]

    const sortByElements = sortByButtons.map((button) => (
        <DivButton key={button.key} text={button.text} className={button.className} onclick={button.onclick} />
    ))

    return (
        <div className="p-1 mb-2 bg-dark text-white d-flex flex-row">
            <div className="d-inline-flex flex-row justify-content-center">
                <div className="p-2">
                    <span>Sort by:</span>
                </div>
                {sortByElements}
            </div>
            <div className="d-inline-flex flex-row">
                <div className="p-2">
                    Show:
                </div>
                {filterElements}
            </div>
            <div className="d-inline-flex flex-row">
                <DivButton text={textForToggleIndividualViewButton} onclick={toggleIndividualSubmissions} />
            </div>
        </div>
    )
}

interface DivButtonProps {
    text: string | React.JSX.Element
    key?: string
    className?: string
    onclick: () => void
}

const DivButton = ({ text, key, className, onclick }: DivButtonProps) => {
    return (
        <div key={key} className={`${className ?? ""} p-2`} role="button" aria-hidden="true" onClick={onclick}>
            {text}
        </div>
    )
}

export default TableSort
