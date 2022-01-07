import React, { Dispatch, SetStateAction } from 'react'
import { useAppState } from '../../overmind'

/** ProfileInfo displays the user's profile information. */
const ProfileInfo = ({ setEditing }: { setEditing: Dispatch<SetStateAction<boolean>> }): JSX.Element => {
    const self = useAppState().self

    return (
        <>
            <div className='card-text text-center'>
                <h2 className='mb-4'>
                    {self.getName()}
                </h2>
            </div>
            <div className='card-text text-center'>
                <i className='fa fa-envelope text-muted'></i>
                <span className='ml-3'>{self.getEmail()}</span>
            </div>
            <div className='card-text text-center'>
                <i className='fa fa-graduation-cap text-muted'></i>
                <span className='ml-3'>{self.getStudentid()}</span>
            </div>
            <span className="badge float-right clickable" onClick={() => setEditing(true)}><i className='fa fa-edit'></i></span>
        </>
    )
}

export default ProfileInfo