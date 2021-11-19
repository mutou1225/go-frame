package toolkit

import (
	"fmt"
	"os"
)

// 不存在返回0
func GetFileSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err == nil {
		return fileInfo.Size()
	}
	if os.IsNotExist(err) {
		return 0
	}
	return 0
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func GetExportExcelFileName(prefix string, flags ...string) string {
	flagStr := "Export"
	if len(flags) > 0 {
		for _, f := range flags {
			flagStr += f + "-"
		}
	}
	_, md5, _ := GetUniqId(flagStr)
	fileName := prefix + "-" + md5 + ".xlsx"
	return fileName
}

// descFile 文件名
// desc 描述
// reload 是否自动刷新
// rTime 刷新时间（默认3000）
func UpdateImportTaskDescFile(descFile, desc string, reload bool, rTime int) error {
	if rTime == 0 {
		rTime = 3000
	}

	reloadFun := ""
	if reload {
		reloadFun = fmt.Sprintf("<script>setTimeout(function(){location.reload();},%d);</script>", rTime)
	}

	formStr := `<!DOCTYPE html> 
	<html>
	<head>
		<meta charset="UTF-8">
		%s
	</head>
	<body>%s
	</body>
	</html>
	`
	htmlStr := fmt.Sprintf(formStr, desc, reloadFun)
	f, err := os.Create(descFile) //创建文件
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(htmlStr)
	if err != nil {
		return err
	}

	f.Sync()
	return nil
}
