package admin

type validator interface {
	IsValid() bool
}

// IsValid checks whether OrgRequest fields are valid
func (req *OrgRequest) IsValid() bool {
	return req.GetOrgName() != ""
}

// IsValid checks that either ID or path field is set
func (org *Organization) IsValid() bool {
	id, path := org.GetID(), org.GetPath()
	return id > 0 || path != ""
}
