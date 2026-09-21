import type { ReactNode } from "react"

interface PreviewFrameProps {
    /** Describes the preview for assistive technology, in place of an image's alt text. */
    label: string
    /** When set, the preview is framed as a browser window showing this address. */
    url?: string
    children: ReactNode
}

/**
 * PreviewFrame presents a sample of the QuickFeed interface on the About page.
 *
 * The frame is built from daisyUI components, so it follows the active theme
 * along with the preview it wraps, and the About page needs no theme-specific
 * assets. The previews are illustrations rather than working UI, so the inner
 * element is `inert` to keep their dead buttons and sortable headers out of the
 * tab order. Because `inert` also hides its subtree from assistive technology,
 * the description sits on the outer element, which presents the whole preview as
 * a single image the way the screenshots it replaces did.
 */
const PreviewFrame = ({ label, url, children }: PreviewFrameProps) => (
    <div role="img" aria-label={label} className="w-full">
        {url ? (
            <div inert className="mockup-browser bg-base-100 border border-base-300 shadow-xl w-full">
                <div className="mockup-browser-toolbar">
                    <div className="input text-base-content/70">{url}</div>
                </div>
                <div className="border-t border-base-300 overflow-x-auto p-4">
                    {children}
                </div>
            </div>
        ) : (
            <div inert>{children}</div>
        )}
    </div>
)

export default PreviewFrame
