import * as React from "react";
import { Enrollment } from '../../../proto/ag_pb';
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink } from '../../managers/NavigationManager';

interface VisibilityViewProps {
    enrollments: Enrollment[];
    onChangeClick: (enrol: Enrollment) => Promise<boolean>;
}

interface VisibilityViewState {
    editing: boolean;
}

export class CourseVisibilityView extends React.Component<VisibilityViewProps, VisibilityViewState> {

    private activateLink = {
        name: "Activate",
        uri: "activate",
    }
    private archivateLink = {
        name: "Archivate",
        uri: "archivate",
    }
    private makeFavoriteLink = {
        name: "Make favorite",
        uri: "favorite",
    }
    private activeLink = {
        name: "Active",
        extra: "light"
    }
    private archivedLink = {
        name: "Archived",
        extra: "light",
    }
    private favoriteLink = {
        name: "Favorite",
        extra: "light",
    }
    constructor(props: VisibilityViewProps) {
        super(props);
        this.state = {
            editing: false,
        }
    }

    public render() {
        return <div>
            <Search className="input-group"
                    placeholder="Search for courses"
                    onChange={(query) => this.handleSearch(query)}
                />
            <DynamicTable
            data={this.props.enrollments}
            header={["Course code", "Course Name", "State"]}
            selector={(enrol: Enrollment) => this.createCourseRow(enrol)}>
        </DynamicTable></div>;

    }

    private generateCourseStateLinks(status: Enrollment.DisplayState): ILink[] {
        const buttonLinks: ILink[] = [];
        switch (status) {
            case Enrollment.DisplayState.ACTIVE:
                this.state.editing ?
                    buttonLinks.push(this.archivateLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.activeLink);
                break;
            case Enrollment.DisplayState.ARCHIVED:
                this.state.editing ?
                    buttonLinks.push(this.activateLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.archivedLink);
                break;
            case Enrollment.DisplayState.FAVORITE:
                this.state.editing ?
                    buttonLinks.push(this.activateLink, this.archivateLink) :
                    buttonLinks.push(this.favoriteLink);
                break;
            default:
                console.log("Got unexpected display status: " + status);
        }
        return buttonLinks;
    }

    private createCourseRow(enrol: Enrollment): (string | JSX.Element)[] {
        const course = enrol.getCourse();
        if (!course) {
            return [];
        }
        const base: (string | JSX.Element)[] = [course.getCode(), course.getName()];
        const links = this.generateCourseStateLinks(enrol.getState());
        const linkButtons = links.map((v, i) => {
            let action: Enrollment.DisplayState;
            switch (v.uri) {
                case "activate":
                    action = Enrollment.DisplayState.ACTIVE;
                case "archivate":
                    action = Enrollment.DisplayState.ARCHIVED;
                case "favorite":
                    action = Enrollment.DisplayState.FAVORITE;
                default:
                    console.log("Got unexpected link uri: " + v.uri);
            }

            return <BootstrapButton
                key={i}
                classType={v.extra ? v.extra as BootstrapClass : "default"}
                type={v.description}
                onClick={(link) => { this.handleStateChange(enrol, action)}}
            >{v.name}
            </BootstrapButton>;
            });

        const btnGroup: JSX.Element = <div className="btn-group action-btn">{linkButtons}</div>
        base.push(btnGroup);
        return base;
    }

    private handleSearch(query: string) {
        // TODO: search by course name, code, year or semester/tag
        return;
    }

    private async toggleEdit() {
        this.setState({
            editing: !this.state.editing,
        })
    }

    // TODO: pass to buttons as an onclick function
    private async handleStateChange(enrol: Enrollment, state: Enrollment.DisplayState) {
        const baseState = enrol.getState();
        enrol.setState(state);
        const ans = await this.props.onChangeClick(enrol);
        if (!ans) {
            enrol.setState(baseState);
        }
    }

}