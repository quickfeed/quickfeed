import { json } from "overmind"
import { useParams } from "react-router"
import { Assignment, Course, Enrollment, EnrollmentLink, GradingBenchmark, Group, Review, Submission, SubmissionLink, User } from "../proto/ag/ag_pb"
import { Score } from "../proto/kit/score/score_pb"
import { Row, RowElement } from "./components/DynamicTable"
import { useActions, useAppState } from "./overmind"
import { UserCourseSubmissions } from "./overmind/state"

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
// layoutTime = "2021-03-20T23:59:00"
export const timeFormatter = (deadline: string): Deadline => {
    const timeToDeadline = new Date(deadline).getTime() - new Date().getTime()
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
                x = x ? func.call(x) : func.call(a)
                y = y ? func.call(y) : func.call(b)
            })
        } else {
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

// TODO: Could be computed on the backend (https://github.com/quickfeed/quickfeed/issues/420)
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


export const isValid = (elm: User | EnrollmentLink): boolean => {
    if (elm instanceof User) {
        return elm.getName().length > 0 && elm.getEmail().length > 0 && elm.getStudentid().length > 0
    }
    if (elm instanceof EnrollmentLink) {
        return elm.getEnrollment()?.getUser() !== undefined && elm.getSubmissionsList().length > 0
    }
    return true
}

/** hasEnrollment returns true if any of the provided has been approved */
export const hasEnrollment = (enrollments: Enrollment[]): boolean => {
    return enrollments.some(enrollment => enrollment.getStatus() > Enrollment.UserStatus.PENDING)
}

export const isStudent = (enrollment: Enrollment): boolean => { return hasStudent(enrollment.getStatus()) }
export const isTeacher = (enrollment: Enrollment): boolean => { return hasTeacher(enrollment.getStatus()) }
export const isPending = (enrollment: Enrollment): boolean => { return hasPending(enrollment.getStatus()) }

export const isPendingGroup = (group: Group): boolean => { return group.getStatus() === Group.GroupStatus.PENDING }
export const isApprovedGroup = (group: Group): boolean => { return group.getStatus() === Group.GroupStatus.APPROVED }

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

export const isCourseCreator = (user: User, course: Course): boolean => { return user.getId() === course.getCoursecreatorid() }
export const isAuthor = (user: User, review: Review): boolean => { return user.getId() === review.getReviewerid() }

export const isManuallyGraded = (assignment: Assignment): boolean => {
    return assignment.getReviewers() > 0
}

export const isApproved = (submission: Submission): boolean => { return submission.getStatus() === Submission.Status.APPROVED }
export const isRevision = (submission: Submission): boolean => { return submission.getStatus() === Submission.Status.REVISION }
export const isRejected = (submission: Submission): boolean => { return submission.getStatus() === Submission.Status.REJECTED }

export const hasReviews = (submission: Submission): boolean => { return json(submission).getReviewsList().length > 0 }
export const hasBenchmarks = (obj: Review | Assignment): boolean => { return json(obj).getGradingbenchmarksList().length > 0 }
export const hasCriteria = (benchmark: GradingBenchmark): boolean => { return json(benchmark).getCriteriaList().length > 0 }
export const hasEnrollments = (obj: Group): boolean => { return json(obj).getEnrollmentsList().length > 0 }

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

// Helper functions for default values for new courses
export const defaultTag = (date: Date): string => {
    return date.getMonth() >= 10 || date.getMonth() < 4 ? "Spring" : "Fall"
}

export const defaultYear = (date: Date): number => {
    return (date.getMonth() <= 11 && date.getDate() <= 31) && date.getMonth() > 10 ? (date.getFullYear() + 1) : date.getFullYear()
}

export const generateSubmissionRows = (links: UserCourseSubmissions[], cellGenerator: (s: SubmissionLink, e?: Enrollment) => RowElement, groupName?: boolean, assignmentID?: number): Row[] => {
    const state = useAppState()
    const course = state.courses.find(c => c.getId() === state.activeCourse)
    return links?.map((link) => {
        const row: Row = []
        if (link.enrollment && link.user) {
            const url = course ? userRepoLink(course, link.user) : userLink(link.user)
            row.push({ value: link.user.getName(), link: url })
            groupName && row.push(link.enrollment.getGroup()?.getName() ?? "")
        } else if (link.group) {
            const data: RowElement = course ? { value: link.group.getName(), link: groupRepoLink(course, link.group) } : link.group.getName()
            row.push(data)
        }
        if (link.submissions) {
            for (const submissionLink of link.submissions) {
                if (state.review.assignmentID > 0 && submissionLink.getAssignment()?.getId() != state.review.assignmentID) {
                    continue
                }
                row.push(cellGenerator(submissionLink, link.enrollment))
            }
        }
        return row
    })
}

export const generateAssignmentsHeader = (base: RowElement[], assignments: Assignment[], group: boolean, assignmentID?: number): Row => {
    const actions = useActions()
    for (const assignment of assignments) {
        if (assignmentID && assignment.getId() !== assignmentID) {
            continue
        }
        if (group && assignment.getIsgrouplab()) {
            base.push({ value: `${assignment.getName()} (g)`, onClick: () => actions.review.setAssignmentID(assignment.getId()) })
        }
        if (!group) {
            base.push({ value: assignment.getIsgrouplab() ? `${assignment.getName()} (g)` : assignment.getName(), onClick: () => actions.review.setAssignmentID(assignment.getId()) })
        }
    }
    return base
}

const slugify = (str: string): string => {
    str = str.replace(/^\s+|\s+$/g, "").toLowerCase()

    // Remove accents, swap ñ for n, etc
    const from = "ÁÄÂÀÃÅČÇĆĎÉĚËÈÊẼĔȆÍÌÎÏŇÑÓÖÒÔÕØŘŔŠŤÚŮÜÙÛÝŸŽáäâàãåčçćďéěëèêẽĕȇíìîïňñóöòôõøðřŕšťúůüùûýÿžþÞĐđßÆaæ·/,:;&"
    const to = "AAAAAACCCDEEEEEEEEIIIINNOOOOOORRSTUUUUUYYZaaaaa-cccdeeeeeeeeiiiinnooooo-orrstuuuuuyyzbBDdBAa-------"
    for (let i = 0; i < from.length; i++) {
        str = str.replace(new RegExp(from.charAt(i), "g"), to.charAt(i))
    }

    // Remove invalid chars, replace whitespace by dashes, collapse dashes
    return str.replace(/[^a-z0-9 -_]/g, "").replace(/\s+/g, "-").replace(/-+/g, "-")
}

export const userLink = (user: User): string => {
    return `https://github.com/${user.getLogin()}`
}

export const userRepoLink = (course: Course, user: User): string => {
    return `https://github.com/${course.getOrganizationpath()}/${user.getLogin()}-labs`
}

export const groupRepoLink = (course: Course, group: Group): string => {
    course.getOrganizationpath()
    return `https://github.com/${course.getOrganizationpath()}/${slugify(group.getName())}`
}
/* Use this function to simulate a delay in the loading of data */
/* Used in development to simulate a slow network connection */
const delay = (ms: number) => {
    return new Promise(resolve => setTimeout(resolve, ms))
}
