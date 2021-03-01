import { useEffect, useState } from "react"
import { RouteComponentProps, RouteProps, RouterProps } from "react-router"
import { useOvermind } from "../overmind"

import { Courses, Enrollment } from "../proto/ag_pb"


interface MatchProps {
    id: string
}

const Course = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions } = useOvermind()

    const [enrollment, setEnrollment] = useState(new Enrollment())
    useEffect(() => {
        actions.getEnrollmentsByUser()
        .then(success => {
            if (success){
                const enrol = actions.getEnrollmentByCourseId(Number(props.match.params.id))
                if (enrol !== null) {
                    setEnrollment(enrol)
                    actions.getSubmissions(enrol.getCourseid())
                }
            }
        })
    }, [])

    const getSubmissions = state.submissions.map(submission => {
        return (
            <div>
                <h1>{submission.getScore()} / 100</h1>
                <code>{submission.getBuildinfo()}</code>
            </div>
        )
    })

    if (enrollment !== null){
        return (
        <div className="box">
            <h1>Welcome to {enrollment.getCourse()?.getName()}, {enrollment.getUser()?.getName()}! You are a {enrollment.getStatus() == Enrollment.UserStatus.STUDENT ? ("student") : ("teacher")}</h1>
            {getSubmissions}
        </div>)
    }
    return <h1>404 Not Found</h1>
}

export default Course