import React from "react"
import { useNavigate } from "react-router-dom"

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
                <div key={icon.text} className={`${icon.classname} ml-2 w-6 h-[22px] flex items-center justify-center`}>
                    <span className="text-xs">{icon.text}</span>
                </div>
            )
        }
    })

    return (
        <li className="w-full">
            <button
                type="button"
                onClick={() => navigate(to)}
                className="flex justify-between items-center w-full h-16 px-4 hover:bg-base-100 rounded-none"
            >
                <span className="flex-1 text-left">{text}</span>
                <div className="flex items-center gap-1">
                    {iconElements}
                    {jsx ?? null}
                </div>
            </button>
        </li>
    )
}

export default NavBarLink
