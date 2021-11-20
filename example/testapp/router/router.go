package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mutou1225/go-frame/example/testapp/router/api"
)

// 初始化路由
func InitAppRouter(r *gin.Engine) {
	evaApiG := r.Group("/test")
	{
		evaApiG.POST("/test_redis_get", api.TestRedisGetApi)
		evaApiG.POST("/test_redis_set", api.TestRedisSetApi)
		evaApiG.POST("/test_es", api.TestEsApi)
		evaApiG.POST("/test_mysql_get", api.TestMysqlGetApi)
	}
}
