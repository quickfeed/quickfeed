import React, { Dispatch, SetStateAction } from 'react'
import { useAppState } from '../../overmind'

/** ProfileInfo displays the user's profile information. */
const ProfileInfo = ({ setEditing }: { setEditing: Dispatch<SetStateAction<boolean>> }) => {
    const self = useAppState().self

    return (
        <>
            <h2 className='text-2xl font-bold mb-6'>
                {self.Name}
            </h2>

            <div className="space-y-4 w-full">
                <div className="flex items-center gap-4 p-3 bg-base-200 rounded-lg hover:bg-base-300 transition-colors">
                    <div className="flex-shrink-0 w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
                        <i className='fa fa-envelope text-primary' />
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="text-xs text-base-content/60 uppercase font-semibold">Email</p>
                        <p className="text-sm font-medium truncate">{self.Email}</p>
                    </div>
                </div>

                <div className="flex items-center gap-4 p-3 bg-base-200 rounded-lg hover:bg-base-300 transition-colors">
                    <div className="flex-shrink-0 w-10 h-10 rounded-full bg-secondary/20 flex items-center justify-center">
                        <i className='fa fa-graduation-cap text-secondary' />
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="text-xs text-base-content/60 uppercase font-semibold">Student ID</p>
                        <p className="text-sm font-medium">{self.StudentID}</p>
                    </div>
                </div>
            </div>

            <button
                className="btn btn-outline btn-primary mt-6 gap-2"
                onClick={() => setEditing(true)}
            >
                <i className='fa fa-edit' />
                Edit Profile
            </button>
        </>
    )
}

export default ProfileInfo
