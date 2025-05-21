import React from "react"
import { useLocation, useNavigate } from "react-router"


const RedirectButton = ({ to }: { to: string }) => {
    const navigate = useNavigate()
    const location = useLocation()

    // The button is hidden if user is currently at the location the button redirects to
    const isHidden = location.pathname === to

    return (
        <button
            className="btn btn-dark redirectButton"
            type="button"
            onClick={() => navigate(to)}
            hidden={isHidden}
        >
            <i className="fa fa-arrow-left" />
        </button>
    )
}

export default RedirectButton
