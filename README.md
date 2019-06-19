<img src="https://i.imgur.com/gtXdkk6.png" align="right" width="256"/>

# [Swap](https://github.com/alxarno/swap)


![](https://img.shields.io/badge/price-free-%235F2FE1.svg)
![](https://img.shields.io/badge/version-0.0.1-green.svg)

Instant messenger for local network 

> Current standalone application is messenger's server, there isn't [UI](https://github.com/alxarno/swap-ui)

## System Capabilities
* Working without a setup
* Auto producing self-signed SSL certificates
* Embedded database
* Files holding
* Declarative settings
* Logging to a file

## Messenger Capabilities
* User's sing in and sing up
* Creating chat
* Invite users to a chat
* Block users in the chat
* Leave chat
* Upload docs
* Instant messaging via WebSocket


## Installation and run

Getting repository 

```
go get github.com/alxarno/swap
```
> All deps are in `vendor` folder

Running only API server, without web UI.

For running with web UI look [below](#build)

```
go run main.go logo.go  router.go  router_without_ui.go
```

> Also you can start by fresh, with auto rebuild
>```
> go get github.com/pilu/fresh
> ./fresh.sh
>```

## Config

After starting swap server, it create `swap.json` file, if it doesnt exist. The `swap.json` contains [swap's settings](https://github.com/alxarno/swap/blob/master/settings/settings.go).

```
{
	"backend":{
		"fileLogs": true, // swap will logging to swap.log file, instead stdout
		"host": "80", // base host for http (will auto redirect to sslhost)
		"sslhost": "443", // https host
		"secret_key_for_token": "CGli0F5jNe3RhLCfVyEBTw==", // auto produced random string for creating some tokens
		"files_path": "./files/" // folder for containing files
	},
	"cert":{
		"org": "Example Co", // self-signed cert's company name 
		"hosts": ["192.168.1.38","localhost","127.0.0.1"], // self-signed cert's hosts
		"rsa-bits":2048 // key bits
	},
	"service":{
		"max_file_size_byte": 104857600, // max uploaded file size (1MB)
		"max_minutes_available_for_files_download": 5, // temporary donwload link time
		"cors": false // cross-domain requests
	},
	"db":{
		"sqlite":{
			"file_path": "swap.db" // DB file name
		}
	}
}
```

## Certificates

After starting swap server, it create `swap.crt` and `swap.key` files, if they don't exist.

They are generated based on information in `swap.json`

You can generate your own certificate and key, and put it like `swap.crt` and `swap.key`


## Build 

For build executable install [packr](https://github.com/gobuffalo/packr) and run
```
./build.sh
```

It download [UI](https://github.com/alxarno/swap-ui), make go classes by [packr](https://github.com/gobuffalo/packr) and compile all

> Now supporting building only for current OS

Result will appear in releases folder

## Built With
* [Go](https://github.com/golang/go)
* [GORM](https://github.com/jinzhu/gorm)
* [SQLite](https://www.sqlite.org/index.html)
* [Gorilla web toolkit](https://github.com/gorilla)
* [Packr ](https://github.com/gobuffalo/packr)

## Download
You can [download](https://github.com/alxarno/swap/releases) the latest portable version of Swap

License
----
GPL-3.0