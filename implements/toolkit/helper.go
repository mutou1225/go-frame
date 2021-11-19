package toolkit

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"time"
	"unicode"
)

const (
	randChar = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Int64Arr []int64

func (a Int64Arr) Len() int           { return len(a) }
func (a Int64Arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Int64Arr) Less(i, j int) bool { return a[i] < a[j] }

func RandomLowerLetterString(l int) string {
	var result bytes.Buffer
	var temp rune = 'a'
	for i := 0; i < l; {
		randX := RandomInt(97, 122)
		if rune(randX) != temp {
			temp = rune(randX)
			strChar := string(temp)
			result.WriteString(strChar)
			i++
		}
	}
	return result.String()
}

func RandomString(l int) string {
	rand.NewSource(time.Now().UnixNano()) // 产生随机种子
	var s bytes.Buffer
	for i := 0; i < l; i++ {
		s.WriteByte(randChar[rand.Int63()%int64(len(randChar))])
	}
	return s.String()
}

// 16进制字符串
func RandomHexadecimal() string {
	var rnd = rand.New(rand.NewSource(time.Now().UnixNano())) // 产生随机种子
	return strconv.FormatUint(rnd.Uint64(), 16)
}

func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func Md5Sum(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	}

	return value
}

func ConvertToString(value interface{}) string {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v {
			return "1"
		} else {
			return "0"
		}
	default:
		return ""
	}

	return ""
}

func StrAtoi(value string) int {
	if value == "" {
		return 0
	}
	i, e := strconv.Atoi(value)
	if e != nil {
		return 0
	}
	return i
}

func StrAtoi64(value string) int64 {
	if value == "" {
		return 0
	}
	i, e := strconv.ParseInt(value, 10, 64)
	if e != nil {
		return 0
	}
	return i
}

func StrAtoiCheckFloat(value string) int {
	if value == "" {
		return 0
	}

	if !StrIsAllNum(value) {
		f, e := strconv.ParseFloat(value, 32)
		if e == nil {
			return int(f)
		}
	}

	i, e := strconv.Atoi(value)
	if e != nil {
		return 0
	}

	return i
}

func StrIsAllNum(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func SliceInt2String(iList []int) []string {
	ret := []string{}
	for _, i := range iList {
		ret = append(ret, strconv.Itoa(i))
	}
	return ret
}

func SliceString2int(sList []string) []int {
	ret := []int{}
	for _, s := range sList {
		ret = append(ret, StrAtoi(s))
	}
	return ret
}

func Interface2IntSlice(sList []interface{}) ([]int, error) {
	ret := []int{}
	for _, s := range sList {
		value := reflect.ValueOf(s)
	BEGIN:
		switch value.Kind() {
		case reflect.Ptr:
			if value.IsNil() {
				return nil, errors.New("Convert Error Ptr <nil>")
			}

			value = value.Elem()
			goto BEGIN
		case reflect.Int:
			ret = append(ret, int(value.Int()))
		case reflect.String:
			ret = append(ret, StrAtoi(value.String()))
		default:
			return nil, errors.New(fmt.Sprintf("Cannot convert type[%s] to int", value.Kind().String()))
		}
	}
	return ret, nil
}

func EqualSliceElem(obj, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	objValue := reflect.ValueOf(obj)
	targetType := reflect.TypeOf(target).Kind()
	objType := reflect.TypeOf(obj).Kind()
	if targetType != objType {
		return false
	}
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice:
		if targetValue.Len() != objValue.Len() {
			return false
		}
		if targetValue.Len() > 0 {
			typeOfObjElem := reflect.TypeOf(targetValue.Index(0).Interface())
			typeOfTarElem := reflect.TypeOf(targetValue.Index(0).Interface())
			if typeOfObjElem != typeOfTarElem {
				return false
			}
			switch typeOfObjElem.Kind() {
			case reflect.Int:
				sort.Ints(target.([]int))
				sort.Ints(obj.([]int))
			case reflect.String:
				sort.Strings(target.([]string))
				sort.Strings(obj.([]string))
			case reflect.Float64:
				sort.Float64s(target.([]float64))
				sort.Float64s(obj.([]float64))
			default:
			}
			for i := 0; i < targetValue.Len(); i++ {
				if objValue.Index(i).Interface() != targetValue.Index(i).Interface() {
					return false
				}
			}
		}
	default:
		return false
	}
	return true
}
