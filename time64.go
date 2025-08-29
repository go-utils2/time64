package time64

import (
	"bytes"
	"fmt"

	"reflect"
	"strconv"
	"time"

	"github.com/go-utils2/time2"
	"github.com/pkg/errors"
)

// TimePeriod 时间段
type TimePeriod struct {
	MinutesFromZero int64  // 从0点开始的分钟数
	Info            string // 文本信息，格式为12:00
}

// NowInDefault 默认时区的当前时间
func NowInDefault() time.Time {
	return time2.Now().In(time.Local)
}

// GetStartOfDayInLocation 获取指定时区的指定天的第一秒
func GetStartOfDayInLocation(date int, location *time.Location) (start time.Time, err error) {
	var (
		day time.Time
	)
	if day, err = time.Parse(`20060102`, fmt.Sprintf(`%d`, date)); err != nil {
		return start, errors.Wrapf(err, `解析天[%d]`, date)
	}
	start = time.Unix(day.Unix()/(secondsInDay)*secondsInDay, 0)
	start = start.Add(time.Duration(int64(time.Second) * -1 * int64(Offset(location))))
	return start, nil
}

// Offset 在指定时区偏移了多少秒
func Offset(location *time.Location) int {
	_, offset := time2.Now().In(location).Zone()
	return offset
}

const (
	secondsInDay = 24 * 60 * 60
	layout       = time.DateTime
)

// GetDate 获取int类型的日期  当前天 20060102的格式，4位年2位月2位日
func GetDate() int {
	now := NowInDefault()
	return GetDate2(now)
}

func GetDate2(now time.Time) int {
	return now.Year()*before4Mask + int(now.Month())*after2Mask + now.Day()
}

// GetMonth 获取默认时区的当前月 200601的格式，4位年2位月
func GetMonth() int {
	now := NowInDefault()
	return now.Year()*after2Mask + int(now.Month())
}

// FormatDateTime 转换时间格式 date格式为yyyymmdd
func FormatDateTime(date int) time.Time {
	return time.Date(date/before4Mask, time.Month(date%before4Mask/after2Mask), date%after2Mask, 0, 0, 0, 0, time.Local)
}

// FormatDateTimeInDefault 转换时间格式 date格式为yyyymmdd
func FormatDateTimeInDefault(date int) time.Time {
	return time.Date(date/before4Mask, time.Month(date%before4Mask/after2Mask), date%after2Mask, 0, 0, 0, 0, time.Local)
}

// GetDateByTime 获取时间戳在默认时区的int类型日期
func GetDateByTime(t int64) int {
	now := time.Unix(t, 0).In(time.Local)
	return now.Year()*before4Mask + int(now.Month())*after2Mask + now.Day()
}

// GetMonthByTime 获取时间戳在默认时区的int类型日期
func GetMonthByTime(t int64) int {
	now := time.Unix(t, 0).In(time.Local)
	return now.Year()*after2Mask + int(now.Month())
}

// GetDatesByRange 通过时间戳获取中间的日期
func GetDatesByRange(from, to int64) []int {
	if from >= to {
		return nil
	}
	fromDate := GetDateByTime(from)
	toDate := GetDateByTime(to)

	result := make([]int, 0, 10)

	for ; fromDate <= toDate; fromDate = DataCal(fromDate, 1) {
		result = append(result, fromDate)
	}

	return result
}

// Time 特别定义的时间
type Time int64

// Now 获取当前时间戳
func Now() Time {
	return Time(time2.Now().UnixMilli())
}

// MarshalText 序列化方法
func (t Time) MarshalText() ([]byte, error) {
	if t == 0 {
		return nil, nil
	}
	return []byte(time.UnixMilli(int64(t)).In(time.Local).Format(time.DateTime)), nil
}

// IsZero 是否为0值
func (t Time) IsZero() bool {
	return t <= 0
}

// Unix 时间戳
func (t Time) Unix() int64 {
	return int64(t / 1000)
}

// UnixMilli 时间戳
func (t Time) UnixMilli() int64 {
	return int64(t)
}

// ToTime 转换为标准time.Time
func (t Time) ToTime() time.Time {
	return time.UnixMilli(int64(t))
}

// In 返回在指定时区的时间
func (t Time) In(loc *time.Location) time.Time {
	return t.ToTime().In(loc)
}

// Add 添加时间间隔
func (t Time) Add(d time.Duration) Time {
	return Time(t.ToTime().Add(d).UnixMilli())
}

// Format 格式化时间
func (t Time) Format(layout string) string {
	return t.ToTime().Format(layout)
}

// Sub 计算时间差
func (t Time) Sub(u Time) time.Duration {
	return t.ToTime().Sub(u.ToTime())
}

const (
	before4Mask = 10000 // 前4位
	after2Mask  = 100   // 后2位
)

// DataCal 日期增加或者减少天
func DataCal(date, add int) int {
	now := time.Date(date/before4Mask, time.Month(date%before4Mask/after2Mask), date%after2Mask, 0, 0, 0, 0, time.Local)

	now = now.AddDate(0, 0, add)

	return now.Year()*before4Mask + int(now.Month())*after2Mask + now.Day()
}

