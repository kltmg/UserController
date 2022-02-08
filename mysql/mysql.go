package mysql

import (
	"UserController/config"
	"UserController/utils"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var (
	createAccountSt    *sql.Stmt
	loginAuthSt        *sql.Stmt
	createProfileSt    *sql.Stmt
	getProfileSt       *sql.Stmt
	updateProfileSt    *sql.Stmt
	updateNickNameSt   *sql.Stmt
	updateProfilePicSt *sql.Stmt
)

func Connect() error {
	DATABASE := "usercontroller"
	Config := config.GetConfig()
	MysqlDB := Config.MYSQL.UserName + ":" + Config.MYSQL.PassWord + "@(" + Config.MYSQL.Host + ":" + Config.MYSQL.Port + ")/" + DATABASE + "?charset=utf8"

	db, err := sql.Open("mysql", MysqlDB)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(Config.MYSQL.ConnMaxLifetime)
	db.SetMaxIdleConns(Config.MYSQL.MaxIdleConns)
	db.SetMaxOpenConns(Config.MYSQL.MaxOpenConns)

	if err = db.Ping(); err != nil {
		return err
	}

	loginAuthSt = dbPrepare(db, "SELECT password FROM login_info WHERE user_name = ?")
	getProfileSt = dbPrepare(db, "SELECT nick_name, pic_name FROM user_info WHERE user_name = ?")
	createAccountSt = dbPrepare(db, "INSERT INTO login_info (user_name, password) values (?, ?)")
	createProfileSt = dbPrepare(db, "INSERT INTO user_info (user_name, nick_name) values (?, ?)")
	return nil

}

func dbPrepare(db *sql.DB, query string) *sql.Stmt {
	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	return stmt
}

func LoginAuth(username string, password string) (bool, error) {
	var pwd string
	row := loginAuthSt.QueryRow(username)
	err := row.Scan(&pwd)

	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if pwd == utils.Sha256(password) {
		return true, nil
	}
	return false, err
}

func GetProfile(userName string) (nickName string, picName string, hasData bool, err error) {
	rows, err := getProfileSt.Query(userName)
	if err != nil {
		return nickName, picName, hasData, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&nickName, &picName)
	}
	if err != nil {
		return nickName, picName, hasData, err
	}
	if nickName != "" {
		hasData = true
	}
	return nickName, picName, hasData, nil
}

func CreateAccount(userName string, passWord string) error {
	pwd := utils.Sha256(passWord)
	_, err := createAccountSt.Exec(userName, pwd)
	if err != nil {
		return err
	}
	return nil
}

func CreateProfile(userName string, nickName string) error {
	_, err := createProfileSt.Exec(userName, nickName)
	if err != nil {
		return err
	}
	return nil
}
