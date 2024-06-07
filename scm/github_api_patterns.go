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
	postReposIssuesByOwnerByRepo                    = "POST /repos/{owner}/{repo}/issues"
	patchReposIssuesByOwnerByRepoByIssueNumber      = "PATCH /repos/{owner}/{repo}/issues/{issue_number}"
	getReposIssuesByOwnerByRepoByIssueNumber        = "GET /repos/{owner}/{repo}/issues/{issue_number}"
)
