package database

import (
	"errors"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// ErrEmptyNoteID is returned by GetNote when the query has no ID.
var ErrEmptyNoteID = errors.New("cannot get note with empty ID")

// GetNote returns the note with the ID set on the query.
func (db *GormDB) GetNote(query *qf.Note) (*qf.Note, error) {
	// Reject a zero ID: Gorm ignores zero-value fields in a struct query, so
	// Where(query).First would otherwise return an arbitrary first row.
	if query.GetID() == 0 {
		return nil, ErrEmptyNoteID
	}
	var note qf.Note
	if err := db.conn.Where(query).First(&note).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

// GetNotes returns all internal notes relevant to the given target, ordered by
// creation time. Exactly one of submissionID, groupID, or enrollmentID should
// be non-zero. When submissionID is set, the result also includes notes
// attached to the submission's group and the submitter's enrollment, so that a
// note written for any group member surfaces for the whole group.
func (db *GormDB) GetNotes(courseID, submissionID, groupID, enrollmentID uint64) ([]*qf.Note, error) {
	tx := db.conn.Where("course_id = ?", courseID)

	if submissionID > 0 {
		// Surface the submission's own notes plus its group and enrollment notes.
		groupIDs, enrollmentIDs, err := db.relatedTargets(courseID, submissionID)
		if err != nil {
			return nil, err
		}
		conditions := db.conn.Where("submission_id = ?", submissionID)
		if len(groupIDs) > 0 {
			conditions = conditions.Or("group_id IN ?", groupIDs)
		}
		if len(enrollmentIDs) > 0 {
			conditions = conditions.Or("enrollment_id IN ?", enrollmentIDs)
		}
		tx = tx.Where(conditions)
	} else if groupID > 0 {
		tx = tx.Where("group_id = ?", groupID)
	} else if enrollmentID > 0 {
		tx = tx.Where("enrollment_id = ?", enrollmentID)
	}

	var notes []*qf.Note
	if err := tx.Order("created_at").Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

// relatedTargets returns the group and enrollment IDs associated with a
// submission, used to surface group and enrollment notes during review.
func (db *GormDB) relatedTargets(courseID, submissionID uint64) (groupIDs, enrollmentIDs []uint64, _ error) {
	submission, err := db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return nil, nil, err
	}
	if uid := submission.GetUserID(); uid > 0 {
		// Individual submission: surface the submitter's enrollment and group notes.
		enrollment, err := db.GetEnrollmentByCourseAndUser(courseID, uid)
		if err == nil {
			enrollmentIDs = append(enrollmentIDs, enrollment.GetID())
			if gid := enrollment.GetGroupID(); gid > 0 {
				groupIDs = append(groupIDs, gid)
			}
		}
	}
	if gid := submission.GetGroupID(); gid > 0 {
		// Group submission: surface the group's notes and every member's enrollment notes.
		groupIDs = append(groupIDs, gid)
		group, err := db.GetGroup(gid)
		if err == nil {
			for _, enrollment := range group.GetEnrollments() {
				enrollmentIDs = append(enrollmentIDs, enrollment.GetID())
			}
		}
	}
	return groupIDs, enrollmentIDs, nil
}

// CreateNote adds a new internal note, stamping its creation time.
func (db *GormDB) CreateNote(note *qf.Note) error {
	note.CreatedAt = timestamppb.Now()
	note.EditedAt = note.GetCreatedAt()
	return db.conn.Create(note).Error
}

// UpdateNote updates the body of an existing internal note and bumps its edited time.
func (db *GormDB) UpdateNote(note *qf.Note) error {
	if note.GetID() == 0 {
		return gorm.ErrMissingWhereClause
	}
	return db.conn.Model(&qf.Note{ID: note.GetID()}).
		Select("Body", "EditedAt").
		Updates(&qf.Note{Body: note.GetBody(), EditedAt: timestamppb.Now()}).Error
}

// DeleteNote removes the internal note matching the given query.
func (db *GormDB) DeleteNote(note *qf.Note) error {
	if note.GetID() == 0 {
		return gorm.ErrMissingWhereClause
	}
	return db.conn.Delete(&qf.Note{}, note.GetID()).Error
}
