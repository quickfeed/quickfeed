package qf

func (cs *CourseSubmissions) For(id uint64) []*Submission {
	submissions := cs.GetSubmissions()
	if submissions == nil {
		return nil
	}
	return submissions[id].GetSubmissions()
}
