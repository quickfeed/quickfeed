import { useParams } from "react-router"
import { Assignment, Course, Enrollment, GradingBenchmark, Group, Review, Submission, User, Enrollment_UserStatus, Group_GroupStatus, Enrollment_DisplayState, Submission_Status, Submissions, Grade } from "../proto/qf/types_pb"
import { Score } from "../proto/kit/score/score_pb"
import { CourseGroup, SubmissionOwner } from "./overmind/state"
import { Timestamp } from "@bufbuild/protobuf"
import { CourseSubmissions } from "../proto/qf/requests_pb"

export enum Color {
    RED = "danger",
    BLUE = "primary",
    GREEN = "success",
    YELLOW = "warning",
    GRAY = "secondary",
    WHITE = "light",
    BLACK = "dark",
}

export enum Sort {
    NAME,
    STATUS,
    ID
}

// ConnStatus indicates the status of streaming connection to the server
export enum ConnStatus {
    CONNECTED,
    DISCONNECTED,
    RECONNECTING,
}

const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"]

/** Returns a string with a prettier format for a deadline */
export const getFormattedTime = (timestamp: Timestamp | undefined): string => {
    if (!timestamp) {
        return "N/A"
    }
    const deadline = timestamp.toDate()
    const minutes = deadline.getMinutes()
    const zero = minutes < 10 ? "0" : ""
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} ${deadline.getHours()}:${zero}${minutes}`
}

export interface Deadline {
    className: string,
    message: string,
    daysUntil: number,
}

/**
 * Utility function for LandingPageTable to format the output string and class/css
 * depending on how far into the future the deadline is.
 *
 * layoutTime = "2021-03-20T23:59:00"
 */
export const timeFormatter = (deadline: Timestamp): Deadline => {
    const timeToDeadline = deadline.toDate().getTime()
    const days = Math.floor(timeToDeadline / (1000 * 3600 * 24))
    const hours = Math.floor(timeToDeadline / (1000 * 3600))
    const minutes = Math.floor((timeToDeadline % (1000 * 3600)) / (1000 * 60))

    if (timeToDeadline < 0) {
        const daysSince = -days
        const hoursSince = -hours
        return { className: "table-danger", message: `Expired ${daysSince > 0 ? `${daysSince} days ago` : `${hoursSince} hours ago`}`, daysUntil: 0 }
    }

    if (days === 0) {
        return { className: "table-danger", message: `${hours} hours and ${minutes} minutes to deadline!`, daysUntil: 0 }
    }

    if (days < 3) {
        return { className: "table-warning", message: `${days} day${days === 1 ? " " : "s"} to deadline`, daysUntil: days }
    }

    if (days < 14) {
        return { className: "table-primary", message: `${days} days`, daysUntil: days }
    }

    return { className: "", message: "", daysUntil: days }
}

// Used for displaying enrollment status
export const EnrollmentStatus = {
    0: "None",
    1: "Pending",
    2: "Student",
    3: "Teacher",
}

// TODO: Could be computed on the backend (https://github.com/quickfeed/quickfeed/issues/420)
/** getPassedTestCount returns a string with the number of passed tests and the total number of tests */
export const getPassedTestsCount = (score: Score[]): string => {
    let totalTests = 0
    let passedTests = 0
    score.forEach(s => {
        if (s.Score === s.MaxScore) {
            passedTests++
        }
        totalTests++
    })
    if (totalTests === 0) {
        return ""
    }
    return `${passedTests}/${totalTests}`
}

/** hasEnrollment returns true if any of the provided has been approved */
export const hasEnrollment = (enrollments: Enrollment[]): boolean => {
    return enrollments.some(enrollment => enrollment.status > Enrollment_UserStatus.PENDING)
}

export const isStudent = (enrollment: Enrollment): boolean => { return hasStudent(enrollment.status) }
export const isTeacher = (enrollment: Enrollment): boolean => { return hasTeacher(enrollment.status) }
export const isPending = (enrollment: Enrollment): boolean => { return hasPending(enrollment.status) }

export const isPendingGroup = (group: Group): boolean => { return group.status === Group_GroupStatus.PENDING }
export const isApprovedGroup = (group: Group): boolean => { return group.status === Group_GroupStatus.APPROVED }

/** isEnrolled returns true if the user is enrolled in the course, and is no longer pending. */
export const isEnrolled = (enrollment: Enrollment): boolean => { return enrollment.status >= Enrollment_UserStatus.STUDENT }

export const hasNone = (status: Enrollment_UserStatus): boolean => { return status === Enrollment_UserStatus.NONE }
export const hasPending = (status: Enrollment_UserStatus): boolean => { return status === Enrollment_UserStatus.PENDING }
export const hasStudent = (status: Enrollment_UserStatus): boolean => { return status === Enrollment_UserStatus.STUDENT }
export const hasTeacher = (status: Enrollment_UserStatus): boolean => { return status === Enrollment_UserStatus.TEACHER }

/** hasEnrolled returns true if user has enrolled in course, or is pending approval. */
export const hasEnrolled = (status: Enrollment_UserStatus): boolean => { return status >= Enrollment_UserStatus.PENDING }

export const isVisible = (enrollment: Enrollment): boolean => { return enrollment.state === Enrollment_DisplayState.VISIBLE }
export const isFavorite = (enrollment: Enrollment): boolean => { return enrollment.state === Enrollment_DisplayState.FAVORITE }

export const isAuthor = (user: User, review: Review): boolean => { return user.ID === review.ReviewerID }

export const isManuallyGraded = (assignment: Assignment): boolean => {
    return assignment.reviewers > 0
}

export const isApproved = (submission: Submission): boolean => { return submission.Grades.every(grade => grade.Status === Submission_Status.APPROVED) }
export const isRevision = (submission: Submission): boolean => { return submission.Grades.every(grade => grade.Status === Submission_Status.REVISION) }
export const isRejected = (submission: Submission): boolean => { return submission.Grades.every(grade => grade.Status === Submission_Status.REJECTED) }

export const hasAllStatus = (submission: Submission, status: Submission_Status): boolean => {
    return submission.Grades.every(grade => grade.Status === status)
}

export const hasReviews = (submission: Submission): boolean => { return submission.reviews.length > 0 }
export const hasBenchmarks = (obj: Review | Assignment): boolean => { return obj.gradingBenchmarks.length > 0 }
export const hasCriteria = (benchmark: GradingBenchmark): boolean => { return benchmark.criteria.length > 0 }
export const hasEnrollments = (obj: Group): boolean => { return obj.enrollments.length > 0 }

export const getStatusByUser = (submission: Submission | null, userID: bigint): Submission_Status => {
    if (!submission) {
        return Submission_Status.NONE
    }
    const grade = submission.Grades.find(grade => grade.UserID === userID)
    if (!grade) {
        return Submission_Status.NONE
    }
    return grade.Status
}

export const setStatusByUser = (submission: Submission, userID: bigint, status: Submission_Status): Submission => {
    const grades = submission.Grades.map(grade => {
        if (grade.UserID === userID) {
            return new Grade({ ...grade, Status: status })
        }
        return grade
    })
    return new Submission({ ...submission, Grades: grades })
}

export const setStatusAll = (submission: Submission, status: Submission_Status): Submission => {
    const grades = submission.Grades.map(grade => {
        return new Grade({ ...grade, Status: status })
    })
    return new Submission({ ...submission, Grades: grades })
}

/** getCourseID returns the course ID determined by the current route */
export const getCourseID = (): bigint => {
    const route = useParams<{ id?: string }>()
    return route.id ? BigInt(route.id) : BigInt(0)
}

export const isHidden = (value: string, query: string): boolean => {
    return !value.toLowerCase().includes(query) && query.length > 0
}

/** getSubmissionsScore calculates the total score of all submissions */
export const getSubmissionsScore = (submissions: Submission[]): number => {
    let score = 0
    submissions.forEach(submission => {
        score += submission.score
    })
    return score
}

/** getNumApproved returns the number of approved submissions */
export const getNumApproved = (submissions: Submission[]): number => {
    let num = 0
    submissions.forEach(submission => {
        if (isApproved(submission)) {
            num++
        }
    })
    return num
}

export const EnrollmentStatusBadge = {
    0: "",
    1: "badge badge-info",
    2: "badge badge-primary",
    3: "badge badge-danger",
}

/** SubmissionStatus returns a string with the status of the submission, given the status number, ex. Submission.Status.APPROVED -> "Approved" */
export const SubmissionStatus = {
    0: "None",
    1: "Approved",
    2: "Rejected",
    3: "Revision",
}

// TODO: This could possibly be done on the server. Would need to add a field to the proto submission/score model.
/** assignmentStatusText returns a string that is used to tell the user what the status of their submission is */
export const assignmentStatusText = (assignment: Assignment, submission: Submission, status: Submission_Status): string => {
    // If the submission is not graded, return a descriptive text
    if (status === Submission_Status.NONE) {
        // If the assignment requires manual approval, and the score is above the threshold, return Await Approval
        if (!assignment.autoApprove && submission.score >= assignment.scoreLimit) {
            return "Awaiting approval"
        }
        if (submission.score < assignment.scoreLimit) {
            return `Need ${assignment.scoreLimit}% score for approval`
        }
    }
    // If the submission is graded, return the status
    return SubmissionStatus[status]
}

// Helper functions for default values for new courses
export const defaultTag = (date: Date): string => {
    return date.getMonth() >= 10 || date.getMonth() < 4 ? "Spring" : "Fall"
}

export const defaultYear = (date: Date): number => {
    return (date.getMonth() <= 11 && date.getDate() <= 31) && date.getMonth() > 10 ? (date.getFullYear() + 1) : date.getFullYear()
}

export const userLink = (user: User): string => {
    return `https://github.com/${user.Login}`
}

