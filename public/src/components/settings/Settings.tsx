import React from 'react'
import { useActions, useAppState } from '../../overmind'
import ResultsColorPicker from './ResultsColorPicker'
import TableColorPicker from './TableColorPicker'


const Settings = () => {
    const { settings } = useAppState()
    const actions = useActions()

    const handleRangeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        actions.settings.updateSettings({ barWidth: parseInt(event.target.value) })
    }

    return (
        <div className='container mt-3'>
            <h1>Settings</h1>
            <p>Change the colors and width of the bar</p>

            <form className='mb-3'>
                <ResultsColorPicker />
                <div className="form-group mb-3">
                    <label htmlFor='barWidth' className='form-label'>Bar Width</label>
                    <input type='range' className='custom-range' id='barWidth' onChange={handleRangeChange} value={settings.settings.barWidth} min='0' max='20' step={1} />
                </div>
                <button type='button' className='btn btn-primary m-2' onClick={() => actions.settings.resetSettings()}>Reset to Default</button>
            </form>

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

            <TableColorPicker />
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
        </div>

    )
}

export default Settings
