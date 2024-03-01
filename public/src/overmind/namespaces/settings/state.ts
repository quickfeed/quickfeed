export interface UserSettings {
    [key: string]: string | undefined | UserSettings[keyof UserSettings]

    'selected-color'?: string
    /* Colors for the criterion/test bar indicator */
    'passed-color'?: string
    'failed-color'?: string
    'bar-width'?: string;
    /* Colors for the review/results table */
    'approved-color'?: string
    'revision-color'?: string
    'rejected-color'?: string

    // Add other settings as needed
}

export const defaultSettings: UserSettings = {
    /* Default colors for the criterion/test bar indicator */
    'selected-color': '#FFFFFF',
    'passed-color': '#006F00',
    'failed-color': '#FF0000',
    /* Default width (px) for the criterion/test bar indicator */
    'bar-width': '8px',
    /* Default colors for the review/results table */
    'approved-color': '#ccffcc',
    'revision-color': '#ffc',
    'rejected-color': '#fcc',
}

export const state = {
    settings: defaultSettings,
}
