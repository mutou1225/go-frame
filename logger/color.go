package logger

import (
	"go-frame/config"
)

// \033[32;1m
// \033[0m 关闭所有属性
//\033[1m 设置高亮度
//\033[4m 下划线
//\033[5m 闪烁
//\033[7m 反显
//\033[8m 消隐
//\033[30m 至 \33[37m 设置前景色
//\033[40m 至 \33[47m 设置背景色
//\033[nA 光标上移n行
//\033[nB 光标下移n行
//\033[nC 光标右移n行
//\033[nD 光标左移n行
//\033[y;xH设置光标位置
//\033[2J 清屏
//\033[K 清除从光标到行尾的内容
//\033[s 保存光标位置
//\033[u 恢复光标位置
//\033[?25l 隐藏光标
//\033[?25h 显示光标<br>

// 使用后要用Reset重置
var (
	Black     = "\033[30;1m"
	Red       = "\033[31;1m"
	Green     = "\033[32;1m"
	Yellow    = "\033[33;1m"
	Blue      = "\033[34;1m"
	Purple    = "\033[35;1m"
	DarkGreen = "\033[36;1m"
	White     = "\033[37;1m"
	Reset     = "\033[0m"
)

func (l *MyLogger)InitColor() {
	if !config.GLogConfig.LogColor {
		Black     = ""
		Red       = ""
		Green     = ""
		Yellow    = ""
		Blue      = ""
		Purple    = ""
		DarkGreen = ""
		White     = ""
		Reset     = ""
	}
}
