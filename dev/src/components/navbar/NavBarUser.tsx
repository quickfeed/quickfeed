import React from "react"
import { Link } from "react-router-dom"
import { useActions, useAppState } from "../../overmind"

const NavBarUser = ():JSX.Element =>{
    const logout = useActions().logout
    const { self, isLoggedIn } = useAppState()


    const AboutButton = () => {
        return (
            <li key="about">
                <a className="dropdown-item bg-dark" >
                    <Link to="/about" className="Sidebar-items-link" style={{color: "#d4d4d4"}}>
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
                    <a className="dropdown-item bg-dark" >
                        <Link to="/admin" className="Sidebar-items-link" style={{color: "#d4d4d4"}}>
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
                       <a href="/logout" className="Sidebar-items-link" style={{color: "#d4d4d4"}} onClick={() => logout()}>Log out</a> 
                    </a>
                </li>
            )
        }
        return (
            <li>
                <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px" }}>
                    <i className="fa fa-2x fa-github" id="github"/>
                </a>
            </li>
        )
    }
    return (
        <div className="navbar-collapse ml-auto" id="main_nav">
            <ul className="navbar-nav ml-auto">
                <li className="nav-item dropdown ml-auto">
                    <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px", marginLeft: "40px" }}>
                        <img className="rounded-circle" src={self.getAvatarurl()} id="avatar" style={{ height: "40px", borderRadius: "50%" }}/>
                    </a>
                    <ul className="dropdown-menu dropdown-menu-center bg-dark">
                        <AboutButton/>
                        <AdminButton/>
                        <LoginButton/>
                    </ul>
                </li>
            </ul>
        </div> 
    )
}

export default NavBarUser