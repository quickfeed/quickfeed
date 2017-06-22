import * as React from "react";

interface NavHeaderBarProps{
    brandName: string;
    id: string;
    brandClick: () => void;
}

class NavHeaderBar extends React.Component<NavHeaderBarProps, undefined>{
    componentDidMount(){
        console.log(this.refs.button);
        let temp = this.refs.button as HTMLElement;
        temp.setAttribute("data-toggle", "collapse");
        temp.setAttribute("data-target", "#" + this.props.id);
        temp.setAttribute("aria-expanded", "false");        
    }
    
    render(){
        return <div className="navbar-header">
            <button ref="button" type="button" className="navbar-toggle collapsed" >
                <span className="sr-only">Toggle navigation</span>
                <span className="icon-bar"></span>
                <span className="icon-bar"></span>
                <span className="icon-bar"></span>
            </button>
            <a className="navbar-brand" onClick={ (e) => { e.preventDefault(); this.props.brandClick() }} href="/">{this.props.brandName}</a>
        </div>
    }
}

export{ NavHeaderBar, NavHeaderBarProps}