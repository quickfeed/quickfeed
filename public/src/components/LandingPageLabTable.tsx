import React from 'react'
import { Link, useHistory } from 'react-router-dom'
import { getFormattedDeadline, timeFormatter } from '../Helpers'
import { useOvermind, useReaction } from '../overmind'

const LandingPageLabTable = () => {
    const { state } = useOvermind()
    
    const history = useHistory()
    const handleRowClick = (crsid:string,assignid:number) => {
      history.push(`/course/${crsid}/${assignid}`)
    }  
    //replace {} with a type of dictionary/record
    
    const makeTable = (): JSX.Element[] => {
        let table: JSX.Element[] = []
        const now = new Date()
        for (const courseID in state.assignments) {
            let crsName = state.courses.find(course => course.getId() === Number(courseID))?.getCode()
            state.assignments[courseID].map(assignment => {
                if(state.submissions[courseID]) {
                    let submission = state.submissions[courseID].find(submission => assignment.getId() === submission.getAssignmentid())
                    if(submission){
                        
                        let timetoDeadline = timeFormatter(new Date(assignment.getDeadline()).getTime(), now)
                        table.push(
                            
                            // eslint-disable-next-line quotes
                            <tr key = {assignment.getId()} className= {`clickable-row ${timetoDeadline[0] ? String(timetoDeadline[1]):''}`} onClick={()=>{handleRowClick(courseID,assignment.getId())}} >
                                <th scope="row">{crsName}</th>
                                <td>{assignment.getName()}</td>
                                <td>{submission.getScore()} / {assignment.getScorelimit()}</td>
                                <td>{getFormattedDeadline(assignment.getDeadline())}</td>
                                <td>{timetoDeadline[0] ? String(timetoDeadline[2]):''}</td>
                                <td>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? 'Awating approval':(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? 'Approved(Auto approve)(shouldn\'t be in final version)':'Score not high enough'}</td>
                                <td>{assignment.getIsgrouplab() ? 'Yes': 'No'}</td>
                            </tr>
                            
                        )
                    }
                }
            })
        }
        return table

    }
    
    return (
        <div>
            <table className="table table-dark table-hover">
                <thead>
                    <tr>
                        <th scope="col">Course</th>
                        <th scope="col">Assignment Name</th>
                        <th scope="col">Progress</th>
                        <th scope="col">Deadline</th>
                        <th scope="col">Time Left</th>
                        <th scope="col">Status</th>
                        <th scope="col">Grouplab</th>
                    </tr>
                </thead>
                <tbody>
                    {makeTable()}
                </tbody>
            </table>
        </div>
    )
}

export default LandingPageLabTable