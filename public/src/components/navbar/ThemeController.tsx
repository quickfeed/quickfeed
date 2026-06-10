import React from "react"
import { useActions, useAppState } from "../../overmind"

interface Theme {
    name: string
    displayName: string
}

const themes: Theme[] = [
    { name: "light", displayName: "Light" },
    { name: "dark", displayName: "Dark" },
    { name: "cupcake", displayName: "Cupcake" },
    { name: "synthwave", displayName: "Synthwave" },
    { name: "cyberpunk", displayName: "Cyberpunk" },
    { name: "forest", displayName: "Forest" },
    { name: "aqua", displayName: "Aqua" }
]

/** ThemeController allows users to select from multiple DaisyUI themes */
const ThemeController = () => {
    const state = useAppState()
    const actions = useActions().global
    const currentTheme = state.theme || "light"

    const handleThemeChange = (themeName: string) => {
        actions.setTheme(themeName)
    }

    return (
        <div className="dropdown dropdown-end">
            <div
                tabIndex={0}
                role="button"
                className="btn btn-ghost btn-circle"
                aria-label="Select theme"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    className="w-5 h-5 stroke-current"
                >
                    <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
                    />
                </svg>
            </div>
            <ul
                tabIndex={0}
                className="menu dropdown-content bg-base-100 rounded-box z-[1] mt-3 w-52 p-2 shadow max-h-96 overflow-y-auto"
            >
                {themes.map((theme) => (
                    <li key={theme.name}>
                        <button
                            onClick={() => handleThemeChange(theme.name)}
                            className={currentTheme === theme.name ? "active" : ""}
                        >
                            <span className="flex items-center justify-between w-full">
                                {theme.displayName}
                                {currentTheme === theme.name && (
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        className="h-4 w-4"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        stroke="currentColor"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M5 13l4 4L19 7"
                                        />
                                    </svg>
                                )}
                            </span>
                        </button>
                    </li>
                ))}
            </ul>
        </div>
    )
}

export default ThemeController
