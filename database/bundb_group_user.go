package database

import (
	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// groupUserJoin is the bun join model for the Group-User many2many relation.
type groupUserJoin struct {
	bun.BaseModel `bun:"table:group_users,alias:gu"`
	GroupID       uint64    `bun:"group_id,pk"`
	Group         *qf.Group `bun:"rel:belongs-to,join:group_id=id"`
	UserID        uint64    `bun:"user_id,pk"`
	User          *qf.User  `bun:"rel:belongs-to,join:user_id=id"`
}
