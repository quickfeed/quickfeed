import * as React from "react";

import { Course, Enrollment, Group, Status, User } from "../../../proto/ag_pb";
import { Search } from "../../components";
import { CourseManager } from "../../managers/CourseManager";
import { NavigationManager } from "../../managers/NavigationManager";
import { UserManager } from "../../managers/UserManager";
import {
    IUserRelation,
} from "../../models";

interface IGroupProps {
    className: string;
    students: IUserRelation[];
    freeStudents: IUserRelation[];
    curUser: User;
    courseMan: CourseManager;
    userMan: UserManager;
    navMan: NavigationManager;
    pagePath: string;
    course: Course;
    groupData?: Group;
}
interface IGroupState {
    name: string;
    students: IUserRelation[];
    selectedStudents: IUserRelation[];
    curUser: IUserRelation | undefined;
    errorFlash: JSX.Element | null;
}

export class GroupForm extends React.Component<IGroupProps, IGroupState> {
    constructor(props: any) {
        super(props);
        const currentUser = this.props.students.find((v) => v.user.getId() === this.props.curUser.getId());
        const as: IUserRelation[] = this.getAvailableStudents(currentUser);
        const ss: IUserRelation[] = this.getSelectedStudents(currentUser);
        this.state = {
            name: this.props.groupData ? this.props.groupData.getName() : "",
            students: as,
            selectedStudents: ss,
            curUser: currentUser,
            errorFlash: null,
        };
    }

