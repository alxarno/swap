package db_api

import (
	"encoding/json"
	"crypto/sha256"
	"encoding/base64"
)

func (c chatUser) GetDeletePoints()([][]int64,error){
	var points [][]int64
	err:=json.Unmarshal([]byte(c.Delete_points), &points)
	if err!=nil{
		return points,err
	}
	return points,nil
}

func (c chatUser) SetDeletePoints(data [][]int64)(error){
	m_data,err:= json.Marshal(data);if err!=nil{
		return err
	}
	c.Delete_points = string(m_data)
	return err
}

func (u User) CheckPass(pass string) bool{
	h := sha256.New()
	h.Write([]byte(pass))
	return u.Pass ==  base64.StdEncoding.EncodeToString(h.Sum(nil))
}




