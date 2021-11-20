package app

import (
	"github.com/robfig/cron/v3"
	"go-frame/logger"
	"time"
)

const (
	Cron = "0 */10 * * * *"
)

var (
	crontab *cron.Cron
)

type CronTask struct {
	Cron     string // 定时参数的格式 Second | Minute | Hour | Dom | Month | Week
	TaskFunc func()
}

// 开启定时任务
func StartCronTask(taskList []CronTask) error {
	if len(taskList) == 0 {
		logger.PrintInfo("Task List Empty, CronTask Do Nothing.")
		return nil
	}

	go func() {
		// New crontab
		crontab = cron.New(cron.WithSeconds(), cron.WithLogger(CronLog{}))
		if crontab == nil {
			logger.PrintError("cron.New() Fail!")
			return
		}
		// 添加定时任务
		for _, tFunc := range taskList {
			if id, err := crontab.AddFunc(tFunc.Cron, tFunc.TaskFunc); err != nil {
				logger.PrintError("crontab AddFunc[%d] err[%s]", id, err.Error())
			}
		}
		// 启动定时器
		crontab.Start()
		logger.PrintInfo("crontab Start ...")
		select {}
	}()

	return nil
}

func StopCronTask() {
	if crontab != nil {
		ctx := crontab.Stop()
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
		}
	}
	logger.PrintInfo("Cron was done")
}

type CronLog struct{}

// Info logs routine messages about cron's operation.
func (c CronLog) Info(msg string, keysAndValues ...interface{}) {
	logger.PrintInfoCalldepth(3, "%s %+v", msg, keysAndValues)
}

// Error logs an error condition.
func (c CronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		logger.PrintErrorCalldepth(3, "Cron Error: %s", err.Error())
	} else {
		logger.PrintErrorCalldepth(3, "%s %+v", msg, keysAndValues)
	}
}
