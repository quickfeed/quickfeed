import React from "react"
import { Link } from "react-router-dom"

interface NavMenuItemProps {
    to?: string
    href?: string
    onClick?: () => void
    children: React.ReactNode
}

/** NavMenuItem is a reusable menu item for the navbar dropdown menu.
 *  Use `to` for internal links (React Router), `href` for external links.
 */
const NavMenuItem = ({ to, href, onClick, children }: NavMenuItemProps) => {
    if (to) {
        return (
            <li>
                <Link to={to} onClick={onClick}>
                    {children}
                </Link>
            </li>
        )
    }

    return (
        <li>
            <a href={href} onClick={onClick}>
                {children}
            </a>
        </li>
    )
}

export default NavMenuItem
