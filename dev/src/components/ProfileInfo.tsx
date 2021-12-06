import React, { Dispatch, SetStateAction } from 'react'
import { useAppState } from '../overmind'

const ProfileInfo = ({setEditing}: {setEditing: Dispatch<SetStateAction<boolean>>}): JSX.Element => {
    const self = useAppState().self

    return (
        <div className="box">
                <div className="card well">
                <div className="card-header">Your Information</div>
                    <ul className="list-group list-group-flush">
                        <li className="list-group-item">
                            <div>Name:</div> 
                            <div>{self.getName()}</div>
                        </li>
                        <li className="list-group-item">Email: {self.getEmail()}</li>
                        <li className="list-group-item">Student ID: {self.getStudentid()}</li>
                    </ul>
                </div>
            <button className="btn btn-primary" onClick={() => setEditing(true)}>Edit Profile</button>
        </div>
        )
}

export default ProfileInfo