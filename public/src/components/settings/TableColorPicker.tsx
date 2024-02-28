import React from "react";
import { useActions, useAppState } from "../../overmind";

const colorPairs = [
    {
        name: 'Default',
        approvedColor: '#CCFFCC', // Light Green
        revisionColor: '#FFFFCC', // Light Yellow
        rejectedColor: '#FFCCCC', // Light Red
    },
    {
        name: 'Color Blind Friendly 1',
        approvedColor: '#56B4E9', // Sky Blue
        revisionColor: '#E69F00', // Orange
        rejectedColor: '#D41159', // Pink
    },
    {
        name: 'Color Blind Friendly 2',
        approvedColor: '#009E73',
        revisionColor: '#F0E442',
        rejectedColor: '#CC79A7',
    }
]

const TableColorPicker = () => {
    const { settings } = useAppState()
    const actions = useActions()
    const [customColors, setCustomColors] = React.useState(false)

    const handleColorChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        // update either the passed or failed color depending on which input was changed
        switch (event.target.name) {
            case 'approved':
                actions.settings.updateSettings({ approvedColor: event.target.value })
                break
            case 'revision':
                actions.settings.updateSettings({ revisionColor: event.target.value })
                break
            case 'rejected':
                actions.settings.updateSettings({ rejectedColor: event.target.value })
                break
        }
    }
    const handleColorPairSelect = (approvedColor: string, revisionColor: string, rejectedColor: string) => {
        actions.settings.updateSettings({ approvedColor, revisionColor, rejectedColor });
    }

    return (
        <div className="form-group mb-3">
            {colorPairs.map((pair) => (
                <button
                    type='button'
                    key={pair.name}
                    onClick={() => handleColorPairSelect(pair.approvedColor, pair.revisionColor, pair.rejectedColor)}
                    className={`btn uniform-button m-2 ${settings.settings.approvedColor === pair.approvedColor && settings.settings.revisionColor === pair.revisionColor && settings.settings.rejectedColor === pair.rejectedColor ? 'active' : 'disabled'}`}
                    style={{
                        display: 'inline-block',
                        margin: '10px',
                        padding: '10px',
                        border: (settings.settings.approvedColor === pair.approvedColor && settings.settings.revisionColor === pair.revisionColor && settings.settings.rejectedColor === pair.rejectedColor) ? '2px solid black' : '1px solid gray',
                    }}
                >
                    <div className="mb-2">{pair.name}</div>
                    <div style={{ display: 'flex', justifyContent: 'center' }}>
                        <span style={{ backgroundColor: pair.approvedColor, width: '20px', height: '20px', marginRight: '5px' }}></span>
                        <span style={{ backgroundColor: pair.revisionColor, width: '20px', height: '20px', marginRight: '5px' }}></span>
                        <span style={{ backgroundColor: pair.rejectedColor, width: '20px', height: '20px' }}></span>
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
                            <div className="input-group-text">Approved Color</div>
                        </div>
                        <input type='color' className='form-control' id='approved' name='approved' value={settings.settings.approvedColor} onChange={handleColorChange} />
                    </div>

                    <div className="input-group mb-2">
                        <div className="input-group-prepend">
                            <div className="input-group-text">Revision Color</div>
                        </div>
                        <input type='color' className='form-control' id='failed' name='revision' value={settings.settings.revisionColor} onChange={handleColorChange} />
                    </div>
                    <div className="input-group mb-2">
                        <div className="input-group-prepend">
                            <div className="input-group-text">Rejected Color</div>
                        </div>
                        <input type='color' className='form-control' id='passed' name='rejected' value={settings.settings.rejectedColor} onChange={handleColorChange} />
                    </div>
                </>
            )}
        </div>
    )
}

export default TableColorPicker