import { json } from "overmind"
import React, { useEffect } from "react"
import { Group, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb"
import { getCourseID, isTeacher } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import DynamicTable, { CellElement } from "./DynamicTable"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Search from "./Search"


const Results = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const { getAllCourseSubmissions } = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            getAllCourseSubmissions(courseID)
        }
        return actions.setActiveSubmission(undefined)
    }, [state.courseSubmissions])

    const Header: (string | JSX.Element)[] = ["Name", "Group"]

    const AssignmentsHeader = (state.assignments[courseID].map(assignment => {
        return assignment.getName()
    }))

    if (!state.courseSubmissions[courseID] || !isTeacher(state.enrollmentsByCourseId[courseID])) {
        return <h1>Nothing</h1>
    }

    const getSubmissionCell = (submissionLink: SubmissionLink, user: User) => {
        if (submissionLink.hasSubmission() && submissionLink.hasAssignment()) {
            return ({   
                value: `${submissionLink.getSubmission()?.getScore()}%`, 
                className: submissionLink.getSubmission()?.getStatus() === Submission.Status.APPROVED ? "result-approved" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmission(json(submissionLink.getSubmission()))
                    actions.setSelectedUser(json(user))
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

    const results = state.courseSubmissions[courseID].map(link => {
        const data: (string | JSX.Element | CellElement)[] = []
        data.push(link.user ? {value: link.user.getName(), link: `https://github.com/${link.user.getLogin()}`} : "")
        data.push(link.enrollment && link.enrollment.hasGroup() ? (link.enrollment.getGroup() as Group)?.getName() : "")
        if (link.submissions && link.user) {
            for (const submissionLink of link.submissions) {
                data.push(getSubmissionCell(submissionLink, link.user))
            }
        }
        return data
    })

    return (
        <div className="box">
            <div className="row">
                <div className="col">
                    <Search />
                    <DynamicTable header={Header.concat(AssignmentsHeader)} data={results} />
                </div>
                <div className="col reviewLab">
                    {state.activeSubmission ?
                    <>
                    <ManageSubmissionStatus />
                    <div className="reviewLabResult mt-2">
                        <Lab teacherSubmission={state.activeSubmission} />
                    </div>
                    </>
                    : null}  
                </div>
            </div>
        </div>

    )
}

export default Results