package redis

import (
	"UserController/config"

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
