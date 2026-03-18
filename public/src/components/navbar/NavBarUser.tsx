import React from "react"
import { useActions, useAppState } from "../../overmind"
import StreamStatus from "./StreamStatus"
import ThemeController from "./ThemeController"
import NavMenuItem from "../navbar-buttons/NavMenuItem"
import { nextURL } from "../../Helpers"

const NavBarUser = () => {
    const { self, isLoggedIn } = useAppState()
    const actions = useActions().global

    if (!isLoggedIn) {
        return (
            <button className="btn bg-black text-white border-black">
                <i className="fa fa-github" />
                <a href={`/auth/github?next=${nextURL()}`} className="ml-2">Sign in with GitHub</a>
            </button>
        )
    }

    return (
        <div className="flex items-center gap-2">
            <ThemeController />
            <StreamStatus />

            <div className="dropdown dropdown-end">
                <div
                    tabIndex={0}
                    role="button"
                    className="btn btn-ghost btn-circle avatar"
                >
                    <div className="w-10 rounded-full">
                        <img src={self.AvatarURL} alt="User avatar" />
                    </div>
                </div>
                <ul
                    tabIndex={0}
                    className="menu dropdown-content bg-base-100 rounded-box z-[1] mt-3 w-52 p-2 shadow"
                >
                    <NavMenuItem to="/profile">Profile</NavMenuItem>
                    <NavMenuItem to="/about">About</NavMenuItem>
                    {self.IsAdmin && <NavMenuItem to="/admin">Admin</NavMenuItem>}
                    <NavMenuItem href="/logout" onClick={() => actions.logout()}>Log out</NavMenuItem>
                </ul>
            </div>
        </div>
    )
}

export default NavBarUser
