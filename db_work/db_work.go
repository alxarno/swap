package spatium_db_work

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"crypto/sha256"
	"os"
	"fmt"
	"encoding/base64"
	models "github.com/AlexeyArno/Spatium/models"
	"time"
	"crypto/rand"
	"errors"
	"encoding/json"
	"strconv"
	"github.com/AlexeyArno/Spatium/settings"
	//"strings"
	//engine "github.com/AlexArno/spatium/src/message_engine"
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
		h := sha256.New()
		h.Write([]byte(data["pass"]))
		query := rows.QueryRow(data["login"], h.Sum(nil))

		err = query.Scan(&user.ID, &user.Login, &user.Pass, &user.Name)
		//make hash of user's password
		rows.Close()
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
		//fmt.Println(data["id"])
		err = query.Scan(&user.ID, &user.Login, &user.Name)
		rows.Close()
		if err == sql.ErrNoRows{
			return nil, err
		}
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
		fmt.Println(err.Error())
		panic(nil)
	}
	query := rows.QueryRow(login).Scan(&id_now)
	rows.Close()
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
	_, err = statement.Exec(login, h.Sum(nil), u_name)
	statement.Close()
	if err != nil {
		return "",err.Error(),err
	}
	rows, err = activeConn.Prepare("SELECT id FROM people WHERE login=?")
	if err != nil {
		return "","DB failed query",err
	}
	err = rows.QueryRow(login).Scan(&id_now)
	rows.Close()
	if err == sql.ErrNoRows{
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
	rows.Close()
	if query != sql.ErrNoRows{

		var list_delete int
		rows, err := activeConn.Prepare("SELECT list_delete FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
		rows.QueryRow(user_id, chat_id).Scan(&list_delete)
		if list_delete == 0{
			return errors.New("User already in chat")
		}
		rows.Close()
		stmt, err := activeConn.Prepare("UPDATE people_in_chats SET list_delete=? WHERE user_id=? and chat_id=?")
		if err != nil {
			//fmt.Println("Fail delete", err)
			return errors.New("Failed return user in chat")
		}
		stmt.Exec(0, user_id, chat_id)
		stmt.Close()
		return nil
	}
	statement, err := activeConn.Prepare("INSERT INTO people_in_chats (user_id, chat_id, blocked, start, deltimes) VALUES (?, ?, ?, ?,?)")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password

	deltime:= [1][2]int64{}
	deltime[0][0] = 0
	deltime[0][1] = 0
	s_deltime,_:= json.Marshal(deltime)

	statement.Exec(user_id, chat_id, 0, time.Now().Unix()-1,string(s_deltime))
	statement.Close()
	statement, err = activeConn.Prepare("UPDATE chats SET lastmodify=? WHERE id=?")


	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password
	statement.Exec(time.Now().Unix(), chat_id)
	statement.Close()
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
	statement.Close()
	if err != nil {
		return "",errors.New("Failed exec statement")
	}
	id, _ := res.LastInsertId()
	err = InsertUserInChat(author_id, id)
	if err != nil {
		return "",err
		//fmt.Println(fin)
	}
	mess_mss := "создал этот чат"
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
	_,err = AddMessage(f_id, float64(id), string(data))
	if err != nil{
		return "", err
	}
	return string(id), nil
}

func CreateChannel(name string, author_id string)(string,  error){
	if !activeConnIsReal{
		OpenDB()
	}
	statement, err := activeConn.Prepare("INSERT INTO chats (name,  author_id, moders_ids, type, lastmodify) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return "",errors.New("Failed permanent statement")
	}
	//make hash of user's password
	res, err := statement.Exec(name,  author_id,"[]", 2, time.Now().Unix())
	statement.Close()
	if err != nil {
		return "",errors.New("Failed exec statement")
	}
	id, _ := res.LastInsertId()
	err = InsertUserInChat(author_id, id)
	if err != nil {
		return "",err
		//fmt.Println(fin)
	}
	mess_mss := "создал этот каннал"
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
	_,err = AddMessage(f_id, float64(id), string(data))
	if err != nil{
		return "", err
	}
	return string(id), nil
}

func GetMyChats(user_id float64)([]*models.UserChatInfo, error){
	type chatInfo struct{
		Id string
		Name string
		UModers string
		Author_id string
		Delete_a int64
		Deltime int64
		C_Type int64
	}
	var chats_ids []*models.UserChatInfo
	var middle []map[string]string
	var chat_infos []chatInfo
	rows, err := activeConn.Query("SELECT chats.id, chats.name,  chats.type, chats.author_id, chats.moders_ids, people_in_chats.delete_a, people_in_chats.deltime  FROM people_in_chats INNER JOIN chats ON people_in_chats.chat_id = chats.id WHERE user_id=? AND list_delete=0", user_id)
	if err != nil {
		fmt.Println("Outside", err)
		return nil,err
	}
	for rows.Next() {
		var id, name, un_moders string
		var author_id string
		var delete_a, deltime, c_type int64
		//var moders []string
		if err := rows.Scan(&id, &name, &c_type, &author_id, &un_moders, &delete_a, &deltime); err != nil {
			fmt.Println("scan 1")
			return nil, err
		}
		chat_infos = append(chat_infos, chatInfo{id, name, un_moders, author_id, delete_a, deltime, c_type})
	}
	rows.Close()
	var name_buf string
	for _,v := range chat_infos{
		name_buf = v.Name
		if v.C_Type == 1{
			n_rows, err := activeConn.Query("SELECT  people.u_name FROM people INNER JOIN people_in_chats ON people_in_chats.user_id = people.id WHERE (people_in_chats.chat_id=?) and (people_in_chats.user_id<>?)", v.Id,user_id)
			if err != nil {
				//fmt.Println("scan 1")
				return nil,err
			}
			//defer n_rows.Close()
			for n_rows.Next(){
				if err := n_rows.Scan(&name_buf); err != nil {
					//fmt.Println("scan 2")
					return  nil,err
				}
			}
			n_rows.Close()
		}
		middle=append(middle, map[string]string{"id": v.Id, "type": strconv.FormatInt(v.C_Type,10), "name": name_buf, "author": v.Author_id, "moders": v.UModers,
			"delete": strconv.FormatInt(v.Delete_a,10), "deltime": strconv.FormatInt(v.Deltime,10)})
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
		message.Close()
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
		//online,err:=getOnlineUsersIntChat(f_id)
		//if err!=nil{
		//	fmt.Println("content")
		//	return nil, err
		//}
		id,err:= strconv.ParseInt(i["type"],10,64)
		if err!=nil{
			return nil, err
		}
		chats_ids=append(chats_ids, &models.UserChatInfo{f_id,i["name"],id, author_name, f_a_id,
															moders,&m_content,i_time,0,
															i_delete, 0})

		//chats_ids
	}
	if err := rows.Err(); err != nil {
		return nil,err
	}
	return chats_ids, nil
}

func AddMessage(user_id float64, chat_id float64, content string)(int64, error){
	if !activeConnIsReal{
		OpenDB()
	}
	// Is user in chat?
	err := CheckUserINChat(user_id, chat_id)
	if err != nil{
		return -1,err
	}
	err= CheckUserInChatDelete(user_id, chat_id)
	if err != nil{
		return -1,err
	}
	//	Create message
	statement, err := activeConn.Prepare("INSERT INTO messages (user_id, chat_id, content, time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return -1,errors.New("DB failed query")
	}
	//make hash of user's password
	res, err := statement.Exec(user_id, chat_id, content, time.Now().Unix())
	if err != nil {
		//fmt.Println(err.Error())
		return -1, err
	}
	statement.Close()
	id, _ := res.LastInsertId()
	return id,nil
}

func AddForceMessage(user_id float64, chat_id float64, content string)(int64, error){
	if !activeConnIsReal{
		OpenDB()
	}

	//	Create message
	statement, err := activeConn.Prepare("INSERT INTO messages (user_id, chat_id, content, time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return -1,errors.New("DB failed query")
	}
	//make hash of user's password
	res, err := statement.Exec(user_id, chat_id, content, time.Now().Unix())
	if err != nil {
		return -1,errors.New("Failed exec statement")
	}
	id, _ := res.LastInsertId()
	return id,nil
}

func CheckUserInChatDelete(user_id float64, chat_id float64)(error){
	//var id_now string
	var delete_a,  deltype  int64
	rows, err := activeConn.Prepare("SELECT  delete_a, delete_by_admin FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(user_id, chat_id).Scan( &delete_a, &deltype)
	defer rows.Close()
	if query == sql.ErrNoRows{
		return errors.New("User aren't in chat")
	}
	if delete_a == 1{
		return errors.New("User aren't in chat")
	}
	if deltype == 1{
		return errors.New("User banned by admin ")
	}
	return nil
}

func CheckUserINChat(user_id float64, chat_id float64)(error){
	var id_now string
	var delete_a int64
	rows, err := activeConn.Prepare("SELECT chat_id, delete_a FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		//panic(nil)
		return errors.New("Fail prepare get chat_id ")
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

func GetFileInformation(file_id string)(map[string]interface{}, error){
	final := make(map[string]interface{})
	//var getFileBD struct{filename string; path string; uses int}
	var filename string
	var path string
	var uses int
	var ratio_size string
	var size int
	rows, err := activeConn.Prepare("SELECT filename, path, uses, ratio_size, size  FROM files  WHERE id=?")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(file_id).Scan(&filename, &path, &uses, &ratio_size, &size)
	defer rows.Close()
	if query == sql.ErrNoRows{
		return final,errors.New("File is undefine")
	}
	final["name"] = filename
	final["path"] = path
	final["ratio_size"] = ratio_size
	final["file_id"] = file_id
	final["size"] = size
	return final, nil
}

func GetChatType(chat_id float64)(int, error){
	var id int
	rows, err := activeConn.Prepare("SELECT  type FROM chats WHERE id=?")
	if err != nil {
		//panic(nil)
		return 0,err
	}
	rows.QueryRow(chat_id).Scan(&id)
	rows.Close()
	return id, nil
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
	rows.Close()
	if query == sql.ErrNoRows{
		return "", errors.New("You are haven't rights for this file")
	}
	return path, nil

}

func GetMessages(user_id float64, chat_id float64, add bool, last_index int)([]models.NewMessageToUser, error){
	const MAX_TIME = 9999999999
	chat_type, err:=GetChatType(chat_id)

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
	if query == sql.ErrNoRows{
		return nil, errors.New("User isn't in chat")
	}
	rows_user_in_chat.Close()
	var messages []models.NewMessageToUser
	i_start,_ := strconv.ParseInt(start,10,64)
	basic_query := "SELECT messages.id, messages.user_id, messages.content, messages.chat_id,  people.u_name, messages.time,  people.login  FROM messages " +
		"INNER JOIN people ON messages.user_id = people.id WHERE (messages.chat_id=?) and ("
	if chat_type==2{
		basic_query = "SELECT messages.id, messages.user_id, messages.content, messages.chat_id,  people.u_name, messages.time,  people.login  FROM messages " +
			"INNER JOIN people ON messages.user_id = people.id WHERE (messages.chat_id=?) ORDER BY messages.time DESC LIMIT 80 "
	}
	messages_queries:= []string{}
	if chat_type!=2 {
		for i := 0; i < len(r_deltimes); i++ {
			if i == 0 && r_deltimes[0][0] == 0 {
				messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", i_start, MAX_TIME))
			} else {
				if i == 0 {
					messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and (messages.time<=%d)) ", i_start, r_deltimes[i][0]))
				} else if i > 0 {
					messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and (messages.time<=%d)) ", r_deltimes[i-1][1], r_deltimes[i][0]))
					if r_deltimes[i][0] == 0 {
						messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and (messages.time<=%d)) ", r_deltimes[i-1][1], MAX_TIME))
					}

				}
			}
		}

		for i, v := range messages_queries {
			if i == 0 {
				basic_query += v
			} else {
				basic_query += "or " + v
			}
		}

		// send yet early messages
		if add {
			basic_query += fmt.Sprintf(") and ((messages.id < %d)", last_index)
		}
		basic_query += ") ORDER BY messages.time DESC LIMIT 80"
	}
	err = getMessageByQuery(basic_query, chat_id, &messages)
	if err != nil{
		return messages,err
	}
	return messages,nil
}

func getMessageByQuery(query string, chat_id float64, messages *[]models.NewMessageToUser)(error){
	rows, err := activeConn.Query(query, chat_id)
	defer rows.Close()
	if err == sql.ErrNoRows{
		return err
	}
	if err!= nil{
		return err
	}
	for rows.Next() {
		var m_id, u_id, content, u_name, c_id, login string
		var m_time int64
		if err := rows.Scan(&m_id, &u_id, &content, &c_id, &u_name, &m_time, &login); err != nil {
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
		f64_uid, err := strconv.ParseFloat(u_id, 64)
		if err != nil {
			return  err
		}
		im_id, err := strconv.ParseInt(m_id, 10,64)
		if err != nil {
			return  err
		}
		*messages = append(*messages, models.NewMessageToUser{
			&im_id,
			&f64_c_id,
			f_content,
			&f64_uid,
			&u_name,
			&login,
			&m_time})
	}
	return  nil
}

func CreateFile(filename string, size int64, user_id float64, chat_id string, ratio_size string)(int64, string, error){
	if !activeConnIsReal{
		OpenDB()
	}
	now_time := strconv.FormatInt(time.Now().Unix(),10)
	f_size :=strconv.FormatInt(size,10)
	if len(filename)>20{
		runes := []rune(filename)
		filename = string(runes[len(runes)-20:])
	}
	// Generate random string, because create different path for two similar files
	b := make([]byte, 20)
	_, err := rand.Read(b)

	if err != nil {
		return -1,"",err
	}

	addName := base64.URLEncoding.EncodeToString(b)

	path := now_time+f_size+addName+filename

	statement, err := activeConn.Prepare("INSERT INTO files (author_id, chat_id, filename, path, time, uses, ratio_size, size) VALUES (?, ?, ?, ?, ?, ?, ?,?)")
	if err != nil {
		statement.Close()
		return -1,"",errors.New("Fail insert file")
	}
	res,err := statement.Exec(user_id, chat_id, filename ,path, now_time, 0, ratio_size, f_size)
	if err != nil {
		fmt.Println(err.Error())
		statement.Close()
		return -1,"",errors.New("Fail exec BD")
	}
	statement.Close()
	id, _ := res.LastInsertId()
	return id,path, nil
}

func DeleteFile(user_id float64, file_id string)(string, error){
	if !activeConnIsReal{
		OpenDB()
	}
	var path string
	message, err := activeConn.Prepare("SELECT path FROM files where (id=?) ")
	defer message.Close()
	if err != nil {
		message.Close()
		return "", err
	}
	query := message.QueryRow(file_id)

	err = query.Scan(&path)
	if err == sql.ErrNoRows{
		message.Close()
		return "", err
	}
	stmt, err := activeConn.Prepare("delete from files where (id=?) and (uses = 0) and (author_id=?)")
	defer stmt.Close()
	if err != nil{
		message.Close()
		return "",errors.New("Fail prepare delete file")
	}
	_, err = stmt.Exec(file_id, user_id)
	if err != nil{
		message.Close()
		return "",errors.New("Fail exec delete file")
	}
	message.Close()
	return  path, nil
}

func createDB_structs(database *sql.DB)(error){
	//Create user structs
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, login TEXT, pass TEXT, u_name TEXT)")
	statement.Exec()
	//Create people in chat structs

	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS people_in_chats (id INTEGER PRIMARY KEY,"+
											" user_id INTEGER, chat_id INTEGER," +
		"									 blocked INTEGER DEFAULT 0, start INTEGER DEFAULT 0, delete_a INTEGER DEFAULT 0," +
			"								 deltime INTEGER DEFAULT 0, deltimes TEXT, delete_by_admin INTEGER DEFAULT 0, list_delete INTEGER DEFAULT 0)")
	statement.Exec()

	//Create messages structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY , user_id INTEGER, chat_id INTEGER, content TEXT, time INTEGER)")
	statement.Exec()

	//Create files structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY , author_id INTEGER, chat_id INTEGER, filename TEXT, path Text, time INTEGER, uses INTEGER, ratio_size TEXT, size INTEGER )")
	statement.Exec()

	//create dialogs info
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS dialogs_info (id INTEGER PRIMARY KEY ,  chat_id INTEGER, user_1 INTEGER , user_2 INTEGER,  delete_users TEXT)")
	if err!=nil{
		fmt.Println(err.Error())
		return err
	}
	statement.Exec()

	//Create chat structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY , name TEXT,  author_id INTEGER , moders_ids TEXT, type INTEGER DEFAULT 0,  lastmodify INTEGER)")
	statement.Exec()

	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS sys (id INTEGER PRIMARY KEY, version TEXT, data_instance TEXT)")
	statement.Exec()

	statement, _ = database.Prepare("INSERT INTO sys (version, data_instance) VALUES (?, ?)")
	current_time := time.Now().UTC().Format("Mon Jan 2 15:04:05 MST 2006")
	_, err = statement.Exec("0.0.1", current_time)
	if err!=nil{
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func FindUserByName(name string, chat_id string)([]map[string]string,error){
	var middle []map[string]string
	var logins []string
	var names []string
	//get logins and names how already in chat
	rows, err := activeConn.Query("SELECT  people.u_name, people.login FROM people INNER JOIN people_in_chats ON people_in_chats.user_id = people.id WHERE people_in_chats.chat_id=? and people_in_chats.list_delete=0", chat_id)
	if err != nil {
		fmt.Println("FindUserByName: 1",err)
		return nil,err
	}
	for rows.Next(){
		var name, login string
		if err := rows.Scan(&name,  &login); err != nil {
			fmt.Println("FindUserByName: 2",err)
			return nil,err
		}
		logins=append(logins,login)
		names= append(names, name)
	}
	rows.Close()
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
	//fmt.Println(query_names)
	//fmt.Println(query_logins)
	//"SELECT  messages.content, people.u_name FROM messages INNER JOIN people ON messages.user_id = people.id WHERE chat_id=? ORDER BY time DESC"
	rows, err = activeConn.Query("SELECT id , u_name, login FROM people  WHERE  u_name NOT IN " +
		"(SELECT  people.u_name FROM people INNER JOIN people_in_chats ON people_in_chats.user_id = people.id WHERE people_in_chats.chat_id=? and people_in_chats.list_delete=0) and ((u_name LIKE (?)) or (login LIKE (?))) ",chat_id, "%"+name+"%", "%"+name+"%")
	if err != nil {
		fmt.Println("FindUserByName: 3",err)
		return nil,err
	}
	for rows.Next(){
		var id, name, login string
		if err := rows.Scan(&id, &name,  &login); err != nil {
			fmt.Println("FindUserByName: 4",err)
			return nil,err
		}
		middle=append(middle, map[string]string{"id": id,"name": name, "login": login})
	}
	rows.Close()
	if len(middle) == 0{
		middle = []map[string]string{}
	}
	return middle, nil
}

func GetUsersChatsIds(user_id float64)([]float64,error){
	var ids []float64
	rows, err := activeConn.Query("SELECT chats.id FROM people_in_chats INNER JOIN chats ON people_in_chats.chat_id = chats.id WHERE user_id=?", user_id)
	if err != nil {
		fmt.Println("GetUsersChatsIds 1:",err)
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id float64
		//var moders []string
		if err := rows.Scan(&id); err != nil {
			fmt.Println("GetUsersChatsIds 2:",err)
			return nil,err
		}
		ids=append(ids, id)
	}
	return  ids, nil
}

func GetUsersIdsForUpdateChatsInfoOnline(chats_ids *[]float64, users_online_ids *[]float64)([]float64, error){
	//Get chats ids for get users for send update
	var final []float64
	var users_ids []string
	var chats_ids_s []string
	for _,v:= range *chats_ids{
		chats_ids_s = append(chats_ids_s, strconv.FormatFloat(v,'f',-1,64))
	}
	for _,v:= range *users_online_ids{
		users_ids = append(users_ids, strconv.FormatFloat(v,'f',-1,64))
	}
	user_s_ids := strings.Join(users_ids, ",")
	chat_s_ids := strings.Join(chats_ids_s, ",")
	//fmt.Println(user_s_ids)
	//fmt.Println(chat_s_ids)
	query := fmt.Sprintf("SELECT DISTINCT user_id FROM people_in_chats WHERE chat_id in (%s) and user_id in (%s) and delete_a = 0", chat_s_ids, user_s_ids)
	stmt, err := activeConn.Query(query)
	if err != nil {
		//fmt.Println("GetUsersIdsForUpdateChatsInfoOnline  :", err)
		return nil,err
	}
	var id string
	for stmt.Next(){
		err = stmt.Scan(&id)
		if err != nil {
			//fmt.Println("GetUsersIdsForUpdateChatsInfoOnline :", err)
			return nil,err
		}
		f64_id,err:= strconv.ParseFloat(id, 64)
		if err != nil {
			//fmt.Println("GetUsersIdsForUpdateChatsInfoOnline : (694) :", err)
			return nil,err
		}
		final = append(final, f64_id)
	}
	return final, nil
}

func GetChatsUsers(chat_id float64)([]float64,error){
	var ids []float64
	rows, err := activeConn.Query("SELECT user_id FROM people_in_chats  WHERE chat_id=? and delete_a = 0 and list_delete=0", chat_id)
	if err != nil {
		fmt.Println("GetChatsUsers 1:",err)
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id float64
		//var moders []string
		if err := rows.Scan(&id); err != nil {
			fmt.Println("GetChatsUsers 2:",err)
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
	rows, err := activeConn.Query("SELECT people.id, people.login, people.u_name, people_in_chats.blocked, people_in_chats.delete_a"+
		" FROM people_in_chats INNER JOIN people ON people_in_chats.user_id = people.id WHERE people_in_chats.chat_id=? and people_in_chats.list_delete = 0", chat_id)
	if err != nil {
		fmt.Println("GetChatUsersInfo 1:",err)
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
			fmt.Println("GetChatUsersInfo 2:",err)
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
	//fmt.Println("CHAT ID", chat_id)
	final := false
	var moders_ids = []float64{}
	var moder_ids_s  string
	var admin_id int64
	rows, err := activeConn.Query("SELECT author_id, moders_ids FROM chats WHERE id=?", chat_id)
	if err != nil {
		fmt.Println("GetChatUsersInfo 1:",err)
		return err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&admin_id, &moder_ids_s)
	if err != nil{
		fmt.Println("GetChatUsersInfo 2:",err)
		return err
	}
	if err == sql.ErrNoRows{
		fmt.Println("GetChatUsersInfo 3:",err)
		return  err
	}
	err = json.Unmarshal([]byte(moder_ids_s), &moders_ids)
	if err != nil {
		fmt.Println("GetChatUsersInfo 4:",err)
		return  err
		//panic(err)
	}
	for _,v:= range moders_ids{
		if v==user_id{
			final=true
		}
	}
	//fmt.Println(admin_id, user_id)

	f64_admin_id:= float64(admin_id)
	if f64_admin_id ==user_id{
		final = true
	}
	if final==true{
		fmt.Println("GetChatUsersInfo 5:",err)
		return nil
	}
	fmt.Println("GetChatUsersInfo 6:")
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

func DeleteUsersInChat(users_ids []float64, chat_id string, delete_yourself bool)(error){
	//var query_str string

	//s_ids := []string{}
	for _,v := range users_ids{
		//s_ids = append(s_ids, strconv.FormatFloat(v,'f',0,64))
		var deltimes string
		r_deltimes:= [][]int64{}
		rows_user_in_chat, err := activeConn.Prepare("SELECT  deltimes FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
		if err != nil {
			return errors.New("Cant prove user isnt in chat")
			//panic(nil)
		}
		rows_user_in_chat.QueryRow(v, chat_id).Scan(&deltimes)
		err = json.Unmarshal([]byte(deltimes), &r_deltimes)
		if err != nil {
			return errors.New("Cant decode delete times")
			//panic(nil)
		}
		rows_user_in_chat.Close()
		//for i:=0;i<len(r_deltimes);i++{
		//
		//}
		s_id := strconv.FormatFloat(v,'f',0,64)
		if r_deltimes[len(r_deltimes)-1][0]==0 {
			r_deltimes[len(r_deltimes)-1][0] = time.Now().Unix()

			query := fmt.Sprintf("UPDATE people_in_chats SET delete_a = ?, deltime = ?, deltimes = ?, delete_by_admin = ? where (user_id = %s) and (chat_id = %s)", s_id, chat_id)
			//fmt.Println(query)
			statement, err := activeConn.Prepare(query)
			if err != nil {
				fmt.Println("DeleteUsersInChat 1:",err)
				return errors.New("DB failed query")
			}
			s_deltime, err := json.Marshal(r_deltimes)
			if err != nil {
				fmt.Println("DeleteUsersInChat 2:",err)
				return errors.New("Fail encode r_deltimes")
			}
			//make hash of user's password
			delete_by_admin := 1
			if delete_yourself{
				delete_by_admin = 0
			}
			_, err = statement.Exec(1, time.Now().Unix()+1, string(s_deltime), delete_by_admin)
			if err != nil {
				fmt.Println("DeleteUsersInChat 3:",err)
				return errors.New("Failed exec statement")
			}
			statement.Close()
		}
	}

	return nil
}

func RecoveryUsersInChat(users_ids []float64, chat_id string)(error) {
	for _, v := range users_ids {
		//s_ids = append(s_ids, strconv.FormatFloat(v,'f',0,64))
		var deltimes string
		r_deltimes := [][]int64{}
		rows_user_in_chat, err := activeConn.Prepare("SELECT  deltimes FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
		if err != nil {
			return errors.New("Cant prove user isnt in chat")
			//panic(nil)
		}
		rows_user_in_chat.QueryRow(v, chat_id).Scan(&deltimes)
		err = json.Unmarshal([]byte(deltimes), &r_deltimes)
		if err != nil {
			return errors.New("Cant decode delete times")
			//panic(nil)
		}
		rows_user_in_chat.Close()
		s_id := strconv.FormatFloat(v, 'f', 0, 64)
		if r_deltimes[len(r_deltimes)-1][1] == 0 {
			r_deltimes[len(r_deltimes)-1][1] = time.Now().Unix()
			r_deltimes = append(r_deltimes, []int64{0,0})
			query := fmt.Sprintf("UPDATE people_in_chats SET delete_a = ?, deltime = ?, deltimes = ? where (user_id = %s) and (chat_id = %s)", s_id, chat_id)
			//fmt.Println(query)
			statement, err := activeConn.Prepare(query)
			if err != nil {
				fmt.Println("RecoveryUsersInChat 1:",err)
				return errors.New("DB failed query")
			}
			s_deltime, err := json.Marshal(r_deltimes)
			if err != nil {
				fmt.Println("RecoveryUsersInChat 1:",err)
				return errors.New("Fail encode r_deltimes")
			}
			//make hash of user's password
			_, err = statement.Exec(0, 0, string(s_deltime))
			statement.Close()
			if err != nil {
				fmt.Println("RecoveryUsersInChat 1:",err)
				return errors.New("Failed exec statement")
			}

		}
	}
	return nil
}

func GetSettings(chat_id string)(string, []map[string]string, error){
	var name,author_id, moders_ids string
	var moders []int64
	var r_moders []map[string]string
	rows_user_in_chat, err := activeConn.Prepare("SELECT  name, author_id, moders_ids FROM chats WHERE id=?")
	if err != nil {
		//return errors.New("Cant prove user isnt in chat")
		//panic(nil)
	}
	rows_user_in_chat.QueryRow(chat_id).Scan(&name, &author_id, &moders_ids)
	err = json.Unmarshal([]byte(moders_ids), &moders)
	if err != nil {
		return "", nil,errors.New("Cant unmarshal moders")
		//panic(nil)
	}
	rows_user_in_chat.Close()
	for i:=0;i<len(moders_ids);i++{
		rows_moders, _ := activeConn.Query("SELECT  u_name, login FROM people WHERE id=?", moders_ids[i])
		for rows_moders.Next() {
			var u_name, login string
			//var m_time int64
			if err := rows_moders.Scan(&u_name, &login); err != nil {
				return "", nil,err
			}
			r_moders = append(r_moders, map[string]string{"name": u_name, "login": login})
		}
		rows_moders.Close()
	}
	return name,r_moders,nil
}

func SetNameChat(chat_id string, name string)(error){
	query := fmt.Sprintf("UPDATE chats SET name = ? where id = %s",  chat_id)
	//fmt.Println(query)
	statement, err := activeConn.Prepare(query)
	if err != nil {
		fmt.Println("SetNameChat 1:",err)
		return errors.New("DB failed query")
	}
	//make hash of user's password
	_, err = statement.Exec(name)
	statement.Close()
	if err != nil {
		fmt.Println("SetNameChat 2:",err)
		return errors.New("Failed exec statement")
	}
	return nil
}

func DeleteMessages(chat_id string, user_id float64, users_ids []string)(error){
	chat_ids := strings.Join(users_ids, ", ")
	query := fmt.Sprintf("delete from messages where (id IN (%s)) and (user_id=?) and (chat_id=?)", chat_ids)
	//fmt.Println(query)
	stmt, err := activeConn.Prepare(query)
	if err != nil{
		fmt.Println("DeleteMessages 1:",err)
		return errors.New("Fail prepare delete ids")
	}
	_, err = stmt.Exec(user_id, chat_id)
	stmt.Close()
	if err != nil{
		fmt.Println("DeleteMessages 2:",err)
		return errors.New("Fail exec delete messages")
	}
	return  nil
}

func GetUserSettings(user_id float64)(map[string]interface{}, error){
	final := map[string]interface{}{}
	rows, err := activeConn.Prepare("SELECT login, u_name FROM people WHERE id=?")
	if err != nil {
		return nil, err
	}
	var login, u_name string
	//make hash of user's password
	query := rows.QueryRow(user_id)

	err = query.Scan(&login, &u_name)
	rows.Close()
	if err != nil {
		return nil, err
	}
	final["login"] = login
	final["name"]=u_name
	return final, nil
}

func SetUserSettings(user_id float64, name string)(error){
	query := fmt.Sprintf("UPDATE people SET u_name = ? where id = %s",  strconv.FormatFloat(user_id,'f',-1,64))
	statement, err := activeConn.Prepare(query)
	if err != nil {
		fmt.Println("SetUserSettings 1:",err)
		return errors.New("DB failed query")
	}
	//make hash of user's password
	_, err = statement.Exec(name)
	statement.Close()
	if err != nil {
		fmt.Println("SetUserSettings 2:",err)
		return errors.New("Failed exec statement")
	}
	return nil
}

//func GetUsersByName
func GetUsersForCreateDialog(user_id float64, name string)([]map[string]interface{},error){
	middle := make([]map[string]interface{},0)
	var dialogs_ids []string
	//var other_chats_ids []string
	query1:= fmt.Sprintf("SELECT  people_in_chats.user_id FROM chats INNER JOIN people_in_chats ON" +
		"		 people_in_chats.chat_id = chats.id WHERE (chats.id in "+
		"(SELECT  chats.id FROM chats INNER JOIN people_in_chats ON people_in_chats.chat_id = chats.id WHERE (chats.type=1) and (people_in_chats.user_id = ?)  and (people_in_chats.list_delete = 0)) )"+
		" and (chats.type=1) and (people_in_chats.user_id <> ?)")

	rows, err := activeConn.Query(query1, user_id, user_id)
	if err != nil {
		fmt.Println("GetUsersForCreateDialog 1:", err.Error())
		return nil,err
	}

	for rows.Next(){
		var id string
		if err := rows.Scan(&id); err != nil {
			fmt.Println("GetUsersForCreateDialog 2:", err.Error())
			return nil,err
		}
		dialogs_ids = append(dialogs_ids, id)
	}
	rows.Close()
	str_dialogs_ids := strings.Join(dialogs_ids, ",")

	//fmt.Println(str_dialogs_ids)
	name = "%"+name+"%"

	users_id_in_user_dialogs := fmt.Sprintf("SELECT DISTINCT id,u_name,login from people where (id not in (%s))" +
		" and (id <> ?) and (u_name LIKE ? or login LIKE ?)",str_dialogs_ids)

	rows, err = activeConn.Query(users_id_in_user_dialogs ,user_id, name, name)
	if err != nil {
		fmt.Println("GetUsersForCreateDialog 3:", err.Error())
		return nil,err
	}
	for rows.Next(){
		var id, u_name, login string
		if err := rows.Scan(&id, &u_name,  &login); err != nil {
			fmt.Println("GetUsersForCreateDialog 4:", err.Error())
			return nil,err
		}
		ret:= map[string]interface{}{}
		ret["id"],_ = strconv.ParseInt(id,10,64)
		ret["name"] = u_name
		ret["login"] = login
		middle=append(middle, ret)
	}
	rows.Close()
	return middle, nil
}

func HaveAlreadyDialog(user_id float64, another_user_id float64)(error, float64){
	var chat_id float64
	var s_delete_users string
	p_rows, err := activeConn.Prepare("SELECT  chat_id, delete_users FROM dialogs_info WHERE ((user_1=?) or (user_1=?)) and ((user_2=?) or (user_2=?))")
	if err != nil {
		fmt.Println("HaveAlreadyDialog 1:", err)
		return  err, 0
	}

	query := p_rows.QueryRow(user_id, another_user_id, user_id, another_user_id).Scan(&chat_id, &s_delete_users)
	//fmt.Println(s_chat_id)
	defer p_rows.Close()
	if query == sql.ErrNoRows{
		fmt.Println("HaveAlreadyDialog 2:", err)
		return nil, 0
	}
	return errors.New("That dialog already created"), chat_id
}

func CreateDialog(user_id float64, another_user_id float64)( *models.MessageContent, float64,int64,error){
	// Search, maybe db already have this dialog
	err, chat_id :=  HaveAlreadyDialog(user_id, another_user_id)

	if err != nil{
		//return nil,0,0,err
		//fmt.Println(chat_id)
		stmt, err := activeConn.Prepare("UPDATE people_in_chats SET list_delete=0 WHERE chat_id=? and user_id=?")
		if err != nil {
			fmt.Println("CreateDialog 1:", err)
			return  nil,0,0,errors.New("Failed permanent statement")
		}
		_, err = stmt.Exec(chat_id, user_id)
		stmt.Close()
		return  nil, chat_id,0,nil
	}
	//var dialogs_ids []float64
	//var other_chats_ids []string
	//may_be_users:= GetUsersForCreateDialog(user_id,)
	//for _,v:= range dialogs_ids{
	//	if v==another_user_id{
	//		return  nil,0,0,errors.New("Dialog with this user already exist")
	//	}
	//}
	statement, err := activeConn.Prepare("INSERT INTO chats (name,  author_id,moders_ids, type, lastmodify) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return  nil,0,0,err
	}
	//make hash of user's password
	res, err := statement.Exec("",  user_id,"[]", 1,time.Now().Unix())
	statement.Close()
	if err != nil {
		fmt.Println("Create Dialog 1: ", err.Error())
		return  nil,0,0,err
	}
	id, _ := res.LastInsertId()
	err = InsertUserInChat(strconv.FormatFloat(user_id,'f',-1,64), id)
	if err != nil {
		fmt.Println("Create Dialog 2: ", err.Error())
		return  nil,0,0,err
		//fmt.Println(fin)
	}
	err = InsertUserInChat(strconv.FormatFloat(another_user_id,'f',-1,64), id)
	if err != nil {
		fmt.Println("Create Dialog 3: ", err.Error())
		return  nil,0,0,err
		//fmt.Println(fin)
	}
	mess_mss := "начал эту беседу"
	docs := []string{}
	m_type := "a_msg"
	mess := models.MessageContent{&mess_mss, &docs, &m_type}
	data ,err := json.Marshal(mess)
	if err != nil{
		fmt.Println("Create Dialog: ", err.Error())
		return   nil,0,0,err
	}
	//f_id,err := strconv.ParseFloat(strconv.FormatFloat(user_id,'f',-1,64), 64)
	//if err != nil{
	//	return  err
	//}
	last_id,err := AddMessage(user_id, float64(id), string(data))
	if err != nil{
		return   nil,0,0,err
	}
	//last_id,_:= a_res.LastInsertId()
	statement, err = activeConn.Prepare("INSERT INTO dialogs_info (chat_id, user_1,user_2, delete_users) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Println("Create Dialog: ", err.Error())
		return  nil,0,0,err
	}
	//make hash of user's password
	res, err = statement.Exec(id, user_id, another_user_id, "[]")
	statement.Close()
	if err != nil {
		fmt.Println("Create Dialog: ", err.Error())
		return  nil,0,0,err
	}
	return  &mess, float64(id),last_id,nil

}

//func DeleteUserFromDialog(user_id float64, chat_id float64)(error){
//	var deltimes string
//	r_deltimes:= [][]int64{}
//	rows_user_in_chat, err := activeConn.Prepare("SELECT  deltimes FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
//	if err != nil {
//		return errors.New("Cant prove user isnt in chat")
//		//panic(nil)
//	}
//	rows_user_in_chat.QueryRow(user_id, chat_id).Scan(&deltimes)
//	err = json.Unmarshal([]byte(deltimes), &r_deltimes)
//	if err != nil {
//		return errors.New("Cant decode delete times")
//		//panic(nil)
//	}
//
//}



func FullDeleteUserFromChat(user_id float64, chat_id float64)(error){
	delete:= CheckUserInChatDelete(user_id,chat_id)
	//fmt.Println(user_id, chat_id)
	if delete == nil{
		fmt.Println("FullDeleteUserFromChat 1:", delete)
		return errors.New("User yet not delete")
	}
	rows, err := activeConn.Query("UPDATE people_in_chats SET list_delete=? WHERE user_id=? and chat_id=?",1, user_id, chat_id)
	if err != nil {
		fmt.Println("FullDeleteUserFromChat 2:", delete)
		//fmt.Println("Fail delete", err)
		return err
	}
	rows.Next()
	rows.Close()
	//p_rows, err := activeConn.Prepare("SELECT delete_users FROM dialogs_info WHERE chat_id=?")
	//if err != nil {
	//	fmt.Println("Fail update dialog info", err)
	//	return err
	//}
	//var del_user string
	////var del_user_b byte
	//var del_user_2 []float64
	//query := p_rows.QueryRow(user_id, chat_id).Scan(&del_user)
	//p_rows.Close()
	//if query == sql.ErrNoRows{
	//	return errors.New("User aren't in chat")
	//}
	//err = json.Unmarshal([]byte(del_user),del_user_2)
	//if err != nil {
	//	fmt.Println("Fail unmarhal delete  info", err)
	//	return err
	//}
	//del_user_2 = append(del_user_2, user_id)
	//del_user_b,err :=json.Marshal(del_user_2)
	//if err != nil {
	//	fmt.Println("Fail unmarhal delete  info", err)
	//	return err
	//}
	//del_user = string(del_user_b)
	//statement, err := activeConn.Prepare("UPDATE dialogs_info SET delete_users=? WHERE chat_id=?")
	//if err != nil {
	//	return errors.New("DB failed query")
	//}
	////make hash of user's password
	//statement.Exec(del_user, chat_id)
	//statement.Close()

	return nil
}

func check_instance_db(db *sql.DB)(bool){
	var date, version string
	rows, err := db.Query("SELECT version, data_instance FROM sys")
	if err!=nil{
		//fmt.Println(err.Error())
		return false
	}
	rows.Next()
	err = rows.Scan(&version, &date)
	if err!=nil{
		//fmt.Println(err.Error())
		return false
	}
	rows.Close()
	fmt.Println("DB date instance: "+date)
	fmt.Println("Spatium version: "+version)
	return true
}

func openSQLite(path string)(*sql.DB, error){
	newDB := false
	_, err := os.Open(path)
	if err != nil{
		newDB = true
		file, err := os.Create(path)
		if err != nil {
			fmt.Println("Cant create database...")
			return nil, err
		}
		defer file.Close()
		fmt.Println("Create database")
	}
	database, err := sql.Open("sqlite3", path)
	if err!=nil{
		return nil, err
	}
	if !check_instance_db(database) || newDB{
		createDB_structs(database)
		check_instance_db(database)
	}
	return database, nil
}

func OpenDB()(error){
	var database *sql.DB = nil
	var err error
	switch settings.ServiceSettings.DB.DataBaseType {
	case "l": database, err = openSQLite(settings.ServiceSettings.DB.SQLite.Path); break
	}
	if database!=nil{
		activeConn = database
		activeConnIsReal=true
	}else{
		return err
	}
	return nil

}

