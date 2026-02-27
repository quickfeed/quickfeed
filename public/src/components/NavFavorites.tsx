import React from "react"
import { useAppState } from "../overmind"
import { useNavigate } from "react-router-dom"
import NavBarCourse from "./navbar/NavBarCourse"
import { isEnrolled, isVisible } from "../Helpers"

const NavFavorites = () => {
    const state = useAppState()
    const navigate = useNavigate()

    const visible = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment))

    const courses = visible.map((enrollment) => {
        return <NavBarCourse key={enrollment.ID.toString()} enrollment={enrollment} />
    })

    return (
        <nav
            className={`
                fixed left-0 w-64 h-screen bg-base-300 shadow-xl overflow-y-auto
                transition-transform duration-200 ease-in-out
                ${state.showFavorites ? "translate-x-0" : "-translate-x-full"}
                scrollbar-hide
            `}
            style={{ top: 'var(--navbar-height)' }}
        >
            <ul className="menu [&_li>*]:rounded-none p-0 w-full">
                {courses}
                {state.isLoggedIn && (
                    <li key="all" className="w-full mt-2">
                        <button
                            onClick={() => navigate("/courses")}
                            className="flex justify-center items-center gap-2 h-16 font-bold hover:bg-base-100 rounded-none w-full cursor-pointer"
                        >

                            <span>View all courses</span>
                        </button>
                    </li>
                )}
            </ul>
        </nav>
    )
}

export default NavFavorites
