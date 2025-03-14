import React from "react"
import { Link, useHistory } from "react-router-dom"

export interface NavLink {
    link: { text: string, to: string }
    icons?: ({ text: string | number, classname: string } | null)[],
    jsx?: React.JSX.Element
}

const NavBarLink = (props: NavLink) => {
    const history = useHistory()

    const icons: React.JSX.Element[] = []
    if (props.icons) {
        props.icons.forEach((icon, index) => {
            if (icon) {
                icons.push(
                    <div key={index} id="icon" className={icon.classname + " ml-2"}>
                        {icon.text}
                    </div>
                )
            }
        })
    }
    return (
        <li onClick={() => history.push(props.link.to)} role="button" aria-hidden="true">
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
