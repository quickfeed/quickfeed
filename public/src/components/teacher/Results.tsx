import * as React from "react";
import {IAssignment, ICourse, IUser} from "../../models";

import {DynamicTable, Row, Search, StudentLab} from "../../components";

interface IResultsProp {
    course: ICourse;
    students: IUser[];
    labs: IAssignment[];
}
interface IResultsState {
    assignment: IAssignment;
    selectedStudent: IUser;
    students: IUser[];
}
class Results extends React.Component<IResultsProp, IResultsState> {
    constructor(props: any) {
        super(props);
        this.state = {
            assignment: this.props.labs[0],
            selectedStudent: this.props.students[0],
            students: this.props.students,
        };
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        if (this.props.students.length > 0) {
            studentLab = <StudentLab course={this.props.course}
                                     assignment={this.state.assignment}
                                     student={this.state.selectedStudent}
            />;
        }

        const searchIcon: JSX.Element = <span className="input-group-addon">
            <i className="glyphicon glyphicon-search"></i>
        </span>;

        return (
            <div>
                <h1>Result: {this.props.course.name}</h1>
                <Row>
                    <div className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                                addonBefore={searchIcon}
                                placeholder="Search for students"
                                onChange={(query) => this.handleOnchange(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                                      data={this.state.students}
                                      selector={(item: IUser) => this.getResultSelector(item)}
                        />
                    </div>
                    <div className="col-lg-6 col-md-6 col-sm-12">
                        {studentLab}
                    </div>
                </Row>
            </div>
        );
    }

    private getResultHeader(): string[] {
        let headers: string[] = ["Name", "Slipdays"];
        headers = headers.concat(this.props.labs.map((e) => e.name));
        return headers;
    }

    private getResultSelector(student: IUser): Array<string | JSX.Element> {
        let selector: Array<string | JSX.Element> = [student.firstName + " " + student.lastName, "5"];
        selector = selector.concat(this.props.labs.map((e) => <a className="lab-result-cell"
                                                                 onClick={() => this.handleOnclick(student, e)}
                                                                 href="#">
            {Math.floor((Math.random() * 100) + 1).toString() + "%"}</a>));
        return selector;
    }

    private handleOnclick(std: IUser, lab: IAssignment): void {
        this.setState({
            selectedStudent: std,
            assignment: lab,
        });
    }

    private handleOnchange(query: string):void {
        query = query.toLowerCase();
        let filteredData: IUser[] = [];
        this.props.students.forEach((std)=> {
            if (std.firstName.toLowerCase().indexOf(query) != -1
                || std.lastName.toLowerCase().indexOf(query) != -1
                || std.email.toLowerCase().indexOf(query) != -1
            ){
                filteredData.push(std);
            }
        });

        this.setState({
            students: filteredData,
        })
    }

}
export {Results};
