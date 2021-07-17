/* eslint-disable quotes */

import { useParams } from "react-router"
import { EnrollmentLink, User } from "../proto/ag/ag_pb"

export interface IBuildInfo {
    builddate: string;
    buildid: number;
    buildlog: string;
    execTime: number
}

export const getBuildInfo = (buildString: string) => {
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

export const getScoreObjects = (scoreString: string) => {
    let scoreObjects: IScoreObjects[] = []
    if (scoreString.length > 0) {
        const parsedScoreObjects = JSON.parse(scoreString)
        for (const scoreObject in parsedScoreObjects) {
            scoreObjects.push(parsedScoreObjects[scoreObject])
        }
    }
    return scoreObjects
    
}


/** Returns a string with a prettier format for a deadline */
export const getFormattedTime = (deadline_string: string) => {
    const months = ['January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December']
    let deadline = new Date(deadline_string)
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} at ${deadline.getHours()}:${deadline.getMinutes() < 10 ? '0' + deadline.getMinutes() : deadline.getMinutes()}`
}

export const formatBuildInfo = (buildInfo: string) => {
    console.log(buildInfo.split('/\n/'))
}

/** Utility function for LandingpageTable functionality. To format the output string and class/css based on how far the deadline is in the future */
export const timeFormatter = (deadline:string , now: Date) => {
    const timeOfDeadline = new Date(deadline)
    const timeToDeadline =  timeOfDeadline.getTime() - now.getTime()
    let days = Math.floor(timeToDeadline / (1000 * 3600 * 24))
    let hours = Math.floor(timeToDeadline / (1000 * 3600))
    let minutes = Math.floor((timeToDeadline % (1000 * 3600)) / (1000*60))
    
    if (days<14){
        if(days<7){
            if (days<3){
                if (timeToDeadline<0){
                    return [true,'table-danger', `deadline was ${-days > 0 ? -days+" days" : -hours+" hours"} ago`,0]
                }
                if (days==0){
                    return [true,'table-danger', `${hours} hours and ${minutes} minutes to deadline!`,0]
                }

                return [true,'table-warning', `${days} day${days==1?'':'s'} to deadline`,days]
            }
        }
        return[true,'table-primary',`${days} days until deadline`,days]
    }
    return [false,'','',days]
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


export const isValid = (element: any) => {
    if (element instanceof User){
        if (element.getName().length === 0 || element.getEmail().length === 0 || element.getStudentid().length === 0) {
            return false
        }
    }
    if (element instanceof EnrollmentLink) {
        console.log(element.getSubmissionsList())
        if (!element.getEnrollment() && !element.getEnrollment()?.getUser() && element.getSubmissionsList().length === 0) {
            return false
        }
    }
    return true
}


export const getCourseID = () => {
    const route = useParams<{id?: string}>()
    return Number(route.id)
}
