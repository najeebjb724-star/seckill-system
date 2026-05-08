package svc

import (
	"seckill-system/internal/config"
	"seckill-system/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config            config.Config
	Redis             *redis.Redis
	ActivityModel     model.ActivityModel
	SeckillOrderModel model.SeckillOrderModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql("root:password@tcp(127.0.0.1:3306)/seckill?parseTime=true")


	redisClient, err := redis.NewRedis(redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
	})
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:            c,
		Redis:             redisClient,
		ActivityModel:     model.NewActivityModel(conn),
		SeckillOrderModel: model.NewSeckillOrderModel(conn),
	}
}
