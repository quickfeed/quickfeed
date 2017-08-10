import * as React from "react";
import { IAssignment, ICourse, IStudentSubmission, IUser, IUserCourseWithUser } from "../../models";

import { DynamicTable, Row, Search, StudentLab } from "../../components";

interface IResultsProp {
    course: ICourse;
    students: IUserCourseWithUser[];
    labs: IAssignment[];
}
interface IResultsState {
    assignment?: IStudentSubmission;
    students: IUserCourseWithUser[];
}
class Results extends React.Component<IResultsProp, IResultsState> {
    constructor(props: IResultsProp) {
        super(props);

        if (this.props.students[0] && this.props.students[0].course.assignments[0]) {
            this.state = {
                assignment: this.props.students[0].course.assignments[0],
                students: this.props.students,
            };
        } else {
            this.state = {
                assignment: undefined,
                students: this.props.students,
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        if (this.props.students.length > 0 && this.state.assignment) {
            studentLab = <StudentLab
                course={this.props.course}
                assignment={this.state.assignment}
            />;
        }

        return (
            <div>
                <h1>Result: {this.props.course.name}</h1>
                <Row>
                    <div className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
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
        let selector: Array<string | JSX.Element> = [student.user.name, "5"];
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
            if (std.user.name.toLowerCase().indexOf(query) !== -1
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
