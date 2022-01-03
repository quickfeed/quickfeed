import React, { useEffect } from "react"
import { Group, Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { getCourseID, isTeacher } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import DynamicTable, { CellElement } from "./DynamicTable"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Search from "./Search"


const Results = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return actions.setActiveSubmission(undefined)
    }, [state.courseSubmissions])

    const header = ["Name", "Group"].concat(state.assignments[courseID].map(assignment => {
        return assignment.getName()
    }))

    if (!state.courseSubmissions[courseID] || !isTeacher(state.enrollmentsByCourseId[courseID])) {
        return <h1>Nothing</h1>
    }

    const getSubmissionCell = (submissionLink: SubmissionLink, enrollmentId: number | undefined): CellElement => {
        if (submissionLink.hasSubmission() && submissionLink.hasAssignment()) {
            return ({   
                value: `${submissionLink.getSubmission()?.getScore()}%`, 
                className: submissionLink.getSubmission()?.getStatus() === Submission.Status.APPROVED ? "result-approved" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmission(submissionLink.getSubmission()?.getId())
                    actions.setSelectedEnrollment(enrollmentId)
                    actions.setActiveSubmissionLink(submissionLink)
                }
            })
        }
        else {
            return ({
                value: "N/A", 
                onClick: () => actions.setActiveSubmission(undefined)
            })
        }
    } 

    const results = state.courseSubmissionsList[courseID].map(link => {
        const data: (string | JSX.Element | CellElement)[] = []
        data.push(link.user ? {value: link.user.getName(), link: `https://github.com/${link.user.getLogin()}`} : "")
        data.push(link.enrollment && link.enrollment.hasGroup() ? (link.enrollment.getGroup() as Group)?.getName() : "")
        if (link.submissions && link.user) {
            for (const submissionLink of link.submissions) {
                data.push(getSubmissionCell(submissionLink, link.enrollment?.getId()))
            }
        }
        return data
    })

    return (
        <div>
            <div className="row">
                <div className="col">
                    <Search />
                    <DynamicTable header={header} data={results} />
                </div>
                <div className="col reviewLab">
                    {state.currentSubmission ?
                    <>
                        <ManageSubmissionStatus />
                        <div className="reviewLabResult mt-2">
                            <Lab teacherSubmission={state.currentSubmission} />
                        </div>
                    </>
                    : null}  
                </div>
            </div>
        </div>

    )
}

export default Results