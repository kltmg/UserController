package main

import (
	"UserController/config"
	"UserController/mysql"
	"UserController/redis"
	"fmt"
)

func initConn() {
	err := config.ReadConfig("..\\config\\config.yaml")
	if err != nil {
		panic(err)
	}

	//connect mysql
	if err = mysql.Connect(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("mysql connection successed.")
	}

	//connect redis
	if err = redis.Connect(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("redis connection successed.")
	}
}

func main() {
	initConn()
}
