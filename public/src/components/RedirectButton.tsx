import React, { useEffect } from "react"
import { useHistory } from "react-router"
import { useAppState } from "../overmind"


const RedirectButton = ({to}: {to: string}) => {
    const state = useAppState()
    const history = useHistory()
    const hide = history.location.pathname == to ? true : false
    
    useEffect(() => {

    }, [])

    return (
        <div className={"btn btn-dark redirectButton"} onClick={() => history.push(to)} hidden={hide}>
            <i className="fa fa-arrow-left"></i>
        </div>
    )
}

export default RedirectButton