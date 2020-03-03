import * as React from "react";

export function showLoader(): JSX.Element {
    return (<div className="load-text"><div className="lds-ripple"><div></div><div></div></div></div>);
}