import React from "react"


const Loading = (): JSX.Element => {
    return (
        <div className="centered">
            <i className="fa fa-refresh fa-spin fa-3x fa-fw"></i>
            <p><strong>Loading...</strong></p>
        </div>
    )
}

export default Loading
