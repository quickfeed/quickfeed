import type { ReactNode } from "react"

interface LogCardProps {
    title: string
    controls?: ReactNode
    className?: string
    children: ReactNode
}

const LogCard = ({ title, controls, className = "", children }: LogCardProps) => (
    <div className={`card bg-base-200 shadow-xl rounded-2xl ${className}`}>
        <div className="relative z-20 flex flex-wrap shrink-0 items-center justify-between gap-2 bg-base-300 px-4 py-3 rounded-t-2xl border-b border-base-content/10">
            <h3 className="text-sm font-semibold flex items-center gap-2">
                <i className="fas fa-terminal" />
                <span>{title}</span>
            </h3>
            {controls}
        </div>
        {children}
    </div>
)

export default LogCard
