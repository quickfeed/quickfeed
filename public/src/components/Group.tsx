import React, { useEffect, useState } from "react"
import { useOvermind } from "../overmind"


export const Group = () => {
    const {state, actions} = useOvermind()

    const [search, setSearch] = useState("")

    useEffect(() => {
        actions.getEnrollmentsByCourse(1)
        console.log(state.users)
    }, [])

    const handleKeyPress = (e: React.FormEvent<HTMLInputElement>) => {
        actions.updateSearch(e.currentTarget.value)
    }

    return(
        <div className="box">
            <input onKeyUp={handleKeyPress}></input>
            <div className='box'>
                {state.userSearch.map(user => {
                    return (
                        <div>{user.getUser()?.getName()}</div>
                    )
                })} 
            </div>
        </div>
        )
}

export default Group