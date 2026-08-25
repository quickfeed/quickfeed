import { useState } from "react"
import { Repository_Type } from "../../../proto/qf/types_pb"
import { useCourseID } from "../../hooks/useCourseID"
import { useAppState } from "../../overmind"


const Command = ({ children }: { children: string }) => {
    const [copyStatus, setCopyStatus] = useState<"idle" | "copied" | "failed">("idle")
    const isSingleLine = !children.includes("\n")

    const copyCommand = async () => {
        try {
            await navigator.clipboard.writeText(children)
            setCopyStatus("copied")
        } catch {
            setCopyStatus("failed")
        }
    }

    const copyLabel = copyStatus === "copied" ? "Copied" : copyStatus === "failed" ? "Copy failed" : "Copy"

    return (
        <div className="relative my-3">
            <pre className={`overflow-x-auto rounded-lg bg-neutral p-4 text-sm text-neutral-content${isSingleLine ? " pr-24" : ""}`}>
                <code>{children}</code>
            </pre>
            {isSingleLine && (
                <button
                    type="button"
                    className="btn btn-ghost btn-xs absolute right-2 top-2 text-neutral-content"
                    aria-label={`Copy command: ${children}`}
                    onClick={() => void copyCommand()}
                >
                    <i className="fas fa-copy" aria-hidden="true" />
                    {copyLabel}
                </button>
            )}
        </div>
    )
}

const ExternalLink = ({ href, children }: { href: string, children: React.ReactNode }) => (
    <a href={href} target="_blank" rel="noopener noreferrer" className="link link-primary">
        {children}
    </a>
)

