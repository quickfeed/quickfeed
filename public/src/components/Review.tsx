import React, { useEffect } from "react"
import { useOvermind } from "../overmind"



export const Review = () => {
    const {actions} = useOvermind()
    useEffect(() => {
        console.log("REVIEW TEST")
        actions.getAllCourseSubmissions(2)
    }, [])
    return (
        <div>Test</div>
    )
}

export default Review