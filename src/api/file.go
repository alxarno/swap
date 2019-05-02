package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/robbert229/jwt"
	"github.com/swap-messenger/Backend/db"
	"github.com/swap-messenger/Backend/settings"
)

type fileInfo struct {
	ratioSize float64
	token     string
	fileType  string
	chatId    int64
	name      string
}
type fileInfoBuff struct {
	ratioSize string
	token     string
	fileType  string
	chatId    string
	name      string
}

func rebuildFileDataTypes(buff fileInfoBuff) (fileInfo, error) {
	var res fileInfo

	rs, err := strconv.ParseFloat(buff.ratioSize, 64)
	if err != nil {
		return res, err
	}
	res.ratioSize = rs
	cid, err := strconv.ParseInt(buff.chatId, 10, 64)
	if err != nil {
		return res, err
	}
	res.chatId = cid
	res.name = buff.name
	res.fileType = buff.fileType
	return res, nil
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(settings.ServiceSettings.Service.MaxFileSize)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	var buff fileInfoBuff
	buff.ratioSize = r.FormValue("ratio_size")
	buff.token = r.FormValue("token")
	buff.chatId = r.FormValue("chat_id")
	buff.name = r.FormValue("name")
	buff.fileType = r.FormValue("type")

	// fmt.Println(buff)

	user, err := TestUserToken(buff.token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendAnswerError(err.Error(), 2, w)
		return
	}
	defer file.Close()
	fD, err := rebuildFileDataTypes(buff)
	if err != nil {
		sendAnswerError(err.Error(), 3, w)
		return
	}
	id, path, err := db.CreateFile(fD.name, handler.Size, user.Id, fD.chatId, fD.ratioSize)
	if err != nil {
		sendAnswerError(err.Error(), 4, w)
		return
	}

	f, err := os.OpenFile(settings.ServiceSettings.Backend.FilesPath+path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		sendAnswerError(err.Error(), 5, w)
		log.Println(err)
		return
	}
	io.Copy(f, file)
	f.Close()
	compressionImage(fD.fileType, fD.ratioSize, path)
	var x = make(map[string]string)
	x["result"] = "Success"
	x["FileId"] = strconv.FormatInt(id, 10)
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token  string `json:"token"`
		FileID int64  `json:"file_id,integer"`
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		sendAnswerError("Failed decode data", 0, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError("Failed decode data", 0, w)
		return
	}
	path, err := db.DeleteFile(user.Id, data.FileID)
	if err != nil {
		sendAnswerError("Failed delete from db", 0, w)
		return
	}
	err = os.Remove(settings.ServiceSettings.Backend.FilesPath + path)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = os.Remove(settings.ServiceSettings.Backend.FilesPath + "min/" + path)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	var x = make(map[string]string)
	x["result"] = "Success"
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func getDisposableFileLink(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token  string `json:"token"`
		FileID int64  `json:"file_id,integer"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		sendAnswerError("Failed decode r.Body", 0, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}
	path, err := db.CheckFileRights(user.Id, data.FileID)
	if err != nil {
		sendAnswerError(err.Error(), 2, w)
		return
	}
	sett, err := settings.GetSettings()
	if err != nil {
		sendAnswerError(err.Error(), 3, w)
		return
	}
	secret := sett.Backend.SecretKeyForToken
	algorithm := jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("path", path)
	claims.Set("user_id", user.Id)
	claims.Set("time", time.Now().Unix()+60)
	link, err := algorithm.Encode(claims)
	if err != nil {
		sendAnswerError(err.Error(), 4, w)
		fmt.Fprintf(w, "%s", "failed encode link")
		return
	}
	var x = make(map[string]string)
	x["link"] = link
	x["result"] = "Success"
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))

}

func getFile(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token  string `json:"token"`
		FileID int64  `json:"file_id,integer"`
		Min    bool   `json:"min"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		sendAnswerError("Failed decode data", 0, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}
	path, err := db.CheckFileRights(user.Id, data.FileID)
	if err != nil {
		sendAnswerError(err.Error(), 2, w)
		return
	}
	file := settings.ServiceSettings.Backend.FilesPath + path
	if data.Min {
		file = settings.ServiceSettings.Backend.FilesPath + "min/" + path
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		sendAnswerError(err.Error(), 3, w)
		return
		// if data.Min {
		// 	file = settings.ServiceSettings.Backend.FilesPath + path
		// 	if _, err := os.Stat(file); os.IsNotExist(err) {
		// 		log.Println(err, 3)
		// 		return
		// 	}
		// }
	}
	http.ServeFile(w, r, file)
	return
}

func FileApi(var1 string, w http.ResponseWriter, r *http.Request) {
	switch var1 {
	case "uploadFile":
		uploadFile(w, r)
	case "deleteFile":
		deleteFile(w, r)
	case "getFile":
		getFile(w, r)
	case "getFileLink":
		getDisposableFileLink(w, r)
	default:
		sendAnswerError("Not found", 0, w)
	}
}
