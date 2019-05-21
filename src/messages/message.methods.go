package messageengine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/swap-messenger/swap/models"
	"github.com/swap-messenger/swap/src/api"
)

func systemMsg(msg string) (map[string]interface{}, error) {
	var final = make(map[string]interface{})
	var userMessageSystem = struct {
		Type    string
		Content struct {
			Type  string
			Token string
			Key   struct {
				e string
				n string
			}
		}
	}{}
	if err := json.Unmarshal([]byte(msg), &userMessageSystem); err != nil {
		return nil, err
	}
	if userMessageSystem.Content.Type == messageActionAuth {
		_, err := api.TestUserToken(userMessageSystem.Content.Token)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		final["Action"] = messageActionAuth
		final["Payload"] = userMessageSystem.Content.Token
		final["e"] = userMessageSystem.Content.Key.e
		final["n"] = userMessageSystem.Content.Key.n
		return final, nil

	}
	return nil, nil
}

func userMsg(msg string) (*models.NewMessageToUser, error) {
	message, err := newMessageAnother(msg)
	if err != nil {
		return nil, err
	}
	message.Time = time.Now().Unix()
	return &message, nil

}
