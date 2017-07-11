import * as React from "react";
import { IAssignment, ICourse, IStudentSubmission, IUser, IUserCourseWithUser } from "../../models";

import { DynamicTable, Row, Search, StudentLab } from "../../components";

interface IResultsProp {
    course: ICourse;
    students: IUserCourseWithUser[];
    labs: IAssignment[];
}
interface IResultsState {
    assignment: IStudentSubmission;
    students: IUserCourseWithUser[];
}
class Results extends React.Component<IResultsProp, IResultsState> {
    constructor(props: IResultsProp) {
        super(props);
        this.state = {
            assignment: this.props.students[0].course.assignments[0],
            students: this.props.students,
        };
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        if (this.props.students.length > 0) {
            studentLab = <StudentLab
                course={this.props.course}
                assignment={this.state.assignment}
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
                            selector={(item: IUserCourseWithUser) => this.getResultSelector(item)}
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

    private getResultSelector(student: IUserCourseWithUser): Array<string | JSX.Element> {
        let selector: Array<string | JSX.Element> = [student.user.firstname + " " + student.user.lastname, "5"];
        selector = selector.concat(student.course.assignments.map((e, i) => <a className="lab-result-cell"
            onClick={() => this.handleOnclick(e)}
            href="#">
            {e.latest ? (e.latest.score + "%") : "N/A"}</a>));
        return selector;
    }

    private handleOnclick(item: IStudentSubmission): void {
        this.setState({
            assignment: item,
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IUserCourseWithUser[] = [];
        this.props.students.forEach((std) => {
            if (std.user.firstname.toLowerCase().indexOf(query) !== -1
                || std.user.lastname.toLowerCase().indexOf(query) !== -1
                || std.user.email.toLowerCase().indexOf(query) !== -1
            ) {
                filteredData.push(std);
            }
        });

        this.setState({
            students: filteredData,
        });
    }

}
export { Results };
