package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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
		PubKey            string `json:"public_key"`
		PrivateKey        string `json:"private_key"`
		Host              string `json:"host"`
		SecretKeyForToken string `json:"secret_key_for_token"`
		FilesPath         string `json:"files_path"`
	} `json:"backend"`
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
				log.Println("Create config error")
				return err
			}

			defaultConfig := `{
				"backend":{
					"test": false,
					"encryption":false,
					"public_key": "./id_rsa.pub",
					"private_key": "./id_rsa",
					"host": "3030",
					"secret_key_for_token": "MY SECRET",
					"files_path": "./public/files/"
				},
				"service":{
					"max_file_size_byte": 104857600
				},
				"db":{
					"db_type":"l",
					"sqlite":{
						"file_path": "swap.db"
					}
				}
			}`

			_, err = f.Write([]byte(defaultConfig))
			f.Close()
			if err != nil {
				log.Println("Cannot write config")
				return err
			}

			os.RemoveAll("./public/files/")
			os.Mkdir("./public", os.ModePerm)
			os.Mkdir("./public/files", os.ModePerm)
			os.Mkdir("./public/files/min", os.ModePerm)
			b, err = ioutil.ReadFile(fileName)
		} else {
			return errors.New("Strange config problem: " + err.Error())
		}

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
