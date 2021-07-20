import React, { useEffect } from "react"
import { useOvermind } from "../overmind"

/* Lists all groups for a given course. */
export const Groups = (props: {courseID: number}) => {
    const {state, actions} = useOvermind()
    const courseID = props.courseID

    const style = {width: "40px", borderRadius: "50%", marginRight: "10px"}

    useEffect(() => {
        actions.getGroupsByCourse(courseID)
    }, [state.groups])

    // Generates JSX.Element array containing all groups for the course
    const Groups = state.groups[courseID]?.map(group => {
        return (
            <ul key={group.getId()} className="list-group list-group-flush">
                <li className="list-group-item">{group.getName()}</li>
                <li className="list-group-item">
                    {group.getStatus() == 0 ? "Pending" : "Approved"}
                    <span className="badge badge-primary float-right">Approve</span>
                </li>
                { // Populates the unordered list with list elements for every user in the group
                    group.getEnrollmentsList().map(enrol => 
                        <li key={enrol.getId()} className="list-group-item">
                            <img src={enrol.getUser()?.getAvatarurl()} style={style}></img>
                            {enrol.getUser()?.getName()} 
                            <span>{enrol.getSlipdaysremaining()}</span>
                        </li>
                )}
            </ul>
        )
    })

    return (
        <div className="box">
            <div className="card well" style={{width: "400px"}}>
                <div className="card-header">Groups</div>
                {Groups}
            </div>
        </div>
    )
}

export default Groups