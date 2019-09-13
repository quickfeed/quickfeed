import * as React from "react";

export function Row(props: { children: any, className?: string }) {
    return <div className={props.className ? "row " + props.className : "row"}>
        {props.children}
    </div>;
}
