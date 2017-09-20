package spatium_db_work

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
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

func GetUser(login string, pass string)(*models.User, error){
	user := new(models.User)
	if !activeConnIsReal{
		OpenDB()
	}
	rows, err := activeConn.Prepare("SELECT id, login, pass, u_name FROM people WHERE (login=?) AND (pass=?)")
	if err != nil {
		panic(nil)
	}
	defer rows.Close()

	err = rows.QueryRow(login, pass).Scan(&user.ID, &user.Login, &user.Pass, &user.Name)
	if err != nil {
		return nil, err
	}
	return user,nil
}

func createDB_structs(database *sql.DB){
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, login TEXT, pass TEXT, u_name TEXT)")
	statement.Exec()
	statement, _ = database.Prepare("INSERT INTO people (login, pass, u_name) VALUES (?, ?, ?)")
	statement.Exec("pussy", "1111","Alex")
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

func main(){
	fmt.Println("DB is here")
}