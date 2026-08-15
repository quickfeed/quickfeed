package hooks

import (
	"context"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

func (wh GitHubWebHook) handleInstallationCreated(ctx context.Context, event *github.InstallationEvent) {
	logger := qlog.FromContext(ctx)
	installerID := uint64(event.GetSender().GetID())
	courseCreator, err := wh.db.GetUserByRemoteIdentity(installerID)
	if err != nil {
		logger.Error("failed to get user by remote identity", label.Error, err)
		return
	}

	if !courseCreator.GetIsAdmin() {
		logger.Error("non-administrator attempted course installation", label.User, courseCreator.GetLogin())
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
		CourseCreatorID:     courseCreator.GetID(),
		Year:                defaultYear(now),
	}

	ctx, logger = qlog.WithLogger(
		ctx,
		label.CourseCode, course.GetCode(),
		label.Organization, orgName,
		label.User, courseCreator.GetLogin(),
	)
	sc, err := wh.scmMgr.GetOrCreateSCM(ctx, orgName)
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return
	}
	c, err := createCourse(ctx, wh.db, sc, course, courseCreator)
	if err != nil {
		// This may be an scm.ErrAlreadyExists error
		logger.Error("failed to create course", label.Error, err)
		return
	}
	// The course now exists, so its records can be copied to its log.
	_, logger = qlog.WithCourse(ctx, c)
	logger.Info("created course")

	if err := wh.tm.Add(courseCreator.GetID()); err != nil {
		logger.Error("failed to schedule token refresh", label.Error, err)
	}
}

func (wh GitHubWebHook) handleInstallationDeleted(ctx context.Context, event *github.InstallationEvent) {
	logger := qlog.FromContext(ctx)
	orgName := event.GetInstallation().GetAccount().GetLogin()
	logger.Info("removing SCM client due to installation deletion", label.Organization, orgName)
	wh.scmMgr.DeleteSCM(orgName)
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
