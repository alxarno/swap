package spatium_db_work

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"crypto/sha256"
	"os"
	"fmt"
	models "github.com/AlexArno/spatium/models"
	"time"
	"errors"
	"encoding/json"
	"strconv"
	"strings"
)
var (
	activeConn *sql.DB
	activeConnIsReal bool
)


func GetInfo() string{
	return "Info"
}

func GetUser(s_type string, data map[string]string)(*models.User, error){
	user := new(models.User)
	if !activeConnIsReal{
		OpenDB()
	}
	if s_type == "login"{
		rows, err := activeConn.Prepare("SELECT id, login, pass, u_name FROM people WHERE (login=?) AND (pass=?)")
		if err != nil {
			panic(nil)
		}
		//make hash of user's password
		h := sha256.New()
		h.Write([]byte(data["pass"]))
		query := rows.QueryRow(data["login"], h.Sum(nil))

		err = query.Scan(&user.ID, &user.Login, &user.Pass, &user.Name)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		return user,nil
	}else{
		rows, err := activeConn.Prepare("SELECT id, login, u_name FROM people WHERE id=?")
		if err != nil {
			panic(nil)
		}
		query := rows.QueryRow(data["id"])
		fmt.Println(data["id"])
		err = query.Scan(&user.ID, &user.Login, &user.Name)
		if err == sql.ErrNoRows{
			return nil, err
		}
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		return user,nil
	}

}

func CreateUser(login string, pass string, u_name string)(string, string, error){
	if !activeConnIsReal{
		OpenDB()
	}
	//test for equals logins
	var id_now string
	rows, err := activeConn.Prepare("SELECT id FROM people WHERE login=?")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(login).Scan(&id_now)
	defer rows.Close()
	if query != sql.ErrNoRows{
		return "","Login is busy",err
	}

	statement, err := activeConn.Prepare("INSERT INTO people (login, pass, u_name) VALUES (?, ?, ?)")
	if err != nil {
		return "","DB failed query",err
	}
	//make hash of user's password
	h := sha256.New()
	h.Write([]byte(pass))
	statement.Exec(login, h.Sum(nil), u_name)
	rows, err = activeConn.Prepare("SELECT id FROM people WHERE login=?")
	if err != nil {
		return "","DB failed query",err
	}
	query = rows.QueryRow(login).Scan(&id_now)
	if query == sql.ErrNoRows{
		return "","Some is fail",err
	}
	return id_now,"Success", nil
}

func InsertUserInChat(user_id string, chat_id int64)( error){
	if !activeConnIsReal{
		OpenDB()
	}
	var id_now string
	rows, err := activeConn.Prepare("SELECT chat_id FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		return errors.New("Cant prove user isnt in chat")
		//panic(nil)
	}
	query := rows.QueryRow(user_id, chat_id).Scan(&id_now)
	defer rows.Close()
	if query != sql.ErrNoRows{
		return errors.New("User already in chat")
	}
	statement, err := activeConn.Prepare("INSERT INTO people_in_chats (user_id, chat_id, blocked, start, deltimes) VALUES (?, ?, ?, ?,?)")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password

	deltime:= [1][1]int64{}
	deltime[0][0] = 0
	s_deltime,_:= json.Marshal(deltime)
	statement.Exec(user_id, chat_id, 0, time.Now().Unix()-1,string(s_deltime))
	statement, err = activeConn.Prepare("UPDATE chats SET lastmodify=? WHERE id=?")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password
	statement.Exec(time.Now().Unix(), chat_id)

	return nil
}

