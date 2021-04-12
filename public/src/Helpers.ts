/* eslint-disable quotes */
import { Assignment, Submission } from './proto/ag_pb'

export interface IBuildInfo {
    builddate: string;
    buildid: number;
    buildlog: string;
    execTime: number
}

export const getBuildInfo = (buildString: string) => {
    let buildinfo: IBuildInfo
    buildinfo = JSON.parse(buildString)
    return buildinfo
    
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
    const parsedScoreObjects = JSON.parse(scoreString)
    for (const scoreObject in parsedScoreObjects) {
        scoreObjects.push(parsedScoreObjects[scoreObject])
    }
    return scoreObjects
    
}


/** Returns a string with a prettier format for a deadline */
export const getFormattedDeadline = (deadline_string: string) => {
    const months = ['January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December']
    let deadline = new Date(deadline_string)
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} by ${deadline.getHours()}:${deadline.getMinutes() < 10 ? '0' + deadline.getMinutes() : deadline.getMinutes()}`
}

export const formatBuildInfo = (buildInfo: string) => {
    console.log(buildInfo.split('/\n/'))
}

/** Utility function for LangindpageTable functionality. To format the output string and class/css based on how far the deadline is in the future */
export const timeFormatter = (deadline:number , now: Date) => {
    console.log(deadline)
    const timeToDeadline = deadline - now.getTime()
    let days = Math.floor(timeToDeadline / (1000 * 3600 * 24))
    let hours = Math.floor(timeToDeadline / (1000 * 3600))
    let minutes = Math.floor(timeToDeadline)
    
    if (days<14){
        if(days<7){
            if (days<3){
                if (timeToDeadline<0){
                    return [true,'table-danger', `deadline was ${-days > 0 ? -days+" days" : -hours+" hours"} ago`]
                }
                if (days==0){
                    return [true,'table-warning', `${hours} hours to deadline!`]
                }

                return [true,'table-warning', `${days} days to deadline`]
            }
        }
        return[true,'table-primary',`${days} days until deadline`]
    }
    return [false]
}
export const layoutTime = "2021-03-20T23:59:00"
// Used for displaying enrollment status
export const EnrollmentStatus = {
    0 : "NONE",
    1 : "PENDING",
    2 : "STUDENT",
    3 : "TEACHER",
}
export const EnrollmentStatusColors = {
    1: "",
    2: "",
    3: "",
}