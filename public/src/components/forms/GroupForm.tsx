import * as React from "react";

import { Course, Enrollment, Group, Status, User } from "../../../proto/ag/ag_pb";
import { Search } from "../../components";
import { CourseManager } from "../../managers/CourseManager";
import { NavigationManager } from "../../managers/NavigationManager";
import { UserManager } from "../../managers/UserManager";
import { searchForStudents } from "../../componentHelper";

interface GroupFormProps {
    className: string;
    students: Enrollment[];
    freeStudents: Enrollment[];
    curUser: User;
    courseMan: CourseManager;
    userMan: UserManager;
    navMan: NavigationManager;
    pagePath: string;
    course: Course;
    groupData?: Group;
}
interface GroupFormState {
    name: string;
    students: Enrollment[];
    selectedStudents: Enrollment[];
    curUser: Enrollment | undefined;
    errorFlash: JSX.Element | null;
    actionReady: boolean;
}

export class GroupForm extends React.Component<GroupFormProps, GroupFormState> {
    constructor(props: any) {
        super(props);
        const currentUser = this.props.students.find((v) => v.getUserid() === this.props.curUser.getId());
        const as: Enrollment[] = this.getAvailableStudents();
        const ss: Enrollment[] = this.getSelectedStudents();
        this.state = {
            name: this.props.groupData ? this.props.groupData.getName() : "",
            students: as,
            selectedStudents: ss,
            curUser: currentUser,
            errorFlash: null,
            actionReady: true,
        };
    }

    public render() {
        const studentSearchBar: JSX.Element = <Search className="input-group"
            placeholder="Search for students"
            onChange={(query) => this.handleSearch(query)} />;

        const selectableStudents: JSX.Element[] = [];
        for (const student of this.state.students) {
            selectableStudents.push(
                <li key={student.getUserid()} className="box-item">
                    <label>{student.getUser()?.getName()}</label>
                    <button type="button"
                        className="btn btn-outline-success add-btn"
                        onClick={() => this.handleAddToGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-plus-sign" />
                    </button>
                </li>);
        }

        const selectedStudents: JSX.Element[] = [];
        for (const student of this.state.selectedStudents) {
            selectedStudents.push(
                <li key={student.getUserid()} className="box-item">
                    <button className="btn btn-outline-primary rm-btn"
                        onClick={() => this.handleRemoveFromGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-minus-sign" />
                    </button>
                    <label>{student.getUser()?.getName()}</label>
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
                            Warning: Group names cannot be changed once created.
                        </div>
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
                                </legend>
                                <ul className="student-group list-group">
                                    {selectableStudents}
                                </ul>
                                {studentSearchBar}

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
                                type="submit">{this.generateButtonString()}
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
            this.setState({
                actionReady: false,
            });
            const userids = this.state.selectedStudents.map((u, i) => u.getUserid());
            const result = this.props.groupData
                ? await this.updateGroup(this.state.name, userids, this.props.groupData.getId())
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
                const isTeacher = await this.props.userMan.isTeacher(this.props.course.getId());
                // if current user is a course teacher, redirect to the groups list
                const redirectTo: string = isTeacher
                    ? "/app/teacher/courses/" + this.props.course.getId() + "/groups"
                    : this.props.pagePath + "/courses/" + this.props.course.getId() + "/members";

                this.props.navMan.navigateTo(redirectTo);
            }
        }
    }

    private async createGroup(name: string, users: number[]): Promise<Group | Status> {
        return this.props.courseMan.createGroup(this.props.course.getId(), name, users);
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

    private handleAddToGroupOnClick(student: Enrollment) {
        const index = this.state.students.indexOf(student);
        if (index >= 0) {
            const newSelectedArr = this.state.selectedStudents.slice();
            newSelectedArr.push(student);
            const newStudentArr = this.state.students.slice();
            newStudentArr.splice(index, 1);
            this.setState({
                students: newStudentArr,
                selectedStudents: newSelectedArr,
            });
        }
    }

    private handleRemoveFromGroupOnClick(student: Enrollment) {
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
        this.setState({
            students: searchForStudents(this.props.freeStudents, query),
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

        if (!this.userValidate(this.state.curUser)) {
            errors.push("You must be a member of the group");
        }
        return errors;
    }

    private userValidate(curUser: Enrollment | undefined): boolean {
        if (!curUser) {
            return false;
        }
        const status = curUser.getStatus();
        return status === Enrollment.UserStatus.TEACHER || (status === Enrollment.UserStatus.STUDENT && this.isCurrentStudentSelected(curUser));
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

    private isCurrentStudentSelected(student: Enrollment): boolean {
        let foundSelected = false;
        this.state.selectedStudents.forEach((user) => {
            if (user.getUserid() === student.getUserid()) {
                foundSelected = true;
            }
        })
        return foundSelected;
    }

    private getSelectedStudents(): Enrollment[] {
        const ss: Enrollment[] = [];
        if (this.props.groupData) {
            // add group members to the list of selected students
            for (const user of this.props.groupData.getUsersList()) {
                const guser = this.props.students.find((v) => v.getUserid() === user.getId());
                if (guser) {
                    ss.push(guser);
                }
            }
        }
        return ss;
    }

    private getAvailableStudents(): Enrollment[] {
        const as: Enrollment[] = this.props.freeStudents.slice();
        if (this.props.groupData) {
            // remove group members from the list of available students
            for (const user of this.props.groupData.getUsersList()) {
                const guser = as.find((v) => v.getUserid() === user.getId());
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

    private generateButtonString(): string {
        if (this.props.groupData) {
            return this.state.actionReady ? "Update" : "Updating";
        }
        return this.state.actionReady ? "Create" : "Creating";
    }
}
