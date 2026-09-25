import type { ReactNode } from "react"
import { useState } from "react"
import LogCard from "./LogCard"

/** LogOutput shows what the run produced outside any test: the run script's own
 *  output, the compilation phase, and whatever followed the tests. Each test's
 *  own output is shown with that test, so this is the rest of the run, and it
 *  starts closed unless the run produced nothing else to read.
 */
const LogOutput = ({ children, defaultOpen = false }: { children: ReactNode; defaultOpen?: boolean }) => {
    const [open, setOpen] = useState(defaultOpen)
    return (
        <LogCard
            title="Build log"
            controls={
                <button
                    type="button"
                    className="btn btn-sm btn-ghost gap-1"
                    aria-expanded={open}
                    onClick={() => setOpen(prev => !prev)}
                >
                    <i className={`fas fa-chevron-${open ? "down" : "right"} text-xs`} />
                    {open ? "Hide" : "Show"}
                </button>
            }
        >
            {open ? (
                <div className="max-h-[70vh] overflow-auto rounded-b-2xl">
                    <pre className="p-4 text-sm leading-relaxed font-mono bg-base-200 m-0">
                        <code className="text-error whitespace-pre-wrap break-words">{children}</code>
                    </pre>
                </div>
            ) : null}
        </LogCard>
    )
}

export default LogOutput
