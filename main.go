package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	swapcrypto "github.com/alxarno/swap/crypto"
	logger "github.com/alxarno/swap/logger"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/models"
	"github.com/alxarno/swap/settings"
	engine "github.com/alxarno/swap/src/messages"
)

type proveConnection struct {
	Login string
	Pass  string
}

type requestGetMessage struct {
	Author string
	ChatID int64
}

type errorAnswer struct {
	Result string
	Type   string
}

type client chan<- models.NewMessageToUser

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// func redirectToHttps(w http.ResponseWriter, r *http.Request) {
// 	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the Backend from your local machine.
// 	http.Redirect(w, r, "https://192.168.56.1:1235"+r.RequestURI, http.StatusMovedPermanently)
// }

func main() {
	_, err := settings.GetSettings()
	if err != nil {
		fmt.Println("Settings -> ", err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	logger.Init(settings.ServiceSettings.Backend.FileLogs)

	swapcrypto.InitCert()

	err = db.BeginDB(nil)
	if err != nil {
		fmt.Println("DB -> ", err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	engine.StartCoreMessenger()
	engine.ConnectActionsToDB()

	router := newRouter()
	myAddres := ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myAddres += ipnet.IP.String()
				myAddres += ":" + settings.ServiceSettings.Backend.Host + "\t"
			}
		}
	}

	printLogo()
	os.Stderr.WriteString("Swap started on \t" + myAddres + "\n")
	logger.Logger.Println("Swap started ...")

	tlsconfig := tls.Config{Certificates: []tls.Certificate{*swapcrypto.Cert}}
	tlsconfig.Rand = rand.Reader
	server := http.Server{
		TLSConfig: &tlsconfig,
		Handler:   router,
		Addr:      ":" + settings.ServiceSettings.Backend.Host,
	}
	log.Fatal("ListenAndServeTLS: ", server.ListenAndServeTLS("", ""))

}
