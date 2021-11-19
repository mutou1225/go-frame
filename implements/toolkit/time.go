package toolkit

import (
	"strconv"
	"strings"
	"time"
)

func GetCurrentTime() time.Time {
	return time.Now().UTC()
}

// 获取时间，自定义时区
func GetCurrentTimeAdd(add int) time.Time {
	return time.Now().UTC().Add(time.Duration(add) * time.Hour)
}

func GetTimeStamp() int64 {
	return time.Now().UTC().Unix()
}

func GetNanoTimeStamp() int64 {
	return time.Now().UTC().UnixNano()
}

func GetMSTimeStamp() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

func GetSecTimeStamp() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Second)
}

func Time2MSTimeStamp(t *time.Time) int64 {
	ts := t.UTC().UnixNano() / int64(time.Millisecond)
	if ts > 0 {
		return ts
	} else {
		return 0
	}
}

func GetCurrTimeStr() string {
	return time.Now().Format(CtLayoutStr)
}

// 字符串转时间（带时区）
func StrToTime(t string) (time.Time, error) {
	tm, err := time.Parse(CtLayoutStr, t)
	if err != nil {
		return time.Time{}, err
	}
	return tm, nil
}

type CustomTime struct {
	time.Time
}

const (
	CtLayoutStr = "2006-01-02 15:04:05"

	ctLayout       = "2006/01/02|15:04:05"
	ctLayoutDayKey = "20060102"
)

var nilTime = (time.Time{}).UnixNano()

func (ct *CustomTime) GetCurrTimeStr() string {
	return ct.Time.Format(CtLayoutStr)
}

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	ct.Time, err = time.Parse(ctLayout, string(b))
	return
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(ct.Time.Format(ctLayout)), nil
}

func (ct *CustomTime) IsSet() bool {
	return ct.UnixNano() != nilTime
}

//获取相差时间
func GetDateDifferDay(d time.Time, days int) time.Time {
	return d.AddDate(0, 0, days)
}

//获取当月开始时间
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTimeOfDay(d)
}

//获取当月结束时间
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

//获取当周开始时间
func GetFirstDateOfWeek(d time.Time) time.Time {
	d = d.AddDate(0, 0, int(-d.Weekday())+1)
	return GetZeroTimeOfDay(d)
}

//获取当周结束时间
func GetLastDateOfWeek(d time.Time) time.Time {
	return GetFirstDateOfWeek(d).AddDate(0, 0, 7)
}

func GetCurrentDayStr(d time.Time) string {
	return d.Format(ctLayoutDayKey)
}

//获取当天的0点时间
func GetZeroTimeOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func ParseTimeOfStr(unixT int64) string {
	return time.Unix(unixT, 0).Format(CtLayoutStr)
}

func ParseTimeOfCustom(unixT int64, layStr string) string {
	return time.Unix(unixT, 0).Format(layStr)
}

func GetCurrentDay(d time.Time, hour, min int) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), hour, min, 0, 0, d.Location())
}

func GetActivityBeginTime(beginTimeEv string) (int64, error) {
	beginTsli := strings.Split(beginTimeEv, ":")
	beginHour := beginTsli[0]
	beginHourInt, err := strconv.Atoi(beginHour)
	if err != nil {
		return 0, err
	}
	beginMin := beginTsli[1]
	beginMinInt, err := strconv.Atoi(beginMin)
	if err != nil {
		return 0, err
	}
	beginTime := GetCurrentDay(time.Now(), beginHourInt, beginMinInt)

	return beginTime.Unix(), nil
}

func GetExpireTimeDay(expireUnixT, nowUnix int64) int32 {
	return int32(time.Unix(expireUnixT, 0).Sub(time.Unix(nowUnix, 0)).Hours() / 24)
}

//获取时间段内的每天日期
func GetDayPoint(bTime, eTime int64) (point []string) {
	bT := time.Unix(bTime, 0)
	eT := time.Unix(eTime, 0)
	for {
		point = append(point, GetFirstDateOfWeek(bT).Format("2006/01/02"))
		if bT.Unix() >= eT.Unix() {
			return
		}
		bT = time.Unix(bT.Unix()+24*3600, 0)
	}
}

//获取时间段内的每周日期
func GetWeekPoint(bTime, eTime int64) (point []string) {
	bT := time.Unix(bTime, 0)
	eT := time.Unix(eTime, 0)
	for {
		point = append(point, GetFirstDateOfWeek(bT).Format("2006/01/02"))
		bT = time.Unix(bT.Unix()+24*3600*7, 0)
		if bT.Unix() >= eT.Unix() {
			return
		}
	}
}

//获取时间段内的每月日期
func GetMonthPoint(bTime, eTime int64) (point []string) {
	bT := time.Unix(bTime, 0)
	eT := time.Unix(eTime, 0)
	for {
		point = append(point, GetFirstDateOfMonth(bT).Format("2006/01"))
		lastDay := GetLastDateOfMonth(bT).Format("2006/01/02")
		days := strings.Split(lastDay, "/")
		d := days[len(days)-1]
		n, _ := strconv.Atoi(d)
		bT = time.Unix(bT.Unix()+24*3600*int64(n), 0)
		if bT.Unix() >= eT.Unix() {
			return
		}
	}
}

