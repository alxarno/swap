package spatium_db_work

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"crypto/sha256"
	"os"
	"fmt"
	models "github.com/AlexArno/spatium/models"
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
		panic(nil)
	}
	query = rows.QueryRow(login).Scan(&id_now)
	if query == sql.ErrNoRows{
		return "","Some is fail",err
	}
	return id_now,"Success", nil
}

func createDB_structs(database *sql.DB){
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, login TEXT, pass TEXT, u_name TEXT)")
	statement.Exec()
	_, fin, err := CreateUser("god","1111", "Alex")
	if err!= nil{
		fmt.Println(fin)
	}
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

