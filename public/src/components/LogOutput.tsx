import type { ReactNode } from "react"
import LogCard from "./LogCard"

const LogOutput = ({ children }: { children: ReactNode }) => (
    <LogCard title="Build Log">
        <div className="max-h-[70vh] overflow-auto rounded-b-2xl">
            <pre className="p-4 text-sm leading-relaxed font-mono bg-base-200 m-0">
                <code className="text-error whitespace-pre-wrap break-words">{children}</code>
            </pre>
        </div>
    </LogCard>
)

export default LogOutput
