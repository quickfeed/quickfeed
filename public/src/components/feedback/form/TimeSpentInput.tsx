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
        <div className="w-full">
            <div className="mb-2">
                <label htmlFor="hours" className="block text-base font-semibold text-base-content mb-1">
                    How much time did you spend on this assignment? <span className="text-error">*</span>
                </label>
                <div className="text-xs text-base-content/60">
                    Enter hours and minutes (max 100 hours)
                </div>
            </div>
            <div className="flex gap-3">
                <div className="flex-1">
                    <div className="join w-full">
                        <input
                            id="hours"
                            type="number"
                            className="input input-bordered join-item w-full focus:input-primary"
                            value={hours}
                            onChange={onHoursChange}
                            placeholder="0"
                            min="0"
                            max="100"
                        />
                        <span className="bg-base-200 join-item px-4 flex items-center text-sm text-base-content/70">hours</span>
                    </div>
                </div>
                <div className="flex-1">
                    <div className="join w-full">
                        <input
                            id="minutes"
                            type="number"
                            className="input input-bordered join-item w-full focus:input-primary"
                            value={minutes}
                            onChange={onMinutesChange}
                            placeholder="0"
                            min="0"
                            max="59"
                        />
                        <span className="bg-base-200 join-item px-4 flex items-center text-sm text-base-content/70">minutes</span>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default TimeSpentInput
