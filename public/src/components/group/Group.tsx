import React from "react"
import { useOvermind } from "../../overmind"
import SubmissionsTable from "../SubmissionsTable"

const Group = (props: {courseID: number}) => {
    const {state} = useOvermind()
    return (
        <React.Fragment>
        <div className="box">
            <div className="jumbotron">
                <div className="centerblock container">
                    <h1>{state.userGroup[props.courseID].getName()}</h1>
                    {state.enrollmentsByCourseId[props.courseID].getCourse()?.getName()}
                </div>
            </div>
            {state.userGroup[props.courseID].getUsersList().map(user => 
                <li key={user.getId()} className="list-group-item">
                                    <img src={user.getAvatarurl()} style={{width: "23px", marginRight: "10px", borderRadius: "50%"}}></img>
                        {user.getName()} 
                </li>
            )}
            <br />
            <SubmissionsTable courseID={props.courseID} group={true} />
        </div>
        </React.Fragment>
    )
}
export default Group