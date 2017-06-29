import * as React from "react";

function Row(props: { children: any, className?: string }) {
    return <div className={props.className ? "row " + props.className : "row"}>
        {props.children}
    </div>;
}

export {Row};
