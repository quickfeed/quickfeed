import * as React from "react";

interface ISearchProps {
    className?: string;
    placeholder?: string;
    onChange?: (val: string) => void;
}
interface ISearchState {
    query: string;
}
export class Search extends React.Component<ISearchProps, ISearchState> {

    constructor(props: any) {
        super(props);
        this.state = {
            query: "",
        };
    }

    public render() {
        return (
            <div className={this.props.className ? this.props.className : ""} >
                <span className="input-group-addon">
                    <i className="glyphicon glyphicon-search"></i>
                </span>
                {this.props.children}
                <input
                    className="form-control"
                    type="text"
                    placeholder={this.props.placeholder ? this.props.placeholder : ""}
                    onChange={(e) => this.onChange(e)}
                    value={this.state.query}
                />
            </div>
        );
    }

    private onChange(e: any) {
        this.setState({
            query: e.target.value,
        });
        if (this.props.onChange) {
            this.props.onChange(e.target.value);
        }
    }
}
