import { useParams } from "react-router"
import { Assignment, Enrollment, EnrollmentLink, Submission, User } from "../proto/ag/ag_pb"
import { Score } from "../proto/kit/score/score_pb"

export enum Color {
    RED = "danger",
    BLUE = "primary",
    GREEN = "success",
    YELLOW = "warning",
    GRAY = "secondary",
}

export enum Sort {
    NAME,
    STATUS,
    ID
}

/** Returns a string with a prettier format for a deadline */
export const getFormattedTime = (deadline_string: string): string => {
    const months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']
    const deadline = new Date(deadline_string)
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} ${deadline.getHours()}:${deadline.getMinutes() < 10 ? '0' + deadline.getMinutes() : deadline.getMinutes()}`
}

export interface Deadline {
    className: string,
    message: string,
    daysUntil: number,
}

/** Utility function for LandingpageTable functionality. To format the output string and class/css based on how far the deadline is in the future */
export const timeFormatter = (deadline: string, now: Date): Deadline => {
    const timeOfDeadline = new Date(deadline)
    const timeToDeadline = timeOfDeadline.getTime() - now.getTime()
    const days = Math.floor(timeToDeadline / (1000 * 3600 * 24))
    const hours = Math.floor(timeToDeadline / (1000 * 3600))
    const minutes = Math.floor((timeToDeadline % (1000 * 3600)) / (1000 * 60))

    if (timeToDeadline < 0) {
        return { className: "table-danger", message: `Expired ${-days > 0 ? -days + " days ago" : -hours + " hours"}`, daysUntil: 0 }
    }

    if (days == 0) {
        return { className: "table-danger", message: `${hours} hours and ${minutes} minutes to deadline!`, daysUntil: 0 }
    }

    if (days < 3) {
        return { className: "table-warning", message: `${days} day${days == 1 ? ' ' : 's'} to deadline`, daysUntil: days }
    }

    if (days < 14) {
        return { className: "table-primary", message: `${days} days`, daysUntil: days }
    }

    return { className: "", message: "", daysUntil: days }
}
export const layoutTime = "2021-03-20T23:59:00"

// Used for displaying enrollment status
export const EnrollmentStatus = {
    0: "None",
    1: "Pending",
    2: "Student",
    3: "Teacher",
}

/* 
    arr: Any array, ex. Enrollment[], User[],    
    funcs: an array of functions that will be applied in order to reach the field to sort on
    by: A function returning an element to sort on

    Example:
        To sort state.enrollmentsByCourseId[2].getUser().getName() by name, call like
        (state.enrollmentsByCourseId[2], [Enrollment.prototype.getUser], User.prototype.getName)

    Returns an array of the same type as arr, sorted by the by-function
*/
export const sortByField = (arr: any[], funcs: Function[], by: Function, descending?: boolean) => {
    const unsortedArray = Object.assign([], arr)
    const sortedArray = unsortedArray.sort((a, b) => {
        let x: any
        let y: any
        if (!a || !b) {
            return 0
        }
        if (funcs.length > 0) {
            funcs.forEach(func => {
                if (!x) {
                    x = func.call(a)
                } else {
                    x = func.call(x)
                }
                if (!y) {
                    y = func.call(b)
                } else {
                    y = func.call(y)
                }
            })
        }
        else {
            x = a
            y = b
        }
        if (by.call(x) === by.call(y)) {
            return 0
        }
        if (by.call(x) < by.call(y)) {
            return descending ? 1 : -1
        }
        if (by.call(x) > by.call(y)) {
            return descending ? -1 : 1
        }
        return -1
    })
    return sortedArray
}


/** getPassedTestCount returns a string with the number of passed tests and the total number of tests */
export const getPassedTestsCount = (score: Score[]): string => {
    let totalTests = 0
    let passedTests = 0
    score.forEach(score => {
        if (score.getScore() === score.getMaxscore()) {
            passedTests++
        }
        totalTests++
    })
    if (totalTests === 0) {
        return ""
    }
    return `${passedTests}/${totalTests}`
}


export const isValid = (element: unknown): boolean => {
    if (element instanceof User) {
        if (element.getName().length === 0 || element.getEmail().length === 0 || element.getStudentid().length === 0) {
            return false
        }
    }
    if (element instanceof EnrollmentLink) {
        if (!element.getEnrollment() && !element.getEnrollment()?.getUser() && element.getSubmissionsList().length === 0) {
            return false
        }
    }
    return true
}

/** hasEnrollment returns true if the user has any approved enrollments, false otherwise */
export const hasEnrollment = (enrollments: Enrollment[]): boolean => {
    for (const enrollment of enrollments) {
        if (enrollment.getStatus() > Enrollment.UserStatus.PENDING) {
            return true
        }
    }
    return false
}

export const isStudent = (enrollment: Enrollment): boolean => { return hasStudent(enrollment.getStatus()) }
export const isTeacher = (enrollment: Enrollment): boolean => { return hasTeacher(enrollment.getStatus()) }
export const isPending = (enrollment: Enrollment): boolean => { return hasPending(enrollment.getStatus()) }

/** isEnrolled returns true if the user is enrolled in the course, and is no longer pending. */
export const isEnrolled = (enrollment: Enrollment): boolean => { return enrollment.getStatus() >= Enrollment.UserStatus.STUDENT }

/** toggleUserStatus switches between teacher and student status. */
export const toggleUserStatus = (enrollment: Enrollment): Enrollment.UserStatus => {
    return isTeacher(enrollment) ? Enrollment.UserStatus.STUDENT : Enrollment.UserStatus.TEACHER
}

export const hasNone = (status: Enrollment.UserStatus): boolean => { return status === Enrollment.UserStatus.NONE }
export const hasPending = (status: Enrollment.UserStatus): boolean => { return status === Enrollment.UserStatus.PENDING }
export const hasStudent = (status: Enrollment.UserStatus): boolean => { return status === Enrollment.UserStatus.STUDENT }
export const hasTeacher = (status: Enrollment.UserStatus): boolean => { return status === Enrollment.UserStatus.TEACHER }

/** hasEnrolled returns true if user has enrolled in course, or is pending approval. */
export const hasEnrolled = (status: Enrollment.UserStatus): boolean => { return status >= Enrollment.UserStatus.PENDING }

export const isVisible = (enrollment: Enrollment): boolean => { return enrollment.getState() === Enrollment.DisplayState.VISIBLE }
export const isFavorite = (enrollment: Enrollment): boolean => { return enrollment.getState() === Enrollment.DisplayState.FAVORITE }

export const isManuallyGraded = (assignment: Assignment): boolean => {
    return assignment.getReviewers() > 0
}

/** getCourseID returns the course ID determined by the current route */
export const getCourseID = (): number => {
    const route = useParams<{ id?: string }>()
    return Number(route.id)
}

export const isHidden = (value: string, query: string): boolean => {
    return !value.toLowerCase().includes(query) && query.length > 0
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
export const assignmentStatusText = (assignment: Assignment, submission: Submission): string => {
    // If the submission is not graded, return a descriptive text
    if (submission.getStatus() === Submission.Status.NONE) {
        // If the assignment requires manual approval, and the score is above the threshold, return Await Approval
        if (!assignment.getAutoapprove() && submission.getScore() >= assignment.getScorelimit()) {
            return "Awaiting approval"
        }
        if (submission.getScore() < assignment.getScorelimit() && submission.getStatus() !== Submission.Status.APPROVED) {
            return `Need ${assignment.getScorelimit()}% score for approval`
        }
    }
    // If the submission is graded, return the status
    return SubmissionStatus[submission.getStatus()]
}

/* Use this function to simulate a delay in the loading of data */
/* Used in development to simulate a slow network connection */
const delay = (ms: number) => {
    return new Promise(resolve => setTimeout(resolve, ms));
}
