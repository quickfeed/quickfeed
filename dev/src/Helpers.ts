/* eslint-disable quotes */

import { useParams } from "react-router"
import { Assignment, Enrollment, EnrollmentLink, Submission, User } from "../proto/ag/ag_pb"
import { Score } from "../proto/kit/score/score_pb"

export interface IBuildInfo {
    builddate: string;
    buildid: number;
    buildlog: string;
    execTime: number
}

export const getBuildInfo = (buildString: string): IBuildInfo => {
    let buildinfo: IBuildInfo
    if (buildString.length === 0) {
        buildinfo = {builddate: "", buildid: 0, buildlog: "", execTime: 0}
    }
    else {
        buildinfo = JSON.parse(buildString)
    }
    return buildinfo
    
}

export enum AlertType {
    INFO,
    DANGER,
    SUCCESS,
    PRIMARY
}

export enum Sort {
    NAME,
    STATUS,
    ID
}

export interface IScoreObjects {
    Secret: string;
    TestName: string;
    Score: number;
    MaxScore: number;
    Weight: number;
}

export const getScoreObjects = (scoreString: string): IScoreObjects[] => {
    const scoreObjects: IScoreObjects[] = []
    if (scoreString.length > 0) {
        const parsedScoreObjects = JSON.parse(scoreString)
        for (const scoreObject in parsedScoreObjects) {
            scoreObjects.push(parsedScoreObjects[scoreObject])
        }
    }
    return scoreObjects
    
}


/** Returns a string with a prettier format for a deadline */
export const getFormattedTime = (deadline_string: string): string => {
    const months = ['January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December']
    const deadline = new Date(deadline_string)
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} ${deadline.getHours()}:${deadline.getMinutes() < 10 ? '0' + deadline.getMinutes() : deadline.getMinutes()}`
}

export interface Deadline {
    className: string,
    message: string,
    daysUntil: number,
}

/** Utility function for LandingpageTable functionality. To format the output string and class/css based on how far the deadline is in the future */
export const timeFormatter = (deadline: string , now: Date): Deadline => {
    const timeOfDeadline = new Date(deadline)
    const timeToDeadline =  timeOfDeadline.getTime() - now.getTime()
    const days = Math.floor(timeToDeadline / (1000 * 3600 * 24))
    const hours = Math.floor(timeToDeadline / (1000 * 3600))
    const minutes = Math.floor((timeToDeadline % (1000 * 3600)) / (1000*60))

    if (timeToDeadline < 0){
        return {className: "table-danger", message: `deadline was ${-days > 0 ? -days+" days" : -hours+" hours"}`, daysUntil: 0}
    }

    if (days == 0) {
        return {className: "table-danger", message: `${hours} hours and ${minutes} minutes to deadline!`, daysUntil: 0}
    }

    if (days < 3){
        return {className: "table-warning", message: `${days} day${days==1?'':'s'} to deadline`, daysUntil: days}
    }
    
    if (days < 14){
        return {className: "table-primary", message: `${days} days until deadline`, daysUntil: days}
    }

    return {className: "", message: "", daysUntil: days}
}
export const layoutTime = "2021-03-20T23:59:00"

// Used for displaying enrollment status
export const EnrollmentStatus = {
    0 : "None",
    1 : "Pending",
    2 : "Student",
    3 : "Teacher",
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

export const SubmissionStatus = {
    0: "None",
    1: "Approved",
    2: "Rejected",
    3: "Revision",
}

export const getPassedTestsCount = (score: Score[]): string => {
    let totalTests = 0
    let passedTests = 0
    score.forEach(score => {
        if (score.getScore() === score.getMaxscore()) {
            passedTests++
        } 
        totalTests++
    })
    return `${passedTests}/${totalTests}`
}


export const isValid = (element: unknown): boolean => {
    if (element instanceof User){
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

export const hasEnrollment = (enrollments: Enrollment[]): boolean => {
    for (const enrollment of enrollments) {
        if (enrollment.getStatus() > Enrollment.UserStatus.PENDING) {
            return true
        }
    }
    return false
}

export const isTeacher = (enrollment: Enrollment): boolean => {
    return enrollment.getStatus() >= Enrollment.UserStatus.TEACHER
}

export const isEnrolled = (enrollment: Enrollment): boolean => {
    return enrollment.getStatus() >= Enrollment.UserStatus.STUDENT
}

export const isManuallyGraded = (assignment: Assignment): boolean => {
    return assignment.getReviewers() > 0
}

export const getCourseID = (): number => {
    const route = useParams<{id?: string}>()
    return Number(route.id)
}

export const isHidden = (value: string, query: string): boolean => {
    return !value.toLowerCase().includes(query) && query.length > 0
}

export const EnrollmentStatusBadge = {
    0 : "",
    1 : "badge badge-info",
    2 : "badge badge-primary",
    3 : "badge badge-danger",
}

/**
 * const test = data.sort((a, b) => {
        const x = isCellElement(a[index])
        const y = isCellElement(b[index])
        if (x && y) {
           return (a as CellElement[])[index].value.localeCompare((b as CellElement[])[index].value)
        }
        if (y && !x) {
            return (a[index] as string).localeCompare((b as CellElement[])[index].value)
        }
        if (x && !y) {
            return ((a as CellElement[])[index].value).localeCompare(b[index] as string)
        }
        return (a[index] as string).localeCompare((b[index] as string))
    })
 */

export const generateStatusText = (assignment: Assignment, submission: Submission) => {
        if (!assignment.getAutoapprove() && submission.getScore() >= assignment.getScorelimit()) {
            return "Awating approval"
        }
        if (submission.getScore() < assignment.getScorelimit() && submission.getStatus() !== Submission.Status.APPROVED) {
            return `Need ${assignment.getScorelimit()}% score for approval`
        }
        return SubmissionStatus[submission.getStatus()]
    }