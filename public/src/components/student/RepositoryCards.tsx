import { Link } from "react-router-dom"
import { Repository_Type } from "../../../proto/qf/types_pb"
import { useCourseID } from "../../hooks/useCourseID"
import { useAppState } from "../../overmind"

interface RepositoryLinkConfig {
    type: Repository_Type
    label: string
    group: "repositories" | "resources"
}

const repositoryLinks: RepositoryLinkConfig[] = [
    { type: Repository_Type.USER, label: "User Repo", group: "repositories" },
    { type: Repository_Type.GROUP, label: "Group Repo", group: "repositories" },
    { type: Repository_Type.ASSIGNMENTS, label: "Assignments", group: "resources" },
    { type: Repository_Type.INFO, label: "Course Info", group: "resources" },
    // Users only see the tests repo if they are enrolled in the course as a teacher.
    // If they are enrolled as students, the tests repo is not included from the backend at all, so it won't show up in the UI.
    { type: Repository_Type.TESTS, label: "Tests", group: "resources" }
]

interface RepoLinkGroupProps {
    title: string
    links: Array<{ label: string; url: string }>
}

const RepoLinkGroup = ({ title, links }: RepoLinkGroupProps) => {
    if (links.length === 0) { return null }

    return (
        <div className="flex items-center gap-2">
            <span className="text-xs font-semibold text-base-content/50 uppercase tracking-wider whitespace-nowrap">{title}</span>
            {links.map((link) => (
                <a
                    key={link.label}
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="btn btn-xs btn-ghost border border-base-content/20"
                >
                    {link.label}
                </a>
            ))}
        </div>
    )
}

/** RepositoryCards displays grouped repository links for a course as a compact inline strip */
export const RepositoryCards = () => {
    const state = useAppState()
    const courseID = useCourseID()
    const courseIDStr = courseID.toString()
    const repositories = state.repositories[courseIDStr]
    const enrollment = state.enrollmentsByCourseID[courseIDStr]
    const hasGroup = state.hasGroup(courseIDStr)
    const groupName = enrollment?.group ? `(${enrollment.group.name})` : ""

    const linksForGroup = (group: RepositoryLinkConfig["group"]) =>
        repositoryLinks
            .filter(config => config.group === group && repositories?.[config.type])
            .map(config => ({
                // If the type is GROUP and the user has a group, include the group name in the label. Otherwise, use the default label.
                // All other types just use the default label.
                label: config.type === Repository_Type.GROUP && groupName
                    ? `Group Repo ${groupName}`
                    : config.label,
                url: repositories[config.type]
            }))

    const repositoryGroupLinks = linksForGroup("repositories")
    const resourcesGroupLinks = linksForGroup("resources")

    return (
        <div className="flex flex-wrap items-center gap-x-6 gap-y-2 mt-3 mb-4 px-3 py-2 bg-base-200 rounded-lg">
            <RepoLinkGroup title="Repos" links={repositoryGroupLinks} />
            <RepoLinkGroup title="Resources" links={resourcesGroupLinks} />
            <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-base-content/50 uppercase tracking-wider whitespace-nowrap">Group</span>
                <Link
                    to={`/course/${courseID}/group`}
                    className="btn btn-xs btn-ghost border border-base-content/20"
                >
                    {hasGroup ? `View ${groupName}` : "Create Group"}
                </Link>
            </div>
        </div>
    )
}
