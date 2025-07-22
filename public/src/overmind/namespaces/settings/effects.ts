// src/overmind/effects.ts
import { UserSettings } from './state'

export const settings = {
    saveSettings(settings: UserSettings) {
        localStorage.setItem('userSettings', JSON.stringify(settings))
    },
    loadSettings(): UserSettings {
        const settings = localStorage.getItem('userSettings')
        return settings ? JSON.parse(settings) : null
    },
}
