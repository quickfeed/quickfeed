import React from "react"
import { useHistory } from "react-router"



const RedirectButton = ({ to }: { to: string }): JSX.Element => {
    const history = useHistory()

    // The button is hidden if user is currently at the location the button redirects to
    const isHidden = history.location.pathname == to ? true : false

    return (
        <div className={"btn btn-dark redirectButton"} onClick={() => history.push(to)} hidden={isHidden}>
            <i className="fa fa-arrow-left"></i>
        </div>
    )
}

export default RedirectButton
