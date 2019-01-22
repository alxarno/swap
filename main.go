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

	db "github.com/Spatium-Messenger/Server/db"
	"github.com/Spatium-Messenger/Server/models"
	"github.com/Spatium-Messenger/Server/settings"

	engine "github.com/Spatium-Messenger/Server/src/message_engine"
	// "github.com/AlexeyArno/Gologer"
)

type ProveConnection struct {
	Login string
	Pass  string
}

type RequestGetMessage struct {
	Author  string
	Chat_Id float64
}

type ErrorAnswer struct {
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
// 	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the server from your local machine.
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
		//removeContents("./public/files")
		//os.Mkdir("./public/files/min", os.ModePerm)
	}

	err = db.BeginDB()
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	//go broadcaster()
	engine.StartCoreMessenger()

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
				myAddres += ":" + settings.ServiceSettings.Server.Host + "\t"
				if *test {
					clearIPs = ipnet.IP.String() + ":" + settings.ServiceSettings.Server.Host
				}
			}
		}
	}

	if *test {
		os.Stderr.WriteString(clearIPs)
	} else {
		printLogo()
		os.Stderr.WriteString("Swap started on \t" + myAddres + "\n")
	}

	if settings.ServiceSettings.Server.Encryption {
		log.Fatal("ListenAndServeTLS: ", http.ListenAndServeTLS(
			":"+settings.ServiceSettings.Server.Host,
			settings.ServiceSettings.Server.CertFile,
			settings.ServiceSettings.Server.KeyFile,
			router))
	} else {
		log.Fatal("ListenAndServe: ", http.ListenAndServe(
			":"+settings.ServiceSettings.Server.Host,
			router))
	}

}