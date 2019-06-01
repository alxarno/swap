package db2

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

func (c ChatUser) getDeletePoints() ([][]int64, error) {
	var points [][]int64
	err := json.Unmarshal([]byte(c.DeletePoints), &points)
	if err != nil {
		return points, err
	}
	return points, nil
}

func (c *ChatUser) setDeletePoints(data [][]int64) error {
	mData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.DeletePoints = string(mData)
	return err
}

func (u User) checkPass(pass string) bool {
	h := sha256.New()
	h.Write([]byte(pass))
	return u.Pass == base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func intToString(a *[]int64) string {
	b := make([]string, len(*a))
	for i, v := range *a {
		b[i] = strconv.FormatInt(v, 10)
	}

	return strings.Join(b, ",")
}
