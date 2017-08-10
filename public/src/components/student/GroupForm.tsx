import * as React from "react";

import { CourseManager } from "../../managers/CourseManager";
import { NavigationManager } from "../../managers/NavigationManager";
import { CourseUserState, ICourse, INewGroup, isError, IUser, IUserRelation } from "../../models";

import { Search } from "../../components";

interface IGroupProp {
    className: string;
    students: IUserRelation[];
    curUser: IUser;
    courseMan: CourseManager;
    navMan: NavigationManager;
    pagePath: string;
    course: ICourse;
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
        this.state = {
            name: "",
            students: this.props.students,
            selectedStudents: [],
            curUser: this.props.students.find((v) => v.user.id === this.props.curUser.id),
            errorFlash: null,
        };
    }

    public componentDidMount() {
        if (this.state.curUser) {
            this.handleAddToGroupOnClick(this.state.curUser);
        }
    }

    public render() {
        const studentSearchBar: JSX.Element = <Search className="input-group"
            placeholder="Search for students"
            onChange={(query) => this.handleSearch(query)} />;

        const selectableStudents: JSX.Element[] = [];
        for (const student of this.state.students) {
            selectableStudents.push(
                <li key={student.user.id} className="list-group-item">
                    {student.user.name}
                    <button type="button"
                        className="btn btn-outline-success" onClick={() => this.handleAddToGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-plus-sign" />
                    </button>
                </li>);
        }

        const selectedStudents: JSX.Element[] = [];
        for (const student of this.state.selectedStudents) {
            selectedStudents.push(
                <li key={student.user.id} className="list-group-item">
                    {student.user.name}
                    <button className="btn btn-outline-primary"
                        onClick={() => this.handleRemoveFromGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-minus-sign" />
                    </button>
                </li>);
        }

        return (
            <div className="student-group-container">
                <h1>Create a Group</h1>
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
                                type="submit">Create
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
                userids: this.state.selectedStudents.map((u, i) => u.user.id),
            };
            const result = await this.props.courseMan.createGroup(formData, this.props.course.id);
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
                const redirectTo: string = this.props.pagePath + "/course/" + this.props.course.id + "/members";
                this.props.navMan.navigateTo(redirectTo);
            }
        }
    }

    private handleInputChange(e: React.FormEvent<any>) {
        const target: any = e.target;
        const value = target.type === "checkbox" ? target.checked : target.value;
        const name = target.name;

        this.setState({
            [name]: value,
        });
    }

    private handleAddToGroupOnClick(student: IUserRelation) {
        const index = this.state.students.indexOf(student);
        if (index >= 0) {
            const newSelectedArr = this.state.selectedStudents.concat(student);
            this.setState({
                students: this.state.students.filter((_, i) => i !== index),
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
            if ((student.user.name.toLowerCase().indexOf(query) !== -1
                || student.user.email.toString().indexOf(query) !== -1)
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
        if (this.state.curUser && this.state.curUser.link.state === CourseUserState.student &&
            !this.isCurrentStudentSelected(this.state.curUser)) {
            errors.push("You must be a member of the group");
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
}
export { GroupForm };
