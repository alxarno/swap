package api2

import (
	"net/http"
	"github.com/Spatium-Messenger/Server/settings"
	"fmt"
	"github.com/Spatium-Messenger/Server/db_api"
	"strconv"
	"os"
	"io"
	"encoding/json"
	"time"
	"github.com/robbert229/jwt"
)

type fileInfo struct{
	ratioSize float64
	token string
	fileType string
	chatId int64
	name string
}
type fileInfoBuff struct{
	ratioSize string
	token string
	fileType string
	chatId string
	name string
}

func rebuildFileDataTypes(buff fileInfoBuff) (fileInfo,error){
	var res fileInfo

	rs,err := strconv.ParseFloat(buff.ratioSize,64);if err !=nil{
		return res,err
	}
	res.ratioSize = rs
	cid,err:=strconv.ParseInt(buff.chatId,10,64);if err!=nil{
		return res,err
	}
	res.chatId = cid
	return res,nil
}

func uploadFile(w http.ResponseWriter, r*http.Request){
	r.ParseMultipartForm(settings.ServiceSettings.Service.MaxFileSize)
	var buff fileInfoBuff
	buff.ratioSize = r.FormValue("ratio_size")
	buff.token = r.FormValue("token")
	buff.chatId = r.FormValue("chat_id")
	buff.name = r.FormValue("fileName")
	buff.fileType = r.FormValue("type")

	user,err:= TestUserToken(buff.token);if err!=nil{
		sendAnswerError(err.Error(),0,w);return
	}
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		sendAnswerError(err.Error(),0, w);return
	}
	defer file.Close()
	fD,err:=rebuildFileDataTypes(buff);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	id,path,err:=db_api.CreateFile(fD.name, handler.Size, user.Id,fD.chatId,fD.ratioSize);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	f, err := os.OpenFile("./public/files/"+path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		sendAnswerError(err.Error(), 0,w)
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	go compressionImage(fD.fileType,fD.ratioSize,path)
	var x = make(map[string]string)
	x["result"]="Success"
	x["FileId"]= strconv.FormatInt(id,10)
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func deleteFile(w http.ResponseWriter,r *http.Request){
	var rData struct{Token string`json:"token"`; FileId float64`json:"file_id"`}
	var data struct{Token string`json:"token"`; FileId int64`json:"file_id"`}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&rData);if err != nil {
		sendAnswerError("Failed decode data",0, w);return
	}
	TypeChanger(rData,&data)
	user,err:= TestUserToken(data.Token);if err!=nil{
		sendAnswerError("Failed decode data",0, w);return
	}
	path,err:= db_api.DeleteFile(user.Id, data.FileId);if err!=nil{
		sendAnswerError("Failed delete from db",0, w);return
	}
	err = os.Remove("./public/files/"+path);if err != nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err = os.Remove("./public/files/min/"+path);if err != nil{
		sendAnswerError(err.Error(),0, w);return
	}
	var x = make(map[string]string)
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func getDisposableFileLink(w http.ResponseWriter, r *http.Request){
	var rData struct{
		Token string`json:"token"`;
		FileId float64`json:"file_id"`}
	var data struct{
		Token string`json:"token"`;
		FileId int64`json:"file_id"`}
	decoder:= json.NewDecoder(r.Body);defer r.Body.Close()
	err := decoder.Decode(&rData);if err != nil {
		sendAnswerError("Failed decode r.Body",0, w);return
	}
	TypeChanger(rData, &data)
	user,err:= TestUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	path,err := db_api.CheckFileRights(user.Id, data.FileId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	sett,err:= settings.GetSettings();if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	secret := sett.Server.SecretKeyForToken
	algorithm :=  jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("path", path)
	claims.Set("user_id", user.Id)
	claims.Set("time", time.Now().Unix()+60)
	link, err := algorithm.Encode(claims)
	if err!=nil{
		fmt.Fprintf(w, "%s","failed encode link");return
	}
	var x = make(map[string]string)
	x["link"]=link
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))

}

func getFile(w http.ResponseWriter, r *http.Request){
	var rData struct{
		Token string`json:"token"`
		FileId float64`json:"file_id"`
		Min bool`json:"min"`}
	var data struct{
		Token string`json:"token"`
		FileId int64`json:"file_id"`
		Min bool`json:"min"`}
	decoder:= json.NewDecoder(r.Body);defer r.Body.Close()
	err := decoder.Decode(&rData);if err != nil {
		sendAnswerError("Failed decode data",0, w);return
	}
	TypeChanger(rData,&data)
	user,err:= TestUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	path,err := db_api.CheckFileRights(user.Id, data.FileId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	file := "./public/files/"+path
	if data.Min{
		file =	"./public/files/min/"+path
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if data.Min {
			file = "./public/files/"+path
			if _, err := os.Stat(file); os.IsNotExist(err) {return}
		}
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
	}
}