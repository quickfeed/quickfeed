import * as React from "react";

interface IButtonProps {
    classType?: BootstrapClass;
    className?: string;
    type?: string;
    disabled?: boolean;
    tooltip?: string;
    onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
}

export type BootstrapClass = "default" | "primary" | "success" | "info" | "warning" | "danger" | "link";
export class BootstrapButton extends React.Component<IButtonProps> {
    public render() {
        const classType = this.props.classType ? this.props.classType : "default" + this.props.type;
        let className = "btn btn-" + classType;

        if (this.props.className) {
            className += " " + this.props.className;
        }

        if (this.props.tooltip) {
            return (<button className={className}
            onClick={(e) => this.handleOnclick(e)}
            data-toggle={"tooltip"}
            data-html={"true"}
            title={this.props.tooltip}
            disabled={this.props.disabled}>
            {this.props.children}
            </button>);
        }
        return (<button className={className}
        onClick={(e) => this.handleOnclick(e)}
        disabled={this.props.disabled}>
        {this.props.children}
        </button>);

    }

    private handleOnclick(e: React.MouseEvent<HTMLButtonElement>): void {
        if (this.props.onClick) {
            this.props.onClick(e);
        }
    }
}
