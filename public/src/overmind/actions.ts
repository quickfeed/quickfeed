import { Context } from "."
import { Color } from "../Helpers"

export const onInitializeOvermind = async ({ actions, effects }: Context) => {
    // Initialize the API client. *Must* be done before accessing the client.
    effects.global.api.init(actions.global.errorHandler)
    await actions.global.fetchUserData()
    // Currently this only alerts the user if they are not logged in after a page refresh
    const alert = localStorage.getItem("alert")
    if (alert) {
        actions.global.alert({ text: alert, color: Color.RED })
        localStorage.removeItem("alert")
    }
}
