import React, { useState } from "react"
import { Link, useHistory } from "react-router-dom"
import { useActions, useAppState } from "../../overmind"

const NavBarUser = ():JSX.Element =>{
    const logout = useActions().logout
    const { self, isLoggedIn } = useAppState()
    const history = useHistory()

    const [hidden, setHidden] = useState<boolean>(true)

    const ProfileButton = () => {
        if (isLoggedIn) {
            return (
                <li onClick={() => history.push("/profile")} onMouseEnter={() => setHidden(false)}>
                    <div><img src={self.getAvatarurl()} id="avatar"></img></div>
                </li>
            )
        }
        return null
    }

    const AboutButton = () => {
        return (
            <li key="about">
                <a className="dropdown-item bg-dark">

                
                <Link to="/about" className="Sidebar-items-link">
                    About
                </Link>
                </a>
            </li>
        )
    }

    const AdminButton = () => {
        if (self.getIsadmin()) {
            return (
                <li>
                    <a className="dropdown-item bg-dark">
                     <Link to="/admin" className="Sidebar-items-link">
                        Admin
                    </Link>                       
                    </a>

                </li>
            )
        }
        return null
    }
    const LoginButton = () => {
        if (isLoggedIn) {
            return (
                <li>
                    <a className="dropdown-item bg-dark">
                       <a href="/logout" className="Sidebar-items-link" onClick={() => logout()}>Log out</a> 
                    </a>
                    
                </li>
            )
        }
        return (
            <li>
                <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px" }}>
                    <i className="fa fa-2x fa-github" id="github"></i>
                </a>
            </li>
        )
    }
    return (
        <div className="collapse navbar-collapse ml-auto " id="main_nav">
        <ul className="navbar-nav ml-auto">
            <li className="nav-item dropdown ml-auto " >
            <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px" }}>
            <img className="mrounded-circle" src={self.getAvatarurl()} id="avatar" style={{ height: "40px", borderRadius: "50%" }}></img>
                </a>
                <ul className="dropdown-menu dropdown-menu-center bg-dark" >
                <AboutButton></AboutButton>
                <AdminButton></AdminButton>
                <LoginButton></LoginButton>

                </ul>
            </li>
        </ul>
        </div> 
    )
}
export default NavBarUser