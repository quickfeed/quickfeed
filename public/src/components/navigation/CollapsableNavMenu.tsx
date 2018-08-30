import * as React from "react";
import { ILink, ILinkCollection } from "../../managers";
import { NavigationHelper } from "../../NavigationHelper";

interface ICollapsableNavMenuProps {
    links: ILinkCollection[];
    onClick?: (link: ILink) => void;
}

class CollapsableNavMenu extends React.Component<ICollapsableNavMenuProps, {}> {

    private topItems: HTMLElement[] = [];

    public render() {
        const children = this.props.links.map((e, i) => {
            return this.renderTopElement(i, e);
        });

        return <ul className="nav nav-list">
            {children}
        </ul>;
    }

    private toggle(index: number) {
        const animations: Array<(() => void)> = [];
        this.topItems.forEach((temp, i) => {
            if (i === index) {
                if (this.collapseIsOpen(temp)) {
                    // Just stay opend
                } else {
                    animations.push(this.openCollapse(temp));
                }
            } else {
                animations.push(this.closeIfOpen(temp));
            }
        });
        setTimeout(() => {
            animations.forEach((e) => {
                e();
            });
        }, 10);
    }

    private collapseIsOpen(ele: HTMLElement) {
        return ele.classList.contains("in");
    }

    private closeIfOpen(ele: HTMLElement): () => void {
        if (this.collapseIsOpen(ele)) {
            return this.closeCollapse(ele);
        }
        return () => {
            "do nothing";
        };
    }

    private openCollapse(ele: HTMLElement): () => void {
        ele.classList.remove("collapse");
        ele.classList.add("collapsing");
        return () => {
            ele.style.height = ele.scrollHeight + "px";
            setTimeout(() => {
                ele.classList.remove("collapsing");
                ele.classList.add("collapse");
                ele.classList.add("in");
                ele.style.height = null;
            }, 350);
        };
    }

    private closeCollapse(ele: HTMLElement): () => void {
        ele.style.height = ele.clientHeight + "px";
        ele.classList.add("collapsing");
        ele.classList.remove("collapse");
        ele.classList.remove("in");
        return () => {
            ele.style.height = null;
            setTimeout(() => {
                ele.classList.remove("collapsing");
                ele.classList.add("collapse");
                ele.style.height = null;
            }, 350);
        };
    }

    private handleClick(e: React.MouseEvent<HTMLAnchorElement>, link: ILink) {
        NavigationHelper.handleClick(e, () => {
            if (this.props.onClick) {
                this.props.onClick(link);
            }
        });
    }

    private renderChilds(index: number, link: ILink): JSX.Element {
        const isActive = link.active ? "active" : "";

        if (link.uri) {
            if (link.absolute) {
                return <li key={index} className={isActive}>
                    <a target="_blank" href={link.uri}>{link.name}</a>
                </li>;
            } else {
                return <li key={index} className={isActive}>
                    <a onClick={(e) => this.handleClick(e, link)}
                        href={"/" + link.uri}>{link.name}</a>
                </li>;
            }
            
        } else {
            return <li key={index} className={isActive}>
                <span className="header">{link.name}</span>
            </li>;
        }
    }

    private renderTopElement(index: number, links: ILinkCollection): JSX.Element {
        const isActive = links.item.active ? "active" : "";
        const subClass = "nav nav-sub collapse " + (links.item.active ? "in" : "");
        let children: JSX.Element[] = [];
        if (links.children) {
            children = links.children.map((e, i) => {
                return this.renderChilds(i, e);
            });
        }
        return <li key={index} className={isActive}>
            <a
                onClick={(e) => {
                    // this.toggle(index);
                    this.handleClick(e, links.item);
                }}
                href={"/" + links.item.uri}>
                {links.item.name}
                <span style={{ float: "right" }}>
                    <span className="glyphicon glyphicon-menu-down"></span>
                </span>
            </a>
            <ul ref={(ele) => {
                if (ele) {
                    this.topItems[index] = ele;
                }
            }}
                className={subClass}>
                {children}
            </ul>
        </li>;
    }
}

export { CollapsableNavMenu };
