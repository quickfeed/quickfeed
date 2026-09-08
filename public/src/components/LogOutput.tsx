import { useEffect, useRef, useState, type ReactNode } from "react"

type LogOutputProps = {
    title?: string
    controls?: ReactNode
    codeClassName?: string
    /** fill grows the scrollable body down to the bottom of the viewport instead
     *  of capping it at a fixed 70vh, for a page (like the course log) whose card
     *  is the last thing on the page and has nothing below it competing for space. */
    fill?: boolean
    /** variant selects how children are wrapped: "code" (default) renders them inside
     *  a monospaced <pre><code> block, for a plain-text log; "table" renders them as-is
     *  inside a horizontally scrollable container, for a column-oriented view. */
    variant?: "code" | "table"
    children: ReactNode
}

/** LogOutput is the card shell shared by the submission build log and the course log:
 *  a header with a terminal icon, a title, and an optional right-hand slot for controls,
 *  followed by a monospaced body. The body scrolls within the card rather than growing
 *  the page, so that a long log leaves the caller's own controls reachable. */
const LogOutput = ({ title = "Build Log", controls, codeClassName = "text-error", fill = false, variant = "code", children }: LogOutputProps) => {
    const bodyRef = useRef<HTMLDivElement>(null)
    const [fillHeight, setFillHeight] = useState<number>()

    useEffect(() => {
        if (!fill) {
            return
        }
        const el = bodyRef.current
        if (!el) {
            return
        }
        // Recomputed on window resize and on any layout change above this card
        // (filters wrapping to another row, an alert appearing), since either
        // moves the body's own top and would otherwise leave a stale height.
        const update = () => {
            const available = window.innerHeight - el.getBoundingClientRect().top - 16
            setFillHeight(Math.max(available, 200))
        }
        update()
        // ResizeObserver is unavailable in some test environments (e.g. jsdom);
        // window resize alone still keeps the height reasonably in sync there.
        const observer = typeof ResizeObserver === "undefined" ? undefined : new ResizeObserver(update)
        observer?.observe(document.body)
        window.addEventListener("resize", update)
        return () => {
            observer?.disconnect()
            window.removeEventListener("resize", update)
        }
    }, [fill])

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
                <div
                    ref={bodyRef}
                    className={fill ? "overflow-auto" : "max-h-[70vh] overflow-auto"}
                    style={fill ? { maxHeight: fillHeight } : undefined}
                >
                    {variant === "table" ? (
                        <div className="overflow-x-auto">{children}</div>
                    ) : (
                        <pre className="p-4 text-sm leading-relaxed font-mono bg-base-200 m-0">
                            <code className={codeClassName} style={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
                                {children}
                            </code>
                        </pre>
                    )}
                </div>
            </div>
        </div>
    )
}

export default LogOutput
