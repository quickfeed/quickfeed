import React from 'react'
import { useActions, useAppState } from '../../overmind'
import { UserSettings } from '../../overmind/namespaces/settings/state'

export interface ColorOption {
    name: string
    colors: Partial<UserSettings>
}

interface ColorPickerProps {
    colorOptions: ColorOption[]
}

const ColorPicker: React.FC<ColorPickerProps> = ({ colorOptions }) => {
    const { settings } = useAppState()
    const actions = useActions()
    const [customColors, setCustomColors] = React.useState(false)

    const handleColorChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = event.target
        actions.settings.updateSettings({ [name]: value })
    }

    const handleColorOptionSelect = (colors: Partial<UserSettings>) => {
        actions.settings.updateSettings(colors)
    }

    // Convert a setting key to a user-friendly label
    const toLabel = (key: string) => key.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')

    // Get all keys included in the color options
    const colorKeys = Object.keys(colorOptions[0].colors)

    return (
        <div className="form-group mb-3">
            {colorOptions.map(({ name, colors }) => (
                <button
                    key={name}
                    onClick={() => handleColorOptionSelect(colors)}
                    className={`btn uniform-button m-2 ${Object.keys(colors).every(key => settings.settings[key] === colors[key]) ? 'active' : ''
                        }`}
                    style={{
                        display: 'inline-block',
                        margin: '10px',
                        padding: '10px',
                        border: '1px solid gray',
                    }}
                >
                    <div className="mb-2">{name}</div>
                    <div style={{ display: 'flex', justifyContent: 'center' }}>
                        {Object.entries(colors).map(([key, color]) => (
                            <span
                                key={key}
                                style={{ backgroundColor: color, width: '20px', height: '20px', marginRight: '5px' }}
                            />
                        ))}
                    </div>
                </button>
            ))}
            <button
                className="btn uniform-button m-2"
                onClick={() => setCustomColors(!customColors)}
            >
                Custom Colors
            </button>

            {customColors &&
                colorKeys.map(key => (
                    <div className="input-group mb-2" key={key}>
                        <div className="input-group-prepend">
                            <span className="input-group-text fixed-prepend">{toLabel(key)}</span>
                        </div>
                        <input
                            type="color"
                            className="form-control"
                            id={key}
                            name={key}
                            value={settings.settings[key]}
                            onChange={handleColorChange}
                        />
                    </div>
                ))}
        </div>
    )
}

export default ColorPicker
