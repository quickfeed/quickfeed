package tokens

import (
	"fmt"
	"time"
)

// UpdateRequired returns true if JWT update is needed for this user ID
// because the user's role has changed or the JWT is about to expire.
func (tm *TokenManager) UpdateRequired(claims *Claims) bool {
	for _, token := range tm.updateTokens {
		if claims.UserID == token {
			return true
		}
	}
	return claims.ExpiresAt-time.Now().Unix() < int64(refreshTime.Seconds())
}

// Update removes user ID from the manager and updates user record in the database
func (tm *TokenManager) Remove(userID uint64) error {
	if !tm.exists(userID) {
		return nil
	}
	if err := tm.update(userID, false); err != nil {
		return err
	}
	var updatedTokenList []uint64
	for _, id := range tm.updateTokens {
		if id != userID {
			updatedTokenList = append(updatedTokenList, id)
		}
	}
	tm.updateTokens = updatedTokenList
	return nil
}

// Add adds a new UserID to the manager and updates user record in the database
func (tm *TokenManager) Add(userID uint64) error {
	if tm.exists(userID) {
		return nil
	}
	if err := tm.update(userID, true); err != nil {
		return err
	}
	tm.updateTokens = append(tm.updateTokens, userID)
	return nil
}

// updateTokenList fetches IDs of users who need token updates from the database
func (tm *TokenManager) updateTokenList() error {
	users, err := tm.db.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to update JWT tokens from database: %w", err)
	}
	var tokens []uint64
	for _, user := range users {
		if user.UpdateToken {
			tokens = append(tokens, user.ID)
		}
	}
	tm.updateTokens = tokens
	return nil
}

// update changes the status of token update of a user record in the database.
func (tm *TokenManager) update(userID uint64, updateToken bool) error {
	user, err := tm.db.GetUser(userID)
	if err != nil {
		return err
	}
	user.UpdateToken = updateToken
	return tm.db.UpdateUser(user)
}

// exists checks if the ID is in the list.
func (tm *TokenManager) exists(id uint64) bool {
	for _, token := range tm.updateTokens {
		if id == token {
			return true
		}
	}
	return false
}
