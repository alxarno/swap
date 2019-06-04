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
	E       string
	N       string
}

// type Encrypte

func encryptedMsg(msg string) (final models.EncryptedMessage, err error) {
	if err = json.Unmarshal([]byte(msg), &final); err != nil {
		return
	}
	return
	//
}

func systemMsg(msg string) (final SystemMessage, err error) {
	var userMessageSystem = struct {
		Type    string `json:"mtype"`
		Content struct {
			Type   string `json:"type"`
			Token  string `json:"token"`
			PubKey struct {
				E string `json:"e"`
				N string `json:"n"`
			} `json:"key"`
		} `json:"content"`
	}{}
	if err = json.Unmarshal([]byte(msg), &userMessageSystem); err != nil {
		return
	}
	if userMessageSystem.Content.Type == messageActionAuth {
		_, err = api.TestUserToken(userMessageSystem.Content.Token)
		if err != nil {
			fmt.Println(err)
			return
		}
		final.Action = messageActionAuth
		final.Payload = userMessageSystem.Content.Token
		final.E = userMessageSystem.Content.PubKey.E
		final.N = userMessageSystem.Content.PubKey.N
		return final, nil

	}
	return
}

func userMsg(msg string) (*models.NewMessageToUser, error) {
	message, err := newMessageAnother(msg)
	if err != nil {
		return nil, err
	}
	message.Time = time.Now().UnixNano() / 1000000
	return &message, nil

}
