import { useState } from "react"
import type { Score } from "../../../proto/kit/score/score_pb"

/** containerPath matches the directory a test run records in a source position,
 *  which is the path inside the container, e.g. "/quickfeed/tests/lab1/". */
const containerPath = /(^|\s)\/\S*\/(\S+:\d+:)/g

/** trimContainerPaths shortens the source positions a run records to the file
 *  name the student sees in their editor. The container's directory layout is
 *  noise to them, and it pushes the message itself off the line. */
export const trimContainerPaths = (text: string): string => text.replace(containerPath, "$1$2")

const Pre = ({ text, className = "" }: { text: string; className?: string }) => (
    <pre className="text-sm leading-relaxed font-mono m-0 overflow-auto max-h-96">
        <code className={`whitespace-pre-wrap break-words ${className}`}>{text}</code>
    </pre>
)

/** TestOutputPanel shows what one test produced. Why it failed comes first,
 *  since that is what a student opens a submission to find, and the rest of the
 *  test's output follows behind a disclosure so that it cannot bury the
 *  failure. A test that failed silently has only its output, which is then
 *  shown directly rather than behind a second click.
 */
const TestOutputPanel = ({ score }: { score: Score }) => {
    const [showOutput, setShowOutput] = useState(false)

    const details = trimContainerPaths(score.TestDetails.trimEnd())
    const output = trimContainerPaths(score.TestOutput.trimEnd())
    if (!details && !output) {
        return <p className="text-sm opacity-60 m-0">This test recorded no output.</p>
    }

    const copy = () => {
        // navigator.clipboard is undefined outside a secure context; there is
        // nowhere to report that here, and the text is on screen to select.
        void navigator.clipboard?.writeText([details, output].filter(Boolean).join("\n\n"))
    }

    return (
        <div className="flex flex-col gap-2">
            <div className="flex items-center justify-between gap-2">
                <h4 className="text-sm font-semibold m-0">{details ? "Why it failed" : "Test output"}</h4>
                <button type="button" className="btn btn-xs" onClick={copy}>Copy</button>
            </div>
            {details ? <Pre text={details} className="text-error" /> : <Pre text={output} />}
            {details && output ? (
                <div>
                    <button
                        type="button"
                        className="btn btn-xs btn-ghost px-1 gap-1"
                        aria-expanded={showOutput}
                        onClick={() => setShowOutput(prev => !prev)}
                    >
                        <i className={`fas fa-chevron-${showOutput ? "down" : "right"} text-xs`} />
                        Test output ({output.split("\n").length} lines)
                    </button>
                    {showOutput ? <Pre text={output} /> : null}
                </div>
            ) : null}
        </div>
    )
}

export default TestOutputPanel
