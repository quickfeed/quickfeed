import React from "react"
import { Enrollment_UserStatus } from "../../proto/qf/types_pb"

type Color = 'red' | 'yellow' | 'green' | 'blue' | 'gray' | 'cyan'

/* Note that these can be moved to the safelist in tailwind.config.js */
/* And loaded dynamically based on props.color */
/* Or defined in the helpers.ts file */
const colorClasses: Record<Color, string> = {
    red: 'badge-error',
    yellow: 'badge-warning',
    green: 'badge-success',
    blue: 'badge-primary',
    cyan: 'badge-info',
    gray: 'badge-neutral',
}

const BadgeStyle = {
    outline: 'badge-outline',
    solid: 'badge-solid',
    ghost: 'badge-ghost',
    dash: 'badge-dash',
}

type BadgeStyleType = keyof typeof BadgeStyle

const RoleColor: Record<Enrollment_UserStatus, Color | null> = {
    [Enrollment_UserStatus.NONE]: null,
    [Enrollment_UserStatus.PENDING]: 'cyan',
    [Enrollment_UserStatus.STUDENT]: 'blue',
    [Enrollment_UserStatus.TEACHER]: 'red',
}

const Badge = (props: { type: BadgeStyleType, color: Color | Enrollment_UserStatus; text: string; className?: string }) => {
    let color: Color | null
    if (typeof props.color === 'number') {
        color = RoleColor[props.color]
        if (!color) return null // Don't render NONE status
    } else {
        color = props.color
    }
    return (
        <span className={`${props.className ?? ""} badge ${BadgeStyle[props.type]} ${colorClasses[color]} `}>
            {props.text}
        </span>
    )
}

export default Badge
