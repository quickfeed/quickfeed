package scm

const (
	getOrganizationsByID                                      = "GET /organizations/{id}"                                            // GetOrganization
	getOrgsByOrg                                              = "GET /orgs/{org}"                                                    // GetOrganization
	patchOrgsByOrg                                            = "PATCH /orgs/{org}"                                                  // CreateCourse
	getOrgsReposByOrg                                         = "GET /orgs/{org}/repos"                                              // GetRepositories
	postOrgsReposByOrg                                        = "POST /orgs/{org}/repos"                                             // CreateCourse, createCourseRepo
	postReposForksByOwnerByRepo                               = "POST /repos/{owner}/{repo}/forks"                                   // createForkedRepo
	getOrgsMembershipsByOrgByUsername                         = "GET /orgs/{org}/memberships/{username}"                             // GetOrganization
	putOrgsMembershipsByOrgByUsername                         = "PUT /orgs/{org}/memberships/{username}"                             // UpdateEnrollment, DemoteTeacherToStudent
	deleteOrgsMembersByOrgByUsername                          = "DELETE /orgs/{org}/members/{username}"                              // RejectEnrollment
	getReposByOwnerByRepo                                     = "GET /repos/{owner}/{repo}"                                          // CreateCourse, CreateGroup, getRepository, getRepo
	deleteReposByOwnerByRepo                                  = "DELETE /repos/{owner}/{repo}"                                       // DeleteGroup, RejectEnrollment, deleteRepository
	getRepositoriesByID                                       = "GET /repositories/{repository_id}"                                  // getRepository, deleteRepository
	getReposContentsByOwnerByRepoByPath                       = "GET /repos/{owner}/{repo}/contents/{path...}"                       // RepositoryIsEmpty
	getReposCollaboratorsByOwnerByRepo                        = "GET /repos/{owner}/{repo}/collaborators"                            // UpdateGroupMembers
	putReposCollaboratorsByOwnerByRepoByUsername              = "PUT /repos/{owner}/{repo}/collaborators/{username}"                 // CreateCourse, UpdateEnrollment, CreateGroup, UpdateGroupMembers, createStudentRepo, grantPullAccessToCourseRepos
	deleteReposCollaboratorsByOwnerByRepoByUsername           = "DELETE /repos/{owner}/{repo}/collaborators/{username}"              // UpdateGroupMembers
	postReposIssuesByOwnerByRepo                              = "POST /repos/{owner}/{repo}/issues"                                  // CreateIssue
	patchReposIssuesByOwnerByRepoByIssueNumber                = "PATCH /repos/{owner}/{repo}/issues/{issue_number}"                  // UpdateIssue
	getReposIssuesByOwnerByRepoByIssueNumber                  = "GET /repos/{owner}/{repo}/issues/{issue_number}"                    // GetIssue
	getReposIssuesByOwnerByRepo                               = "GET /repos/{owner}/{repo}/issues"                                   // GetIssues
	postReposIssuesCommentsByOwnerByRepoByIssueNumber         = "POST /repos/{owner}/{repo}/issues/{issue_number}/comments"          // CreateIssueComment
	patchReposIssuesCommentsByOwnerByRepoByCommentID          = "PATCH /repos/{owner}/{repo}/issues/comments/{comment_id}"           // UpdateIssueComment
	postReposPullsRequestedReviewersByOwnerByRepoByPullNumber = "POST /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers" // RequestReviewers
	postReposMergeUpstreamByOwnerByRepo                       = "POST /repos/{owner}/{repo}/merge-upstream"                          // SyncFork
	postAppManifestsByCodeConversions                         = "POST /app-manifests/{code}/conversions"                             // CreateCourse
)
