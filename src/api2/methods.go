package api2

import (
	"net/http"
	"fmt"
	"encoding/json"
	"time"
	"github.com/robbert229/jwt"
	"github.com/Spatium-Messenger/Server/settings"
	"github.com/Spatium-Messenger/Server/db_api"
	"errors"
)

var secret = settings.ServiceSettings.Server.SecretKeyForToken

func sendAnswerError(eType string, errCode int, w http.ResponseWriter){
	var answer = make(map[string]interface{})
	answer["result"] = "Error"
	answer["code"] = errCode
	answer["type"]=eType
	finish, _:=json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func generateToken(id int64) (string,error){
	algorithm :=  jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("id", id)
	claims.Set("time", time.Now().AddDate(0,0,30).Unix())
	token, err := algorithm.Encode(claims); if err!=nil{
		return "",err
	}
	return token,nil
}

func getJson(target interface{}, r*http.Request) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func testUserToken(token string)(*db_api.User,error){
	algorithm :=  jwt.HmacSha256(secret)
	claims, err := algorithm.Decode(token);if err != nil {
		return nil, errors.New("token is wrong")
	}
	id,err :=claims.Get("id");if err!=nil{
		return nil,  errors.New("token is wrong")
	}
	tokenTime,err :=claims.Get("time");if err!=nil{
		return nil,  errors.New("token is wrong")
	}
	if tokenTime.(int64)<time.Now().Unix(){
		//u:=db_api.User{Id: id.(int64)}
		u,err:=db_api.GetUser("id",map[string]interface{}{"id":id});if err!=nil{
			return nil,err
		}
		return u,nil
	}
	return nil, errors.New("token time is done")
}

func getUserByToken(r *http.Request)(*db_api.User,error){
	var data struct{
		Token string`json:"token"`
	}
	err:=getJson(&data, r);if err!=nil{
		return nil,err
	}
	u,err:= testUserToken(data.Token);if err!=nil{
		return nil,err
	}
	return u,nil
}
