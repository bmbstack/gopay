package common

import (
	"reflect"
	"strings"
	"runtime"
	"strconv"
	"time"
	"crypto/tls"
	"golang.org/x/crypto/pkcs12"
	"fmt"
	"encoding/pem"
	"bytes"
	"encoding/json"
	"net/url"
)

const (
	DateShortLayout            = "2006-01-02"
	DateFullLayout             = "2006-01-02 15:04:05"
	DateFullLayoutWithoutSplit = "20060102150405"
	TimeLocationName           = "Asia/Chongqing"
)

// GetTypeName returns a string representing the name of the object typ.
// if the name is defined then it is used, otherwise, the name is derived from the
// Stringer interface.
//
// the stringer returns something like *somepkg.MyStruct, so skip
// the *somepkg and return MyStruct
func GetTypeName(typ reflect.Type) string {
	if typ.Name() != "" {
		return typ.Name()
	}
	split := strings.Split(typ.String(), ".")
	return split[len(split)-1]
}

func IsEmpty(i interface{}) bool {
	return isEmpty(reflect.ValueOf(i))
}

func IsNotEmpty(i interface{}) bool {
	return !isEmpty(reflect.ValueOf(i))
}

func isEmpty(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool() == false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0

	case reflect.Float32, reflect.Float64:
		return v.Float() == 0

	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0

	case reflect.Ptr, reflect.Interface:
		return isEmpty(v.Elem())

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isEmpty(v.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Slice, reflect.String, reflect.Map:
		return v.Len() == 0

	case reflect.Struct:
		for i, n := 0, v.NumField(); i < n; i++ {
			if !isEmpty(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v.IsNil()
	}
}

func Substring(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func CurrentMethodName() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(1, pc)
	f := runtime.FuncForPC(pc[1])
	arr := strings.Split(f.Name(), ".")
	return arr[len(arr)-1]
}

func MapToXml(params map[string]string) string {
	var buf bytes.Buffer
	buf.WriteString(`<xml>`)
	for k, v := range params {
		buf.WriteString(`<`)
		buf.WriteString(k)
		buf.WriteString(`><![CDATA[`)
		buf.WriteString(v)
		buf.WriteString(`]]></`)
		buf.WriteString(k)
		buf.WriteString(`>`)
	}
	buf.WriteString(`</xml>`)

	return buf.String()
}

// 用时间戳生成随机字符串
func NonceStr() string {
	return strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
}

// 将Pkcs12转成Pem
func Pkcs12ToPem(p12 []byte, password string) tls.Certificate {
	blocks, err := pkcs12.ToPEM(p12, password)

	// 从恐慌恢复
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
		}
	}()

	if err != nil {
		panic(err)
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		panic(err)
	}
	return cert
}

func Marshal(value interface{}) string {
	var bytes, err = json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func MapToUrlValues(params map[string]string) url.Values {
	values := make(url.Values)
	for key, value := range params {
		values.Add(key, fmt.Sprintf(value))
	}
	return values
}
