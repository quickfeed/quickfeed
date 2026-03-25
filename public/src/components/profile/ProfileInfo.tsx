import React, { Dispatch, SetStateAction } from 'react'
import { useAppState } from '../../overmind'

const Field = ({ icon, label, value, color }: { icon: string; label: string; value: string; color: string }) => (
    <div className="flex items-center gap-3 p-3 rounded-xl bg-base-100 border border-base-300">
        <div className={`w-9 h-9 rounded-lg ${color} flex items-center justify-center flex-shrink-0`}>
            <i className={`fa ${icon} text-sm`} />
        </div>
        <div className="min-w-0 text-left">
            <p className="text-[10px] uppercase font-bold tracking-widest text-base-content/40">{label}</p>
            <p className="text-sm font-medium truncate">{value}</p>
        </div>
    </div>
)

/** ProfileInfo displays the user's profile information. */
const ProfileInfo = ({ setEditing }: { setEditing: Dispatch<SetStateAction<boolean>> }) => {
    const self = useAppState().self

    return (
        <>
            <div className="mb-1">
                <h2 className="text-2xl font-bold">
                    {self.Name}
                </h2>
            </div>

            <div className="divider my-3" />

            <div className="space-y-2 w-full">
                <Field icon="fa-envelope" label="Email" value={self.Email} color="bg-primary/10 text-primary" />
                <Field icon="fa-id-card" label="Student ID" value={self.StudentID} color="bg-secondary/10 text-secondary" />
            </div>

            <button
                className="btn btn-primary btn-sm mt-6 gap-2 w-full"
                onClick={() => setEditing(true)}
            >
                <i className="fa fa-edit" />
                Edit Profile
            </button>
        </>
    )
}

export default ProfileInfo
