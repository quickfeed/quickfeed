package hooks

import (
	"context"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// handleInstallationCreated handles installation created events.
// This event is triggered when a user installs the QuickFeed GitHub app on their organization.
// The event causes QuickFeed to create a course for the organization and create repositories for the course.
func (wh GitHubWebHook) handleInstallationCreated(event *github.InstallationEvent) {
	remoteID := event.GetSender().GetID()
	user, err := wh.db.GetUserByRemoteIdentity(uint64(remoteID))
	if err != nil {
		wh.logger.Errorf("Installation created event: could not find user with remote ID %d: %v", remoteID, err)
		return
	}

	orgName := event.GetInstallation().GetAccount().GetLogin()
	orgID := uint64(event.GetInstallation().GetAccount().GetID())
	course := &qf.Course{
		ScmOrganizationID:   orgID,
		ScmOrganizationName: orgName,
		Name:                orgName,
		CourseCreatorID:     user.ID,
	}

	// TODO: Not sure if a timeout is needed here.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sc, err := wh.scmMgr.GetOrCreateSCM(ctx, wh.logger, orgName)
	if err != nil {
		wh.logger.Errorf("Installation created event: could not get scm client: %v", err)
		return
	}
	// TODO: The following code is more or less duplicated from web/courses_new.go.
	// TODO: Refactor to avoid duplication.
	repos, err := sc.CreateCourse(ctx, &scm.CourseOptions{
		CourseCreator:  user.Login,
		OrganizationID: orgID,
	})
	if err != nil {
		wh.logger.Errorf("Installation created event: could not create course: %v", err)
		return
	}
	for _, repo := range repos {
		dbRepo := qf.Repository{
			ScmOrganizationID: orgID,
			ScmRepositoryID:   repo.ID,
			HTMLURL:           repo.HTMLURL,
			RepoType:          qf.RepoType(repo.Path),
		}
		if dbRepo.IsUserRepo() {
			dbRepo.UserID = user.ID
		}
		if err := wh.db.CreateRepository(&dbRepo); err != nil {
			wh.logger.Errorf("Installation created event: failed to create database record for repository %s: %w", repo.Path, err)
			return
		}
	}

	if err := wh.db.CreateCourse(user.ID, course); err != nil {
		wh.logger.Errorf("Installation created event: failed to create database record for course %s: %w", course.Name, err)
	}
}
