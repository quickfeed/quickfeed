import React, { useState, ReactNode } from "react"

interface CollapsibleProps {
    title: string
    children: ReactNode
    defaultOpen?: boolean
    className?: string
}

const Collapsible: React.FC<CollapsibleProps> = ({ title, children, defaultOpen = false, className }) => {
    const [open, setOpen] = useState(defaultOpen)
    return (
        <div className={`${className || ""}`.trim()}>
            <div
                className="d-flex align-items-center"
                onClick={() => setOpen(!open)}
                aria-expanded={open}
                tabIndex={0}
                onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { setOpen(!open) } }}
                role="button"
            >
                <h2 className="mb-0 flex-grow-1">{title}</h2>
                <span className="text-primary">
                    {open ? 'Hide' : 'Show'}
                </span>
            </div>
            {open && (
                <div className="mt-3">
                    {children}
                </div>
            )}
        </div>
    )
}

export default Collapsible
