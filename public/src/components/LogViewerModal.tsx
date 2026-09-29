import { useEffect, useRef, useState } from "react"
import { copyText, downloadText } from "../textActions"
import LogCard from "./LogCard"

interface LogViewerModalProps {
    /** What the text is, e.g. the name of the column it came from */
    title: string
    /** The text to show, as it was logged */
    text: string
    /** Basename for the downloaded file, without an extension */
    filename: string
    /** Called when the user requests the modal to close (Esc, ✕, or backdrop) */
    onClose: () => void
}

/**
 * Overlay showing one log value in full. A build log or a test run's output is
 * too tall and too wide for a table cell, so the cell shows its first line and
 * defers the rest to this. Closes on Esc, the ✕ button, or a backdrop click.
 */
// FOCUSABLE matches the controls a dialog can hand focus to. The overlay's own
// are all buttons, but the value it shows is scrollable and so may be in the
// tab order too, and a selector is what keeps the two ends of the loop honest.
const FOCUSABLE = 'a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex="-1"])'

const LogViewerModal = ({ title, text, filename, onClose }: LogViewerModalProps) => {
    const [notice, setNotice] = useState<string | null>(null)
    const box = useRef<HTMLDivElement>(null)

    // The overlay covers the table it was opened from, so focus has to come
    // with it: left on the Show button behind an aria-modal, a keyboard or
    // screen-reader user tabs through a table they can no longer see. Focus
    // goes back where it came from on close, since that is the row they were
    // reading.
    useEffect(() => {
        const opener = document.activeElement as HTMLElement | null
        box.current?.querySelector<HTMLElement>(FOCUSABLE)?.focus()
        return () => opener?.focus()
    }, [])

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                onClose()
                return
            }
            if (e.key !== "Tab" || !box.current) {
                return
            }
            // Tab off either end wraps to the other, so focus cannot leave the
            // dialog for the obscured page behind it.
            const focusable = Array.from(box.current.querySelectorAll<HTMLElement>(FOCUSABLE))
            const [first] = focusable
            const last = focusable[focusable.length - 1]
            if (!first) {
                return
            }
            if (e.shiftKey && document.activeElement === first) {
                e.preventDefault()
                last.focus()
            } else if (!e.shiftKey && document.activeElement === last) {
                e.preventDefault()
                first.focus()
            }
        }
        window.addEventListener("keydown", handleKeyDown)
        return () => window.removeEventListener("keydown", handleKeyDown)
    }, [onClose])

    // A Copy that failed has to say so: the text is on screen to select by
    // hand, but only if the reader knows the button did not do it for them.
    const handleCopy = async () => {
        setNotice(await copyText(text) ? null : "Could not copy; the browser denied access to the clipboard")
    }

    return (
        // The title names the dialog for a screen reader, which otherwise
        // announces it as an unnamed one.
        <div className="modal modal-open" role="dialog" aria-modal="true" aria-label={title}>
            <div ref={box} className="modal-box max-w-5xl w-full max-h-[90vh] p-0 overflow-hidden flex flex-col">
                <LogCard
                    title={title}
                    className="flex-1 min-h-0 shadow-none rounded-none"
                    controls={
                        <div className="flex items-center gap-2">
                            <button type="button" className="btn btn-sm" onClick={() => void handleCopy()}>Copy</button>
                            <button type="button" className="btn btn-sm" onClick={() => downloadText(text, filename)}>Download</button>
                            <button type="button" className="btn btn-sm btn-ghost" aria-label="Close" onClick={onClose}>✕</button>
                        </div>
                    }
                >
                    {notice && <div className="alert alert-error rounded-none shrink-0"><span>{notice}</span></div>}
                    <div className="flex-1 min-h-0 overflow-auto">
                        <pre className="p-4 text-sm leading-relaxed font-mono bg-base-200 m-0">
                            <code className="whitespace-pre-wrap break-words">{text}</code>
                        </pre>
                    </div>
                </LogCard>
            </div>
            <div className="modal-backdrop" onClick={onClose} />
        </div>
    )
}

export default LogViewerModal
