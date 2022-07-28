package database

import (
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/quickfeed/quickfeed/qf"
	"golang.org/x/crypto/bcrypt"
)

// CreateAPIKey creates a new API key.
func (db *GormDB) CreateApplication(app *qf.Application) (*qf.Application, error) {
	// Generate unique key.
	uniqueID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	if uniqueID.String() == "" {
		return nil, fmt.Errorf("could not generate unique ID")
	}
	// Hash the unique ID.
	hash, err := bcrypt.GenerateFromPassword([]byte(uniqueID.String()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate a 32-byte hex string.
	randID := rand.Uint32()
	base64ID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", randID)))

	app.Secret = base64.StdEncoding.EncodeToString(hash)
	app.ClientID = base64ID
	if err := db.conn.Create(&app).Error; err != nil {
		return nil, err
	}
	app.Secret = uniqueID.String()
	return app, nil
}

func (db *GormDB) GetApplication(userID uint64, clientID, secret string) (*qf.Application, error) {
	var app qf.Application

	if err := db.conn.Where("client_id = ? AND user_id = ?", clientID, userID).First(&app).Error; err != nil {
		return nil, err
	}
	if app.Secret == "" {
		return nil, fmt.Errorf("application not found")
	}

	appSecret, err := base64.StdEncoding.DecodeString(app.Secret)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(appSecret, []byte(secret)); err != nil {
		return nil, err
	}
	return &qf.Application{
		Name:        app.Name,
		UserID:      app.UserID,
		ClientID:    app.ClientID,
		Description: app.Description,
	}, nil
}

func (db *GormDB) GetApplications(userID uint64) (*qf.Applications, error) {
	var apps []*qf.Application
	if err := db.conn.Where("user_id = ?", userID).Find(&apps).Error; err != nil {
		return nil, err
	}
	for _, app := range apps {
		app.Secret = ""
	}
	return &qf.Applications{
		Applications: apps,
	}, nil
}

func (db *GormDB) DeleteApplication(app *qf.Application) error {
	return db.conn.Delete(app).Error
}