func CreateChat(name string, author_id string)(string,  error){
	if !activeConnIsReal{
		OpenDB()
	}
	statement, err := activeConn.Prepare("INSERT INTO chats (name,  author_id,moders_ids, lastmodify) VALUES (?, ?, ?, ?)")
	if err != nil {
		return "",errors.New("Failed permanent statement")
	}
	//make hash of user's password
	res, err := statement.Exec(name,  author_id,"[]", time.Now().Unix())
	if err != nil {
		return "",errors.New("Failed exec statement")
	}
	id, _ := res.LastInsertId()
	err = InsertUserInChat(author_id, id)
	if err != nil {
		return "",err
		//fmt.Println(fin)
	}
	mess_mss := "Я создал этот чат"
	docs := []string{}
	m_type := "a_msg"
	mess := models.MessageContent{&mess_mss, &docs, &m_type}
	data ,err := json.Marshal(mess)
	if err != nil{
		return "", err
	}
	f_id,err := strconv.ParseFloat(author_id, 64)
	if err != nil{
		return "", err
	}
	err = AddMessage(f_id, float64(id), string(data))
	if err != nil{
		return "", err
	}
	return string(id), nil

}

func GetMyChats(user_id float64)([]*models.UserChatInfo, error){
	var chats_ids []*models.UserChatInfo
	var middle []map[string]string
	rows, err := activeConn.Query("SELECT chats.id, chats.name, chats.author_id, chats.moders_ids, people_in_chats.delete_a, people_in_chats.deltime  FROM people_in_chats INNER JOIN chats ON people_in_chats.chat_id = chats.id WHERE user_id=?", user_id)
	if err != nil {
		fmt.Println("Outside", err)
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id, name, un_moders string
		var author_id string
		var delete_a,deltime int64
		//var moders []string
		if err := rows.Scan(&id,  &name, &author_id, &un_moders, &delete_a, &deltime); err != nil {
			fmt.Println("scan 1")
			return nil,err
		}
		middle=append(middle, map[string]string{"id": id, "name": name, "author": author_id, "moders": un_moders,
		"delete": strconv.FormatInt(delete_a,10), "deltime": strconv.FormatInt(deltime,10)})


	}
	for _,i := range middle{
		var author_name, content, msg_time string
		i_delete,_ := strconv.ParseInt(i["delete"],10,64)
		//deltime, _:= strconv.ParseInt(i["deltime"],10,64)
		message, err := activeConn.Query("SELECT  messages.content, people.u_name, messages.time  FROM messages INNER JOIN people ON messages.user_id = people.id WHERE chat_id=? ORDER BY time DESC", i["id"])
		if i_delete  == 1{
			message, err = activeConn.Query("SELECT  messages.content, people.u_name, messages.time FROM messages INNER JOIN people ON messages.user_id = people.id WHERE (chat_id=?) and (messages.time<?) ORDER BY time DESC", i["id"], i["deltime"])
		}
		if err != nil {
			fmt.Println("Inside")
			return nil,err
		}
		//query := message.QueryRow(i["id"])
		message.Next()
		err = message.Scan(&content, &author_name, &msg_time)
		if err == sql.ErrNoRows{
			//return nil, err
			msg_time = ""
			content = ""
			author_name = ""
		}
		f_id,err := strconv.ParseFloat(i["id"], 64)
		if err != nil {
			return nil, err
		}
		i_time,err := strconv.ParseInt(msg_time, 10,64)
		if err != nil {
			return nil, err
		}
		f_a_id,err := strconv.ParseFloat(i["author"], 64)
		if err != nil {
			fmt.Println("f_a_id")
			return nil, err
		}
		var m_content models.MessageContent
		var moders []float64
		err = json.Unmarshal([]byte(i["moders"]), &moders)
		if err!=nil{
			fmt.Println("moders")
			return nil, err
		}
		err = json.Unmarshal([]byte(content), &m_content)
		if err!=nil{
			fmt.Println("content")
			return nil, err
		}

		chats_ids=append(chats_ids, &models.UserChatInfo{f_id,i["name"], author_name, f_a_id, moders,&m_content,i_time,0,i_delete})
		defer message.Close()
		//chats_ids
	}
	if err := rows.Err(); err != nil {
		return nil,err
	}
	return chats_ids, nil
}

