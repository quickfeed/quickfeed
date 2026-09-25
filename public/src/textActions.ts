/**
 * copyText puts text on the clipboard, reporting whether it got there.
 * navigator.clipboard is undefined outside a secure context, and writeText
 * rejects when the browser denies access; a Copy button that quietly does
 * nothing in either case looks like one that worked, so callers are handed the
 * outcome to say something about.
 */
export const copyText = async (text: string): Promise<boolean> => {
    try {
        await navigator.clipboard.writeText(text)
        return true
    } catch {
        return false
    }
}

/** downloadText offers text to the user as filename.txt. */
export const downloadText = (text: string, filename: string) => {
    const url = URL.createObjectURL(new Blob([text], { type: "text/plain" }))
    const link = document.createElement("a")
    link.href = url
    link.download = `${filename}.txt`
    link.click()
    // Revoking the URL before the browser has read it cancels the download the
    // click just started, so leave that to the next tick.
    setTimeout(() => URL.revokeObjectURL(url), 0)
}
