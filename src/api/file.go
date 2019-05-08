package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/robbert229/jwt"
	db "github.com/swap-messenger/swap/db2"
	"github.com/swap-messenger/swap/settings"
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
	const ref string = "File upload API:"
	err := r.ParseMultipartForm(settings.ServiceSettings.Service.MaxFileSize)
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_DECODE_FORM_DATA, 0, w)
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
		sendAnswerError(ref, err, buff.token, INVALID_TOKEN, 1, w)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_GET_DATA_FROM_FORM, 2, w)
		return
	}
	defer file.Close()
	fD, err := rebuildFileDataTypes(buff)
	if err != nil {
		sendAnswerError(ref, err, buff, FAILED_REBUILD_DATATYPES, 3, w)
		return
	}
	id, path, err := db.CreateFile(fD.name, handler.Size, user.ID, fD.chatId, fD.ratioSize)
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_CREATE_FILE, 4, w)
		return
	}

	f, err := os.OpenFile(settings.ServiceSettings.Backend.FilesPath+path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_OPEN_FILE, 5, w)
		return
	}
	io.Copy(f, file)
	f.Close()
	compressionImage(fD.fileType, fD.ratioSize, path)
	var x = make(map[string]string)
	x["result"] = successResult
	x["FileId"] = strconv.FormatInt(id, 10)
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	const ref string = "File delete API:"
	var data struct {
		Token  string `json:"token"`
		FileID int64  `json:"file_id,integer"`
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}
	path, err := db.DeleteFile(data.FileID, user.ID)
	if err != nil {
		sendAnswerError(ref, err, map[string]interface{}{"userID": user.ID, "fileID": data.FileID}, FAILED_DELETE_FILE_DB, 2, w)
		return
	}

	defPath := settings.ServiceSettings.Backend.FilesPath
	err = os.Remove(defPath + path)
	if err != nil {
		sendAnswerError(ref, err, defPath+path, FAILED_DELETE_FILE_OS, 3, w)
		return
	}
	err = os.Remove(defPath + "min/" + path)
	if err != nil {
		sendAnswerError(ref, err, defPath+"min/"+path, FAILED_DELETE_FILE_OS, 4, w)
		return
	}
	var x = make(map[string]string)
	x["result"] = successResult
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func getDisposableFileLink(w http.ResponseWriter, r *http.Request) {
	const ref string = "File disposable link API:"
	var data struct {
		Token  string `json:"token"`
		FileID int64  `json:"file_id,integer"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}
	path, err := db.CheckFileRights(user.ID, data.FileID)
	if err != nil {
		sendAnswerError(ref, err, map[string]interface{}{"userID": user.ID, "fileID": data.FileID}, HAVENT_RIGHTS_FOR_ACTION, 2, w)
		return
	}
	sett, err := settings.GetSettings()
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_GET_SETTINGS, 3, w)
		return
	}
	secret := sett.Backend.SecretKeyForToken
	algorithm := jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("path", path)
	claims.Set("user_id", user.ID)
	claims.Set("time", time.Now().Unix()+60)
	link, err := algorithm.Encode(claims)
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_ENCODE_DATA, 4, w)
		return
	}
	var x = make(map[string]string)
	x["link"] = link
	x["result"] = successResult
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))

}

func getFile(w http.ResponseWriter, r *http.Request) {
	const ref string = "File get file API:"
	var data struct {
		Token  string `json:"token"`
		FileID int64  `json:"file_id,integer"`
		Min    bool   `json:"min"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}
	path, err := db.CheckFileRights(user.ID, data.FileID)
	if err != nil {
		sendAnswerError(ref, err, map[string]interface{}{"userID": user.ID, "fileID": data.FileID}, HAVENT_RIGHTS_FOR_ACTION, 2, w)
		return
	}
	file := settings.ServiceSettings.Backend.FilesPath + path
	if data.Min {
		file = settings.ServiceSettings.Backend.FilesPath + "min/" + path
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		sendAnswerError(ref, err, file, FILE_DOESNT_EXIST, 3, w)
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
		sendAnswerError("File API Router", nil, nil, END_POINT_NOT_FOUND, 0, w)
	}
}