const repositoryName = (url: string): string => url
    .replace(/^https:\/\/github\.com\//, "")
    .replace(/\.git$/, "")

const directoryName = (courseCode?: string): string => courseCode
    ?.toLowerCase()
    .replace(/[^a-z0-9._-]+/g, "-")
    .replace(/^-|-$/g, "") || "course"

/** SubmissionGuide explains the QuickFeed submission workflow using the active course's repositories. */
const SubmissionGuide = () => {
    const state = useAppState()
    const courseID = useCourseID()
    const course = state.courses.find(candidate => candidate.ID === courseID)
    const repositories = state.repositories[courseID.toString()]
    const userRepository = repositories?.[Repository_Type.USER] ?? ""
    const groupRepository = repositories?.[Repository_Type.GROUP] ?? ""
    const assignmentsRepository = repositories?.[Repository_Type.ASSIGNMENTS] ?? ""
    const workingDirectory = directoryName(course?.code)

    const cloneCommand = userRepository
        ? `gh repo clone ${repositoryName(userRepository)} ${workingDirectory}`
        : "gh repo clone ORGANIZATION/username-labs course"
    const gitCloneCommand = userRepository
        ? `git clone ${userRepository} ${workingDirectory}`
        : "git clone https://github.com/ORGANIZATION/username-labs course"
    const upstreamCommand = assignmentsRepository
        ? `git remote add upstream ${assignmentsRepository}`
        : "git remote add upstream https://github.com/ORGANIZATION/assignments"

    return (
        <main className="mx-auto w-full max-w-5xl pb-16">
            <header className="mb-8">
                <p className="mb-2 text-sm font-semibold uppercase tracking-wider text-primary">Student guide</p>
                <h1 className="mb-4 text-4xl font-bold">Submitting assignments</h1>
                <p className="max-w-3xl text-lg leading-relaxed text-base-content/75">
                    QuickFeed treats a push to your repository&apos;s default branch as a submission.
                </p>
            </header>

            <div role="alert" className="alert alert-info mb-10 items-start">
                <i className="fas fa-circle-info mt-1" aria-hidden="true" />
                <span>
                    Your course may have additional rules about deadlines, group work, commit messages, or manual review.
                    Follow those course rules in addition to this QuickFeed guide.
                </span>
            </div>

            <div className="space-y-10 text-base leading-relaxed">
                <section>
                    <h2 className="mb-3 text-2xl font-bold">1. Choose the right repository</h2>
                    <p>
                        Use <strong>User Repo</strong> on the course page for individual work.
                        For approved group work, use <strong>Group Repo</strong> instead.
                        The <strong>Assignments</strong> repository is maintained by the teaching staff and is read-only for students.
                    </p>
                    <p className="mt-3">
                        QuickFeed normally synchronizes new and updated assignments into student and group repositories automatically.
                    </p>
                </section>

                <section>
                    <h2 className="mb-3 text-2xl font-bold">2. Clone your repository</h2>
                    <p>
                        First configure GitHub authentication using the{" "}
                        <ExternalLink href="https://cli.github.com/">GitHub CLI</ExternalLink> or{" "}
                        <ExternalLink href="https://docs.github.com/en/authentication/connecting-to-github-with-ssh">an SSH key</ExternalLink>.
                        To authenticate with the GitHub CLI, run:
                    </p>
                    <Command>gh auth login</Command>
                    <p>Then clone your repository:</p>
                    <Command>{cloneCommand}</Command>
                    <p>Alternatively, clone over HTTPS:</p>
                    <Command>{gitCloneCommand}</Command>
                </section>

                <section>
                    <h2 className="mb-3 text-2xl font-bold">3. Commit and push your work</h2>
                    <p>From your local repository, use the normal Git workflow:</p>
                    <Command>{`git status\ngit add <files you changed>\ngit commit -m "Describe your changes"\ngit push`}</Command>
                    <ul className="ml-6 list-disc space-y-2">
                        <li>Only committed changes are pushed and submitted.</li>
                        <li>Push individual work to the repository&apos;s default branch so QuickFeed processes it.</li>
                        <li>You may push again whenever you need to resubmit.</li>
                        <li>Open the assignment in QuickFeed to see the latest score, status, and available build output.</li>
                    </ul>
                </section>

                <section>
                    <h2 className="mb-3 text-2xl font-bold">4. Work on a group assignment</h2>
                    {groupRepository ? (
                        <>
                            <p>Your group repository for this course is available now.</p>
                            <Command>{`gh repo clone ${repositoryName(groupRepository)}`}</Command>
                        </>
                    ) : (
                        <p>
                            After your group is approved, <strong>Group Repo</strong> appears on the course page.
                            Clone that repository and push shared work there.
                        </p>
                    )}
                    <p className="mt-3">
                        Pull before starting work and before pushing so you incorporate changes from other group members.
                    </p>
                    <Command>git pull</Command>
                </section>

                <section>
                    <h2 className="mb-3 text-2xl font-bold">5. Resolve assignment update conflicts</h2>
                    <p>
                        QuickFeed cannot automatically synchronize an assignment update when it conflicts with changes in your repository.
                        In that case, check whether an <code className="rounded bg-base-200 px-1 py-0.5">upstream</code> remote already exists:
                    </p>
                    <Command>git remote -v</Command>
                    <p>If it is missing, add the course assignments repository:</p>
                    <Command>{upstreamCommand}</Command>
                    <p>Then merge the latest assignment changes:</p>
                    <Command>git pull upstream main</Command>
                    <p>
                        If Git reports conflicts, edit the affected files, remove the conflict markers, test the combined result, then commit and push it.
                    </p>
                    <Command>{`git add <resolved files>\ngit commit\ngit push`}</Command>
                </section>

                <section>
                    <h2 className="mb-3 text-2xl font-bold">Deadlines and slip days</h2>
                    <p>
                        QuickFeed records the submission when it processes your push.
                        If your course uses slip days, a submission after an assignment deadline may consume them.
                        Check the deadline shown in QuickFeed and your course&apos;s policy for the exact rules.
                    </p>
                </section>
            </div>
        </main>
    )
}

export default SubmissionGuide
