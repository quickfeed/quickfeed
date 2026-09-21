import type { ReactNode } from "react"
import React from "react"

interface FeatureBlockProps {
    heading: string
    subheading: string
    content: string
    points?: string[]
    media: ReactNode
    reverse?: boolean
}

/**
* FeatureBlock is a component that displays a feature with an illustration and text.
* @param heading The main heading of the feature block.
* @param subheading The eyebrow label shown above the heading.
* @param content The lead paragraph of the feature block.
* @param points Short supporting statements, listed below the lead paragraph.
* @param media The illustration to display beside the text, typically a
* PreviewFrame wrapping a preview of the QuickFeed interface.
* @param reverse A boolean indicating whether to reverse the order of the media and text.
*/

const FeatureBlock: React.FC<FeatureBlockProps> = ({ heading, subheading, content, points, media, reverse = false }) => {
    return (
        <div className={`flex flex-col ${reverse ? 'md:flex-row-reverse' : 'md:flex-row'} items-center gap-12 my-12`}>
            <div className="flex-1 space-y-4">
                <p className="text-sm font-semibold uppercase tracking-wider text-primary">
                    {subheading}
                </p>
                <h3 className="text-3xl font-bold text-base-content">
                    {heading}
                </h3>
                <p className="text-lg leading-relaxed text-base-content/80">
                    {content}
                </p>
                {points ? <FeaturePoints points={points} /> : null}
            </div>
            <div className="flex-1 min-w-0">
                {media}
            </div>
        </div>
    )
}

const FeaturePoints = ({ points }: { points: string[] }) => (
    <ul className="space-y-2 pt-1">
        {points.map(point => (
            <li key={point} className="flex items-start gap-3 text-base-content/70">
                <i className="fas fa-circle-check mt-1 text-success shrink-0" aria-hidden="true" />
                <span>{point}</span>
            </li>
        ))}
    </ul>
)

interface MiniFeatureBlockProps {
    title: string
    content: string
    media: ReactNode
}

export const MiniFeatureBlock: React.FC<MiniFeatureBlockProps> = ({ title, content, media }) => {
    return (
        <div className="card bg-base-200 border border-base-content/10 shadow-xl p-6 text-center transition-all hover:shadow-2xl hover:-translate-y-1">
            <div className="flex justify-center items-center mb-6 h-40">
                {media}
            </div>
            <h3 className="text-xl font-semibold mb-4 text-base-content">{title}</h3>
            <p className="text-base leading-relaxed text-base-content/70">{content}</p>
        </div>
    )
}

export default FeatureBlock
