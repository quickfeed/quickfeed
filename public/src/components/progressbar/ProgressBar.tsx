import * as React from "react";

interface IProgressBarProps {
    progress: number;
}

export class ProgressBar extends React.Component<IProgressBarProps> {

    public render() {
        const progressBarStyle = {
            width: this.props.progress + "%",
        };

        return (
            <div className="progress">
                <div className="progress-bar" role="progressbar" aria-valuenow={this.props.progress} aria-valuemin={0}
                    aria-valuemax={100} style={progressBarStyle}>{this.props.progress}%
                </div>
            </div>
        );
    }
}
