import React from "react"
import { Link, useHistory } from "react-router-dom"

export interface NavLink {
    text: string
    to: string
    icons?: ({ text: string | number, classname: string } | null)[]
    jsx?: React.JSX.Element
}

const NavBarLink = ({ link: { text, to, icons, jsx } }: { link: NavLink }) => {
    const history = useHistory()

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
        <li onClick={() => history.push(to)} role="button" aria-hidden="true"> {/* skipcq: JS-0761 */}
            <div className="col" id="title">
                <Link to={to}>{text}</Link>
            </div>
            <div className="col">
                {iconElements}
                {jsx ?? null}
            </div>
        </li>
    )
}

export default NavBarLink