export const userRepoLink = (user: User, course?: Course): string => {
    if (!course) {
        return userLink(user)
    }
    return `https://github.com/${course.ScmOrganizationName}/${user.Login}-labs`
}

export const groupRepoLink = (group: Group, course?: Course): string => {
    if (!course) {
        return ""
    }
    return `https://github.com/${course.ScmOrganizationName}/${group.name}`
}

export const getSubmissionCellColor = (submission: Submission): string => {
    if (isApproved(submission)) {
        return "result-approved"
    }
    if (isRevision(submission)) {
        return "result-revision"
    }
    if (isRejected(submission)) {
        return "result-rejected"
    }
    return "clickable"
}

// pattern for group name validation. Only letters, numbers, underscores and dashes are allowed.
const pattern = /^[a-zA-Z0-9_-]+$/
export const validateGroup = (group: CourseGroup): { valid: boolean, message: string } => {
    if (group.name.length === 0) {
        return { valid: false, message: "Group name cannot be empty" }
    }
    if (group.name.length > 20) {
        return { valid: false, message: "Group name cannot be longer than 20 characters" }
    }
    if (group.name.includes(" ")) {
        // Explicitly warn the user that spaces are not allowed.
        // Common mistake is to use spaces instead of underscores.
        return { valid: false, message: "Group name cannot contain spaces" }
    }
    if (!pattern.test(group.name)) {
        return { valid: false, message: "Group name can only contain letters (a-z, A-Z), numbers, underscores and dashes" }
    }
    if (group.users.length === 0) {
        return { valid: false, message: "Group must have at least one user" }
    }
    return { valid: true, message: "" }
}

