import React from "react"
import { Redirect, Route, Switch, useHistory } from "react-router"
import { useOvermind } from "../overmind"
import EditCourse from "./admin/EditCourse"
import Users from "./admin/Users"
import Card from "./Card"
import CourseCreationForm from "./forms/CourseCreationForm"



export const AdminPage = () => {
    
    const {state} = useOvermind()

    // Ideas: Statistics, Create Course, Promote Users
    
    const manageUsers = {title: "Manage Users", text: "View and manage all users.", buttonText: "Manage Users", to: "/admin/manage"}
    const createCourse = {title: "Create Course", text: "Create a new course.", buttonText: "Create Course", to: "/admin/create"}
    const editCourse = {title: "Edit Course", text: "Edit an existing course.", buttonText: "Edit Course", to: "/admin/edit"}
    
    /** Button used to redirect a user, ex. return to top layer of course page */
    const RedirectButton = ({to}: {to: string}) => {
        const history  = useHistory()
        const hide = history.location.pathname == "/admin" ? true : false
        return (
            <div className={"btn btn-dark redirectButton"} onClick={() => history.push(to)} hidden={hide}>
                <i className="fa fa-arrow-left"></i>
            </div>
        )
    }
    
    if (state.self.getIsadmin()) {
        return (
            <div className="box">
                <RedirectButton to={"/admin"}></RedirectButton>
                <div className="row">
                    <Card title={createCourse.title} text={createCourse.text} buttonText={createCourse.buttonText} to={createCourse.to}></Card>
                    <Card title={editCourse.title} text={editCourse.text} buttonText={editCourse.buttonText} to={editCourse.to}></Card>
                    <Card title={manageUsers.title} text={manageUsers.text} buttonText={manageUsers.buttonText} to={manageUsers.to}></Card>
                </div>
                <Switch>
                    <Route path={"/admin/manage"} component={Users}></Route>
                    <Route path={"/admin/create"} component={CourseCreationForm}></Route>
                    <Route path={"/admin/edit"} component={EditCourse}></Route>
                </Switch>
            </div>
        )
    }
    return (
        <Redirect to="/" />
    )
}

export default AdminPage