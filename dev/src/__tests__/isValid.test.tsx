import {isValid} from "../Helpers"
import {User, EnrollmentLink, Enrollment, Submission, SubmissionLink} from "../../proto/ag/ag_pb"

describe("User and enrollment validation", ()=> {
    it("User should be valid", () =>{
        const user = new User().setId(1).setName("Test User").setEmail("mail@mail.com").setStudentid("1234567")
        const isValidUser = isValid(user)
        expect(isValidUser).toBe(true)
    });

    it("User should not be valid if name is empty", () =>{
        const user2 = new User().setId(2).setEmail("mail@mail.com").setStudentid("1234567")
        const isValidUser = isValid(user2)
        expect(isValidUser).toBe(false)
    });
    
    //Should isValid have a function that checks that it is a legit email, and not just a string with length > 0?
    it("User should not be valid if email is empty", ()=>{
        const user3 = new User().setId(1).setName("Test User3").setStudentid("1234567")
        const isValidUser = isValid(user3)
        expect(isValidUser).toBe(false)
    });

    it("User should not be valid if studentId is empty", ()=>{
        const user4 = new User().setId(4).setName("Test User3").setEmail("mail@mail.com")
        const isValidUser = isValid(user4)
        expect(isValidUser).toBe(false)
    });

    it("User should not be valid if name,email and studentId is empty", ()=>{
        const user5 = new User().setId(5)
        const isValidUser = isValid(user5)
        expect(isValidUser).toBe(false)
    });

    it("If enrollment link is valid it should pass", () =>{
        const user = new User().setId(6)
        const enrollment = new Enrollment().setId(1).setUser(user)
        const submission = new Submission().setId(1)
        const submissionLink1 = new SubmissionLink().setSubmission(submission)
        var submissionArray = [submissionLink1]
        const enrollmentLink = new EnrollmentLink().setEnrollment(enrollment).setSubmissionsList(submissionArray)
        const isValidEnrollmentlink = isValid(enrollmentLink)
        expect(isValidEnrollmentlink).toBe(true)
    });
    
    it("If enrollment link has no submission list, enrollment or user it should be invalid", () =>{
        const user2 = new User().setId(6)
        const enrollmentLink2 = new EnrollmentLink()
        const isValidEnrollmentlink2 = isValid(enrollmentLink2)
        expect(isValidEnrollmentlink2).toBe(false)
    })
}); 
