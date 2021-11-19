package app

import (
	"context"
	"eva_services_go/config"
	"eva_services_go/frame/middleware"
	"eva_services_go/frame/servermux"
	"eva_services_go/logger"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"io"
	"os"
)

/*
  RegisterHttpRoute 此处注册http接口
  类似nginx的access、error日志
*/

/*
func RegisterHttpRoute() *gin.Engine {
	accessInfoLogger := &AccessInfoLogger{}
	accessErrLogger := &AccessErrLogger{}
	ginRouter := router.InitRouter(accessInfoLogger, accessErrLogger)
	return ginRouter
}
*/

// 初始化Router
func InitRouter() *gin.Engine {

	// 初始化gin
	gin.DefaultWriter = io.MultiWriter(os.Stdout, &AccessInfoLogger{})
	gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, &AccessErrLogger{})

	if config.IsTest() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if config.IsTest() {
		pprof.Register(r)
	}

	//r.GET("/", app.IndexApi)
	r.GET("/ping", PingApi)
	r.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, middleware.StatsReport())
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusOK)
		//c.File("/etc/nginx/favicon.ico")
	})

	apiDebug := r.Group("/debug")
	{
		apiDebug.GET("/vars", servermux.ExpvarHandler)
		apiDebug.GET("/metrics", servermux.MetricsHandler)
		apiDebug.GET("/heartbeat", Heartbeat)
	}

	r.Use(middleware.RequestStats())
	//r.Use(middleware.InitContext())
	r.Use(middleware.ThrowPanic())
	r.Use(middleware.TimeoutMiddleware(3 * time.Minute))
	r.Use(middleware.PrintPostData())
	r.Use(middleware.CheckCallSign())

	return r
}

/*
// 初始化Monitor
func InitMonitor(port int, handler *http.ServeMux) {
	if port != 0 {
		go func() {
			addr := "0.0.0.0:" + strconv.Itoa(port)
			logger.PrintInfo("App run monitor server addr: %v", addr)
			log.Printf("App run monitor server addr: %v", addr)
			err := http.ListenAndServe(addr, handler)
			if err != nil {
				logger.PrintError("App run monitor server err: %v", err)
				log.Printf("App run monitor server err: %v", err)
			}
		}()
	}
}
*/

type AccessInfoLogger struct{}

// gin Write
func (a *AccessInfoLogger) Write(p []byte) (n int, err error) {
	logger.PrintInfo(" %s", p)
	return 0, nil
}

type AccessErrLogger struct{}

// gin Write
func (a *AccessErrLogger) Write(p []byte) (n int, err error) {
	logger.PrintError(" %s", p)
	return 0, nil
}

func StartServer(router *gin.Engine, appPort int) {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", appPort),
		Handler: router,
	}
	logger.PrintInfo("App run monitor server addr[%s]", server.Addr)

	go QuitServer(server)

	err := server.ListenAndServe()
	if err != nil {
		logger.PrintInfo("App Run Err: %s", err.Error())
		log.Printf("App Run Err: %s", err.Error())
	}
}

func QuitServer(server *http.Server) {
	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	sig := <-quitChan
	logger.PrintInfo("got a signal: %v", sig)
	signal.Stop(quitChan)

	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(cxt); err != nil {
		logger.PrintInfo("server.Shutdown() Err: ", err.Error())
	}

	// 看看实际退出所耗费的时间
	logger.PrintInfo("QuitServer: %v", time.Since(now))
}
