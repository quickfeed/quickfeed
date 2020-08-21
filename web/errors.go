package web

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrGroupNameDuplicate indicates that another group with the same name already exists on this course
	ErrGroupNameDuplicate = status.Errorf(codes.AlreadyExists, "group with this name already exists. Please choose another name")
	ErrUserNotInGroup     = status.Errorf(codes.NotFound, "user is not in group")
	// ErrInvalidUserInfo is returned to user if user information in context is invalid.
	ErrInvalidUserInfo = status.Errorf(codes.PermissionDenied, "authorization failed. please try to logout and sign in again")
	// ErrNotCommentAuthor indicates that the current user attempts to edit a comment created by another user.
	ErrNotCommentAuthor = status.Errorf(codes.PermissionDenied, "users cannot edit comments posted by other users")
	// ErrNotCourseTeacher indicates that the attempted action requires course teacher privileges.
	ErrNotCourseTeacher = status.Errorf(codes.PermissionDenied, "this action can be performed only by a course teacher")
)
