package assignments

import (
	"context"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
	"math/rand"
)

func tasks_parser(contents []byte) ([]*pb.Task, error) {
	//TODO Add parsing of task .md File here
	var t []*pb.Task
	t = append(t, &pb.Task{
		ID:    rand.Uint64(),
		Title: "title",
		Body:  string(contents),
		Issue: &[]pb.Issue{},
	})
	return t, nil
}

func isExists(ID uint64, issues *[]pb.Issue) (issue *pb.Issue) {
	for _, issue = range *issues {
		if issue.ID == ID {
			return issue
		}
	}
	return nil
}

func UpdateIssue(c context.Context, sc scm.SCM, course string, repo string, title string, body string, issueNumber uint32) (issue *scm.Issue, err error) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course,
		Repository:   repo,
		Title:        title,
		Body:         body,
	}

	updateIssue := &scm.IssueOptions{
		Organization: course,
		Repository:   repo,
		IssueNumber:  int(issueNumber),
	}
	Issue, err := sc.EditRepoIssue(c, updateIssue, newIssue)
	return Issue, err
}

func CreateTask(c context.Context, logger *zap.SugaredLogger, sc scm.SCM, course *pb.Course, repo scm.Repository, task *pb.Task, assignment *pb.Assignment) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   repo.Path,
		Title:        task.Title,
		Body:         task.Body,
	}
	logger.Debugf("CreateTask: Creating Task %s on repository %s, Course %s ", task.Title, repo.Path, course.Name)
	issue, err := sc.CreateIssue(c, newIssue)
	if err != nil {
		logger.Debugf("SyncTasks: failed to create new task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
		return
	}
	i := &pb.Issue{
		ID:         issue.ID,
		TaskID:     task.ID,
		State:      issue.Status,
		Repository: repo.Path,
		Assignee:   issue.Assignee,
		//Assignees: issue., // TODO Add assignees list in scm
		IssueNumber: issue.IssueNumber,
	}
	*task.Issue = append(*task.Issue, *i)
}

func SyncTasks(c context.Context, logger *zap.SugaredLogger, sc scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	org, err := sc.GetOrganization(c, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		logger.Debugf("SyncTasks: Failed to Fetch Course %s due to ERROR : %s", course.Name, err)
		return err
	}
	_, err = sc.GetRepositories(c, org)
	if err != nil {
		logger.Debugf("SyncTasks: Failed to Fetch Repositories from Course %s due to ERROR: %s", course.Name, err)
		return err
	}

	//for _, repo := range repos {
	//	fmt.Println(repo.Path)
	//	for _, assignment := range assignments {
	//		for _, task := range assignment.Tasks {
	//
	//			//IssueOpt := &scm.IssueOptions{
	//			//	Organization: course.Name,
	//			//	Repository:   repo.Path,
	//			//}
	//
	//			// check if issue already exist
	//			gitIssues, err := sc.GetRepoIssues(c, &scm.IssueOptions{
	//				Organization: course.Name,
	//				Repository:   repo.Path,
	//			})
	//			if gitIssues != nil {
	//				for _, gitIssue := range gitIssues{
	//					issue := isExists(gitIssue.ID,task.Issue)
	//					if (issue != nil){
	//						isu, err := UpdateIssue(c,sc,course.Name,repo.Path,task.Title,task.Body,issue.IssueNumber)
	//						if err != nil {
	//							fmt.Println(err)
	//						}
	//
	//					}
	//				}
	//			}else{
	//				logger.Debugf("SyncTasks: failed to create new task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
	//			}
	//
	//
	//
	//				//task.Number = issue.Number
	//			} else {
	//				fmt.Println(" came here")
	//				//updateIssue := &scm.IssueOptions{
	//				//	Organization: course.Name,
	//				//	Repository:   repo.Path,
	//				//	IssueNumber:  task.Number,
	//				//}
	//				//issue, err := sc.EditRepoIssue(c, updateIssue, newIssue)
	//				//if err != nil {
	//				//	logger.Debugf("EditRepoIssue: failed to Update  task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
	//				//}
	//				//task.Number = issue.Number
	//			}
	//
	//		}
	//
	//	}
	//return nil
	return nil
}
