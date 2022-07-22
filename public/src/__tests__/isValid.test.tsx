import { isValid } from "../Helpers"
import { User, Enrollment, Submission, EnrollmentLink, SubmissionLink } from "../../proto/qf/types_pb"

describe("User and enrollment validation", () => {
    it("User should be valid", () => {
        const user = new User().setId(1).setName("Test User").setEmail("mail@mail.com").setStudentid("1234567")
        const isValidUser = isValid(user)
        expect(isValidUser).toBe(true)
    })

    it("User should not be valid if name is empty", () => {
        const user2 = new User().setId(2).setEmail("mail@mail.com").setStudentid("1234567")
        const isValidUser = isValid(user2)
        expect(isValidUser).toBe(false)
    })

    it("User should not be valid if email is empty", () => {
        const user3 = new User().setId(1).setName("Test User3").setStudentid("1234567")
        const isValidUser = isValid(user3)
        expect(isValidUser).toBe(false)
    })

    it("Email is a valid email", () => {
        const email = "hei@mail.com"
        const regex = /\S+@\S+\.\S+/
        let valid = false
        const test = email.match(regex)

        if (test !== null) {
            if (test.length > 0) {
                valid = true
            }
        }
        expect(valid).toBe(true)
    })

    it("Email is not a valid email", () => {
        const email = "hei@mail"
        const regex = /\S+@\S+\.\S+/
        let valid = false
        const test = email.match(regex)

        if (test !== null) {
            if (test.length > 0) {
                valid = true
            }
        }
        expect(valid).toBe(false)
    })

    it("User should not be valid if studentId is empty", () => {
        const user4 = new User().setId(4).setName("Test User3").setEmail("mail@mail.com")
        const isValidUser = isValid(user4)
        expect(isValidUser).toBe(false)
    })

    it("User should not be valid if name,email and studentId is empty", () => {
        const user5 = new User().setId(5)
        const isValidUser = isValid(user5)
        expect(isValidUser).toBe(false)
    })

    it("If enrollment link is valid it should pass", () => {
        const user = new User().setId(6)
        const enrollment = new Enrollment().setId(1).setUser(user)
        const submission = new Submission().setId(1)
        const submissionLink = new SubmissionLink().setSubmission(submission)
        const submissionArray = [submissionLink]
        const enrollmentLink = new EnrollmentLink().setEnrollment(enrollment).setSubmissionsList(submissionArray)
        const isValidEnrollmentLink = isValid(enrollmentLink)
        expect(isValidEnrollmentLink).toBe(true)
    })

    it("If enrollment link has no submission list, enrollment or user it should be invalid", () => {
        const enrollmentLink = new EnrollmentLink()
        const isValidEnrollmentLink = isValid(enrollmentLink)
        expect(isValidEnrollmentLink).toBe(false)
    })
})
