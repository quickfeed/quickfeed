import React from 'react'
import { useActions, useAppState } from '../../overmind'
import { hasTeacher } from '../../Helpers'
import ColorPicker, { ColorOption } from './ColorPicker'


const tableColors: ColorOption[] = [
    {
        name: 'Default',
        colors: {
            "approved-color": '#CCFFCC', // Light Green
            "revision-color": '#FFFFCC', // Light Yellow
            "rejected-color": '#FFCCCC', // Light Red
        }
    },
    {
        name: 'Color Blind Friendly 1',
        colors: {
            "approved-color": '#56B4E9', // Sky Blue
            "revision-color": '#E69F00', // Orange
            "rejected-color": '#D41159', // Pink
        }
    },
    {
        name: 'Color Blind Friendly 2',
        colors: {
            "approved-color": '#009E73',
            "revision-color": '#F0E442',
            "rejected-color": '#CC79A7',
        }
    }
]

const resultsColors: ColorOption[] = [
    {
        name: 'Default',
        colors: {
            "passed-color": '#006F00', // Green
            "failed-color": '#FF0000', // Red
        }
    },
    {
        name: 'Color Blind Friendly 1',
        colors: {
            "passed-color": '#56B4E9', // Sky Blue
            "failed-color": '#E69F00', // Orange
        }
    },
    {
        name: 'Color Blind Friendly 2',
        colors: {
            "passed-color": '#1A85FF', // Blue
            "failed-color": '#D41159', // Pink
        }
    }
    // Add more pairs as needed
]

const Settings = () => {
    const { settings, enrollments } = useAppState()
    const actions = useActions()

    // Some settings are only relevant for teachers
    const isTeacher = enrollments.some(enrollment => hasTeacher(enrollment.status))

    const handleRangeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        actions.settings.updateSettings({ 'bar-width': event.target.value + 'px' })
    }

    return (
        <div className='container mt-3'>
            <h1>Settings</h1>
            <p>Change the colors and width of the bar</p>

            <div className='mb-3'>
                <ColorPicker colorOptions={resultsColors} />
                <div className="form-group mb-3">
                    <label htmlFor='barWidth' className='form-label'>Bar Width</label>
                    <input type='range' className='custom-range' id='barWidth' onChange={handleRangeChange} defaultValue={settings.settings['bar-width']} min='0' max='20' step='1' />
                </div>
                <button type='button' className='btn btn-primary m-2' onClick={() => actions.settings.resetSettings()}>Reset to Default</button>
            </div>

            <table className="table">
                <thead className="thead-dark">
                    <tr>
                        <th scope="col">Criterion Preview</th>
                    </tr>
                </thead>
                <tbody>
                    <tr className="align-items-center">
                        <td className='passed'>Passed criterion</td>
                    </tr>
                    <tr className="align-items-center">
                        <td className='failed'>Failed criterion</td>
                    </tr>
                </tbody>
            </table>

            <div className="card bg-light">
                <code className="card-body" style={{ color: "#c7254e", wordBreak: "break-word" }}></code>
            </div>

            {isTeacher ? (
                <>
                    <ColorPicker colorOptions={tableColors} />
                    <table className="table">
                        <tbody>
                            <tr>
                                <td className="result-approved ">100 %</td>
                                <td className="result-approved ">100 %</td>
                                <td className="result-approved ">95 %</td>
                                <td className="result-approved ">99 %</td>
                                <td className="result-approved ">100 %</td>
                                <td className="clickable ">38 %</td>
                                <td className="clickable ">0 %</td>
                            </tr>
                            <tr>
                                <td className="result-approved ">100 %</td>
                                <td className="result-approved ">100 %</td>
                                <td className="result-approved ">97 %</td>
                                <td className="result-approved ">99 %</td>
                                <td className="result-rejected ">98 %</td>
                                <td className="result-rejected ">100 %</td>
                                <td className="clickable ">0 %</td>
                            </tr>
                            <tr>
                                <td className="result-revision ">100 %</td>
                                <td className="result-rejected ">100 %</td>
                                <td className="result-approved ">97 %</td>
                                <td className="result-revision ">99 %</td>
                                <td className="result-approved ">98 %</td>
                                <td className="result-rejected ">100 %</td>
                                <td className="clickable ">0 %</td>
                            </tr>
                        </tbody>
                    </table>
                </>
            ) : null}
        </div>

    )
}

export default Settings
