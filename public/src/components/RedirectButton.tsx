import React from "react"
import { useHistory } from "react-router"


const RedirectButton = ({to}: {to: string}) => {
    const history = useHistory()
    const hide = history.location.pathname == to ? true : false
    return (
        <div className={"btn btn-dark redirectButton"} onClick={() => history.push(to)} hidden={hide}>
            <i className="fa fa-arrow-left"></i>
        </div>
    )
}

export default RedirectButton