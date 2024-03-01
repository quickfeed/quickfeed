import React, { createContext, useContext, ReactNode, useEffect } from 'react'
import { useAppState } from './overmind'

interface ThemeContextProps {
    children: ReactNode
}

const ThemeContext = createContext({})

export const ThemeProvider: React.FC<ThemeContextProps> = ({ children }) => {
    const { settings } = useAppState()

    useEffect(() => {
        const theme = settings.settings
        Object.entries(theme).forEach(([key, value]) => {
            if (value === undefined) return
            document.documentElement.style.setProperty('--' + key, value)

        })
    }, [settings.settings])

    return <ThemeContext.Provider value={{}}>{children}</ThemeContext.Provider>
};

export const useTheme = () => useContext(ThemeContext)
