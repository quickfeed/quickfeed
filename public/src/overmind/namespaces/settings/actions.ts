import { Context } from '../..'
import { UserSettings, defaultSettings } from './state';

export const onInitializeOvermind = async ({ actions, effects }: Context) => {
    // Initialize the API client. *Must* be done before accessing the client.
    const settings = effects.settings.settings.loadSettings()
    if (settings) {
        actions.settings.updateSettings(settings)
    }
}

/* Set the index of the selected review */
export const updateSettings = ({ state, effects }: Context, newSettings: Partial<UserSettings>) => {
    state.settings.settings = { ...state.settings.settings, ...newSettings };
    effects.settings.settings.saveSettings(state.settings.settings)
};

export const resetSettings = ({ state, effects }: Context) => {
    state.settings.settings = defaultSettings;
    effects.settings.settings.saveSettings(state.settings.settings)
};