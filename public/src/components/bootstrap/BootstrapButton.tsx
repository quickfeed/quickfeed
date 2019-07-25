import * as React from "react";

interface IButtonProps {
    classType?: BootstrapClass;
    className?: string;
    type?: string;
    disabled?: boolean;
    onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
}

export type BootstrapClass = "default" | "primary" | "success" | "info" | "warning" | "danger" | "link";
export class BootstrapButton extends React.Component<IButtonProps> {
    public render() {
        const type: BootstrapClass = this.props.classType ? this.props.classType : "default";
        let className = "btn btn-" + type;

        if (this.props.className) {
            className += " " + this.props.className;
        }

        return (
            <button className={className}
                onClick={(e) => this.handleOnclick(e)}
                disabled={this.props.disabled}
            >
                {this.props.children}
            </button>
        );
    }

    private handleOnclick(e: React.MouseEvent<HTMLButtonElement>): void {
        if (this.props.onClick) {
            this.props.onClick(e);
        }
    }
}
