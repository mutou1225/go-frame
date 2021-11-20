package apptask

import (
	"go-frame/frame/appengine/app"
	"go-frame/implements/toolkit"
	"go-frame/logger"
)

func funcTask() {
	logger.PrintInfo("funcTask", toolkit.GetCurrentTime())
}

// 注册定时任务
func AppRegisterTasks() []app.CronTask {
	var tasksList = make([]app.CronTask, 0)

	//添加定时任务
	tasksList = append(tasksList, app.CronTask{ "0 */5 * * * *", funcTask})

	return tasksList
}
