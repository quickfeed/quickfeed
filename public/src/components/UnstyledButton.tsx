import React from "react";

const UnstyledButton = (props: React.PropsWithChildren<{ onClick: () => void }>) => {
    return (
        <button type="button" style={{color: "black"}} onClick={props.onClick} className="btn btn-link p-0">
            {props.children}
        </button>
    )
}

export default UnstyledButton