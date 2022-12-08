import { Context } from "."
import { SubmissionSort } from "../Helpers"


export const resetState = ({ state }: Context) => {
    Object.assign(state.review, {
        selectedReview: -1,
        reviews: {},
        minimumScore: 0,
        assignmentID: -1
    })

    const initialState = {
        activeAssignment: -1,
        activeCourse: -1,
        activeEnrollment: null,
        activeSubmissionLink: null,
        query: "",
        sortSubmissionsBy: SubmissionSort.Approved,
        sortAscending: true,
        submissionFilters: [],
        groupView: false,
        status: [],
        assignments: {},
        repositories: {},

        courseGroup: { courseID: 0, enrollments: [], users: [], groupName: "" },
        alerts: [],
        isLoading: true,
        courseEnrollments: {},
        groups: {},
        users: {},
        allUsers: [],
        courses: [],
        courseSubmissions: [],
        courseGroupSubmissions: {},
        submissions: {},
        userGroup: {},
        enrollments: [],
    }

    Object.assign(state, initialState)
}
