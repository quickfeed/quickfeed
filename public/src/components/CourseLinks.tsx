import React from "react"
import { Link } from "react-router-dom"
import { Repository_Type } from "../../proto/qf/types_pb"
import { getCourseID } from "../Helpers"
import { useAppState } from "../overmind"

/** CourseLinks displays various repository links for the current course, in addition to links to take the user to the group page. */
const CourseLinks = (): JSX.Element => {
    const state = useAppState()
    const courseID = getCourseID()
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    const repo = state.repositories[courseID.toString()]

    return (
        <div className="col-lg-3">
            <div className="list-group width-resize">
                <div className="list-group-item list-group-item-action active text-center">
                    <h6>
                        <strong>Links</strong>
                    </h6>
                </div>

                <a
                    href={repo[Repository_Type.USER]}
                    target={"_blank"}
                    rel="noopener noreferrer"
                    className="list-group-item list-group-item-action"
                >
                    User Repository
                </a>

                {repo[Repository_Type.GROUP] ? (
                    <a
                        href={repo[Repository_Type.GROUP]}
                        target={"_blank"}
                        rel="noopener noreferrer"
                        className="list-group-item list-group-item-action overflow-ellipses"
                        style={{ textAlign: "left" }}
                    >
                        Group Repository ({enrollment.group?.name})
                    </a>
                ) : null}

                <a
                    href={repo[Repository_Type.ASSIGNMENTS]}
                    target={"_blank"}
                    rel="noopener noreferrer"
                    className="list-group-item list-group-item-action"
                >
                    Assignments
                </a>

                <a
                    href={repo[Repository_Type.INFO]}
                    target={"_blank"}
                    rel="noopener noreferrer"
                    className="list-group-item list-group-item-action"
                >
                    Course Info
                </a>

                {state.hasGroup(courseID.toString()) ? (
                    <Link
                        to={`/course/${courseID}/group`}
                        className="list-group-item list-group-item-action"
                    >
                        View Group
                    </Link>
                ) : (
                    <Link
                        to={`/course/${courseID}/group`}
                        className="list-group-item list-group-item-action list-group-item-success"
                    >
                        Create a Group
                    </Link>
                )}
            </div>
        </div>
    )
}

export default CourseLinks
