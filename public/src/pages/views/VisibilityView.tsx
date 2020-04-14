import * as React from "react";
import { Enrollment } from '../../../proto/ag_pb';
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink } from '../../managers/NavigationManager';
import { searchForCourses, sortCoursesByVisibility } from '../../componentHelper';

interface VisibilityViewProps {
    enrollments: Enrollment[];
    onChangeClick: (enrol: Enrollment) => Promise<boolean>;
}

interface VisibilityViewState {
    sortedCourses: Enrollment[];
    editing: boolean;
}

export class CourseVisibilityView extends React.Component<VisibilityViewProps, VisibilityViewState> {

    private activateLink = {
        name: "Show",
        uri: "activate",
        extra: "primary",
    }
    private archiveLink = {
        name: "Hide",
        uri: "archive",
        extra: "primary",
    }
    private makeFavoriteLink = {
        name: "Make favorite",
        uri: "favorite",
        extra: "success",
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
            sortedCourses: sortCoursesByVisibility(this.props.enrollments),
        }
    }

    public render() {
        return <div>
            <Search className="input-group"
                    placeholder="Search for courses"
                    onChange={(query) => this.handleSearch(query)}
                />
            <div>{this.editButton()}</div>
            <DynamicTable
            data={this.state.sortedCourses}
            header={["Course code", "Course Name", "State"]}
            selector={(enrol: Enrollment) => this.createCourseRow(enrol)}>
        </DynamicTable></div>;
    }

    private generateCourseStateLinks(status: Enrollment.DisplayState): ILink[] {
        const buttonLinks: ILink[] = [];
        switch (status) {
            case Enrollment.DisplayState.ACTIVE:
                this.state.editing ?
                    buttonLinks.push(this.archiveLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.activeLink);
                break;
            case Enrollment.DisplayState.UNSET:
                this.state.editing ?
                    buttonLinks.push(this.archiveLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.activeLink);
                break;
            case Enrollment.DisplayState.ARCHIVED:
                this.state.editing ?
                    buttonLinks.push(this.activateLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.archivedLink);
                break;
            case Enrollment.DisplayState.FAVORITE:
                this.state.editing ?
                    buttonLinks.push(this.activateLink, this.archiveLink) :
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
                    break;
                case "archive":
                    action = Enrollment.DisplayState.ARCHIVED;
                    break;
                case "favorite":
                    action = Enrollment.DisplayState.FAVORITE;
                    break;
                default:
                    console.log("Got unexpected link uri: " + v.uri);
                    action = Enrollment.DisplayState.UNSET;
            }

            return <BootstrapButton
                key={i}
                classType={v.extra ? v.extra as BootstrapClass : "default"}
                type={v.description}
                onClick={() => { this.handleStateChange(enrol, action)}}
            >{v.name}
            </BootstrapButton>;
            });

        const btnGroup = <div className="btn-group action-btn">{linkButtons}</div>
        base.push(btnGroup);
        return base;
    }

    private handleSearch(query: string) {
        query.toLowerCase();
        const filteredCourses = searchForCourses(sortCoursesByVisibility(this.props.enrollments), query);

        this.setState({
            sortedCourses: filteredCourses,
        });
    }

    private async toggleEdit() {
        this.setState({
            editing: !this.state.editing,
        })
    }

    private async handleStateChange(enrol: Enrollment, state: Enrollment.DisplayState) {
        if (state) {
            const baseState = enrol.getState();
            enrol.setState(state);
            const ans = await this.props.onChangeClick(enrol);
            if (!ans) {
                enrol.setState(baseState);
            }
        }
    }

    private editButton() {
        return <button type="button"
                id="edit"
                className="btn btn-success member-btn"
                onClick={() => this.toggleEdit()}
        >{this.editButtonString()}</button>;
    }

    private editButtonString(): string {
        return this.state.editing ? "Done" : "Edit";
    }
}