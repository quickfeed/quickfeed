import { ProgressBarView } from "../ProgressBar"
import SubmissionScores from "../submissions/SubmissionScores"
import { previewAssignment, previewScoreLimit, previewSubmission } from "./previewData"

/**
 * LabResultPreview shows the lab result a student sees after a push, using the
 * same progress bar and score table the course pages render.
 */
const LabResultPreview = () => {
    const score = previewSubmission.score
    const remainingToPass = Math.max(0, previewScoreLimit - score)
    return (
        <div className="space-y-2 text-sm">
            <ProgressBarView
                score={score}
                remainingToPass={remainingToPass}
                color="bg-success"
                text={`${score} %`}
                secondaryText={`${remainingToPass} %`}
            />
            <table className="table table-zebra">
                <thead className="bg-base-300 text-base-content">
                    <tr className="text-lg">
                        <th colSpan={2}>Lab information</th>
                        <th>{previewAssignment}</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td colSpan={2} className="passed pl-3!">Status</td>
                        <td>Approved</td>
                    </tr>
                    <tr>
                        <td colSpan={2}>Delivered</td>
                        <td>20 September 2025 07:07</td>
                    </tr>
                    <tr>
                        <td colSpan={2}>Deadline</td>
                        <td>21 September 2025 23:59</td>
                    </tr>
                    <tr>
                        <td colSpan={2}>Tests Passed</td>
                        <td>2/4</td>
                    </tr>
                    <tr>
                        <td colSpan={2}>Execution time</td>
                        <td>7 seconds</td>
                    </tr>
                    <tr>
                        <td colSpan={2}>Slip days</td>
                        <td>7</td>
                    </tr>
                </tbody>
            </table>
            <SubmissionScores submission={previewSubmission} />
        </div>
    )
}

export default LabResultPreview
