import * as React from "react";
import { Enrollment } from "../../../proto/ag_pb";
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink } from "../../managers/NavigationManager";
import { searchForCourses, sortEnrollmentsByVisibility } from "../../componentHelper";

interface CourseListViewProps {
    enrollments: Enrollment[];
    onChangeClick: (enrol: Enrollment) => Promise<boolean>;
}

interface CourseListViewState {
    sortedCourses: Enrollment[];
    editing: boolean;
}

export class CourseListView extends React.Component<CourseListViewProps, CourseListViewState> {

    private showLink = {
        name: "Show",
        uri: "show",
        extra: "primary",
    }
    private hideLink = {
        name: "Hide",
        uri: "hide",
        extra: "primary",
    }
    private makeFavoriteLink = {
        name: "Favorite",
        uri: "favorite",
        extra: "success",
    }
    private visibleLink = {
        name: "Visible",
        extra: "light"
    }
    private hiddenLink = {
        name: "Hidden",
        extra: "light",
    }
    private favoriteLink = {
        name: "Favorite",
        extra: "light",
    }
    constructor(props: CourseListViewProps) {
        super(props);
        this.state = {
            editing: false,
            sortedCourses: sortEnrollmentsByVisibility(this.props.enrollments),
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
            header={["Course code", "Course Name", "Year", "State"]}
            selector={(enrol: Enrollment) => this.createCourseRow(enrol)}>
        </DynamicTable></div>;
    }

    private generateCourseStateLinks(status: Enrollment.DisplayState): ILink[] {
        const buttonLinks: ILink[] = [];
        switch (status) {
            case Enrollment.DisplayState.VISIBLE:
                this.state.editing ?
                    buttonLinks.push(this.hideLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.visibleLink);
                break;
            case Enrollment.DisplayState.UNSET:
                this.state.editing ?
                    buttonLinks.push(this.hideLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.visibleLink);
                break;
            case Enrollment.DisplayState.HIDDEN:
                this.state.editing ?
                    buttonLinks.push(this.showLink, this.makeFavoriteLink) :
                    buttonLinks.push(this.hiddenLink);
                break;
            case Enrollment.DisplayState.FAVORITE:
                this.state.editing ?
                    buttonLinks.push(this.showLink, this.hideLink) :
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
        const base: (string | JSX.Element)[] = [course.getCode(), course.getName(), course.getYear() + "-" + course.getTag()];
        const links = this.generateCourseStateLinks(enrol.getState());
        const linkButtons = links.map((v, i) => {
            let action: Enrollment.DisplayState;
            switch (v.uri) {
                case "show":
                    action = Enrollment.DisplayState.VISIBLE;
                    break;
                case "hide":
                    action = Enrollment.DisplayState.HIDDEN;
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
        this.setState({
            sortedCourses: searchForCourses(sortEnrollmentsByVisibility(this.props.enrollments), query) as Enrollment[],
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