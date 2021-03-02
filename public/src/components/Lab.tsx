import { Submission } from "../proto/ag_pb"




const Lab = (submission : Submission) => {
    return (
        <div>
            <h1>{submission.getScore()} / 100</h1>
            <code>{submission.getBuildinfo()}</code>
        </div>
    )
}

export default Lab