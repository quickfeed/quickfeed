import * as React from "react";

interface ISearchProp {
    className?: string;
    placeholder?: string;
    addonBefore?: JSX.Element;
    onChange?: (val: string) => void;
}
interface ISearchState {
    query: string;
}
class Search extends React.Component<ISearchProp, ISearchState> {

    constructor(props: any) {
        super(props);
        this.state = {
            query: "",
        };
    }

    public render() {
        let addOn: JSX.Element | null = null;
        if (this.props.addonBefore) {
            addOn = this.props.addonBefore;
        }
        return (
            <div className={this.props.className ? this.props.className : ""} >
                {addOn}
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
export { Search };
