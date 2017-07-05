import * as React from "react";
import { IAssignment, ICourse, ILabInfo, IStudentSubmission, ITestCases, IUser } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";

interface IStudentLabProbs {
    course: ICourse;
    assignment: IStudentSubmission;
    student?: IUser;
}

class StudentLab extends React.Component<IStudentLabProbs, {}> {
    public render() {
        // return <h1>{this.props.assignment.name}</h1>;
        // TODO: fetch real data from backend database for corresponding course assignment
        /*const testCases: ITestCases[] = [
            { name: "Test Case 1", score: 60, points: 100, weight: 1 },
            { name: "Test Case 2", score: 50, points: 100, weight: 1 },
            { name: "Test Case 3", score: 40, points: 100, weight: 1 },
            { name: "Test Case 4", score: 30, points: 100, weight: 1 },
            { name: "Test Case 5", score: 20, points: 100, weight: 1 },
        ];

        const labInfo: ILabInfo = {
            lab: this.props.assignment.name,
            courseId: this.props.course.name,
            score: 50,
            weight: 100,
            testCases: testCases,
            passedTests: 10,
            failedTests: 20,
            executetionTime: 0.33,
            buildDate: new Date(2017, 5, 25),
            buildId: 10,
        };
        if (this.props.student) {
            labInfo.student = this.props.student;
        }*/
        return <LabResultView course={this.props.course} labInfo={this.props.assignment}></LabResultView>;
    }
}

export { StudentLab, IStudentLabProbs };
