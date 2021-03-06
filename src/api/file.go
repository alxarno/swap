package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	db "github.com/alxarno/swap/db2"
	logger "github.com/alxarno/swap/logger"
	"github.com/alxarno/swap/settings"
	"github.com/robbert229/jwt"
)

type fileInfo struct {
	ratioSize float64
	token     string
	fileType  string
	chatID    int64
	duration  int64
	name      string
}
type fileInfoBuff struct {
	ratioSize string
	token     string
	fileType  string
	duration  string
	chatID    string
	name      string
}

func registerFileEndpoints(r *Router) {
	r.Route("/upload", uploadFile, "POST")
	r.Route("/{id:[0-9]+}", getFile, "GET")
	r.Route("/{id:[0-9]+}/delete", deleteFile, "DELETE")
	r.Route("/{id:[0-9]+}/link", getDisposableFileLink, "GET")
}

func rebuildFileDataTypes(buff fileInfoBuff) (fileInfo, error) {
	var res fileInfo

	rs, err := strconv.ParseFloat(buff.ratioSize, 64)
	if err != nil {
		return res, err
	}
	res.ratioSize = rs
	cID, err := strconv.ParseInt(buff.chatID, 10, 64)
	if err != nil {
		return res, err
	}
	res.chatID = cID
	res.name = buff.name
	res.fileType = buff.fileType
	duration, err := strconv.ParseInt(buff.duration, 10, 64)
	if err != nil {
		return res, err
	}
	res.duration = duration
	return res, nil
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	const ref string = "File upload API:"
	err := r.ParseMultipartForm(settings.ServiceSettings.Service.MaxFileSize)
	if err != nil {
		sendAnswerError(ref, err, "", failedDecodeFromData, 0, w)
		return
	}
	var buff fileInfoBuff
	buff.ratioSize = r.FormValue("ratio_size")
	buff.chatID = r.FormValue("chat_id")
	buff.name = r.FormValue("name")
	buff.fileType = r.FormValue("type")
	buff.duration = r.FormValue("duration")

	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, r.Header.Get("X-Auth-Token"), invalidToken, 1, w)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendAnswerError(ref, err, "", failedDecodeFromData, 2, w)
		return
	}
	defer file.Close()
	fD, err := rebuildFileDataTypes(buff)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("Data - %v", buff), failedRebuildDataTypes, 3, w)
		return
	}
	id, path, err := db.CreateFile(fD.name, handler.Size, user.ID, fD.chatID, fD.ratioSize, fD.duration)
	if err != nil {
		sendAnswerError(ref, err, "", failedCreatFile, 4, w)
		return
	}

	f, err := os.OpenFile(settings.ServiceSettings.Backend.FilesPath+path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		sendAnswerError(ref, err, "", failedOpenFile, 5, w)
		return
	}
	io.Copy(f, file)
	f.Close()
	compressionImage(fD.fileType, fD.ratioSize, path)
	var answer = struct {
		Result string `json:"result"`
		FileID int64  `json:"file_id"`
	}{
		Result: successResult,
		FileID: id,
	}
	logger.Logger.Printf("User %d uploaded file - %s, %d B \n", user.ID, path, handler.Size)
	finish, _ := json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	const ref string = "File delete API:"
	fileID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, r.Header.Get("X-Auth-Token"), invalidToken, 1, w)
		return
	}
	path, err := db.DeleteFile(fileID, user.ID)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d, fileID - %d", user.ID, fileID), failedDeleteFileDB, 2, w)
		return
	}

	defPath := settings.ServiceSettings.Backend.FilesPath
	err = os.Remove(defPath + path)
	if err != nil {
		sendAnswerError(ref, err, defPath+path, failedDeleteFileOS, 3, w)
		return
	}
	err = os.Remove(defPath + "min/" + path)
	if err != nil {
		sendAnswerError(ref, err, defPath+"min/"+path, failedDeleteFileOS, 4, w)
		return
	}

	var answer = struct {
		Result string `json:"result"`
	}{
		Result: successResult,
	}
	finish, _ := json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func getDisposableFileLink(w http.ResponseWriter, r *http.Request) {
	const ref string = "File disposable link API:"
	fileID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	path, err := db.CheckFileRights(user.ID, fileID)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d, fileID - %d", user.ID, fileID), haventRightsForAction, 2, w)
		return
	}
	sett, err := settings.GetSettings()
	if err != nil {
		sendAnswerError(ref, err, "", failedGetSettings, 3, w)
		return
	}
	timeoff := time.Now().Unix() + (60 * settings.ServiceSettings.Service.MaxMinutesForFile)
	secret := sett.Backend.SecretKeyForToken
	algorithm := jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("path", path)
	claims.Set("user_id", user.ID)
	claims.Set("time", timeoff)
	link, err := algorithm.Encode(claims)
	if err != nil {
		sendAnswerError(ref, err, "", failedEncodeData, 4, w)
		return
	}

	var answer = struct {
		Result  string `json:"result"`
		Link    string `json:"link"`
		TimeOff int64  `json:"timeoff"`
	}{
		Result:  successResult,
		Link:    link,
		TimeOff: timeoff,
	}

	finish, _ := json.Marshal(answer)
	fmt.Fprintf(w, string(finish))

}

func getFile(w http.ResponseWriter, r *http.Request) {
	const ref string = "File get file API:"
	var user *db.User
	min := (r.URL.Query().Get("min") == "")
	fileID := pageNumber(r, 1)
	user, err := UserByHeader(r)
	if err != nil {
		if user, err = UserByCookie(r); err != nil {
			sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
			return
		}
	}
	path, err := db.CheckFileRights(user.ID, fileID)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d, chatID - %d", user.ID, fileID), haventRightsForAction, 2, w)
		return
	}
	file := settings.ServiceSettings.Backend.FilesPath + path
	if min {
		file = settings.ServiceSettings.Backend.FilesPath + "min/" + path
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		sendAnswerError(ref, err, file, fileDoesntExist, 3, w)
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
