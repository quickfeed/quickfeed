import React from "react";
import { useActions, useAppState } from "../../overmind";

const colorPairs = [
    {
        name: 'Default',
        passedColor: '#006F00', // Green
        failedColor: '#FF0000', // Red
    },
    {
        name: 'Color Blind Friendly 1',
        passedColor: '#56B4E9', // Sky Blue
        failedColor: '#E69F00', // Orange
    },
    {
        name: 'Color Blind Friendly 2',
        passedColor: '#1A85FF', // Blue
        failedColor: '#D41159', // Pink
    }
    // Add more pairs as needed
]

const ResultsColorPicker = () => {
    const { settings } = useAppState()
    const actions = useActions()
    const [customColors, setCustomColors] = React.useState(false)

    const handleColorChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        // update either the passed or failed color depending on which input was changed
        if (event.target.name === 'passed') {
            actions.settings.updateSettings({ passedColor: event.target.value })
        } else {
            actions.settings.updateSettings({ failedColor: event.target.value })
        }
    };
    const handleColorPairSelect = (passedColor: string, failedColor: string) => {
        actions.settings.updateSettings({ passedColor, failedColor });
    }
    return (
        <div className="form-group mb-3">
            {colorPairs.map((pair) => (
                <button
                    type='button'
                    key={pair.name}
                    onClick={() => handleColorPairSelect(pair.passedColor, pair.failedColor)}
                    className={`btn uniform-button m-2 ${settings.settings.passedColor === pair.passedColor && settings.settings.failedColor === pair.failedColor ? 'active' : 'disabled'}`}
                    style={{
                        display: 'inline-block',
                        margin: '10px',
                        padding: '10px',
                        border: (settings.settings.passedColor === pair.passedColor && settings.settings.failedColor === pair.failedColor) ? '2px solid black' : '1px solid gray',
                    }}
                >
                    <div className="mb-2">{pair.name}</div>
                    <div style={{ display: 'flex', justifyContent: 'center' }}>
                        <span style={{ backgroundColor: pair.passedColor, width: '20px', height: '20px', marginRight: '5px' }}></span>
                        <span style={{ backgroundColor: pair.failedColor, width: '20px', height: '20px' }}></span>
                    </div>
                </button>
            ))}
            <button
                type='button'
                className='btn uniform-button m-2'
                style={{
                    display: 'inline-block',
                    margin: '10px',
                    padding: '10px',
                    border: customColors ? '2px solid black' : '1px solid gray',
                }}
                onClick={() => setCustomColors(!customColors)}>Custom Colors</button>

            {customColors && (
                <>
                    <div className="input-group mb-2">
                        <div className="input-group-prepend">
                            <div className="input-group-text">Passed Color</div>
                        </div>
                        <input type='color' className='form-control' id='passed' name='passed' value={settings.settings.passedColor} onChange={handleColorChange} />
                    </div>

                    <div className="input-group mb-2">
                        <div className="input-group-prepend">
                            <div className="input-group-text">Failed Color</div>
                        </div>
                        <input type='color' className='form-control' id='failed' name='failed' value={settings.settings.failedColor} onChange={handleColorChange} />
                    </div>
                </>
            )}
        </div>
    )
}

export default ResultsColorPicker