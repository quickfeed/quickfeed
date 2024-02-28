import React, { createContext, useContext, ReactNode, useEffect } from 'react'
import { useAppState } from './overmind'

interface ThemeContextProps {
    children: ReactNode
}

const ThemeContext = createContext({})

export const ThemeProvider: React.FC<ThemeContextProps> = ({ children }) => {
    const { settings } = useAppState()

    useEffect(() => {
        // Update CSS variables when theme settings change
        document.documentElement.style.setProperty('--passed-color', settings.settings.passedColor)
        document.documentElement.style.setProperty('--failed-color', settings.settings.failedColor)
        document.documentElement.style.setProperty('--bar-width', settings.settings.barWidth + 'px')
        document.documentElement.style.setProperty('--approved-color', settings.settings.approvedColor)
        document.documentElement.style.setProperty('--revision-color', settings.settings.revisionColor)
        document.documentElement.style.setProperty('--rejected-color', settings.settings.rejectedColor)
    }, [settings.settings.passedColor, settings.settings.failedColor, settings.settings.barWidth, settings.settings.approvedColor, settings.settings.revisionColor, settings.settings.rejectedColor])

    return <ThemeContext.Provider value={{}}>{children}</ThemeContext.Provider>
};

export const useTheme = () => useContext(ThemeContext)
