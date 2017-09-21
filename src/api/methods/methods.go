package methods

import (
	"net/http"
	"encoding/json"
	"fmt"
	"time"
	db_work "github.com/AlexArno/spatium/db_work"
	"github.com/robbert229/jwt"
	"github.com/AlexArno/spatium/models"
)

type userGetToken struct{
	line string
}

func SendAnswerError(e_type string, w http.ResponseWriter){
	var answer = make(map[string]string)
	answer["result"] = "Error"
	answer["type"]=e_type
	finish, _:=json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func GetJson(target interface{}, r*http.Request) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func TestUserToken(secret string, token_line string)(*models.User,  string){
	algorithm :=  jwt.HmacSha256(secret)
	claims, err := algorithm.Decode(token_line)
	if err != nil {
		return nil, "Token is failed"
	}
	id,err :=claims.Get("id")
	if err != nil{
		return nil, "Token's id is undefined"
	}
	//We need id in string, because sql is necessary
	fl_id :=id.(string)
	//id_int64 := int64(fl_id)
	//u_id:= strconv.FormatInt(id_int64, 10)

	u_time,err :=claims.Get("time")
	if err != nil{
		return nil, "Token's time is undefined"
	}
	token_time,ok := u_time.(float64)
	token_time_int64 := int64(token_time)
	if ok != true{
		return nil, "Token's time is failed"
	}
	if token_time_int64<time.Now().Unix(){
		if err != nil {
			return nil, "Token's time is low"
		}
	}
	now_user, err:= db_work.GetUser("id" , map[string]string{"id": fl_id})
	if err!=nil{
		return nil, "Failed find user with token's id"
	}
	if now_user == nil{
		return nil, "User with token's id isn't found"
	}
	return now_user, ""
}