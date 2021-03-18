import { Action } from "overmind";
import React from "react";
import { useOvermind } from "../overmind";
import { State } from "../overmind/state";

export const ToggleSwitch: React.FC = () =>{ 
    const {state,actions} = useOvermind()

    const changeTheme = () => {
        actions.changeTheme()
        window.localStorage.setItem("theme", state.theme)
    }

    return(
        <div></div>
    )
}