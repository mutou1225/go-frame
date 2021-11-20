package main

import (
	"github.com/mutou1225/go-frame/application/testapp/appstorage"
	"github.com/mutou1225/go-frame/application/testapp/apptask"
	"github.com/mutou1225/go-frame/application/testapp/router"
	"github.com/mutou1225/go-frame/frame/appengine"
)

const (
	//Version 版本
	Version = "010000"
	//VersionEx 版本
	VersionEx = "1.0.0"
	//Update 版本
	Update = "2021-2-19 17:46:00"
	//服务名
	AppName = "TestApp"
)

func main() {
	//配置文件初始化
	logName := AppName
	configFile := "/huishoubao/config/tinyxml2/eva_pro_config.xml"
	serverCfgFile := "/huishoubao/config/TestAppServer.xml"

	// 初始化系统框架
	appengine.InitApplication(configFile, serverCfgFile, logName)
	defer appengine.ExitApplication(appstorage.ExitStorage)

	// app 路由
	appengine.InitAppframe(router.InitAppRouter, appstorage.InitStorage, apptask.AppRegisterTasks)

	// 运行
	appengine.RunApplication()

	// 单独运行一些应用
	// appengine.RunCustomProgram(model.TestMQFunc)
}
