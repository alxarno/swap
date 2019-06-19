package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	swapcrypto "github.com/alxarno/swap/crypto"
	logger "github.com/alxarno/swap/logger"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/settings"
	engine "github.com/alxarno/swap/src/messages"
)

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

	tlsconfig := tls.Config{Certificates: []tls.Certificate{*swapcrypto.Cert}}
	tlsconfig.Rand = rand.Reader
	tlsserver := http.Server{
		TLSConfig: &tlsconfig,
		Handler:   router,
		Addr:      ":" + settings.ServiceSettings.Backend.SSLHost,
	}

	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+strings.Split(r.Host, ":")[0]+":"+settings.ServiceSettings.Backend.SSLHost+r.URL.String(), http.StatusPermanentRedirect)
	})

	httpserver := http.Server{
		Addr:    ":" + settings.ServiceSettings.Backend.Host,
		Handler: r,
	}

	go func() {
		if err := httpserver.ListenAndServe(); err != http.ErrServerClosed {
			logger.Logger.Printf("Listen Error HTTP: %v\n", err)
		}
	}()

	go func() {
		if err := tlsserver.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			logger.Logger.Printf("Listen Error TLS: %v\n", err)
		}
	}()

	printLogo()
	os.Stderr.WriteString("Swap started on \t" + myAddres + "\n")
	logger.Logger.Println("Swap started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	logger.Logger.Println("Swap shutdowned")

	httpctx, httpcancel := context.WithTimeout(context.Background(), 5*time.Second)
	httpsctx, httpscancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer httpcancel()
	defer httpscancel()
	if err = httpserver.Shutdown(httpctx); err != nil {
		logger.Logger.Printf("Could not gracefully shutdown the http server %v\n", err)
	}
	if err = tlsserver.Shutdown(httpsctx); err != nil {
		logger.Logger.Printf("Could not gracefully shutdown the https server %v\n", err)
	}

}
