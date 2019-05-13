package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	db "github.com/swap-messenger/swap/db2"
	"github.com/swap-messenger/swap/models"
	"github.com/swap-messenger/swap/settings"
	engine "github.com/swap-messenger/swap/src/messages"
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
	test := flag.Bool("test", false, "a bool")
	flag.Parse()
	_, err := settings.GetSettings()
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}
	if *test {
		settings.SetTestVar(true)
	} else {

	}

	// engine.ConnectActionsToDB()
	err = db.BeginDB()
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	engine.StartCoreMessenger(*test)

	router := newRouter()
	myAddres := ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	var clearIPs string
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myAddres += ipnet.IP.String()
				myAddres += ":" + settings.ServiceSettings.Backend.Host + "\t"
				if *test {
					clearIPs = ipnet.IP.String() + ":" + settings.ServiceSettings.Backend.Host
				}
			}
		}
	}

	// log.Println(models.MESSAGE_COMMAND_USER_CREATED_CHAT)

	if *test {
		os.Stderr.WriteString(clearIPs)
	} else {
		printLogo()
		os.Stderr.WriteString("Swap started on \t" + myAddres + "\n")
	}

	if settings.ServiceSettings.Backend.Encryption {
		log.Fatal("ListenAndServeTLS: ", http.ListenAndServeTLS(
			":"+settings.ServiceSettings.Backend.Host,
			settings.ServiceSettings.Backend.CertFile,
			settings.ServiceSettings.Backend.KeyFile,
			router))
	} else {
		log.Fatal("ListenAndServe: ", http.ListenAndServe(
			":"+settings.ServiceSettings.Backend.Host,
			router))
	}

}
