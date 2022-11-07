package scm

// CourseOptions contain information about new course.
type CourseOptions struct {
	OrganizationID uint64
	CourseCreator  string
}

func (opt CourseOptions) valid() bool {
	return opt.OrganizationID > 0 && opt.CourseCreator != ""
}