// newID returns a new auto-incrementing ID
// Can be used to generate IDs for client-only objects
// such as the Alert object
export const newID = (() => {
    let id: number = 0
    return () => {
        return id++
    }
})()

/* Use this function to simulate a delay in the loading of data */
/* Used in development to simulate a slow network connection */
export const delay = (ms: number) => {
    return new Promise(resolve => setTimeout(resolve, ms))
}


export enum EnrollmentSort {
    Name,
    Status,
    Email,
    Activity,
    Slipdays,
    Approved,
    StudentID
}

export enum SubmissionSort {
    ID,
    Name,
    Status,
    Score,
    Approved
}

/** Sorting */
const enrollmentCompare = (a: Enrollment, b: Enrollment, sortBy: EnrollmentSort, descending: boolean): number => {
    const sortOrder = descending ? -1 : 1
    switch (sortBy) {
        case EnrollmentSort.Name: {
            const nameA = a.user?.Name ?? ""
            const nameB = b.user?.Name ?? ""
            return sortOrder * (nameA.localeCompare(nameB))
        }
        case EnrollmentSort.Status:
            return sortOrder * (a.status - b.status)
        case EnrollmentSort.Email: {
            const emailA = a.user?.Email ?? ""
            const emailB = b.user?.Email ?? ""
            return sortOrder * (emailA.localeCompare(emailB))
        }
        case EnrollmentSort.Activity:
            if (a.lastActivityDate && b.lastActivityDate) {
                return sortOrder * (a.lastActivityDate.toDate().getTime() - b.lastActivityDate.toDate().getTime())
            }
            return 0
        case EnrollmentSort.Slipdays:
            return sortOrder * (a.slipDaysRemaining - b.slipDaysRemaining)
        case EnrollmentSort.Approved:
            return sortOrder * Number(a.totalApproved - b.totalApproved)
        case EnrollmentSort.StudentID: {
            const aID = a.user?.ID ?? BigInt(0)
            const bID = b.user?.ID ?? BigInt(0)
            return sortOrder * Number(aID - bID)
        }
        default:
            return 0
    }
}

