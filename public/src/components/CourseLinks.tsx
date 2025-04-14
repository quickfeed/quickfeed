import React from "react"
import { Link } from "react-router-dom"
import { Repository_Type } from "../../proto/qf/types_pb"
import { getCourseID } from "../Helpers"
import { useAppState } from "../overmind"

type link = {
    type: Repository_Type,
    text: string
    className?: string,
    style?: React.CSSProperties
}

/** CourseLinks displays various repository links for the current course, in addition to links to take the user to the group page. */
const CourseLinks = () => {
    const state = useAppState()
    const courseID = getCourseID()
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    const repo = state.repositories[courseID.toString()]
    const groupName = enrollment.group ? `(${enrollment.group?.name})` : ""

    const links: link[] = [
        { type: Repository_Type.USER, text: "User Repository" },
        { type: Repository_Type.GROUP, text: `Group Repository ${groupName}`, style: { textAlign: "left" }, className: "overflow-ellipses" },
        { type: Repository_Type.ASSIGNMENTS, text: "Assignments" },
        { type: Repository_Type.INFO, text: "Course Info" }
    ]

    const LinkElement = ({ link }: { link: link }) => {
        if (repo[link.type] === undefined) {
            return null
        }

        return <a
            href={repo[link.type]}
            target={"_blank"}
            rel="noopener noreferrer"
            className={`list-group-item list-group-item-action ${link.className ?? ""}`}
            style={link.style}
        >
            {link.text}
        </a>
    }

    const groupLinkText = state.hasGroup(courseID.toString()) ? "View Group" : "Create a Group"
    const groupLinkClassName = state.hasGroup(courseID.toString()) ? "" : "list-group-item-success"
    return (
        <div className="col-lg-3">
            <div className="list-group width-resize">
                <div className="list-group-item list-group-item-action active text-center">
                    <h6>
                        <strong>Links</strong>
                    </h6>
                </div>

                {links.map(link => { return <LinkElement key={link.type} link={link} /> })}

                <Link
                    to={`/course/${courseID}/group`}
                    className={`list-group-item list-group-item-action ${groupLinkClassName}`}
                >
                    {groupLinkText}
                </Link>

            </div>
        </div>
    )
}

export default CourseLinks
