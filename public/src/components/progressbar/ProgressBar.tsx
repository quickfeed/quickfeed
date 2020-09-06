import * as React from "react";

interface IProgressBarProps {
    progress: number;
    scoreToPass: number;
}

export class ProgressBar extends React.Component<IProgressBarProps> {

    public render() {
        const secondaryBarWidth = this.props.scoreToPass - this.props.progress;

        console.log("Secondary bar length is " + secondaryBarWidth);

        const progressBarStyle = {
            width: this.props.progress + "%",
        };
        const secondaryBarStyle = {
            width: secondaryBarWidth + "%",
        };
        let barClass = "progress-bar";
        if (secondaryBarWidth <= 0) {
            console.log("Primary bar: success!");
            barClass += " progress-bar-success";
        }

        const secondaryBar = <div className="progress-bar progressbar-secondary bg-secondary" role="progressbar" style={secondaryBarStyle}>
            {secondaryBarWidth}% to go</div>

        return (
            <div className="progress">
                <div className={barClass} role="progressbar" style={progressBarStyle}>{this.props.progress}% done
                </div>
                {secondaryBarWidth > 0 ? secondaryBar : null }
            </div>
        );
    }
}
