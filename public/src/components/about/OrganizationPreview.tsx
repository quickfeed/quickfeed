import { previewOrganization, previewRepositories } from "./previewData"

/**
 * OrganizationPreview shows the repositories QuickFeed creates in a course
 * organization on GitHub. It is a diagram of the repository layout rather than a
 * mock of GitHub's own interface: the names come from the defaults in
 * qf/repo.go, so the preview stays accurate as long as those do.
 */
const OrganizationPreview = () => (
    <div className="card bg-base-200 shadow-xl rounded-2xl overflow-hidden">
        <div className="flex items-center gap-2 bg-base-300 px-4 py-3 border-b border-base-content/10">
            <i className="fab fa-github" />
            <span className="font-semibold text-sm">github.com/{previewOrganization}</span>
            <span className="badge badge-sm badge-ghost ml-auto">Organization</span>
        </div>
        <ul className="divide-y divide-base-content/10 text-sm">
            {previewRepositories.map(repository => (
                <li key={repository.name} className="flex items-center gap-3 px-4 py-2">
                    <i className={`${repository.icon} w-4 text-center text-base-content/60`} />
                    <span className="font-mono truncate">{repository.name}</span>
                    <span className="ml-auto whitespace-nowrap text-xs text-base-content/60">{repository.access}</span>
                </li>
            ))}
        </ul>
        <div className="flex items-start gap-2 border-t border-base-content/10 px-4 py-3 text-xs text-base-content/70">
            <i className="fas fa-code-branch mt-0.5" />
            <span>QuickFeed creates and manages these repositories, and treats a push to the default branch as a submission.</span>
        </div>
    </div>
)

export default OrganizationPreview
