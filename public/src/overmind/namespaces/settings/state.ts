export interface UserSettings {
    selectedColor: string
    /* Colors for the criterion/test bar indicator */
    passedColor: string
    failedColor: string
    barWidth: number;
    /* Colors for the review/results table */
    approvedColor: string
    revisionColor: string
    rejectedColor: string

    // Add other settings as needed
}

export const defaultSettings: UserSettings = {
    /* Default colors for the criterion/test bar indicator */
    selectedColor: '#FFFFFF',
    passedColor: '#006F00',
    failedColor: '#FF0000',
    /* Default width (px) for the criterion/test bar indicator */
    barWidth: 5,
    /* Default colors for the review/results table */
    approvedColor: '#ccffcc',
    revisionColor: '#ffc',
    rejectedColor: '#fcc',
}

export const state = {
    settings: defaultSettings,
}
