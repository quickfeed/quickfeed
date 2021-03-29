
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
    const months = ["January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December"];
    let deadline = new Date(deadline_string)
    return `${deadline.getDate()} ${months[deadline.getMonth()]} ${deadline.getFullYear()} by ${deadline.getHours()}:${deadline.getMinutes() < 10 ? "0" + deadline.getMinutes() : deadline.getMinutes()}`
}

export const formatBuildInfo = (buildInfo: string) => {
    console.log(buildInfo.split("/\n/"))
}