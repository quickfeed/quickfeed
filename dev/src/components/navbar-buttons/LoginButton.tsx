import React from "react"
import { useActions, useAppState } from "../../overmind"

const LoginButton = () => {
    const { isLoggedIn } = useAppState()
    const logout = useActions().logout
    if (isLoggedIn) {
        return (
            <li>
                <a className="dropdown-item bg-dark">
                    <a href="/logout" className="sidebar-items-link" style={{ color: "#d4d4d4" }} onClick={() => logout()}>Log out</a>
                </a>
            </li>
        )
    }
    return (
        <li>
            <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px" }}>
                <i className="fa fa-2x fa-github" id="github" />
            </a>
        </li>
    )
}

export default LoginButton
