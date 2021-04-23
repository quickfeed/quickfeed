import React from "react"
import { useEffect, useState } from "react"
import { RouteComponentProps } from "react-router-dom"
import { useOvermind } from "../overmind"
import { Enrollment } from "../proto/ag_pb"


export const Members = (props: RouteComponentProps<{id?: string | undefined}>) => {
    const {state, actions} = useOvermind()
    let courseID = Number(props.match.params.id)

    let [enrol, setEnrol] = useState(new Enrollment())

    useEffect(() => {
        console.log(state.courseEnrollments)
    }, [])



    if (courseID) {
        return (
            <div className='row '>
                <div className="card well  col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>Pending</div>
                        <ul className="list-group list-group-flush">
                        {state.courseEnrollments[courseID].map(user => {
                            if (user.getStatus() === Enrollment.UserStatus.PENDING) {
                            return (
                                <li key={user.getUserid()} className={"list-group-item" }>{user.getUser()?.getName()}, {user.getUser()?.getStudentid()} 
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
                        {state.courseEnrollments[courseID].map(user => {
                            return (
                                <li key={user.getUserid()} className={"list-group-item" }>{user.getUser()?.getName()} <i style={{float: "right"}} className={"badge badge-" + (user.getStatus() === 2 ? "primary" : "danger")}>{user.getStatus() === 2 ? "Student" : "Teacher"}</i></li>
                                )
                        })} 
                        </ul>
                </div>
                </div>
            </div>
        )
    }
    return (<div>Test</div>)
}

export default Members