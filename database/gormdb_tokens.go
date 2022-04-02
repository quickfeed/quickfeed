package database

import pb "github.com/autograde/quickfeed/ag"

// GetTokenRecords returns all update token records
func (db *GormDB) GetTokenRecords() ([]*pb.UpdateTokenRecord, error) {
	var tokens []*pb.UpdateTokenRecord
	if err := db.conn.Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// CreateTokenRecord creates a new update token record
func (db *GormDB) CreateTokenRecord(query *pb.UpdateTokenRecord) error {
	return db.conn.Save(query).Error
}

// DeleteTokenRecord deletes an existing update token record
func (db *GormDB) DeleteTokenRecord(query *pb.UpdateTokenRecord) error {
	return db.conn.Delete(query).Error
}
