package redis

import (
	"UserController/config"
	"time"

	"github.com/go-redis/redis"
)

var client *redis.Client

func Connect() error {
	Config := config.GetConfig()
	client = redis.NewClient(&redis.Options{
		Addr:     Config.REDIS.Host + ":" + Config.REDIS.Port,
		Password: "",
		// 数据库
		DB: 0,
		// 连接池大小 Maximum number of socket connections.
		PoolSize: Config.REDIS.PoolSize,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func CheckToken(userName string, token string) (bool, error) {
	val, err := client.Get("auth_" + userName).Result()
	if err != nil {
		return false, err
	}
	return token == val, nil
}

func GetProfile(userName string) (nickName string, picName string, hasData bool, err error) {
	vals, err := client.HGetAll(userName).Result()
	if err != nil {
		return "", "", false, err
	}
	if vals["vaild"] != "" {
		hasData = true
	}
	return vals["nick_name"], vals["pic_name"], hasData, nil
}

func SetNickNameAndPicName(userName string, nickName string, picName string) error {
	fileds := map[string]interface{}{
		"vaild":     "1",
		"nick_name": nickName,
		"pic_name":  picName,
	}
	err := client.HMSet(userName, fileds).Err()
	if err != nil {
		return err
	}
	return nil
}

func SetToken(userName string, token string, expiration int64) error {
	err := client.Set("auth_"+userName, token, time.Duration(expiration*1e9)).Err()
	if err != nil {
		return err
	}
	return nil
}

func InvaildCache(username string) error {
	err := client.HSet(username, "vaild", "").Err()
	if err != nil {
		return err
	}
	return nil
}
