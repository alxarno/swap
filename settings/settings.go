package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	ServiceSettings *Settings
)

const fileName string = "./swap.json"

type Settings struct {
	Backend struct {
		Test              bool   `json:"test"`
		Encryption        bool   `json:"encryption"`
		CertFile          string `json:"cert_file"`
		KeyFile           string `json:"key_file"`
		Host              string `json:"host"`
		SecretKeyForToken string `json:"secret_key_for_token"`
	} `json:"Backend"`
	Service struct {
		MaxFileSize int64 `json:"max_file_size_byte"`
	} `json:"service"`
	DB struct {
		DataBaseType string `json:"db_type"`
		SQLite       struct {
			Path string `json:"file_path"`
		} `json:"sqlite"`
	} `json:"db"`
}

func SetTestVar(test bool) {
	ServiceSettings.Backend.Test = test
}

func LoadSettings() error {
	//Read settings file
	b, err := ioutil.ReadFile(fileName) // just pass the file name
	if err != nil {
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			f, err := os.Create(fileName)
			if err != nil {
				fmt.Println("Create config error")
				return err
			}

			default_config := `{	"Backend": {		"encryption":false,		"cert_file": "",		"key_file": "",		"host": "1234",		"secret_key_for_token": "MY SECRET"	},	"service":{		"max_file_size_byte": 104857600	}}`

			_, err = f.Write([]byte(default_config))
			if err != nil {
				return err
			}
		}
		return err
	}
	err = json.Unmarshal(b, &ServiceSettings)
	if err != nil {
		fmt.Println("Config unmarshaling error")
		return err
	}
	return nil
}

func GetSettings() (*Settings, error) {
	if ServiceSettings == nil {
		err := LoadSettings()
		if err != nil {
			return nil, err
		}
	}
	return ServiceSettings, nil
}
