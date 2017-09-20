package chat

import (
	"fmt"
	"net/http"
)

func hello()(string){
	return "hello"
}

func MainChatApi(var1 string, w http.ResponseWriter, r *http.Request){
	fmt.Println("User"+var1)
}
