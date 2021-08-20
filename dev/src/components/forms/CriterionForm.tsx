import React from "react";
import { Assignment, GradingCriterion } from "../../../proto/ag/ag_pb";
import { useActions, useGrpc } from "../../overmind";



const CriterionForm = ({criterion, benchmarkID, assignment, setEditing}: {criterion?: GradingCriterion, benchmarkID?: number, assignment?: Assignment, setEditing: React.Dispatch<React.SetStateAction<number>>}): JSX.Element => {
    
    const grpc = useGrpc().grpcMan
    const actions = useActions()
    const newCriterion = new GradingCriterion()

    if (benchmarkID) {
        newCriterion.setBenchmarkid(benchmarkID)
    }

    const handleCriteria = (event: React.FormEvent<HTMLInputElement>) => {
        const {name, value} = event.currentTarget
        switch (name) {
            case "criterion":
                criterion ? criterion.setDescription(value) : newCriterion.setDescription(value)
                break
            case "points":
                criterion ? criterion.setPoints(Number(value)) : newCriterion.setPoints(Number(value))
                break
            case "submit":
                criterion ? grpc.updateCriterion(criterion) : actions.createCriterion({criterion: newCriterion, assignment: assignment as Assignment})
                setEditing(-1)
                break
            default:
                break
        }
    }

    const submitText = criterion ? "Edit" : "Create"

    return (
        <>
        <input type="text" defaultValue={criterion?.getDescription()} name="criterion" onKeyUp={e => {handleCriteria(e)}}></input>
        <input type="number" defaultValue={criterion?.getPoints()} name="points" placeholder="Points" onKeyUp={e => {handleCriteria(e)}}></input>
        <input type="submit" value={submitText} name="submit" onClick={e => handleCriteria(e)}></input>
        </>
    )
}

export default CriterionForm