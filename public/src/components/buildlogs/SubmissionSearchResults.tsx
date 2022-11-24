import React from "react"
import { UserCourseSubmissions } from "../../overmind/state"
import SubmissionLogs from "./SubmissionLogs"


const SubmissionSearchResults = ({ courseSubmissions }: { courseSubmissions: UserCourseSubmissions[] }) => {

    if (courseSubmissions.length === 0) {
        return <p>No results found</p>
    }

    return (
        <ul>
            {courseSubmissions.map((userSubmissions, idx) => {
                return <BuildLogs key={idx} link={userSubmissions} />
            })}
        </ul>
    )
}

const BuildLogs = ({ link }: { link: UserCourseSubmissions }) => {

    if (!link.user || !link.submissions) {
        return null
    }

    const submissions = link.submissions.map((subLink, idx) => {
        return <SubmissionLogs key={idx} subLink={subLink} />
    })

    return (
        <ul>
            <h3>{link.user.Name}</h3>
            {submissions}
        </ul>
    )
}

export default SubmissionSearchResults
