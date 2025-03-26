import React, { useCallback } from "react"
import { useHistory } from "react-router"


const RedirectButton = ({ to }: { to: string }) => {
    const history = useHistory()

    // The button is hidden if user is currently at the location the button redirects to
    const isHidden = history.location.pathname == to ? true : false

    const handleRedirect = useCallback(() => history.push(to), [to, history])

    return (
        <div className={"btn btn-dark redirectButton"} onClick={handleRedirect} hidden={isHidden}>
            <i className="fa fa-arrow-left" />
        </div>
    )
}

export default RedirectButton
