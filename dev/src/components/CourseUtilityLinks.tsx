import React from "react"
import { Link } from "react-router-dom"
import { Repository } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useAppState } from "../overmind"


/** CourseUtilityLinks displays various repository links for the current course, in addition to links to take the user to the group page. */
const CourseUtilityLinks = (): JSX.Element => {
    const state = useAppState()
    const courseID = getCourseID()
    const enrollment = state.enrollmentsByCourseID[courseID]
    const repo = state.repositories[courseID]

    return (
        <div className="col-sm-3" >
            <div className="list-group">
                <div className="list-group-item list-group-item-action active text-center">
                    <h6>
                        <strong>Utility</strong>
                    </h6>
                </div>

                <a href={repo[Repository.Type.USER]} target={"_blank"} rel="noopener noreferrer" className="list-group-item list-group-item-action">
                    User Repository
                </a>

                {repo[Repository.Type.GROUP] ? (
                    <a href={repo[Repository.Type.GROUP]} target={"_blank"} rel="noopener noreferrer" className="list-group-item list-group-item-action overflow-ellipses" style={{ textAlign: "left" }}>
                        Group Repository ({enrollment.getGroup()?.getName()})
                    </a>
                ) : null}

                <a href={repo[Repository.Type.ASSIGNMENTS]} target={"_blank"} rel="noopener noreferrer" className="list-group-item list-group-item-action">
                    Assignments
                </a>

                <a href={repo[Repository.Type.COURSEINFO]} target={"_blank"} rel="noopener noreferrer" className="list-group-item list-group-item-action">
                    Course Info
                </a>

                {enrollment.hasGroup() ?
                    <Link to={"/course/" + courseID + "/group"} className="list-group-item list-group-item-action">
                        View Group
                    </Link>
                    : <Link to={"/course/" + courseID + "/group/create"} className="list-group-item list-group-item-action list-group-item-success">
                        Create a Group
                    </Link>}
            </div>
        </div>
    )
}

export default CourseUtilityLinks
