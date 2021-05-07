import React from "react"
import { useEffect } from "react"
import { Redirect, RouteComponentProps } from "react-router-dom"
import { sortByField } from "../Helpers"
import { useOvermind } from "../overmind"
import { Enrollment } from "../../proto/ag_pb"


export const Members = (props: RouteComponentProps<{id?: string | undefined}>) => {
    const {state, actions} = useOvermind()
    let courseID = Number(props.match.params.id)

    useEffect(() => {

    }, [])

    if (courseID && state.enrollmentsByCourseId[courseID].getStatus() === Enrollment.UserStatus.TEACHER) {
        return (
            <div className='row '>
                <div className="card well  col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>Pending</div>
                        <ul className="list-group list-group-flush">

                        
                        {state.courseEnrollments[courseID].map(user => {
                            if (user.getStatus() === Enrollment.UserStatus.PENDING) {
                                return (
                                    <li key={user.getUserid()} className={"list-group-item" }>{user.getUser()?.getName()} 
                                        <span className={"badge badge-primary"} onClick={() => actions.updateEnrollment({enrollment: user, status: Enrollment.UserStatus.STUDENT})}>
                                            Approve
                                        </span>
                                    </li>
                                )
                            }
                        })} 
                        </ul>
                </div>
                <div className="col">
                    <div className="card well  col-md-offset-2">
                        <div className="card-header" style={{textAlign: "center"}}>Members</div>
                            <ul className="list-group list-group-flush">
                                {sortByField(state.courseEnrollments[courseID], [], Enrollment.prototype.getStatus, true).map((user: Enrollment) => {
                                return (
                                    <li key={user.getUserid()} className={"list-group-item" }>{user.getUser()?.getName()} {user.getId()} <i style={{float: "right"}} className={"badge badge-" + (user.getStatus() === 2 ? "primary" : user.getStatus() === 3 ? "danger" : "info")}>{user.getStatus() === 2 ? "Student" : user.getStatus() === 3 ? "Teacher" : user.getStatus() == 1 ?  "Pending" : "None"}</i></li>
                                )
                                })} 
                            </ul>
                    </div>
                </div>
            </div>
        )
    }
    return (<Redirect to="/" />)
}

export default Members