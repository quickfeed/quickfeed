import type { ReactNode } from "react"
import { Link } from "react-router"
import { Repository_Type } from "../../proto/qf/types_pb"
import { useCourseID } from "../hooks/useCourseID"
import { useAppState } from "../overmind"

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

const stripButton = "btn btn-xs btn-ghost border border-base-content/20"

/** StripGroup is one labelled section of the repository strip.
 *  It is `display: contents` below sm so that its label and its links become cells of
 *  the strip's own grid: every group's label shares one column and every group's links
 *  start at the same offset, instead of each stacked row starting wherever its label ends.
 *  From sm up the strip is a flex row again and the group is an ordinary flex item. */
const StripGroup = ({ title, className = "", children }: { title: string, className?: string, children: ReactNode }) => (
    <div className={`contents sm:flex sm:items-center sm:gap-2 ${className}`}>
        <span className="text-xs font-semibold text-base-content/50 uppercase tracking-wider whitespace-nowrap">{title}</span>
        <div className="flex flex-wrap items-center gap-2">
            {children}
        </div>
    </div>
)

const RepoLinkGroup = ({ title, links }: RepoLinkGroupProps) => {
    if (links.length === 0) { return null }

    return (
        <StripGroup title={title}>
            {links.map((link) => (
                <a
                    key={link.label}
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={stripButton}
                >
                    {link.label}
                </a>
            ))}
        </StripGroup>
    )
}

/** CourseLinks displays the course repository, group, and help links as a compact inline strip */
export const CourseLinks = () => {
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
        <div className="grid grid-cols-[auto_1fr] items-center gap-x-3 gap-y-2 sm:flex sm:flex-wrap sm:gap-x-6 mt-3 mb-4 px-3 py-2 bg-base-200 rounded-lg">
            <RepoLinkGroup title="Repos" links={repositoryGroupLinks} />
            <RepoLinkGroup title="Resources" links={resourcesGroupLinks} />
            <StripGroup title="Group">
                <Link to={`/course/${courseID}/group`} className={stripButton}>
                    {hasGroup ? `View ${groupName}` : "Create Group"}
                </Link>
            </StripGroup>
            {/* Help sits at the far right of the strip, but only while the strip is a
                single flex row; when stacked in narrow viewports, it is just the last row of the grid. */}
            <StripGroup title="Help" className="sm:ml-auto">
                <Link to={`/course/${courseID}/submission-guide`} className={stripButton}>
                    Submission Guide
                </Link>
            </StripGroup>
        </div>
    )
}
