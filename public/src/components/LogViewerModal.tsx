import { useEffect } from "react"
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
const LogViewerModal = ({ title, text, filename, onClose }: LogViewerModalProps) => {
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") { onClose() }
        }
        window.addEventListener("keydown", handleKeyDown)
        return () => window.removeEventListener("keydown", handleKeyDown)
    }, [onClose])

    const handleCopy = () => {
        // navigator.clipboard is undefined outside a secure context; there is
        // nowhere to report that here, and the text is on screen to select.
        void navigator.clipboard?.writeText(text)
    }

    const handleDownload = () => {
        const url = URL.createObjectURL(new Blob([text], { type: "text/plain" }))
        const link = document.createElement("a")
        link.href = url
        link.download = `${filename}.txt`
        link.click()
        // Revoking the URL before the browser has read it cancels the download
        // the click just started, so leave that to the next tick.
        setTimeout(() => URL.revokeObjectURL(url), 0)
    }

    return (
        // The title names the dialog for a screen reader, which otherwise
        // announces it as an unnamed one.
        <div className="modal modal-open" role="dialog" aria-modal="true" aria-label={title}>
            <div className="modal-box max-w-5xl w-full max-h-[90vh] p-0 overflow-hidden flex flex-col">
                <LogCard
                    title={title}
                    className="flex-1 min-h-0 shadow-none rounded-none"
                    controls={
                        <div className="flex items-center gap-2">
                            <button type="button" className="btn btn-sm" onClick={handleCopy}>Copy</button>
                            <button type="button" className="btn btn-sm" onClick={handleDownload}>Download</button>
                            <button type="button" className="btn btn-sm btn-ghost" aria-label="Close" onClick={onClose}>✕</button>
                        </div>
                    }
                >
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
