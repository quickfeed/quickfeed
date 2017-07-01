import * as React from "react";

interface IButtonProps {
    className?: string;
    text?: string;
    type?: string;
    onClick?: () => void;
}
class Button extends React.Component<IButtonProps, undefined> {
    public render() {
        return (
            <button className={this.props.className ? "btn " + this.props.className : "btn"}
                    type={this.props.type ? this.props.type : ""}
                    onClick={() => this.handleOnclick()}
            >
                {this.props.text ? this.props.text : ""}
            </button>
        );
    }

    private handleOnclick(): void {
        if (this.props.onClick) {
            this.props.onClick();
        }
    }
}
export {Button};
