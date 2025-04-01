import React, { JSX } from "react"

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

const FeatureBlock: React.FC<FeatureBlockProps> = ({
    heading,
    subheading,
    content,
    imageSrc,
    reverse = false,
}) => {
    return (
        <div className={`row featurette ${reverse ? "flex-row-reverse" : ""}`}>
            <div className="col-md-7">
                <h2 className="featurette-heading">
                    {heading}: <span className="text-muted">{subheading}</span>
                </h2>
                <p className="lead">{content}</p>
            </div>
            <div className="col-md-5">
                <img
                    className="featurette-image img-responsive about"
                    src={imageSrc}
                    alt="Feature example"
                />
            </div>
        </div>
    )
}

interface MiniFeatureBlockProps {
    title: string
    content: string
    imgSrc?: string
    icon?: JSX.Element
    style?: React.CSSProperties
}

/**
 * MiniFeatureBlock is a smaller version of FeatureBlock, used for displaying features in a grid layout.
 * @param title The title of the feature block.
 * @param content The content of the feature block.
 * @param imageSrc (Optional) The source URL of the image to be displayed.
 * @param icon (Optional) An optional icon to be displayed instead of an image.
 * @param style (Optional) Additional styles to be applied to the parent div of the icon.
*/
export const MiniFeatureBlock: React.FC<MiniFeatureBlockProps> = ({
    title,
    content,
    imgSrc,
    icon,
    style,
}) => {
    let iconOrImg
    if (icon) {
        iconOrImg =
            <div style={{
                width: "140px",
                height: "140px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                margin: "0 auto",
                ...style
            }}>
                {icon}
            </div>
    }
    else if (imgSrc) {
        iconOrImg =
            <div style={{
                width: "140px",
                height: "140px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                margin: "0 auto",
                ...style
            }}>
                <img
                    src={imgSrc}
                    alt={title}
                    style={{ width: "100%", height: "100%", objectFit: "contain" }}
                />
            </div>
    }
    return (
        <div className="col-lg-4 text-center">
            {iconOrImg}
            <h2>{title}</h2>
            <p>{content}</p>
        </div>
    )
}

export default FeatureBlock
