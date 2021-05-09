import React from "react"
import { Link, useHistory } from "react-router-dom"
import { useOvermind } from "../overmind"


const NavBarFooter = () => {
    const {
        state: {user, theme},
        actions: {changeTheme}
    } = useOvermind()
    const history = useHistory()


    const LogInButton = () => {
        if (user.id > 0) {
            return (
            <li>
                <div id="title">
                    <a href="/logout">Log out</a>
                </div>
            </li>
            )
        }
        return (
            <li>
                <a href="/auth/github" style={{textAlign:"center",paddingTop:"15px"}}>
                    Log in with
                    <i className="fa fa-2x fa-github" id="github"></i>
                </a>
            </li>
        )
    }


    return (
        <div className="SidebarFooter">
                
        <li key="about">
            <div id="title">
                <Link to="/about">
                    About
                </Link>
            </div>
        </li>

        <li key="theme" onClick={() => changeTheme()}>
                <i className={theme === "light" ? "icon fa fa-sun-o fa-lg fa-flip-horizontal" : "icon fa fa-moon-o fa-lg flip-horizontal"} style={{color: "white"}}></i> 
        </li>

        <LogInButton />

        {user.isadmin ? 
            <li key="admin">
                <div id="title">
                <Link to="/admin">
                    Admin
                </Link>
                </div>
            </li>
            : ""
        }

        {user.id > 0 ? 
            <li key="profile" onClick={() => history.push("/profile")}>
                <div id="title"><img src={user.avatarurl} id="avatar"></img></div>    
            </li>
        : ""}

        </div>
    )
}

export default NavBarFooter