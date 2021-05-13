import React from "react"
import { Link, useHistory } from "react-router-dom"
import { useOvermind } from "../overmind"


const NavBarFooter = () => {
    const {
        state: {user: {id,avatarurl,isadmin}, theme},
        actions: {changeTheme}
    } = useOvermind()
    const history = useHistory()


    const LogInButton = () => {
        if (id > 0) {
            return (
            <li>
                <a href="/logout" className="Sidebar-items-link">Log out</a>
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
            <Link to="/about" className="Sidebar-items-link">
                About
            </Link>
        </li>

        <li key="theme" onClick={() => changeTheme()}>
                <i className={theme === "light" ? "icon fa fa-sun-o fa-lg fa-flip-horizontal" : "icon fa fa-moon-o fa-lg flip-horizontal"} style={{color: "white"}}></i> 
        </li>

        <LogInButton />

        {isadmin ? 
            <li key="admin">
                <Link to="/admin" className="Sidebar-items-link">
                    Admin
                </Link>
            </li>
            : ""
        }

        {id > 0 ? 
            <li key="profile" onClick={() => history.push("/profile")}>
                <div><img src={avatarurl} id="avatar"></img></div>    
            </li>
        : ""}

        </div>
    )
}

export default NavBarFooter