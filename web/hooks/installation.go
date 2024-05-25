package hooks

import (
	"context"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func (wh GitHubWebHook) handleInstallationCreated(event *github.InstallationEvent) {
	installerID := uint64(event.GetSender().GetID())
	courseCreator, err := wh.db.GetUserByRemoteIdentity(installerID)
	if err != nil {
		wh.logger.Errorf("Could not get user by remote identity: %v", err)
		return
	}

	if !courseCreator.GetIsAdmin() {
		wh.logger.Errorf("User %s is not an admin", courseCreator.Login)
		return
	}

	orgName := event.GetInstallation().GetAccount().GetLogin()
	orgID := uint64(event.GetInstallation().GetAccount().GetID())
	now := time.Now()
	course := &qf.Course{
		ScmOrganizationID:   orgID,
		ScmOrganizationName: orgName,
		Name:                orgName,
		Code:                orgName,
		Tag:                 defaultTag(now),
		CourseCreatorID:     courseCreator.ID,
		Year:                defaultYear(now),
	}

	ctx := context.Background()
	sc, err := wh.scmMgr.GetOrCreateSCM(ctx, wh.logger, orgName)
	if err != nil {
		wh.logger.Errorf("Could not create SCM client for course %s: %v", orgName, err)
		return
	}
	repos, err := sc.CreateCourse(ctx, &scm.CourseOptions{
		CourseCreator:  courseCreator.Login,
		OrganizationID: orgID,
	})
	if err != nil {
		wh.logger.Errorf("Could not create course %s: %v", orgName, err)
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
			dbRepo.UserID = courseCreator.ID
		}
		if err := wh.db.CreateRepository(&dbRepo); err != nil {
			wh.logger.Errorf("Could not create database repository %s: %v", repo.Path, err)
			return
		}
	}

	if err := wh.db.CreateCourse(courseCreator.ID, course); err != nil {
		wh.logger.Errorf("Could not create database record for course %s: %v", orgName, err)
		return
	}

	if err := wh.tm.Add(courseCreator.ID); err != nil {
		wh.logger.Errorf("Could not add user %s for token refresh: %v", courseCreator.Login, err)
	}
}

func defaultYear(now time.Time) uint32 {
	if now.Month() > time.October {
		return uint32(now.Year() + 1)
	}
	return uint32(now.Year())
}

func defaultTag(now time.Time) string {
	if now.Month() > time.October || now.Month() < time.April {
		return "Spring"
	}
	return "Fall"
}
