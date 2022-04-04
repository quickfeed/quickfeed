import React from "react"
import { useActions, useAppState } from "../overmind"
import { Link } from "react-router-dom"
import NavFavorites from "./NavFavorites"
import NavBarUser from "./navbar/NavBarUser"

//TODO Review the NavBar behaviour.
const NavBar = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    const onCourseClick = () => {
        actions.setActiveFavorite(!state.showFavorites)
    }

    return (
        <nav className="navbar navbar-expand-lg" style={{ backgroundColor: "#222", color: "#d4d4d4" }} id="main" >
            {!state.showFavorites &&
                <a className="navbar-brand" style={{ marginLeft: "30px", fontWeight: "bold" }}>
                    <Link to="/" style={{ fontWeight: "bold", fontSize: "30px", color: "#d4d4d4" }}>
                        QuickFeed
                    </Link>
                </a>
            }
            {!state.isLoggedIn &&
                <div className="navbar-collapse ml-auto">
                    <ul className="ms-auto ml-auto list-unstyled">
                        <li className="nav-item">
                            <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px", color: "#d4d4d4", marginLeft: "40px" }}>
                                Sign in with <i className="fa fa-2x fa-github align-middle ms-auto " id="github" />
                            </a>
                        </li>
                    </ul>
                </div>
            }
            {state.isLoggedIn &&
                <ul className="mr-auto me-auto list-unstyled">
                    <a className="clickable" onClick={() => { onCourseClick() }}
                        style={{ paddingTop: "15px", marginLeft: "10px", fontSize: 25 }}>☰</a>
                </ul>
            }
            <ul className="ms-auto ml-auto list-unstyled" style={{ marginRight: "55px", paddingTop: "15px" }}>
                <NavBarUser />
            </ul>
            <div>
                {state.showFavorites &&
                    <NavFavorites />
                }
            </div>
        </nav>
    )
}

export default NavBar