func TimeToDate(t time.Time) int {
	localTime := t.In(time.Local)
	return localTime.Year()*before4Mask + int(localTime.Month())*after2Mask + localTime.Day()
}

const dateStrLen = 8

// UnmarshalJSON 反序列化方法 支持int类型的日期 时间戳 2006-01-02 15:04:05 类型的字符串
func (t *Time) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `""`)
	s := string(data)

	if len(data) == dateStrLen {
		_, err := time.ParseInLocation("20060102", s, time.Local)
		if err != nil {
			return err
		}
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*t = Time(parseInt)
		return nil
	}

	if len(data) == 10 {
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*t = Time(parseInt)
		return nil
	}

	value, err := time.ParseInLocation(layout[:len(data)], s, time.Local)
	if err != nil {
		return err
	}
	*t = Time(value.Unix())
	return nil
}

// Scan 数据库类型变成go类型方法
func (t *Time) Scan(src interface{}) error {
	finished := false
	v := reflect.ValueOf(src)

	switch v.Kind() {
	case reflect.Int64, reflect.Int32:
		*t = Time(v.Int())
		finished = true

	case reflect.Slice:
		data, ok := src.([]uint8)
		if ok {
			value := int64(0)
			for i := range data {
				// 首先是16进制，然后是ASCII码(31对应1)
				value = value*10 + int64(data[i]/16*10+data[i]%16-30)
			}
			*t = Time(value)
			finished = true
		}
	default:
		return errors.Errorf("un support type %T", v.Kind())
	}
	if finished {
		return nil
	}
	return fmt.Errorf(`参数类型不支持,实际是[%s]`, reflect.TypeOf(src).Kind().String())
}

func (t Time) NowInDefault() time.Time {
	return time.Unix(t.Unix(), 0).In(time.Local)
}

// Monday 传入时间本周第一天当前时间
func Monday(now time.Time) time.Time {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = perWeek
	}
	monday := now.AddDate(0, 0, -weekday+1)
	return monday
}

const (
	endOfWeekFromMonday = 6
	perWeek             = 7
)

// EndOfWeek 获取本周日结束时间
func EndOfWeek(now time.Time) time.Time {
	return Monday(now).AddDate(0, 0, endOfWeekFromMonday)
}

/*
LastWeekMonday 获取上周的星期一当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上周的星期一的当前时间
*/
func LastWeekMonday(now time.Time) time.Time {
	return Monday(now).AddDate(0, 0, -perWeek)
}

/*
EndOfLastWeek 获取上周日的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上周日的当前时间
*/
func EndOfLastWeek(now time.Time) time.Time {
	return DecreaseOneDay(Monday(now))
}

/*
Month 获取本月初的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	本月初的当前时间
*/
func Month(now time.Time) time.Time {
	y, m, _ := now.Date()
	month1st := time.Date(y, m, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())

	return month1st
}

/*
EndOfMonth 获取本月底的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	本月底的当前时间
*/
func EndOfMonth(now time.Time) time.Time {
	return DecreaseOneDay(Month(now).AddDate(0, 1, 0))
}

/*
LastMonthFirst 获取上月初的0时
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上月初的0时
*/
func LastMonthFirst(now time.Time) time.Time {
	y, m, _ := now.Date()
	month1st := time.Date(y, m-1, 1, 0, 0, 0, 0, now.Location())
	return month1st
}

/*
EndOfLastMonth 获取上月底的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上月底的当前时间
*/
func EndOfLastMonth(now time.Time) time.Time {
	return DecreaseOneDay(Month(now))
}

/*
Day 获取今天的0时
参数:
返回值:
*	time.Time	time.Time	当前时间
*/
func Day() time.Time {
	now := NowInDefault()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

// DecreaseOneDay 减少1天
func DecreaseOneDay(now time.Time) time.Time {
	return now.AddDate(0, 0, -1)
}

// DateToTimestamp 默认时区格式化的时间转时间戳 format格式为: 2006-01-02 15:04:05
func DateToTimestamp(format string) (tt time.Time, err error) {
	tt, err = time.ParseInLocation(time.DateTime, format, time.Local)
	return
}

// GetTodayStart 获取今天开始时间
func GetTodayStart() time.Time {
	return GetDayStart(time2.Now())
}

// GetDayStart 获取某天开始时间
func GetDayStart(t time.Time) time.Time {
	y, m, d := t.In(time.Local).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// GetDayEnd 获取某天结束时间
func GetDayEnd(t time.Time) time.Time {
	t2 := GetDayStart(t)
	return t2.AddDate(0, 0, 1)
}

// GetMonthStart 获取某月开始时间
func GetMonthStart(t time.Time) time.Time {
	y, m, _ := t.In(time.Local).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
}

// GetMonthEnd 获取某月结束时间
func GetMonthEnd(t time.Time) time.Time {
	t2 := GetDayStart(t)
	return t2.AddDate(0, 1, 0)
}

// GetYearStart 获取某年开始时间
func GetYearStart(t time.Time) time.Time {
	y, _, _ := t.In(time.Local).Date()
	return time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
}

// GetYearEnd 获取某年结束时间
func GetYearEnd(t time.Time) time.Time {
	t2 := GetDayStart(t)
	return t2.AddDate(1, 0, 0)
}
