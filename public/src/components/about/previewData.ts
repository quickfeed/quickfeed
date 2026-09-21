import { create } from "@bufbuild/protobuf"
import { ScoreSchema } from "../../../proto/kit/score/score_pb"
import type { Submission } from "../../../proto/qf/types_pb"
import { SubmissionSchema } from "../../../proto/qf/types_pb"
import { AssignmentsRepo, InfoRepo, TestsRepo, studentRepoName } from "../../Helpers"

// Sample data for the About page previews. The previews render the application's
// own components and markup, so the data is the only thing invented here; the
// colors and chrome follow whichever daisyUI theme is active.

export const previewAssignment = "lab1"
export const previewOrganization = "dat520-2026"

// Weights 1, 1, 1 and 5 sum to 8, which puts the weighted total at 92%.
export const previewSubmission: Submission = create(SubmissionSchema, {
    score: 92,
    Scores: [
        create(ScoreSchema, { ID: 1n, TestName: "TestGitQuestions", Score: 10, MaxScore: 10, Weight: 1 }),
        create(ScoreSchema, { ID: 2n, TestName: "TestMissingSemesterQuestions", Score: 7, MaxScore: 9, Weight: 1 }),
        create(ScoreSchema, { ID: 3n, TestName: "TestShellQuestions", Score: 12, MaxScore: 20, Weight: 1 }),
        create(ScoreSchema, { ID: 4n, TestName: "TestToken", Score: 4, MaxScore: 4, Weight: 5 }),
    ],
})

export const previewScoreLimit = 80

// Mirrors the output of the run script in doc/templates/go-course: the banners
// come from the script, the rest is ordinary `go test -v` output. The two
// failing tests match the "Tests Passed 2/4" row of the lab result preview, and
// the setup and run times add up to its execution time.
export const previewBuildLog = `*** Preparing Test Execution for ${previewAssignment} ***

*** Finished Test Setup in 3 seconds ***

*** Running Tests ***

=== RUN   TestGitQuestions
--- PASS: TestGitQuestions (0.00s)
=== RUN   TestMissingSemesterQuestions
    questions_test.go:41: TestMissingSemesterQuestions: 7/9 cases passed
--- FAIL: TestMissingSemesterQuestions (0.00s)
=== RUN   TestShellQuestions
    questions_test.go:88: TestShellQuestions: 12/20 cases passed
--- FAIL: TestShellQuestions (0.01s)
=== RUN   TestToken
--- PASS: TestToken (0.00s)
FAIL
FAIL	dat520/${previewAssignment}	0.019s

*** Finished Running Tests in 4 seconds ***`

export type PreviewRepository = {
    name: string
    /** Who the repository belongs to or who can reach it. */
    access: string
    icon: string
}

// The repository names QuickFeed gives a course organization; see qf/repo.go.
export const previewRepositories: PreviewRepository[] = [
    { name: InfoRepo, access: "Course information", icon: "fas fa-circle-info" },
    { name: AssignmentsRepo, access: "Read-only for students", icon: "fas fa-file-code" },
    { name: TestsRepo, access: "Teaching staff only", icon: "fas fa-flask" },
    { name: studentRepoName("hfurubotten"), access: "One per student", icon: "fas fa-user" },
    { name: "group-alpha", access: "One per group", icon: "fas fa-users" },
]

// A cell is rendered with the same color classes getSubmissionCellColor assigns
// in the real results table; "none" leaves the cell on the table ground.
export type PreviewStatus = "approved" | "revision" | "rejected" | "none"

export type PreviewMember = {
    name: string
    labs: { score: number, status: PreviewStatus }[]
}

export const previewLabs = ["lab1", "lab2", "lab3", "lab4", "lab5", "lab6"]

export const previewMembers: PreviewMember[] = [
    { name: "Hector Holt", labs: rowOf([100, 100, 100, 100, 100, 100], ["approved", "approved", "approved", "approved", "approved", "approved"]) },
    { name: "Virginia Russo", labs: rowOf([100, 100, 94, 99, 95, 0], ["approved", "approved", "approved", "approved", "approved", "none"]) },
    { name: "Mary Jackson", labs: rowOf([100, 100, 100, 99, 13, 0], ["approved", "approved", "approved", "approved", "rejected", "none"]) },
    { name: "Laura Glass", labs: rowOf([33, 100, 91, 100, 53, 0], ["revision", "approved", "approved", "approved", "revision", "none"]) },
    { name: "Todd Myers", labs: rowOf([92, 100, 94, 99, 99, 0], ["approved", "approved", "approved", "approved", "approved", "none"]) },
    { name: "Laurie Morales", labs: rowOf([94, 95, 97, 93, 36, 0], ["approved", "approved", "approved", "approved", "rejected", "none"]) },
]

function rowOf(scores: number[], statuses: PreviewStatus[]): PreviewMember["labs"] {
    return scores.map((score, i) => ({ score, status: statuses[i] }))
}
