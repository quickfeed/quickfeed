package models

// User represents a user account.
type User struct {
	ID uint64

	RemoteIdentities []RemoteIdentity `json:"remoteidentities,omitempty"`

	Courses []Course `gorm:"many2many:user_courses;"`
}

// RemoteIdentity represents a third-party identity which can be attached to a
// user account.
type RemoteIdentity struct {
	ID uint64

	// TODO: Provider + RemoteID = key.
	Provider string
	RemoteID uint64

	AccessToken string `json:"-"`

	UserID uint64
}

// db.Model(&course).Related(&users)
//// SELECT * FROM "users" INNER JOIN "user_courses" ON "user_courses"."user_id" = "users"."id" WHERE ("user_courses"."course_id" IN ('111'))
