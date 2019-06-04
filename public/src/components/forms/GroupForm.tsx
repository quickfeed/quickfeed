import * as React from "react";

import { CourseManager } from "../../managers/CourseManager";
import { NavigationManager } from "../../managers/NavigationManager";
import {
    ICourse, IError,
    INewGroup, isError, IStatusCode, IUserRelation,
} from "../../models";
import { User, Enrollment, Group } from "../../../proto/ag_pb";
import { Search } from "../../components";
import { UserManager } from "../../managers/UserManager";

interface IGroupProp {
    className: string;
    students: IUserRelation[];
    curUser: User;
    courseMan: CourseManager;
    userMan: UserManager;
    navMan: NavigationManager;
    pagePath: string;
    course: ICourse;
    groupData?: Group;
}
interface IGroupState {
    name: string;
    students: IUserRelation[];
    selectedStudents: IUserRelation[];
    curUser: IUserRelation | undefined;
    errorFlash: JSX.Element | null;
}

class GroupForm extends React.Component<IGroupProp, IGroupState> {
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
            const formData: INewGroup = {
                name: this.state.name,
                userids: this.state.selectedStudents.map((u, i) => u.user.getId()),
            };

            const result = this.props.groupData ?
                await this.updateGroup(formData, this.props.groupData.getId()) : await this.createGroup(formData);
            if (isError(result) && result.data) {
                const errMsg = result.data.message;
                let serverErrors: string[] = [];
                if (errMsg instanceof Array) {
                    serverErrors = errMsg;
                } else {
                    serverErrors.push(errMsg);
                }
                const flashErrors = this.getFlashErrors(serverErrors);
                this.setState({
                    errorFlash: flashErrors,
                });
            } else {
                if (this.props.groupData) {
                    if (this.props.groupData.getUsersList().filter((x) => x.getId() === this.props.curUser.getId()).length > 0) {
                        const redirectTo: string = this.props.groupData ?
                            this.props.pagePath + "/courses/" + this.props.course.id + "/groups"
                            : this.props.pagePath + "/courses/" + this.props.course.id + "/members";

                        this.props.navMan.navigateTo(redirectTo);
                    } else {
                        // There is group data, but cur user is not part of that group
                        const flash = this.genCurUserMissingFromGroupWarn();
                        this.setState({
                            errorFlash: flash,
                        });
                    }
                } else { // Teacher created group, so no group data.
                    const flash = this.genCurUserMissingFromGroupWarn();

                    this.setState({
                        errorFlash: flash,
                    });
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

    private async createGroup(formData: INewGroup): Promise<Group | IError> {
        return await this.props.courseMan.createGroup(formData, this.props.course.id);
    }

    private async updateGroup(formData: INewGroup, gid: number): Promise<IStatusCode | IError> {
        const groupData = new Group();
        groupData.setId(gid);
        groupData.setName(formData.name);
        groupData.setCourseid(this.props.course.id);
        const groupUsers: User[] = [];  
        formData.userids.forEach((ele) => {
            const user = this.props.userMan.getUser(ele);
            console.log("GroupForm: updateGroup pushing users from promise: " + user);
            user.then((user) => groupUsers.push(user)); 
            
        });
        groupData.setUsersList(groupUsers);
        return await this.props.courseMan.updateGroup(groupData);
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
                || student.user.getEmail().toString().indexOf(query) !== -1)
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
            errors.push("Group Name cannot be blank");
        }
        if (this.state.selectedStudents.length === 0) {
            errors.push("Group mush have members.");
        }
        if (this.state.curUser
            && (this.state.curUser.link.state === Enrollment.UserStatus.STUDENT
                || this.state.curUser.link.state === Enrollment.UserStatus.TEACHER)
            && !this.isCurrentStudentSelected(this.state.curUser)) {

            if (this.state.curUser.link.state !== Enrollment.UserStatus.TEACHER) {
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
            for (const user of this.props.groupData.getUsersList()) {
                const guser = this.props.students.find((v) => v.user.getId() === user.getId());
                if (guser) {
                    ss.push(guser);
                }
            }

        } else if (curUser) {
            ss.push(curUser);
        }
        return ss;
    }

    private getAvailableStudents(curUser: IUserRelation | undefined): IUserRelation[] {
        const as: IUserRelation[] = this.props.students.slice();
        if (this.props.groupData) {
            for (const user of this.props.groupData.getUsersList()) {
                const guser = as.find((v) => v.user.getId() === user.getId());
                if (guser) {
                    const index = as.indexOf(guser);
                    if (index >= 0) {
                        as.splice(index, 1);
                    }
                }
            }

        } else if (curUser) {
            const index = as.indexOf(curUser);
            if (index >= 0) {
                as.splice(index, 1);
            }
        }
        return as;
    }
}
export { GroupForm };
