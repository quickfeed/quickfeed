import React from "react"
import { Link, useHistory } from "react-router-dom"
export interface NavLink {
    link: {text: string, to: string}
    icons?: ({ text: string | number, classname: string } | null)[], 
    jsx?: JSX.Element
}

const NavBarLink = (props: NavLink): JSX.Element => {
    const history = useHistory()

    const icons: JSX.Element[] = []
    if (props.icons) {
        props.icons.forEach((icon, index) => {
            if (icon) {
                icons.push(
                    <div key={index} id="icon" style={{marginRight: "5px"}} className={icon.classname}>
                        {icon.text}
                    </div>
                )
            }
        })
    }
    return (
        <li className="activeLabs" onClick={() => history.push(props.link.to)}>
            
            <div className="col" id="title">
                <Link to={props.link.to}>{props.link.text}</Link>
            </div>
            <div className="col">
            {icons ? icons : null}
            {props.jsx ? props.jsx : null}
            </div>
        </li>
    )
}

export default NavBarLink