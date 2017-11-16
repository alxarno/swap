package settings

import (
	"io/ioutil"
	"os"
	"fmt"
	"encoding/json"
)

var (
	ServiceSettings Settings
)

type Settings struct {
	Server struct{
		Encryption bool`json:"encryption"`
		Cert_file string`json:"cert_file"`
		Key_file string`json:"key_file"`
		Host string`json:"host"`
		SecretKeyForToken string`json:"secret_key_for_token"`
	}`json:"server"`
	Service struct{
		MaxFileSize int64 `json:"max_file_size_byte"`
	}`json:"service"`
	DB struct{
		DataBaseType string`json:"db_type"`
		SQLite struct{
			Path string`json:"file_path"`
		}`json:"sqlite"`
		Postgres struct{
			User string`json:"user"`
			Pass string`json:"pass"`
			Name string`json:"name"`
		}`json:"postgres"`
		Mysql struct{
			Url string`json:"url"`
		}`json:"mysql"`
	}`json:"db"`
}

func LoadSettings()(error){
	//Read settings file
	b, err := ioutil.ReadFile("./spatium_config.json") // just pass the file name
	if err != nil {
		if _, err := os.Stat("./spatium_config.json"); os.IsNotExist(err) {
			f, err := os.Create("./spatium_config.json")
			if err!=nil{
				fmt.Println("Create config error")
				return err
			}

			default_config := `{	"server": {		"encryption":false,		"cert_file": "",		"key_file": "",		"host": "1234",		"secret_key_for_token": "MY SECRET"	},	"service":{		"max_file_size_byte": 104857600	}}`

			_, err = f.Write([]byte(default_config))
			if err!=nil{
				return err
			}
		}
		return err
	}
	err = json.Unmarshal(b,&ServiceSettings)
	if err!=nil{
		fmt.Println("Config unmarshaling error")
		return err
	}
	return nil
}

func GetSettings()(Settings){
	return ServiceSettings
}