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
	"reflect"
	//"strconv"
)

func getToken()(string,error){
	secret,err := settings.GetSettings();if err!=nil{
		return "",err
	}
	return secret.Server.SecretKeyForToken,nil
}

func sendAnswerError(eType string, errCode int, w http.ResponseWriter){
	var answer = make(map[string]interface{})
	answer["result"] = "Error"
	answer["code"] = errCode
	answer["type"]=eType
	finish, _:=json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func sendAnswerSuccess(w http.ResponseWriter){
	var x = make(map[string]string)
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func generateToken(id int64) (string,error){
	secret,err:= getToken();if err!=nil{
		return "",err
	}
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

func TestUserToken(token string)(*db_api.User,error){
	secret,err:= getToken();if err!=nil{
		return nil,err
	}
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

	if int64(tokenTime.(float64))>time.Now().Unix(){
		//u:=db_api.User{Id: id.(int64)}
		u,err:=db_api.GetUser("id",map[string]interface{}{"id":int64(id.(float64))});if err!=nil{
			return nil,err
		}
		return u,nil
	}
	//fmt.Println(int64(tokenTime.(float64)))
	//fmt.Println(time.Now().Unix())
	return nil, errors.New("token time is done")
}

func getUserByToken(r *http.Request)(*db_api.User,error){
	var data struct{
		Token string`json:"token"`
	}
	err:=getJson(&data, r);if err!=nil{
		return nil,err
	}
	u,err:= TestUserToken(data.Token);if err!=nil{
		return nil,err
	}
	return u,nil
}


//This function need for transform receive body of already unmarshal json (strings, and float64) to
// (strings, and int64), because json support only float64 values and write every
//time small convert code is laziness...
func TypeChanger(receiver interface{}, sender interface{}){
	for i:=0;i<reflect.TypeOf(receiver).NumField();i++{
		switch reflect.ValueOf(receiver).FieldByIndex([]int{i}).Kind(){
		case reflect.Float64:
			rField :=reflect.ValueOf(sender).Elem().FieldByIndex([]int{i})
			v:=int64(reflect.ValueOf(receiver).FieldByIndex([]int{i}).Float())
			if rField.IsValid() {
				rField.SetInt(v)

			}
		case reflect.String:
			rField :=reflect.ValueOf(sender).Elem().FieldByIndex([]int{i})
			v:=reflect.ValueOf(receiver).FieldByIndex([]int{i}).String()
			if rField.IsValid() {
				rField.SetString(v)
			}
		case reflect.Bool:
			rField :=reflect.ValueOf(sender).Elem().FieldByIndex([]int{i})
			v:=reflect.ValueOf(receiver).FieldByIndex([]int{i}).Bool()
			if rField.IsValid() {
				rField.SetBool(v)
			}
		case reflect.Slice:
			rField :=reflect.ValueOf(sender).Elem().FieldByIndex([]int{i})
			v:=reflect.ValueOf(receiver)
			slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(int64(0))),  v.Len(),  v.Len())
			rField.Set(slice)
			for i := 0; i < v.Len(); i++ {
				if rField.IsValid() {
					rField.Index(i).SetInt(int64(v.Index(i).Float()))
				}
			}
		default:

		}
	}
}

