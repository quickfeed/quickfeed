import * as React from "react";
import { IAssignment, ICourse, ITestCases, ILabInfo } from "../../models";
import {LabResultView} from "../../pages/views/LabResultView";

interface IStudentLabProbs {
    course: ICourse;
    assignment: IAssignment;
}

class StudentLab extends React.Component<IStudentLabProbs, undefined> {
    public render() {
        //return <h1>{this.props.assignment.name}</h1>;
        // TODO: fetch real data from backend database for corresponding course assignment
        let testCases: ITestCases[] = [
            {name: "Test Case 1", score: 60, points: 100, weight: 1},
            {name: "Test Case 2", score: 50, points: 100, weight: 1},
            {name: "Test Case 3", score: 40, points: 100, weight: 1},
            {name: "Test Case 4", score: 30, points: 100, weight: 1},
            {name: "Test Case 5", score: 20, points: 100, weight: 1}
        ];

        let labInfo: ILabInfo = {
            lab: this.props.assignment.name,
            course: this.props.course.name,
            score: 50,
            weight: 100,
            test_cases: testCases,
            pass_tests: 10,
            fail_tests: 20,
            exec_time: 0.33,
            build_time: new Date(2017, 5, 25),
            build_id: 10
        };
        return <LabResultView labInfo={labInfo}></LabResultView>
    }
}

export { StudentLab, IStudentLabProbs };