func AddMessage(user_id float64, chat_id float64, content string)(error){
	if !activeConnIsReal{
		OpenDB()
	}
	// Is user in chat?
	err := CheckUserINChat(user_id, chat_id)
	if err != nil{
		return err
	}
	err= CheckUserInChatDelete(user_id, chat_id)
	if err != nil{
		return err
	}
//	Create message
	statement, err := activeConn.Prepare("INSERT INTO messages (user_id, chat_id, content, time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password
	_, err = statement.Exec(user_id, chat_id, content, time.Now().Unix())
	if err != nil {
		return errors.New("Failed exec statement")
	}
	return nil
}

func CheckUserInChatDelete(user_id float64, chat_id float64)(error){
	//var id_now string
	var delete_a int64
	rows, err := activeConn.Prepare("SELECT  delete_a FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(user_id, chat_id).Scan( &delete_a)
	defer rows.Close()
	if query == sql.ErrNoRows{
		return errors.New("User aren't in chat")
	}
	if delete_a == 1{
		return errors.New("User aren't in chat")
	}
	return nil
}

func CheckUserINChat(user_id float64, chat_id float64)(error){
	var id_now string
	var delete_a int64
	rows, err := activeConn.Prepare("SELECT chat_id, delete_a FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(user_id, chat_id).Scan(&id_now, &delete_a)
	defer rows.Close()
	if query == sql.ErrNoRows{
		return errors.New("User aren't in chat")
	}
	//if delete_a == 1{
	//	return errors.New("User aren't in chat")
	//}
	return nil
}

func GetFileInformation(file_id string)(map[string]string, error){
	final := make(map[string]string)
	//var getFileBD struct{filename string; path string; uses int}
	var filename string
	var path string
	var uses int
	var ratio_size string
	rows, err := activeConn.Prepare("SELECT filename, path, uses, ratio_size  FROM files  WHERE id=?")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(file_id).Scan(&filename, &path, &uses, &ratio_size)
	defer rows.Close()
	if query == sql.ErrNoRows{
		return final,errors.New("File is undefine")
	}
	final["name"] = filename
	final["path"] = path
	final["ratio_size"] = ratio_size
	final["file_id"] = file_id
	return final, nil
}

func GetFileProve(user_id float64, file_id string)(string, error){
	//var file_id string
	var path string
	//var uses int
	rows, err := activeConn.Prepare("SELECT files.path FROM files INNER JOIN people_in_chats ON people_in_chats.chat_id = files.chat_id WHERE (people_in_chats.user_id=?) and (files.id =?)")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(user_id, file_id).Scan( &path)
	if query == sql.ErrNoRows{
		return "", errors.New("You are haven't rights for this file")
	}
	return path, nil

}

func GetMessages(user_id float64, chat_id float64)([]models.NewMessageToUser, error){
	var id_now string
	var start,deltimes string
	var delete_a, deltime int64
	r_deltimes := [][]int64{}
	rows_user_in_chat, err := activeConn.Prepare("SELECT chat_id, start, delete_a, deltime, deltimes FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		return nil, errors.New("Cant prove user isnt in chat")
		//panic(nil)
	}
	query := rows_user_in_chat.QueryRow(user_id, chat_id).Scan(&id_now, &start, &delete_a, &deltime, &deltimes)
	err = json.Unmarshal([]byte(deltimes), &r_deltimes)
	if err != nil {
		return nil, errors.New("Cant decode delete times")
		//panic(nil)
	}
	defer rows_user_in_chat.Close()
	if query == sql.ErrNoRows{
		return nil, errors.New("User isn't in chat")
	}
	var messages []models.NewMessageToUser
	i_start,_ := strconv.ParseInt(start,10,64)
	for i:=0;i<len(r_deltimes);i++{
		if i==0 && r_deltimes[0][0]==0{

			err = getMessageBetweenTime(&messages, i_start,9999999999,chat_id)
		}else{
			if i==0{
				err = getMessageBetweenTime(&messages, i_start, r_deltimes[i+1][0],chat_id)
			}else if i>0{
				//i_start,_ := strconv.ParseInt(start,10,64)
				//i_stop, _ := r_deltimes[]
				err = getMessageBetweenTime(&messages, r_deltimes[i-1][1],r_deltimes[i][0],chat_id)
				if r_deltimes[i][1] != 0 && i!= len(r_deltimes)-1{
					err = getMessageBetweenTime(&messages, r_deltimes[i][1],r_deltimes[i+1][0],chat_id)
				}else if i== len(r_deltimes)-1{
					err = getMessageBetweenTime(&messages, r_deltimes[i][1],9999999999,chat_id)
				}
				//messages,err = getMessageBetweenTime(messages, r_deltimes[i-1][1],r_deltimes[i][0],chat_id)
			}
		}
	}
	return messages,nil
}



func getMessageBetweenTime(messages *[]models.NewMessageToUser, start int64, finish int64, chat_id float64)(error){
	rows, err := activeConn.Query("SELECT messages.user_id, messages.content, messages.chat_id,  people.u_name, messages.time  FROM messages " +
		"INNER JOIN people ON messages.user_id = people.id WHERE (messages.chat_id=?) and (messages.time>?) and (messages.time<?)", chat_id, start, finish)
	for rows.Next() {
		var id, content, u_name, c_id string
		var m_time int64
		if err := rows.Scan(&id, &content, &c_id, &u_name, &m_time); err != nil {
			return err
		}
		//decode content
		//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		var r_content models.MessageContent
		var f_content models.MessageContentToUser
		err = json.Unmarshal([]byte(content), &r_content)
		if err != nil {
			return  err
		}
		f_content.Message = r_content.Message
		f_content.Type = r_content.Type
		documents := *r_content.Documents
		//fmt.Println(documents)
		for i := 0; i < len(documents); i++ {
			//id := *r_content.Documents
			parse_doc, err := GetFileInformation(documents[i])
			if err != nil {
				return  err
			}
			f_content.Documents = append(f_content.Documents, parse_doc)
		}
		f64_c_id, err := strconv.ParseFloat(c_id, 64)
		if err != nil {
			return  err
		}
		f64_id, err := strconv.ParseFloat(id, 64)
		if err != nil {
			return  err
		}
		*messages = append(*messages, models.NewMessageToUser{&f64_c_id, f_content, &f64_id, &u_name, &m_time})
	}
	return  nil
}

func CreateFile(filename string, size int64, user_id float64, chat_id string, ratio_size string)(int64, string, error){
	if !activeConnIsReal{
		OpenDB()
	}
	now_time := strconv.FormatInt(time.Now().Unix(),10)
	f_size :=strconv.FormatInt(size,10)
	path := now_time+f_size+filename

	statement, err := activeConn.Prepare("INSERT INTO files (author_id, chat_id, filename, path, time, uses, ratio_size) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return -1,"",errors.New("Fail insert file")
	}
	res,err := statement.Exec(user_id, chat_id, filename ,path, now_time, 0, ratio_size)
	if err != nil {
		return -1,"",errors.New("Fail exec BD")
	}
	id, _ := res.LastInsertId()
	return id,path, nil
}

func DeleteFile(user_id float64, file_id string)(string, error){
	if !activeConnIsReal{
		OpenDB()
	}
	var path string
	message, err := activeConn.Prepare("SELECT path FROM files where (id=?) ")
	if err != nil {
		return "", err
	}
	query := message.QueryRow(file_id)

	err = query.Scan(&path)
	if err == sql.ErrNoRows{
		return "", err
	}
	stmt, err := activeConn.Prepare("delete from files where (id=?) and (uses = 0) and (author_id=?)")
	if err != nil{
		return "",errors.New("Fail prepare delete file")
	}
	_, err = stmt.Exec(file_id, user_id)
	if err != nil{
		return "",errors.New("Fail exec delete file")
	}
	return  path, nil
}

func createDB_structs(database *sql.DB) {
	//Create user structs
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, login TEXT, pass TEXT, u_name TEXT)")
	statement.Exec()
	user_id, fin, err := CreateUser("god", "1111", "Alex")
	if err != nil {
		fmt.Println(fin)
		return
	}
	//Create people in chat structs

	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS people_in_chats ( user_id INTEGER, chat_id INTEGER, blocked INTEGER DEFAULT 0, start INTEGER DEFAULT 0, delete_a INTEGER DEFAULT 0, deltime INTEGER DEFAULT 0, deltimes TEXT)")
	statement.Exec()

	//Create messages structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, user_id INTEGER, chat_id INTEGER, content TEXT, time INTEGER)")
	statement.Exec()

	//Create files structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, author_id INTEGER, chat_id INTEGER, filename TEXT, path Text, time INTEGER, uses INTEGER, ratio_size TEXT)")
	statement.Exec()

	//Create chat structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY, name TEXT,  author_id INTEGER , moders_ids TEXT, type INTEGER DEFAULT 0,  lastmodify INTEGER)")
	statement.Exec()
	_, err = CreateChat("globalChat",  user_id)
	if err != nil {
		fmt.Println(err.Error())
	}



	}

func FindUserByName(name string, chat_id string)([]map[string]string,error){
	var middle []map[string]string
	var logins []string
	var names []string
	//get logins and names how already in chat
	rows, err := activeConn.Query("SELECT  people.u_name, people.login FROM people INNER JOIN people_in_chats ON people_in_chats.user_id = people.id WHERE people_in_chats.chat_id=?", chat_id)
	if err != nil {
		fmt.Println("scan 1")
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var name, login string
		if err := rows.Scan(&name,  &login); err != nil {
			fmt.Println("scan 2")
			return nil,err
		}
		logins=append(logins,login)
		names= append(names, name)
	}
	query_logins:=""
	query_names:=""
	for i := 0; i < len(logins); i++ {
		if i>=1{
			query_logins+=","
		}
		query_logins += logins[i]
	}
	for i := 0; i < len(names); i++ {
		if i>=1{
			query_names += ","
		}
		query_names += "'"+names[i]+"'"
	}
	fmt.Println(query_names)
	fmt.Println(query_logins)
	//"SELECT  messages.content, people.u_name FROM messages INNER JOIN people ON messages.user_id = people.id WHERE chat_id=? ORDER BY time DESC"
	rows, err = activeConn.Query("SELECT id , u_name, login FROM people  WHERE  u_name NOT IN " +
		"(SELECT  people.u_name FROM people INNER JOIN people_in_chats ON people_in_chats.user_id = people.id WHERE people_in_chats.chat_id=?) and u_name LIKE (?)",chat_id, "%"+name+"%")
	if err != nil {
		fmt.Println("scan 3")
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id, name, login string
		if err := rows.Scan(&id, &name,  &login); err != nil {
			fmt.Println("scan 4")
			return nil,err
		}
		middle=append(middle, map[string]string{"id": id,"name": name, "login": login})
	}
	if len(middle) == 0{
		middle = []map[string]string{}
	}
	return middle, nil
}

func GetUsersChatsIds(user_id float64)([]string,error){
	var ids []string
	rows, err := activeConn.Query("SELECT chats.id FROM people_in_chats INNER JOIN chats ON people_in_chats.chat_id = chats.id WHERE user_id=?", user_id)
	if err != nil {
		fmt.Println("Outside", err)
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id string
		//var moders []string
		if err := rows.Scan(&id); err != nil {
			fmt.Println("scan 1")
			return nil,err
		}
		ids=append(ids, id)
	}
	return  ids, nil
}

func GetChatsUsers(chat_id float64)([]float64,error){
	var ids []float64
	rows, err := activeConn.Query("SELECT user_id FROM people_in_chats  WHERE chat_id=? and delete_a = 0", chat_id)
	if err != nil {
		fmt.Println("Outside", err)
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id float64
		//var moders []string
		if err := rows.Scan(&id); err != nil {
			fmt.Println("scan 1")
			return nil,err
		}
		ids=append(ids, id)
	}
	return  ids, nil
}

func GetChatUsersInfo(chat_id float64)(string, error ){
	type userInfo struct {
		Id int `json:"id"`
		Login string `json:"login"`
		Name string `json:"name"`
		Blocked int `json:"blocked"`
		Delete int `json:"delete"`
	}

	users:=make([]userInfo,0)
	rows, err := activeConn.Query("SELECT people.id, people.login, people.u_name, people_in_chats.blocked, people_in_chats.delete_a FROM people_in_chats INNER JOIN people ON people_in_chats.user_id = people.id WHERE people_in_chats.chat_id=?", chat_id)
	if err != nil {
		fmt.Println("Outside", err)
		return "",err
	}
	defer rows.Close()
	for rows.Next(){
		var login string
		var name string
		var blocked int
		var id int
		var delete_a int
		//var moders []string
		if err := rows.Scan(&id, &login, &name, &blocked, &delete_a); err != nil {
			fmt.Println("scan 1")
			return "",err
		}
		user :=userInfo{id,login, name, blocked, delete_a}
		users=append(users, user)
	}
	finish, _:=json.Marshal(users)
	return  string(finish), nil
}
//func AddUsersInChat
func CheckUserRightsInChat(user_id float64, chat_id float64)(error){
	err:=CheckUserINChat(user_id, chat_id)
	if err!=nil{
		return err
	}
	fmt.Println("CHAT ID", chat_id)
	final := false
	var moders_ids = []float64{}
	var moder_ids_s  string
	var admin_id int64
	rows, err := activeConn.Query("SELECT author_id, moders_ids FROM chats WHERE id=?", chat_id)
	if err != nil {
		fmt.Println("Outside", err)
		return err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&admin_id, &moder_ids_s)
	if err != nil{
		fmt.Println("2", err)
		return err
	}
	if err == sql.ErrNoRows{
		fmt.Println("3", err)
		return  err
	}
	err = json.Unmarshal([]byte(moder_ids_s), &moders_ids)
	if err != nil {
		//panic(err)
	}
	for _,v:= range moders_ids{
		if v==user_id{
			final=true
		}
	}
	fmt.Println(admin_id, user_id)

	f64_admin_id:= float64(admin_id)
	if f64_admin_id ==user_id{
		final = true
	}
	if final==true{
		return nil
	}
	return errors.New("You haven't rights for this action")

	//for rows.Next(){
	//	var id, name, un_moders string
	//	var author_id string
	//	//var moders []string
	//	if err := rows.Scan(&id,  &name, &author_id, &un_moders); err != nil {
	//		fmt.Println("scan 1")
	//		return err
	//	}
	//	middle=append(middle, map[string]string{"id": id, "name": name, "author": author_id, "moders": un_moders})
	//
	//
	//}
}

func DeleteUsersInChat(users_ids []float64, chat_id string)(error){
	//var query_str string
	s_ids := []string{}
	for _,v := range users_ids{
		s_ids = append(s_ids, strconv.FormatFloat(v,'f',0,64))
		//query_str+=  strconv.FormatFloat(v,'E',-1,64)
	}
	asID := strings.Join(s_ids, ",")
	query := fmt.Sprintf("UPDATE people_in_chats SET delete_a = ?, deltime = ? where (user_id in (%s)) and (chat_id=%s)", asID, chat_id)
	fmt.Println(query)
	statement, err := activeConn.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return errors.New("DB failed query")
	}
	//make hash of user's password
	_, err = statement.Exec(1,time.Now().Unix()+1)
	if err != nil {
		fmt.Println(err)
		return errors.New("Failed exec statement")
	}
	return nil
}



func OpenDB(){
	newDB := false
	_, err := os.Open("app.db")
	if err != nil{
		newDB = true
		file, err := os.Create("app.db")
		if err != nil {
			// handle the error here
			fmt.Println("God: i cant create database, your PC is atheist")
			return
		}
		defer file.Close()
		fmt.Println("God: im create database")
	}
	database, _ := sql.Open("sqlite3", "./app.db")
	if newDB{
		createDB_structs(database)
	}
	activeConn = database
	activeConnIsReal=true
}

