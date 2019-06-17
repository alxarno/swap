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

// const FilesDirPath string = "./files/"
const MinFilesDirPath string = "min/"

type settings struct {
	Backend struct {
		Host              string `json:"host"`
		SecretKeyForToken string `json:"secret_key_for_token"`
		FilesPath         string `json:"files_path"`
		FileLogs          bool   `json:"fileLogs"`
	} `json:"backend"`
	Cert struct {
		Org     string   `json:"org"`
		Hosts   []string `json:"hosts"`
		RsaBits int      `json:"rsa-bits"`
	} `json:"cert"`
	Service struct {
		MaxFileSize       int64 `json:"max_file_size_byte"`
		MaxMinutesForFile int64 `json:"max_minutes_available_for_files_download"`
		CORS              bool  `json:"cors"`
	} `json:"service"`
	DB struct {
		SQLite struct {
			Path string `json:"file_path"`
		} `json:"sqlite"`
	} `json:"db"`
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
			"fileLogs": true,
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
			"max_file_size_byte": 104857600,
			"max_minutes_available_for_files_download": 5,
			"cors": false
		},
		"db":{
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
	if _, err := os.Stat(ServiceSettings.Backend.FilesPath + MinFilesDirPath); os.IsNotExist(err) {
		os.MkdirAll(ServiceSettings.Backend.FilesPath+MinFilesDirPath, os.ModePerm)
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
