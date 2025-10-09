import React, { Dispatch, SetStateAction } from 'react'
import { useAppState } from '../../overmind'

/** ProfileInfo displays the user's profile information. */
const ProfileInfo = ({ setEditing }: { setEditing: Dispatch<SetStateAction<boolean>> }) => {
    const self = useAppState().self

    return (
        <>
            <div className='card-text text-center'>
                <h2 className='mb-4'>
                    {self.Name}
                </h2>
            </div>
            <div className='card-text text-center'>
                <i className='fa fa-envelope text-muted' />
                <span className='ml-3'>{self.Email}</span>
            </div>
            <div className='card-text text-center'>
                <i className='fa fa-graduation-cap text-muted' />
                <span className='ml-3'>{self.StudentID}</span>
            </div>
            <span aria-hidden role="button" className="ml-auto clickable" onClick={() => setEditing(true)}><i className='fa fa-edit' /></span> {/* skipcq: JS-0746 */}
        </>
    )
}

export default ProfileInfo
