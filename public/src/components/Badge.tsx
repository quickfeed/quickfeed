import React from "react";
import { Enrollment_UserStatus } from "../../proto/qf/types_pb";

type Color = 'red' | 'yellow' | 'green' | 'blue' | 'gray' | 'cyan';

/* Note that these can be moved to the safelist in tailwind.config.js */
/* And loaded dynamically based on props.color */
/* Or defined in the helpers.ts file */
const colorClasses: Record<Color, string> = {
    red: 'bg-red-700 ring-red-950 text-white',
    yellow: 'bg-yellow-400 ring-yellow-500/60 text-black',
    green: 'bg-green-400 ring-green-600/60 text-black',
    blue: 'bg-blue-500 ring-blue-800/60 text-white',
    cyan: 'bg-cyan-700 ring-cyan-800/60 text-white',
    gray: 'bg-zinc-500 ring-zinc-800/60 text-white',
};

const RoleColor: Record<Enrollment_UserStatus, Color | null> = {
    [Enrollment_UserStatus.NONE]: null,
    [Enrollment_UserStatus.PENDING]: 'cyan',
    [Enrollment_UserStatus.STUDENT]: 'blue',
    [Enrollment_UserStatus.TEACHER]: 'red',
}

const Badge = (props: { color: Color | Enrollment_UserStatus; text: string; className?: string }) => {
    let color: Color | null;
    if (typeof props.color === 'number') {
        color = RoleColor[props.color];
        if (!color) return null; // Don't render NONE status
    } else {
        color = props.color;
    }
    return (
        <span className={`${props.className ?? ""} ${colorClasses[color]} inline-flex items-center rounded-md ml-2 px-2 py-1 text-xs font-medium ring-1 ring-inset`}>
            {props.text}
        </span>
    )
}

export default Badge;
