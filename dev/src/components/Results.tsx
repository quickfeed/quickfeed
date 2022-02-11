import React, { useEffect } from "react"
import { Enrollment, SubmissionLink } from "../../proto/ag/ag_pb"
import { Color, generateAssignmentsHeader, generateSubmissionRows, getCourseID, getSubmissionCellColor, SubmissionSort } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Button, { ButtonType } from "./admin/Button"
import DynamicTable, { CellElement, Row } from "./DynamicTable"
import TableSort from "./forms/TableSort"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Search from "./Search"


const Results = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => {
            actions.setActiveSubmissionLink(undefined)
            actions.setGroupView(false)
            actions.setActiveEnrollment(undefined)
        }
    }, [state.courseSubmissions])


    if (!state.courseSubmissions[courseID]) {
        return <h1>Fetching Submissions...</h1>
    }

    const getSubmissionCell = (submissionLink: SubmissionLink, enrollment: Enrollment): CellElement => {
        const submission = submissionLink.getSubmission()
        if (submission) {
            const isSelected = state.activeSubmission === submission?.getId()
            return ({
                value: `${submission.getScore()} %`,
                className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""}`,
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                    actions.setActiveEnrollment(enrollment)
                }
            })
        } else {
            return ({
                value: "N/A",
                onClick: () => {
                    actions.setActiveSubmissionLink(undefined)
                    actions.setActiveEnrollment(undefined)
                }
            })
        }
    }

    const groupView = state.groupView
    const base: Row = [{ value: "Name", onClick: () => actions.setSubmissionSort(SubmissionSort.Name) }]
    const assignments = state.assignments[courseID].filter(assignment => (state.review.assignmentID < 0) || assignment.getId() === state.review.assignmentID)
    const header = generateAssignmentsHeader(base, assignments, groupView)

    const links = state.sortedAndFilteredSubmissions
    const rows = generateSubmissionRows(links, getSubmissionCell, false)

    const labView = state.currentSubmission ?
        <>
            <ManageSubmissionStatus />
            <div className="reviewLabResult mt-2">
                <Lab />
            </div>
        </>
        : null

    return (
        <div className="row">
            <div className={state.review.assignmentID >= 0 ? "col-md-4" : "col-md-6"}>
                <Search placeholder={"Search by name ..."} >
                    <Button type={ButtonType.BUTTON}
                        text={groupView ? "View by student" : "View by group"}
                        onclick={() => { actions.setGroupView(!groupView); actions.review.setAssignmentID(-1) }}
                        color={groupView ? Color.BLUE : Color.GREEN} />
                </Search>
                <TableSort />
                <DynamicTable header={header} data={rows} />
            </div>
            <div className="col reviewLab">
                {labView}
            </div>
        </div>
    )
}

export default Results
