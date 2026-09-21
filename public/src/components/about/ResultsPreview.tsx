import { Color } from "../../Helpers"
import Button, { ButtonType } from "../admin/Button"
import type { PreviewStatus } from "./previewData"
import { previewLabs, previewMembers } from "./previewData"

// Mirrors the classes getSubmissionCellColor assigns in the results table;
// an ungraded cell is left on the table ground.
const statusColor: Record<PreviewStatus, string> = {
    approved: "bg-success text-success-content",
    revision: "bg-warning text-warning-content",
    rejected: "bg-error text-error-content",
    none: "",
}

const noop = () => { /* the preview is an illustration, not working UI */ }

/**
 * ResultsPreview shows the teacher's view of a course: the grading actions for
 * the selected submission and the results table across students and labs.
 */
const ResultsPreview = () => (
    <div className="space-y-4 text-sm">
        <div className="flex flex-wrap gap-3">
            <Button text="Approve" color={Color.GREEN} className="flex-1" onClick={noop} />
            <Button text="Revision" color={Color.YELLOW} type={ButtonType.OUTLINE} className="flex-1" onClick={noop} />
            <Button text="Reject" color={Color.RED} type={ButtonType.OUTLINE} className="flex-1" onClick={noop} />
            <Button text="Rebuild" color={Color.BLUE} type={ButtonType.OUTLINE} className="flex-1" onClick={noop} />
        </div>
        <table className="table table-zebra">
            <thead className="bg-base-300">
                <tr>
                    <th>Name</th>
                    {previewLabs.map(lab => (
                        <th key={lab}>{lab}</th>
                    ))}
                </tr>
            </thead>
            <tbody>
                {previewMembers.map(member => (
                    <tr key={member.name}>
                        <th className="font-medium">{member.name}</th>
                        {member.labs.map((lab, index) => (
                            <td key={previewLabs[index]} className={statusColor[lab.status]}>
                                {lab.score} %
                            </td>
                        ))}
                    </tr>
                ))}
            </tbody>
        </table>
    </div>
)

export default ResultsPreview
