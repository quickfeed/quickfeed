import React, { ReactNode } from "react"

interface FeatureBlockProps {
    heading: string
    subheading: string
    content: string
    imageSrc: string
    reverse?: boolean
}

/**
* FeatureBlock is a component that displays a feature with an image and text.
* @param heading The main heading of the feature block.
* @param subheading The subheading of the feature block.
* @param content The content of the feature block.
* @param imageSrc The source URL of the image to be displayed.
* @param reverse A boolean indicating whether to reverse the order of the image and text.
*/

const FeatureBlock: React.FC<FeatureBlockProps> = ({ heading, subheading, content, imageSrc, reverse = false }) => {
    return (
        <div className={`flex flex-col ${reverse ? 'md:flex-row-reverse' : 'md:flex-row'} items-center gap-12 my-12`}>
            <div className="flex-1 space-y-4">
                <h2 className="text-3xl font-bold text-base-content">
                    {heading}
                </h2>
                <h3 className="text-xl text-base-content/60 font-medium">
                    {subheading}
                </h3>
                <p className="text-base leading-loose text-base-content/80">
                    {content}
                </p>
            </div>
            <div className="flex-1">
                <img
                    src={imageSrc}
                    alt={heading}
                    className="w-full h-auto rounded-lg shadow-xl"
                />
            </div>
        </div>
    )
}

interface MiniFeatureBlockProps {
    title: string
    content: string
    media: ReactNode
}

export const MiniFeatureBlock: React.FC<MiniFeatureBlockProps> = ({ title, content, media }) => {
    return (
        <div className="card bg-base-200 shadow-xl p-6 text-center hover:shadow-2xl transition-shadow">
            <div className="flex justify-center items-center mb-6 h-40">
                {media}
            </div>
            <h4 className="text-xl font-semibold mb-4 text-base-content">{title}</h4>
            <p className="text-base leading-relaxed text-base-content/70">{content}</p>
        </div>
    )
}

export default FeatureBlock
