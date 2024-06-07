package scm

const (
	getByID                                         = "GET /organizations/{id}"
	getOrgsByOrg                                    = "GET /orgs/{org}"
	getOrgsReposByOrg                               = "GET /orgs/{org}/repos"
	getOrgsMembershipsByOrgByUsername               = "GET /orgs/{org}/memberships/{username}"
	getReposContentsByOwnerByRepoByPath             = "GET /repos/{owner}/{repo}/contents/{path...}"
	getReposCollaboratorsByOwnerByRepo              = "GET /repos/{owner}/{repo}/collaborators"
	putReposCollaboratorsByOwnerByRepoByUsername    = "PUT /repos/{owner}/{repo}/collaborators/{username}"
	deleteReposCollaboratorsByOwnerByRepoByUsername = "DELETE /repos/{owner}/{repo}/collaborators/{username}"
	postIssueByOwnerByRepo                          = "POST /repos/{owner}/{repo}/issues"
	patchIssueByOwnerByRepoByIssueNumber            = "PATCH /repos/{owner}/{repo}/issues/{issue_number}"
	getIssueByOwnerByRepoByIssueNumber              = "GET /repos/{owner}/{repo}/issues/{issue_number}"
	getIssuesByOwnerByRepo                          = "GET /repos/{owner}/{repo}/issues"
	postIssueCommentByOwnerByRepoByIssueNumber      = "POST /repos/{owner}/{repo}/issues/{issue_number}/comments"
)
