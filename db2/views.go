package db2

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

func (c ChatUser) GetDeletePoints() ([][]int64, error) {
	var points [][]int64
	err := json.Unmarshal([]byte(c.DeletePoints), &points)
	if err != nil {
		return points, err
	}
	return points, nil
}

func (c *ChatUser) SetDeletePoints(data [][]int64) error {
	mData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.DeletePoints = string(mData)
	return err
}

func (u User) CheckPass(pass string) bool {
	h := sha256.New()
	h.Write([]byte(pass))
	return u.Pass == base64.StdEncoding.EncodeToString(h.Sum(nil))
}
