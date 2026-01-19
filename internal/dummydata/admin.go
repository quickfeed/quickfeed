package dummydata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) admin(username string) error {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s", username))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Failed to fetch user info. Wrong username")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var info *struct {
		Login     string `json:"login"`
		ID        uint64 `json:"id"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return err
	}
	return g.db.CreateUser(&qf.User{
		ID:          1,
		Login:       info.Login,
		Name:        info.Name,
		ScmRemoteID: info.ID,
		AvatarURL:   info.AvatarURL,
		Email:       fmt.Sprintf("%s@gmail.com", username),
		StudentID:   "999999",
	})
}
