package mysql

import (
	"UserController/config"
	"UserController/utils"
	"fmt"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
)

type account struct {
	User_name string `json:"user_name"`
	Password  string `json:"password"`
}
type profile struct {
	User_name string `json:"user_name"`
	Nick_name string `json:"nick_name"`
}

const (
	TOTAL_INSERT_NUM = 10000000 //共插入多少行数据
	PER_INSERT_NUM   = 10000    //单次向mysql插入多少行数据
	MAX_FAILNUM      = 10       //最大容许插入失败次数
)

//初始化测试数据10 000 000条
func TestInitAccount(t *testing.T) {
	err := config.ReadConfig("..\\config\\config.yaml")
	if err != nil {
		panic(err)
	}
	DATABASE := "usercontroller"
	Config := config.GetConfig()
	MysqlDB := Config.MYSQL.UserName + ":" + Config.MYSQL.PassWord + "@(" + Config.MYSQL.Host + ":" + Config.MYSQL.Port + ")/" + DATABASE + "?charset=utf8"

	createAccountSQL := `INSERT INTO login_info (user_name, password) values (:user_name, :password)`
	createProfileSQL := `INSERT INTO user_info (user_name, nick_name) values (:user_name, :nick_name)`

	db, err := sqlx.Open("mysql", MysqlDB)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	//var accountArray [PER_INSERT_NUM]account
	//var profileArray [PER_INSERT_NUM]profile
	pwd := utils.Sha256("123")
	accountArray := make([]account, PER_INSERT_NUM)
	profileArray := make([]profile, PER_INSERT_NUM)
	for i := 8340000; i < TOTAL_INSERT_NUM; i++ {
		userName := "bot" + strconv.Itoa(i)
		accountArray[i%PER_INSERT_NUM] = account{User_name: userName, Password: pwd}
		profileArray[i%PER_INSERT_NUM] = profile{User_name: userName, Nick_name: "bot"}
		//fmt.Println(accountArray, profileArray)
		if i%PER_INSERT_NUM == PER_INSERT_NUM-1 {
			conn, _ := db.Beginx()
			_, err = conn.NamedExec(createAccountSQL, accountArray)
			if err != nil {
				fmt.Println(err)
			}
			_, err = conn.NamedExec(createProfileSQL, profileArray)
			if err != nil {
				fmt.Println(err)
			}
			conn.Commit()
			fmt.Println("finish 10000...")
		}
	}

}
