import React from "react"
import { Redirect, Route, Switch, useHistory } from "react-router"
import { useAppState } from "../overmind"
import EditCourse from "../components/admin/EditCourse"
import Users from "../components/admin/Users"
import Card from "../components/Card"
import RedirectButton from "../components/RedirectButton"
import Alerts from "../components/alerts/Alerts"
import CreateCourse from "../components/admin/CreateCourse"


// AdminPage is the page containing the admin-only components.
const AdminPage = (): JSX.Element => {
    const state = useAppState()
    const history = useHistory()

    // Objects containing props for the cards in the admin page.
    // TODO: Perhaps make a Card prop type.
    const manageUsers = { title: "Manage Users", text: "View and manage all users.", buttonText: "Manage Users", to: "/admin/manage" }
    const createCourse = { title: "Create Course", text: "Create a new course.", buttonText: "Create Course", to: "/admin/create" }
    const editCourse = { title: "Edit Course", text: "Edit an existing course.", buttonText: "Edit Course", to: "/admin/edit" }

    // If the user is not an admin, redirect to the home page.
    if (!state.self.IsAdmin) {
        return <Redirect to="/" />
    }

    const root = "/admin"
    return (
        <div className="box">
            <RedirectButton to={root} />
            <Alerts />
            <div className="row" hidden={history.location.pathname !== root}>
                <Card {...manageUsers} />
                <Card {...createCourse} />
                <Card {...editCourse} />
            </div>

            <Switch>
                <Route path="/admin/manage">
                    <Users />
                </Route>
                <Route path="/admin/create">
                    <CreateCourse />
                </Route>
                <Route path="/admin/edit">
                    <EditCourse />
                </Route>
            </Switch>
        </div>
    )
}

export default AdminPage
