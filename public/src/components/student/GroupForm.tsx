import * as React from "react";

import { IUserRelation } from "../../models";

import { Search } from "../../components";

interface IGroupProp {
    className: string;
    students: IUserRelation[];
    capacity: number;
}
interface IGroupState {
    name: string;
    students: IUserRelation[];
    selectedStudents: IUserRelation[];
}
class GroupForm extends React.Component<IGroupProp, IGroupState> {
    constructor(props: any) {
        super(props);
        this.state = {
            name: "",
            students: this.props.students,
            selectedStudents: [],
        };
    }

    public render() {
        const searchIcon: JSX.Element = <span className="input-group-addon">
            <i className="glyphicon glyphicon-search"></i></span>;

        const studentSearchBar: JSX.Element = <Search className="input-group"
            addonBefore={searchIcon}
            placeholder="Search for students"
            onChange={(query) => this.handleSearch(query)} />;

        const selectableStudents: JSX.Element[] = [];
        for (const student of this.state.students) {
            selectableStudents.push(
                <li key={student.user.id} className="list-group-item">
                    {student.user.firstname + " " + student.user.lastname}
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
                    {student.user.firstname + " " + student.user.lastname}
                    <button className="btn btn-outline-primary"
                        onClick={() => this.handleRemoveFromGroupOnClick(student)}>
                        <i className="glyphicon glyphicon-minus-sign" />
                    </button>
                </li>);
        }

        return (
            <div className="student-group-container">
                <h1>Create a Group</h1>
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
                                <legend>Available Students <small className="hint">
                                    select {this.props.capacity} students for your group</small>
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
                                className={this.state.selectedStudents.length
                                    === this.props.capacity ? "btn btn-primary active" : "btn btn-primary disabled"}
                                type="submit">Create
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        );
    }

    private handleFormSubmit(e: React.FormEvent<any>) {
        e.preventDefault();
        if (this.state.selectedStudents.length === this.props.capacity) {
            console.log("state", this.state);
            console.log("group ", this.state.name, this.state.selectedStudents);
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
            if ((student.user.firstname.toLowerCase().indexOf(query) !== -1
                || student.user.lastname.toLowerCase().indexOf(query) !== -1
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
}
export { GroupForm };
