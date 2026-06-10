import React from "react"

interface AvatarProps {
    src: string
    alt: string
    /** Tailwind width class, e.g. "w-10" or "w-32". Defaults to "w-10". */
    size?: string
    /**
     * "ring" (default): prominent ring style for profile/group display.
     * "inline": subtle border style for dense table rows.
     */
    variant?: "ring" | "inline"
}

const Avatar = ({ src, alt, size = "w-10", variant = "ring" }: AvatarProps) => {
    if (variant === "inline") {
        return <img src={src} alt={alt} className={`${size} rounded-full border border-base-300`} />
    }
    return (
        <div className="avatar">
            <div className={`${size} rounded-full ring ring-primary ring-offset-base-100 ring-offset-2`}>
                <img src={src} alt={alt} />
            </div>
        </div>
    )
}

export default Avatar
