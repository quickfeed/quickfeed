import type { ReactNode } from "react"

type LogOutputProps = {
    title?: string
    controls?: ReactNode
    codeClassName?: string
    children: ReactNode
}

/** LogOutput is the card shell shared by the submission build log and the course log:
 *  a header with a terminal icon, a title, and an optional right-hand slot for controls,
 *  followed by a monospaced body. The body scrolls within the card rather than growing
 *  the page, so that a long log leaves the caller's own controls reachable. */
const LogOutput = ({ title = "Build Log", controls, codeClassName = "text-error", children }: LogOutputProps) => {
    return (
        <div className="card bg-base-200 shadow-xl rounded-2xl overflow-hidden">
            <div className="card-body p-0">
                <div className="flex items-center justify-between bg-base-300 px-4 py-3 border-b border-base-content/10">
                    <h3 className="text-sm font-semibold flex items-center gap-2">
                        <i className="fas fa-terminal" />
                        <span>{title}</span>
                    </h3>
                    {controls}
                </div>
                <div className="max-h-[70vh] overflow-auto">
                    <pre className="p-4 text-sm leading-relaxed font-mono bg-base-200 m-0">
                        <code className={codeClassName} style={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
                            {children}
                        </code>
                    </pre>
                </div>
            </div>
        </div>
    )
}

export default LogOutput
