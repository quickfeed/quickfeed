import * as React from "react";

interface IProgressBarProps {
    progress: number;
    scoreToPass: number;
}

export class ProgressBar extends React.Component<IProgressBarProps> {

    public render() {
        const secondaryBarWidth = this.props.scoreToPass - this.props.progress;
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
            {this.generateSecondaryBarText(secondaryBarWidth)}</div>;

        return (
            <div className="progress">
                <div className={barClass} role="progressbar" style={progressBarStyle}>{this.generateMainBarText(secondaryBarWidth)}
                </div>
                {secondaryBarWidth > 0 ? secondaryBar : null }
            </div>
        );
    }

    private generateMainBarText(delta: number): string {
        let mainText = this.props.progress + " % completed";
        if (delta < 12 && delta > 0) {
            mainText += " / " + delta + " % to go"
        }
        return this.props.progress >= 10 ? mainText : "";
    }

    generateSecondaryBarText(delta: number): string {
        let secondaryText = delta + " % to go";
        if (this.props.progress < 10) {
            const mainPart = this.props.progress > 0 ? this.props.progress + " % completed" + " / " : "";
            secondaryText = mainPart + secondaryText;
        }
        return delta >= 12 ? secondaryText : "";
    }
}
