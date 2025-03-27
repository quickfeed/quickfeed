import React from "react";

interface FeatureBlockProps {
    id: string;
    heading: string;
    subheading: string;
    content: string;
    imageSrc: string;
    reverse?: boolean;
}

const FeatureBlock: React.FC<FeatureBlockProps> = ({
    id,
    heading,
    subheading,
    content,
    imageSrc,
    reverse = false,
}) => {
    return (
        <section id={id}>
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
        </section>
    );
};

interface MiniFeatureBlockProps {
    title: string;
    content: string;
    imageSrc?: string;
    icon?: JSX.Element;
}

export const MiniFeatureBlock: React.FC<MiniFeatureBlockProps> = ({
    title,
    content,
    imageSrc,
    icon,
}) => {
    return (
        <div className="col-lg-4 text-center">
            {icon ? (
                icon
            ) : imageSrc ? (
                <img
                    className="img-circle"
                    src={imageSrc}
                    alt={title}
                    style={{ width: "140px", height: "140px" }}
                />
            ) : null}
            <h2>{title}</h2>
            <p>{content}</p>
        </div>
    );
};

export default FeatureBlock;
