package types_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf/types"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestHasChanged tests if HasChanged returns the correct value
func TestHasChanged(t *testing.T) {
	tests := map[string]struct {
		task1, task2 *types.Task
		want         bool
	}{
		"No change":             {task1: &types.Task{Title: "Title 1", Body: "Body 1"}, task2: &types.Task{Title: "Title 1", Body: "Body 1"}, want: false},
		"Body change":           {task1: &types.Task{Title: "Title 1", Body: "Body 1"}, task2: &types.Task{Title: "Title 1", Body: "Body 2"}, want: true},
		"Title change":          {task1: &types.Task{Title: "Title 1", Body: "Body 1"}, task2: &types.Task{Title: "Title 2", Body: "Body 1"}, want: true},
		"Body and title change": {task1: &types.Task{Title: "Title 1", Body: "Body 1"}, task2: &types.Task{Title: "Title 2", Body: "Body 2"}, want: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.task1.HasChanged(tt.task2); tt.want != got {
				t.Errorf("\ntask1.HasChanged(task2) = %t, expected %t\ntask1:\t%v\ntask2:\t%v", got, tt.want, tt.task1, tt.task2)
			}
		})
	}
}

func TestMarkDeleted(t *testing.T) {
	const deleteMsg = "\n**The task associated with this issue has been deleted by the teaching staff.**\n"

	tests := map[string]struct {
		task     *types.Task
		wantTask *types.Task
		deleted  bool
	}{
		"Delete task": {
			task:     &types.Task{Title: "Title 1", Body: "Body 1"},
			wantTask: &types.Task{Title: "DELETED Title 1", Body: deleteMsg + "Body 1"},
			deleted:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotTask := tt.task
			gotTask.MarkDeleted()
			if diff := cmp.Diff(tt.wantTask, gotTask, protocmp.Transform()); diff != "" {
				t.Errorf("MarkDeleted() mismatch (-wantTask +gotTask):\n%s", diff)
			}
			if got := gotTask.IsDeleted(); got != tt.deleted {
				t.Errorf("IsDeleted() = %t, expected %t", got, tt.deleted)
			}
		})
	}
}
