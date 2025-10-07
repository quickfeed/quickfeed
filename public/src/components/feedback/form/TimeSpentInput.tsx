import React from 'react'

interface TimeSpentInputProps {
    hours: string
    minutes: string
    onHoursChange: (e: React.ChangeEvent<HTMLInputElement>) => void
    onMinutesChange: (e: React.ChangeEvent<HTMLInputElement>) => void
}

export const TimeSpentInput: React.FC<TimeSpentInputProps> = ({
    hours,
    minutes,
    onHoursChange,
    onMinutesChange
}) => {
    return (
        <div className="mb-3">
            <label htmlFor="hours" className="form-label">
                How much time did you spend on this assignment? <span className="text-danger">*</span>
            </label>
            <p className="text-muted small mb-2">Enter hours and minutes (max 100 hours)</p>
            <div className="row">
                <div className="col-6">
                    <div className="input-group">
                        <input
                            id="hours"
                            type="number"
                            className="form-control"
                            value={hours}
                            onChange={onHoursChange}
                            placeholder="0"
                            min="0"
                            max="100"
                        />
                        <span className="input-group-text">hours</span>
                    </div>
                </div>
                <div className="col-6">
                    <div className="input-group">
                        <input
                            id="minutes"
                            type="number"
                            className="form-control"
                            value={minutes}
                            onChange={onMinutesChange}
                            placeholder="0"
                            min="0"
                            max="59"
                        />
                        <span className="input-group-text">minutes</span>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default TimeSpentInput
