import React from "react"
import { Link, useNavigate } from "react-router-dom"

export interface NavLink {
    text: string
    to: string
    icons?: ({ text: string | number, classname: string } | null)[]
    jsx?: React.JSX.Element
}

const NavBarLink = ({ link: { text, to, icons, jsx } }: { link: NavLink }) => {
    const navigate = useNavigate()

    const iconElements: React.JSX.Element[] = []
    icons?.forEach((icon) => {
        if (icon) {
            iconElements.push(
                <div key={icon.text} id="icon" className={`${icon.classname} ml-2`}>
                    {icon.text}
                </div>
            )
        }
    })

    return (
        <li>
            <button
                type="button"
                onClick={() => navigate(to)}
                className="navbar-link-btn"
                style={{ background: "none", border: "none", padding: 0, width: "100%" }}
            >
                <div className="col" id="title">
                    <Link to={to}>{text}</Link>
                </div>
                <div className="col">
                    {iconElements}
                    {jsx ?? null}
                </div>
            </button>
        </li>
    )
}

export default NavBarLink
