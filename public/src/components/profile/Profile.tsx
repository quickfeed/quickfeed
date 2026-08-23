import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { useAppState } from '../../overmind'
import ProfileCard from './ProfileCard'
import ProfileForm from './ProfileForm'
import ProfileHero from './ProfileHero'
import ProfileInfo from './ProfileInfo'
import SignupText from './SignupText'


const Profile = () => {
    const state = useAppState()
    const navigate = useNavigate()
    const location = useLocation()
    // Holds a local state to check whether the user asked to edit their user
    // information. Editing is forced while that information is incomplete;
    // there is nothing useful to show instead, and the signup text belongs
    // with the form.
    const [editRequested, setEditing] = useState(false)
    const editing = editRequested || !state.isValid

    // Redirect from "/" to "/profile" when user object is invalid
    useEffect(() => {
        if (!state.isLoggedIn) {
            navigate("/")
        } else if (!state.isValid && location.pathname === "/") {
            navigate("/profile")
        }
    }, [state.isLoggedIn, state.isValid, location.pathname, navigate])

    return (
        <div className="min-h-screen">
            <ProfileHero name={state.self.Name} />
            <div className="container mx-auto px-4 py-12 flex justify-center">
                <ProfileCard>
                    {editing ?
                        <ProfileForm setEditing={setEditing} >
                            {state.isValid ? null : <SignupText />}
                        </ProfileForm>
                        : <ProfileInfo setEditing={setEditing} />
                    }
                </ProfileCard>
            </div>
        </div>
    )
}

export default Profile
