import React from "react"
import { Repository_Type } from "../../../proto/qf/types_pb"

interface RepositoryCardsProps {
    repositories: Record<number, string>
    groupName?: string
}

interface RepositoryLinkConfig {
    type: Repository_Type
    label: string
    group: "repositories" | "resources"
}

const repositoryLinks: RepositoryLinkConfig[] = [
    { type: Repository_Type.USER, label: "User Repository", group: "repositories" },
    { type: Repository_Type.GROUP, label: "Group Repository", group: "repositories" },
    { type: Repository_Type.ASSIGNMENTS, label: "Assignments", group: "resources" },
    { type: Repository_Type.INFO, label: "Course Info", group: "resources" }
]

interface RepositoryCardProps {
    title: string
    description: string
    links: Array<{ label: string; url: string }>
}

const RepositoryCard = ({ title, description, links }: RepositoryCardProps) => {
    if (links.length === 0) return null

    return (
        <div className="card bg-base-200 shadow-sm">
            <div className="card-body">
                <h5 className="card-title">{title}</h5>
                <p className="card-text">{description}</p>
                <div className="card-actions justify-start gap-2 flex-wrap">
                    {links.map((link) => (
                        <button
                            key={link.label}
                            className="btn btn-primary"
                            onClick={() => window.open(link.url, '_blank', 'noopener,noreferrer')}
                        >
                            {link.label}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    )
}

/** RepositoryCards displays grouped repository links for a course */
export const RepositoryCards = ({ repositories, groupName }: RepositoryCardsProps) => {
    // Build links for "My Repositories" card
    const repositoryGroupLinks = repositoryLinks
        .filter(config => config.group === "repositories" && repositories?.[config.type])
        .map(config => ({
            label: config.type === Repository_Type.GROUP && groupName
                ? `${config.label} ${groupName}`
                : config.label,
            url: repositories[config.type]
        }))

    // Build links for "Course Resources" card
    const resourcesGroupLinks = repositoryLinks
        .filter(config => config.group === "resources" && repositories?.[config.type])
        .map(config => ({
            label: config.label,
            url: repositories[config.type]
        }))

    return (
        <>
            <RepositoryCard
                title="My Repositories"
                description="Access your personal and group repositories."
                links={repositoryGroupLinks}
            />
            <RepositoryCard
                title="Course Resources"
                description="View assignments and course information."
                links={resourcesGroupLinks}
            />
        </>
    )
}
