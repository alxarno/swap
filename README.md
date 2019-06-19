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

Running 

```
go run main.go logo.go  router.go  router_without_ui.go
```

> Also you can start by fresh, with auto rebuild
>```
> go get github.com/pilu/fresh
> ./fresh.sh
>```

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


License
----
GPL-3.0