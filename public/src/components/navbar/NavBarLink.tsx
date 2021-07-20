import React from "react"
import { Link } from "react-router-dom"
import { NavLink } from "../../Helpers"


const NavBarLink = (props: NavLink) => {
    const icons: JSX.Element[] = []
    if (props.icons) {
        props.icons.forEach((icon, index) => {
            icons.push(
                <div key={index} id="icon" className={icon.classname}>
                    {icon.text}
                </div>
            )
        })
    }
    return (
        <li className="activeLabs">
            {icons}
            <div id="title">
                <Link to={props.link.to}>{props.link.text}</Link>
            </div>
        </li>
    )
}

export default NavBarLink