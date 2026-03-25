import React, { useEffect } from "react"
import { SubmissionSort } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"


/**
 * TableSort displays a widget for sorting and filtering the submissions table.
 * It modifies state values that are used by the table for sorting/filtering.
 */
const TableSort = () => {
    const state = useAppState()
    const actions = useActions().global

    useEffect(() => {
        return () => {
            // Reset sort state to default when component is unmounted
            actions.setSubmissionSort(SubmissionSort.Approved)
            actions.setAscending(true)
            actions.clearSubmissionFilter()
        }
    }, [actions])

    const handleSortChange = (sort: SubmissionSort) => {
        if (state.sortSubmissionsBy === sort) {
            // If clicking the same sort option, toggle direction
            actions.setAscending(!state.sortAscending)
        } else {
            actions.setSubmissionSort(sort)
        }
    }

    const isFilterActive = (filterName: string) => state.submissionFilters.includes(filterName)

    return (
        <div className="flex flex-wrap items-center gap-4 p-3 bg-base-200 rounded-lg">
            {/* Sort Options */}
            <div className="flex items-center gap-1">
                <span className="text-xs font-medium text-base-content/60 mr-1">Sort:</span>
                <div className="join">
                    <SortButton
                        label="Approved"
                        isActive={state.sortSubmissionsBy === SubmissionSort.Approved}
                        ascending={state.sortAscending}
                        onClick={() => handleSortChange(SubmissionSort.Approved)}
                    />
                    <SortButton
                        label="Score"
                        isActive={state.sortSubmissionsBy === SubmissionSort.Score}
                        ascending={state.sortAscending}
                        onClick={() => handleSortChange(SubmissionSort.Score)}
                    />
                </div>
            </div>

            {/* Divider */}
            <div className="divider divider-horizontal mx-0" />

            {/* Filter Options */}
            <div className="flex items-center gap-1">
                <span className="text-xs font-medium text-base-content/60 mr-1">Hide:</span>
                <div className="flex gap-1">
                    <FilterToggle
                        label="Teachers"
                        isActive={isFilterActive("teachers")}
                        onClick={() => actions.setSubmissionFilter("teachers")}
                    />
                    <FilterToggle
                        label="Graded"
                        isActive={isFilterActive("approved")}
                        onClick={() => actions.setSubmissionFilter("approved")}
                    />
                </div>
            </div>
        </div>
    )
}

interface SortButtonProps {
    label: string
    isActive: boolean
    ascending: boolean
    onClick: () => void
}

/** Sort button that shows direction indicator when active */
const SortButton = ({ label, isActive, ascending, onClick }: SortButtonProps) => {
    const directionIcon = ascending ? "fa-arrow-up" : "fa-arrow-down"

    return (
        <button
            className={`btn btn-xs join-item gap-1 ${isActive ? "btn-primary" : "btn-ghost"}`}
            onClick={onClick}
        >
            {label}
            {isActive && <i className={`fa ${directionIcon} text-xs`} />}
        </button>
    )
}

interface FilterToggleProps {
    label: string
    isActive: boolean
    onClick: () => void
}

/** Filter toggle that shows check mark when active (items are hidden) */
const FilterToggle = ({ label, isActive, onClick }: FilterToggleProps) => {
    return (
        <button
            className={`btn btn-xs gap-1 ${isActive ? "btn-error btn-outline" : "btn-ghost"}`}
            onClick={onClick}
            title={isActive ? `Show ${label.toLowerCase()}` : `Hide ${label.toLowerCase()}`}
        >
            {isActive && <i className="fa fa-eye-slash text-xs" />}
            {label}
        </button>
    )
}

export default TableSort
