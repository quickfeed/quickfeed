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
        // set bootstrap class
        const type: BootstrapClass = this.props.classType ? this.props.classType : "default";
        // add custom class when provided
        const fullType = this.props.type ? type + " " + this.props.type : type;
        let className = "btn btn-" + fullType;

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
