import React from "react"
import { Link } from "react-router-dom"
import { Repository_Type } from "../../../proto/qf/types_pb"

interface RepositoryCardsProps {
    repositories: Record<number, string>
    groupName?: string
    hasGroup?: boolean
    groupPath?: string
}

interface RepositoryLinkConfig {
    type: Repository_Type
    label: string
    group: "repositories" | "resources"
}

const repositoryLinks: RepositoryLinkConfig[] = [
    { type: Repository_Type.USER, label: "User Repo", group: "repositories" },
    { type: Repository_Type.GROUP, label: "Group Repo", group: "repositories" },
    { type: Repository_Type.ASSIGNMENTS, label: "Assignments", group: "resources" },
    { type: Repository_Type.INFO, label: "Course Info", group: "resources" }
]

interface RepoLinkGroupProps {
    title: string
    links: Array<{ label: string; url: string }>
}

const RepoLinkGroup = ({ title, links }: RepoLinkGroupProps) => {
    if (links.length === 0) return null

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
export const RepositoryCards = ({ repositories, groupName, hasGroup, groupPath }: RepositoryCardsProps) => {
    const repositoryGroupLinks = repositoryLinks
        .filter(config => config.group === "repositories" && repositories?.[config.type])
        .map(config => ({
            label: config.type === Repository_Type.GROUP && groupName
                ? `Group Repo ${groupName}`
                : config.label,
            url: repositories[config.type]
        }))

    const resourcesGroupLinks = repositoryLinks
        .filter(config => config.group === "resources" && repositories?.[config.type])
        .map(config => ({
            label: config.label,
            url: repositories[config.type]
        }))

    return (
        <>
            <RepoLinkGroup title="Repos" links={repositoryGroupLinks} />
            <RepoLinkGroup title="Resources" links={resourcesGroupLinks} />
            {groupPath && (
                <div className="flex items-center gap-2">
                    <span className="text-xs font-semibold text-base-content/50 uppercase tracking-wider whitespace-nowrap">Group</span>
                    <Link
                        to={groupPath}
                        className="btn btn-xs btn-ghost border border-base-content/20"
                    >
                        {hasGroup ? `View ${groupName}` : "Create Group"}
                    </Link>
                </div>
            )}
        </>
    )
}
