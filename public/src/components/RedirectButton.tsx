import React from "react"
import { useHistory } from "react-router"


const RedirectButton = ({ to }: { to: string }) => {
    const history = useHistory()

    // Path to dashboard when on root of Student or Teacher page
    const path = history.location.pathname == to ? "/" : to

    return (
        <div className={"btn btn-dark redirectButton"} onClick={() => history.push(path)}>
            <i className="fa fa-arrow-left" />
        </div>
    )
}

export default RedirectButton
