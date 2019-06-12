package messageengine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alxarno/swap/models"
	"github.com/alxarno/swap/src/api"
)

type SystemMessage struct {
	Action  string
	Payload string
}

func systemMsg(msg string) (final SystemMessage, err error) {
	var userMessageSystem = struct {
		Type    string `json:"mtype"`
		Content struct {
			Type  string `json:"type"`
			Token string `json:"token"`
		} `json:"content"`
	}{}
	if err = json.Unmarshal([]byte(msg), &userMessageSystem); err != nil {
		return
	}
	if userMessageSystem.Content.Type == messageActionAuth {
		_, err = api.TestUserToken(userMessageSystem.Content.Token)
		if err != nil {
			fmt.Println("systemMsg -> TestUserToken", err)
			return
		}
		final.Action = messageActionAuth
		final.Payload = userMessageSystem.Content.Token
		return final, nil

	}
	return
}

func userMsg(msg string) (*models.NewMessageToUser, error) {
	message, err := newMessageAnother(msg)
	if err != nil {
		return nil, err
	}
	message.Time = time.Now().UnixNano() / 1000000 // nanoSec -> miliSec
	return &message, nil

}
