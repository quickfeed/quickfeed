import React, { useEffect } from "react"
import { RouteComponentProps } from "react-router-dom"
import { useOvermind } from "../overmind"

/* Lists all groups for a given course. */
export const Groups = (props: RouteComponentProps<{id?: string | undefined}>) => {
    const {state, actions} = useOvermind()
    const courseID = Number(props.match.params.id)

    useEffect(() => {
        actions.getGroupsByCourse(courseID)
    }, [state.groups])

    // Generates JSX.Element array containing all groups for the course
    const Groups = state.groups[courseID]?.map(group => {
        return (
            <ul className="list-group list-group-flush">
                <li className="list-group-item">Name: {group.getName()}</li>
                <li className="list-group-item">Status: {group.getStatus()}</li>
                { // Populates the unordered list with list elements for every user in the group
                    group.getEnrollmentsList().map(enrol => 
                        <li key={enrol.getId()} className="list-group-item">
                            <img src={enrol.getUser()?.getAvatarurl()} style={{width: "40px", borderRadius: "50%", marginRight: "10px"}}></img>
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