type StartAndStopTime struct {
	StartTime int64
	StopTime  int64
}

//获取7天日期
func GetNDayPoint(n int, eTime int64) (startT, endT int64, m map[string]StartAndStopTime) {
	m = make(map[string]StartAndStopTime)
	var startTime time.Time
	eT := time.Unix(eTime, 0)

	startTime = GetZeroTimeOfDay(eT)
	s := StartAndStopTime{
		StartTime: startTime.Unix(),
		StopTime:  startTime.Unix() + 24*3600 - 1,
	}
	endT = s.StopTime
	m[startTime.Format("2006/01/02")] = s
	for i := 1; i < n; i++ {
		s = StartAndStopTime{
			StartTime: startTime.AddDate(0, 0, -1*i).Unix(),
		}
		s.StopTime = s.StartTime + 24*3600 - 1
		m[time.Unix(s.StartTime, 0).Format("2006/01/02")] = s
		startT = s.StartTime
	}
	return
}

//获取7周日期
func GetNWeekPoint(n int, eTime int64) (startT, endT int64, m map[string]StartAndStopTime) {
	m = make(map[string]StartAndStopTime)
	var startTime time.Time
	eT := time.Unix(eTime, 0)

	startTime = GetFirstDateOfWeek(eT)
	s := StartAndStopTime{
		StartTime: startTime.Unix(),
		StopTime:  startTime.Unix() + 24*3600*7 - 1,
	}
	endT = s.StopTime
	m[startTime.Format("2006/01/02")] = s
	for i := 1; i < n; i++ {
		s = StartAndStopTime{
			StartTime: startTime.AddDate(0, 0, -7*i).Unix(),
		}
		s.StopTime = s.StartTime + 24*3600*7 - 1
		m[time.Unix(s.StartTime, 0).Format("2006/01/02")] = s
		startT = s.StartTime
	}
	startT = s.StartTime
	return
}

//获取7月日期
func GetNMonthPoint(n int, eTime int64) (startT, endT int64, m map[string]StartAndStopTime) {
	m = make(map[string]StartAndStopTime)
	var startTime time.Time
	eT := time.Unix(eTime, 0)

	startTime = GetFirstDateOfMonth(eT)
	s := StartAndStopTime{
		StartTime: startTime.Unix(),
	}
	s.StopTime = time.Unix(s.StartTime, 0).AddDate(0, 1, 0).Unix() - 1
	endT = s.StopTime
	m[startTime.Format("2006/01")] = s
	for i := 1; i < n; i++ {
		s = StartAndStopTime{
			StartTime: startTime.AddDate(0, -1*i, 0).Unix(),
		}
		s.StopTime = time.Unix(s.StartTime, 0).AddDate(0, 1, 0).Unix() - 1
		m[time.Unix(s.StartTime, 0).Format("2006/01")] = s
		startT = s.StartTime
	}
	return
}

func GetDayPoint2(st, et int64) (points []string) {
	t := time.Unix(et, 0)
	st = GetZeroTimeOfDay(time.Unix(st, 0)).Unix()
	for {
		day := t.Format("2006/01/02")
		points = append(points, day)
		t = t.AddDate(0, 0, -1)
		if t.Unix() < st {
			break
		}
	}
	return
}

func GetWeekPoint2(st, et int64) (points []string) {
	wst := GetFirstDateOfWeek(time.Unix(et, 0))
	for {
		day := wst.Format("2006/01/02")
		if (wst.Unix() < st) && ((wst.AddDate(0, 0, 7).Unix() - 1) < st) {
			break
		}
		wst = wst.AddDate(0, 0, -7)
		points = append(points, day)
	}
	return
}

func GetMonthPoint2(st, et int64) (points []string) {
	wst := GetFirstDateOfMonth(time.Unix(et, 0))
	for {
		day := wst.Format("2006/01")
		if (wst.Unix() < st) && ((wst.AddDate(0, 1, 0).Unix() - 1) < st) {
			break
		}
		wst = wst.AddDate(0, -1, 0)
		points = append(points, day)
	}
	return
}

func GetYesterDayBeginAndEndTimestamp() (int64, int64) {
	now := time.Now()
	//nowStr:= now.Format("2006-01-02 15:04:05")
	//fmt.Println(nowStr)
	lastDay := now.AddDate(0, 0, -1)
	lastDayStart := time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 0, 0, 0, 0, now.Location())
	//lastDayStartStr := lastDayStart.Format("2006-01-02 15:04:05")
	lastDayStartInt := lastDayStart.Unix()
	//fmt.Println(lastDayStartStr)
	lastDayEnd := time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 23, 59, 59, 0, now.Location())
	//lastDayEndStr := lastDayEnd.Format("2006-01-02 15:04:05")
	lastDayEndInt := lastDayEnd.Unix()
	//fmt.Printf("begin:%d end:%d\n",lastDayStartInt,lastDayEndInt)
	//fmt.Println(lastDayEndStr)
	return lastDayStartInt, lastDayEndInt

}

// 获取到第二天0点的秒数
func GetNextDaySecond() time.Duration {
	next := time.Now().Add(time.Hour * 24)
	next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	return next.Sub(time.Now())
}
