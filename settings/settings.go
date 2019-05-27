package settings

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	ServiceSettings *settings
)

const fileName string = "./swap.json"
const FilesDirPath string = "./files/"
const MinFilesDirPath string = FilesDirPath + "min/"

type settings struct {
	Backend struct {
		Test              bool   `json:"test"`
		Host              string `json:"host"`
		SecretKeyForToken string `json:"secret_key_for_token"`
		FilesPath         string `json:"files_path"`
	} `json:"backend"`
	Cert struct {
		Org     string   `json:"org"`
		Hosts   []string `json:"hosts"`
		RsaBits int      `json:"rsa-bits"`
	} `json:"cert"`
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

func createConfig() (data []byte, err error) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return
	}
	randBytes := make([]byte, 16)
	_, err = rand.Read(randBytes)
	if err != nil {
		panic("Settings -> Cannot fill random bytes -> " + err.Error())
	}

	defaultConfig := fmt.Sprintf(`{
		"backend":{
			"test": false,
			"host": "3030",
			"secret_key_for_token": "%s",
			"files_path": "./files/"
		},
		"cert":{
			"org": "Example Co",
			"hosts": ["192.168.1.38","localhost","127.0.0.1"],
			"rsa-bits":2048
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
	}`, base64.StdEncoding.EncodeToString(randBytes))

	_, err = f.Write([]byte(defaultConfig))
	f.Close()
	if err != nil {
		log.Println("Cannot write config")
		return
	}

	if _, err := os.Stat(FilesDirPath); os.IsNotExist(err) {
		os.MkdirAll(MinFilesDirPath, os.ModePerm)
	} else {
		panic("Folder for docs already exist delete it or move")
	}
	b, err := ioutil.ReadFile(fileName)
	return b, err
}

func LoadSettings() error {
	//Read settings file
	b, err := ioutil.ReadFile(fileName) // just pass the file name
	if err != nil {
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			b, err = createConfig()
			if err != nil {
				return errors.New("Strange creating config problem -> " + err.Error())
			}
		} else {
			return errors.New("Strange config problem -> " + err.Error())
		}

	}
	err = json.Unmarshal(b, &ServiceSettings)
	if err != nil {
		fmt.Println("Config unmarshaling error")
		return err
	}
	return nil
}

func GetSettings() (*settings, error) {
	if ServiceSettings == nil {
		err := LoadSettings()
		if err != nil {
			return nil, err
		}
	}
	return ServiceSettings, nil
}
