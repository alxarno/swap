package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func main() {
	_, err := settings.GetSettings()
	if err != nil {
		logger.Logger.Println("Settings loading fatal -> ", err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	logger.Init(settings.ServiceSettings.Backend.FileLogs)

	swapcrypto.InitCert()

	err = db.BeginDB(nil)
	if err != nil {
		logger.Logger.Println("DB setup fatal-> ", err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	engine.StartCoreMessenger()
	engine.ConnectActionsToDB()

	router := newRouter()
	myAddres := ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Logger.Print("Some network error -> ", err.Error())
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myAddres += ipnet.IP.String()
				myAddres += ":" + settings.ServiceSettings.Backend.Host + "/" + settings.ServiceSettings.Backend.SSLHost + "\t"
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
		Addr:      ":" + settings.ServiceSettings.Backend.SSLHost,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+strings.Split(r.Host, ":")[0]+":"+settings.ServiceSettings.Backend.SSLHost+r.URL.String(), http.StatusPermanentRedirect)
	})

	go logger.Logger.Fatal(http.ListenAndServe(":"+settings.ServiceSettings.Backend.Host, nil))

	logger.Logger.Fatal("ListenAndServeTLS: ", server.ListenAndServeTLS("", ""))

}
