export interface UserSettings {
    selectedColor: string;
    passedColor: string;
    failedColor: string;
    barWidth: number;
    // Add other settings as needed
}

export const defaultSettings: UserSettings = {
    selectedColor: '#FFFFFF',
    passedColor: '#006F00',
    failedColor: '#FF0000',
    barWidth: 5,
};

export const state = {
    settings: defaultSettings,
}
