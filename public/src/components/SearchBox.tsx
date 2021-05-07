import React from "react"



export const SearchBox = (props: {courseID: number, assignmentID: number}) => {
    return (
        <div>
            <h1>{props.courseID}</h1>
            <h2>{props.assignmentID}</h2>
        </div>
    )
}

export default SearchBox