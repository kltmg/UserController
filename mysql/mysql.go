package mysql

import (
	"UserController/config"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
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
	return nil

}
