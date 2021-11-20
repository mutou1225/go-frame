package appengine

import (
	"github.com/gin-gonic/gin"
	cfg "go-frame/config"
	"go-frame/frame/appengine/app"
	"go-frame/frame/errcode"
	"go-frame/implements/opentracing"
	"go-frame/logger"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

// Application ...
type Application struct {
	Name      string
	waitGroup *sync.WaitGroup
	//Type           int32
	//LoggerRootPath string
	//SetupVars      func() error
}

// ListenerApplication ...
type WEBApplication struct {
	Application
	AppPort        int
	MonitorEndPort int
	// 监控使用的http server
	// MonitorMux *http.ServeMux
	// RegisterHttpRoute 定义HTTP router
	RegisterHttpRoute func(r *gin.Engine)
	// 系统定时任务
	RegisterTasks func() []app.CronTask
	// 应用存储的初始化
	AppStorageInit func()
}

var (
	application WEBApplication
)

// 初始化整个系统
func InitApplication(configFile, serverCfgFile, appName string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//初始化日志配置系统
	cfg.InitLogConfig()

	//初始化配置系统
	cfg.InitConfig(configFile)

	// 初始化服务的配置文件
	cfg.InitServerConfig(serverCfgFile)

	// 初始化app程序名
	appSetName(appName)

	//初始化日志系统
	logger.InitLogger(appName, cfg.GetSerLogFileName())

	// 打印配置信息
	logger.PrinfInterface(cfg.GetAllConfig())
	logger.PrinfInterface(cfg.GetAllSerConfig())

	// 初始化 OpenTracing
	opentracing.NewOpenTracing(cfg.GetServerName())
}

// 系统退出的释放操作
func ExitApplication(appExitFunc func()) {
	app.StopCronTask()
	appExitFunc()

	// 打印系统最后退出
	logger.ExitLogger()
}

// InitAppframe
func InitAppframe(appHttpRoute func(r *gin.Engine), appDbInit func(), taskFunc func() []app.CronTask) {
	application.AppPort = cfg.GetServerPort()
	application.MonitorEndPort = cfg.GetSerMonitorPort()
	//application.MonitorMux = server_mux.NewServerMux()
	application.RegisterHttpRoute = appHttpRoute
	application.RegisterTasks = taskFunc
	application.AppStorageInit = appDbInit
}

// RunApplication
func RunApplication() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("RunApplication() Err: %s", errcode.GetSystemPanic(err))
			logger.PrintError("RunApplication() Err: %s", errcode.GetSystemPanic(err))
		}
	}()

	if application.Name == "" {
		log.Print("Application name can't not be empty")
		logger.PrintError("Application name can't not be empty")
		application.Name = "unknown-app"
	}

	// 检测是否有端口
	if application.AppPort == 0 && application.RegisterHttpRoute != nil {
		logger.PrintPanic("App Port is 0!")
	}

	// 设置服务器监视器
	//app.InitMonitor(application.MonitorEndPort, application.MonitorMux)

	// 初始化存储
	if application.AppStorageInit != nil {
		application.AppStorageInit()
	}

	// 运行定时任务
	if application.RegisterTasks != nil {
		err := app.StartCronTask(application.RegisterTasks())
		if err != nil {
			log.Printf("StartCronTask err[%s]", err.Error())
			logger.PrintError("StartCronTask err[%s]", err.Error())
		}
	}

	// 初始化gin
	if application.AppPort != 0 {
		router := app.InitRouter()
		if router == nil {
			logger.PrintPanic("App InitRouter() nil !")
		}

		// 检测是否有路由
		if application.RegisterHttpRoute == nil {
			logger.PrintPanic("App RegisterHttpRoute nil ??")
		} else {
			// 加载app的路由
			application.RegisterHttpRoute(router)
		}

		app.StartServer(router, application.AppPort)

	} else {
		log.Printf("AppPort Err: %d", application.AppPort)
	}

	log.Print("Application RunApplication() Exit.")
}

// 启动自定义程序
func RunCustomProgram(program ...func(*sync.WaitGroup, chan struct{})) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("RunCustomProgram() Err: %s", errcode.GetSystemPanic(err))
			logger.PrintError("RunCustomProgram() Err: %s", errcode.GetSystemPanic(err))
		}
	}()

	if application.Name == "" {
		log.Print("Application name can't not be empty")
		logger.PrintError("Application name can't not be empty")
		application.Name = "unknown-app"
	}

	// 设置服务器监视器
	//app.InitMonitor(application.MonitorEndPort, application.MonitorMux)

	// 初始化存储
	if application.AppStorageInit != nil {
		application.AppStorageInit()
	}

	// 运行定时任务
	if application.RegisterTasks != nil {
		err := app.StartCronTask(application.RegisterTasks())
		if err != nil {
			logger.PrintError("StartCronTask err[%s]", err.Error())
		}
	}

	// 运行程序
	endChanList := make([]chan struct{}, len(program))
	application.waitGroup = &sync.WaitGroup{}
	for i, cfunc := range program {
		application.waitGroup.Add(1)
		endChanList[i] = make(chan struct{})
		go cfunc(application.waitGroup, endChanList[i])
	}

	go signalQuitApp(endChanList)
	application.waitGroup.Wait()

	log.Print("Application Exit.")
	logger.PrintInfo("Application RunCustomProgram() Exit.")
}

func signalQuitApp (endChanList []chan struct{}) {
	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	sig := <-quitChan

	for _, end := range endChanList {
		end <- struct{}{}
	}

	logger.PrintInfo("got a signal: %v app ending", sig)
}

// 获取app名
func AppGetName() string {
	return application.Name
}

// 设置app名
func appSetName(name string) {
	application.Name = name
}