export const sortEnrollments = (enrollments: Enrollment[], sortBy: EnrollmentSort, descending: boolean): Enrollment[] => {
    return enrollments.sort((a, b) => {
        return enrollmentCompare(a, b, sortBy, descending)
    })
}

export class SubmissionsForCourse {
    userSubmissions: Map<bigint, Submissions> = new Map()
    groupSubmissions: Map<bigint, Submissions> = new Map()

    /** ForUser returns user submissions for the given enrollment */
    ForUser(enrollment: Enrollment): Submission[] {
        return this.userSubmissions.get(enrollment.ID)?.submissions ?? []
    }

    /** ForGroup returns group submissions for the given group or enrollment */
    ForGroup(group: Group | Enrollment): Submission[] {
        if (group instanceof Group) {
            return this.groupSubmissions.get(group.ID)?.submissions ?? []
        }
        return this.groupSubmissions.get(group.groupID)?.submissions ?? []
    }

    /** ForOwner returns all submissions related to the passed in owner.
     * This is usually the selected group or user. */
    ForOwner(owner: SubmissionOwner): Submission[] {
        if (owner.type === "GROUP") {
            return this.groupSubmissions.get(owner.id)?.submissions ?? []
        }
        return this.userSubmissions.get(owner.id)?.submissions ?? []
    }

    ByID(id: bigint): Submission | undefined {
        for (const submissions of this.userSubmissions.values()) {
            const submission = submissions.submissions.find(s => s.ID === id)
            if (submission) {
                return submission
            }
        }
        for (const submissions of this.groupSubmissions.values()) {
            const submission = submissions.submissions.find(s => s.ID === id)
            if (submission) {
                return submission
            }
        }
        return undefined
    }

    OwnerByID(id: bigint): SubmissionOwner | undefined {
        for (const [key, submissions] of this.userSubmissions.entries()) {
            const submission = submissions.submissions.find(s => s.ID === id)
            if (submission) {
                if (submission.groupID > 0) {
                    return { type: "GROUP", id: submission.groupID }
                }
                return { type: "ENROLLMENT", id: key }
            }
        }
        for (const [key, submissions] of this.groupSubmissions.entries()) {
            const submission = submissions.submissions.find(s => s.ID === id)
            if (submission) {
                return { type: "GROUP", id: key }
            }
        }
        return undefined
    }

    update(owner: SubmissionOwner, submission: Submission) {
        const submissions = this.ForOwner(owner)
        const index = submissions.findIndex(s => s.AssignmentID === submission.AssignmentID)
        if (index === -1) {
            return
        } else {
            submissions[index] = submission
        }
        if (owner.type === "GROUP") {
            const clone = new Map(this.groupSubmissions)
            this.groupSubmissions = clone.set(owner.id, new Submissions({ submissions }))
        } else {
            const clone = new Map(this.userSubmissions)
            this.userSubmissions = clone.set(owner.id, new Submissions({ submissions }))
        }
    }

    setSubmissions(type: "USER" | "GROUP", submissions: CourseSubmissions) {
        const map = new Map<bigint, Submissions>()
        for (const [key, value] of Object.entries(submissions.submissions)) {
            map.set(BigInt(key), value)
        }
        switch (type) {
            case "USER":
                this.userSubmissions = map
                break
            case "GROUP":
                this.groupSubmissions = map
                break
        }
    }
}
