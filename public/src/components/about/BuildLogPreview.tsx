import LogOutput from "../LogOutput"
import { previewBuildLog } from "./previewData"

/** BuildLogPreview shows the build log QuickFeed produces for a student push. */
const BuildLogPreview = () => (
    <LogOutput>{previewBuildLog}</LogOutput>
)

export default BuildLogPreview
