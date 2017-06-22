interface ILink{
    name: string;
    description?: string;
    uri?: string;
    active?: boolean;
}

interface NavHeaderBarProps{
    brandName: string;
    id: string;
}

interface INavBarProps{
    id: string;
    links: ILink[];
    isFluid: boolean;
    isInverse: boolean;
    brandName: string;
    onClick?: (link:ILink) => void;
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
            <a className="navbar-brand" onClick={ (e) => { e.preventDefault(); }} href="#">{this.props.brandName}</a>
        </div>
    }
}

class NavBar extends React.Component<INavBarProps, undefined> {

    private renderIsFluid(){
        let name = "container"
        if (this.props.isFluid){
            name += "-fluid";
        }
        return name;
    }

    private renderNavBarClass(){
        let name = "navbar navbar-absolute-top";
        if (this.props.isInverse){
            name += " navbar-inverse";
        }
        else 
        {
            name += " navbar-default";
        }
        return name;
    }

    private handleClick(link: ILink){
        if (this.props.onClick){
            this.props.onClick(link);
        }
    }

    render(){
        let items = this.props.links.map((v, i) => {
            let active = "";
            if(v.active){
                active = "active";
            }
            return <li className={active} key={i}><a href="#"  onClick={(e) => { e.preventDefault(); this.handleClick(v); }}>{v.name}</a></li>
        });

        return <nav className={this.renderNavBarClass()}>
            <div className={this.renderIsFluid()}>
                <NavHeaderBar id={this.props.id} brandName={this.props.brandName}></NavHeaderBar>

                <div className="collapse navbar-collapse" id={this.props.id}>
                    <ul className="nav navbar-nav">
                        {items}
                    </ul>
                </div>
            </div>
        </nav>
    }
}

interface INavMenuProps{
    links: ILink[];
    onClick?: (link: ILink) => void;
}

interface INavMenuFormatableProps<T>{
    items: T[];
    formater?: (item: T) => string;
    onClick?: (item: T) => void;
}

class NavMenu extends React.Component<INavMenuProps, undefined> {
    render(){
        const items = this.props.links.map((v: ILink, i: number) => {
            let active = "";
            if (v.active){
                active = "active";
            }
            return <li className={active} key={i}><a href="#" onClick={(e) => { e.preventDefault(); if (this.props.onClick) this.props.onClick(v); }}>{v.name}</a></li>
        })
        return <ul className="nav nav-pills nav-stacked">
            {items}
        </ul>
    }
}

class NavMenuFormatable<T> extends React.Component<INavMenuFormatableProps<T>, undefined> {
    renderObj(item: T): string{
        if (this.props.formater){
            return this.props.formater(item);
        }
        return item.toString();
    }

    handleItemClick(item: T): void{
        if (this.props.onClick){
            this.props.onClick(item);
        }
    }

    render(){
        const items = this.props.items.map((v, i) => {
            return <li key={i}><a href="#" onClick={() => { this.handleItemClick(v) }}>{this.renderObj(v)}</a></li>
        })
        return <ul className="nav nav-pills nav-stacked">
            {items}
        </ul>
    }
}

interface IDynamicTableProps<T>{
    header: string[];
    data: T[];
    selector: (item: T) => string[];
}

class DynamicTable<T> extends React.Component<IDynamicTableProps<T>, undefined>{

    renderCells(values: string[]): JSX.Element[]{
        return values.map((v, i) => {
                return <td key={i}>{v}</td>
            });
    }

    renderRow(item: T, i: number): JSX.Element{
        return <tr key={i}>{ this.renderCells(this.props.selector(item)) }</tr>;
    }

    render(){
        let rows = this.props.data.map((v, i) => {
            return this.renderRow(v, i);
        });

        return <table className="table">
            <thead>
                <tr>{this.renderCells(this.props.header)}</tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    }
}

function Row(props: {children: any, className?: string}){
    return <div className={"row " + props.className ? props.className : ""}>
        {props.children}
    </div>
}