    public render() {
        const studentSearchBar: JSX.Element = <Search className="input-group"
            placeholder="Search for students"
            onChange={(query) => this.handleSearch(query)} />;

        const selectableStudents: JSX.Element[] = [];
        for (const student of this.state.students) {
            selectableStudents.push(
                <li key={student.user.getId()} className="list-group-item">
                    {student.user.getName()}
                    <button type="button"
                        className="btn btn-outline-success" onClick={() => this.handleAddToGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-plus-sign" />
                    </button>
                </li>);
        }

        const selectedStudents: JSX.Element[] = [];
        for (const student of this.state.selectedStudents) {
            selectedStudents.push(
                <li key={student.user.getId()} className="list-group-item">
                    {student.user.getName()}
                    <button className="btn btn-outline-primary"
                        onClick={() => this.handleRemoveFromGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-minus-sign" />
                    </button>
                </li>);
        }

        return (
            <div className="student-group-container">
                <h1>{this.props.groupData ? "Edit Group" : "Create Group"}</h1>
                {this.state.errorFlash}
                <form className={this.props.className}
                    onSubmit={(e) => this.handleFormSubmit(e)}>
                    <div className="form-group row">
                    <div className="col-sm-12 alert alert-warning">
                        Choose wisely! Name cannot be changed after the group has been created.</div>
                    </div>
                    <div className="form-group row">
                        <label className="col-sm-1 col-form-label" htmlFor="tag">Name:</label>
                        <div className="col-sm-11">
                            <input type="text"
                                className="form-control"
                                id="name"
                                placeholder="Enter group name"
                                name="name"
                                value={this.state.name}
                                onChange={(e) => this.handleInputChange(e)}
                            />
                        </div>
                    </div>
                    <div className="form-group row">
                        <div className="col-sm-6">
                            <fieldset>
                                <legend>Available Students
                                    {/* <small className="hint">
                                    select {this.props.capacity} students for your group</small> */}
                                </legend>
                                {studentSearchBar} <br />
                                <ul className="student-group list-group">
                                    {selectableStudents}
                                </ul>

                            </fieldset>
                        </div>
                        <div className="col-sm-6">
                            <fieldset>
                                <legend>Selected Students</legend>
                                <ul className="student-group list-group">
                                    {selectedStudents}
                                </ul>

                            </fieldset>
                        </div>
                    </div>
                    <div className="form-group row">
                        <div className="col-sm-offset-5 col-sm-2">
                            <button
                                className="btn btn-primary active"
                                type="submit">{this.props.groupData ? "Update" : "Create"}
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        );
    }

    private async handleFormSubmit(e: React.FormEvent<any>) {
        e.preventDefault();
        const errors: string[] = this.groupValidate();
        if (errors.length > 0) {
            const flashErrors = this.getFlashErrors(errors);
            this.setState({
                errorFlash: flashErrors,
            });
        } else {
            const userids = this.state.selectedStudents.map((u, i) => u.user.getId());
            const result = this.props.groupData ?
                await this.updateGroup(this.state.name, userids, this.props.groupData.getId())
                 : await this.createGroup(this.state.name, userids);
            if ((result instanceof Status) && (result.getCode() > 0)) {
                const errMsg = result.getError();
                const serverErrors: string[] = [];
                serverErrors.push(errMsg);
                const flashErrors = this.getFlashErrors(serverErrors);
                this.setState({
                    errorFlash: flashErrors,
                });
            } else {
                if (this.props.groupData) {
                    if (this.props.groupData.getUsersList().filter((x) =>
                         x.getId() === this.props.curUser.getId()).length > 0) {
                        const redirectTo: string = this.props.groupData ?
                            this.props.pagePath + "/courses/" + this.props.course.getId() + "/groups"
                            : this.props.pagePath + "/courses/" + this.props.course.getId() + "/members";

                        this.props.navMan.navigateTo(redirectTo);
                    }
                } else { // Teacher created group, so no group data.
                    this.props.navMan.refresh();
                }
            }
        }
    }
    private genCurUserMissingFromGroupWarn(): JSX.Element {
        const flash: JSX.Element =
            <div className="alert alert-warning">
                <h4> Group created without you as a member.</h4>
            </div>;

        return flash;
    }

    private async createGroup(name: string, users: number[]): Promise<Group | Status> {
        return this.props.courseMan.createGroup(name, users, this.props.course.getId());
    }

    private async updateGroup(name: string, users: number[], gid: number): Promise<Status> {
        const groupData = new Group();
        groupData.setId(gid);
        groupData.setName(name);
        groupData.setCourseid(this.props.course.getId());
        const groupUsers: User[] = [];
        users.forEach((ele) => {
            const usr = new User();
            usr.setId(ele);
            groupUsers.push(usr);
        });
        groupData.setUsersList(groupUsers);
        return this.props.courseMan.updateGroup(groupData);
    }

    private handleInputChange(e: React.FormEvent<any>) {
        const target: any = e.target;
        const value = target.type === "checkbox" ? target.checked : target.value;
        const name = target.name as "name";

        this.setState({
            [name]: value,
        });
    }

    private handleAddToGroupOnClick(student: IUserRelation) {
        const index = this.state.students.indexOf(student);
        if (index >= 0) {
            const newSelectedArr = this.state.selectedStudents.slice();
            newSelectedArr.push(student);
            const newStudentArr = this.state.students.slice();
            newStudentArr.splice(index, 1);
            this.setState({
                students: newStudentArr, // this.state.students.filter((_, i) => i !== index),
                selectedStudents: newSelectedArr,
            });
        }
    }

    private handleRemoveFromGroupOnClick(student: IUserRelation) {
        const index = this.state.selectedStudents.indexOf(student);
        if (index >= 0) {
            const newStudentsdArr = this.state.students.concat(student);
            this.setState({
                students: newStudentsdArr,
                selectedStudents: this.state.selectedStudents.filter((_, i) => i !== index),
            });
        }
    }

    private handleSearch(query: string): void {
        query = query.toLowerCase();
        const filteredData: IUserRelation[] = [];
        this.props.students.forEach((student) => {
            if ((student.user.getName().toLowerCase().indexOf(query) !== -1
                || student.user.getEmail().indexOf(query) !== -1
                || student.user.getLogin().indexOf(query)) !== -1
                && this.state.selectedStudents.indexOf(student) === -1
            ) {
                filteredData.push(student);
            }
        });

        this.setState({
            students: filteredData,
        });
    }

    private groupValidate(): string[] {
        const errors: string[] = [];
        if (this.state.name === "") {
            errors.push("Group name cannot be blank");
        }
        if (this.state.selectedStudents.length === 0) {
            errors.push("Group mush have members.");
        }
        if (this.state.curUser
            && (this.state.curUser.link.getStatus() === Enrollment.UserStatus.STUDENT
                || this.state.curUser.link.getStatus() === Enrollment.UserStatus.TEACHER)
            && !this.isCurrentStudentSelected(this.state.curUser)) {

            if (this.state.curUser.link.getStatus() !== Enrollment.UserStatus.TEACHER) {
                errors.push("You must be a member of the group");
            }
        }
        return errors;
    }

    private getFlashErrors(errors: string[]): JSX.Element {
        const errorArr: JSX.Element[] = [];
        for (let i: number = 0; i < errors.length; i++) {
            errorArr.push(<li key={i}>{errors[i]}</li>);
        }
        const flash: JSX.Element =
            <div className="alert alert-danger">
                <h4>{errorArr.length} errors prohibited Group from being saved: </h4>
                <ul>
                    {errorArr}
                </ul>
            </div>;
        return flash;
    }

    private isCurrentStudentSelected(student: IUserRelation): boolean {
        const index = this.state.selectedStudents.indexOf(student);
        return index >= 0;
    }

    private getSelectedStudents(curUser: IUserRelation | undefined): IUserRelation[] {
        const ss: IUserRelation[] = [];
        if (this.props.groupData) {
            // add group members to the list of selected students
            for (const user of this.props.groupData.getUsersList()) {
                const guser = this.props.students.find((v) => v.user.getId() === user.getId());
                if (guser) {
                    ss.push(guser);
                }
            }
        }
        return ss;
    }

    private getAvailableStudents(curUser: IUserRelation | undefined): IUserRelation[] {
        const as: IUserRelation[] = this.props.freeStudents.slice();
        if (this.props.groupData) {
            // remove group members from the list of available students
            for (const user of this.props.groupData.getUsersList()) {
                const guser = as.find((v) => v.user.getId() === user.getId());
                if (guser) {
                    const index = as.indexOf(guser);
                    if (index >= 0) {
                        as.splice(index, 1);
                    }
                }
            }
        }
        return as;
    }
}
