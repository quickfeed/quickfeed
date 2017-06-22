import * as React from "react"

function Row(props: {children: any, className?: string}){
    return <div className={"row " + props.className ? props.className : ""}>
        {props.children}
    </div>
}

export {Row}