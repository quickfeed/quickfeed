import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { Enrollment_UserStatus, EnrollmentSchema, ReviewSchema } from "../../proto/qf/types_pb"
import ReviewResult from "../components/ReviewResult"
import { initializeOvermind } from "./TestHelpers"

describe("ReviewResult", () => {
    it.each([
        { name: "teacher", status: Enrollment_UserStatus.TEACHER, want: true },
        { name: "student", status: Enrollment_UserStatus.STUDENT, want: false },
    ])("shows the review's score to a $name: $want", ({ status, want }) => {
        const overmind = initializeOvermind({
            activeCourse: 1n,
            enrollments: [create(EnrollmentSchema, { courseID: 1n, userID: 2n, status })],
        })
        render(
            <Provider value={overmind}>
                <ReviewResult review={create(ReviewSchema, { ID: 1n, score: 6 })} />
            </Provider>
        )
        expect(screen.queryByText("Score:") !== null).toBe(want)
    })
})
