package user

import (
	"fmt"
	"net/http"
	"encoding/json"
	db_work "github.com/AlexArno/spatium/db_work"
	"github.com/robbert229/jwt"
	"time"
)
type ProveConnection struct{
	Login string
	Pass string
}
type userGetToken struct{
	line string
}

var (
	secret = "321312421"
)
func sendAnswerError(e_type string, w http.ResponseWriter){
	var answer = make(map[string]string)
	answer["result"] = "Error"
	answer["type"]=e_type
	finish, _:=json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func getJson(target interface{}, r*http.Request) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func enter( w http.ResponseWriter, r *http.Request){

	var data *ProveConnection
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	fmt.Println(data)
	if data == nil{
		sendAnswerError("Haven't all fields (login, pass)", w)
		return
	}
	now_user, err:= db_work.GetUser("login" , map[string]string{"login":data.Login, "pass":data.Pass})
	if err!=nil{
		sendAnswerError("User is undefined", w)
		return
	}
	if now_user == nil{
		sendAnswerError("User is undefined", w)
		return
	}
	algorithm :=  jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("id", &now_user.ID)
	claims.Set("time", time.Now().Local().Add(time.Hour*24*30).Unix())
	token, err := algorithm.Encode(claims)
	if err!=nil{
		sendAnswerError("Token is failed", w)
		fmt.Println(err)
		return
	}
	var x = make(map[string]string)
	x["token"]=token
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func proveToken(w http.ResponseWriter, r *http.Request){
	algorithm :=  jwt.HmacSha256(secret)

	token := new(userGetToken)
	getJson(token, r)
	claims, err := algorithm.Decode(token.line)
	if err != nil {
		sendAnswerError("Token is failed", w)
		return
	}
	id,err :=claims.Get("id")
	fmt.Println(id)
	if err != nil{
		fmt.Println(err)
	}
	now_user, err:= db_work.GetUser("id" , map[string]string{"id": "1"})
	if err!=nil{
		sendAnswerError("User is undefined", w)
		return
	}
	if now_user == nil{
		sendAnswerError("User is undefined", w)
		return
	}
	var finish = make(map[string]string)
	finish["result"]="Success"
	x, _:=json.Marshal(finish)
	fmt.Fprintf(w, string(x))
}

func MainUserApi(var1 string, w http.ResponseWriter, r *http.Request){
	fmt.Println(var1+"Hello")
	switch var1 {
		case "enter":
			fmt.Println("Switch works")
			enter(w, r)
	}
}
