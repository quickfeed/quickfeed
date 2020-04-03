import * as React from "react";
import { Course, Enrollment } from "../../../proto/ag_pb";
import { DynamicTable, Search } from "../../components";
import { ILink } from '../../managers/NavigationManager';

interface ArchiveViewProps {
    enrollments: Enrollment[];
    onChangeClick: (enrol: Enrollment) => Promise<boolean>;
}

interface ArchiveViewState {
    editing: boolean;
}

export class ArchiveView extends React.Component<ArchiveViewProps, ArchiveViewState> {

    constructor(props: ArchiveViewProps) {
        super(props);
        this.state = {
            editing: false,
        }
    }

    public render() {
        return <div>
            <Search className="input-group"
                    placeholder="Search for users"
                    onChange={(query) => this.handleSearch(query)}
                />
            <DynamicTable
            data={this.props.enrollments}
            header={["Course code", "Course Name", "State"]}
            selector={(enrol: Enrollment) => this.createCourseRow(enrol)}>
        </DynamicTable></div>;

    }

    private createCourseRow(enrol: Enrollment): (string | JSX.Element)[] {
        const course = enrol.getCourse();
        if (!course) {
            return [];
        }
        const activateLink = {
            name: "Activate",
            uri: "activate",
        }
        const archivateLink = {
            name: "Archivate",
            uri: "archivate",
        }
        const makeFavoriteLink = {
            name: "Make favorite",
            uri: "favorite",
        }
        const activeLink = {
            name: "Active",
            extra: "light"
        }
        const archivedLink = {
            name: "Archived",
            extra: "light",
        }
        const favoriteLink = {
            name: "Favorite",
            extra: "light",
        }

        const base: (string | JSX.Element)[] = [course.getCode(), course.getName()];
        const buttonLinks: ILink[] = [];
        switch (enrol.getState()) {
            case Enrollment.DisplayState.ACTIVE:
                this.state.editing ?
                    buttonLinks.push(archivateLink, makeFavoriteLink) :
                    buttonLinks.push(activeLink);
                break;
            case Enrollment.DisplayState.ARCHIVED:
                this.state.editing ?
                    buttonLinks.push(activateLink, makeFavoriteLink) :
                    buttonLinks.push(archivedLink);
                break;
            case Enrollment.DisplayState.FAVORITE:
                this.state.editing ?
                    buttonLinks.push(activateLink, archivateLink) :
                    buttonLinks.push(favoriteLink);
                break;
            default:
                console.log("Got unexpected display status: " + enrol.getState());
        }

        // TODO: generate buttons from the links

        // TODO: add buttons as a single button group to the table row

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