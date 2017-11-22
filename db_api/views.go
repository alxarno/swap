package db_api

import (
	"encoding/json"
)

func (c Chat_User) GetDeletePoints()([][]int64,error){
	var points [][]int64
	err:=json.Unmarshal([]byte(c.Delete_points), &points)
	if err!=nil{
		return points,err
	}
	return points,nil
}


