package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	logging "github.com/op/go-logging"
	"github.com/satori/go.uuid"
)

var log = logging.MustGetLogger("common")

var osDict = map[string]int{
	"centos":  OS_CENTOS,
	"red hat": OS_REDHAT,
	"redhat":  OS_REDHAT,
	"ubuntu":  OS_UBUNTU,
	"suse":    OS_SUSE,
	"windows": OS_WINDOWS,
}

var archDict = map[string]int{
	"x86":   ARCH_X86,
	"amd64": ARCH_X86,
	"i686":  ARCH_X86,
	"i386":  ARCH_X86,
	"aarch": ARCH_ARM,
	"arm":   ARCH_ARM,
}

func GetOsType(os string) int {
	for key, value := range osDict {
		if strings.Contains(strings.ToLower(os), key) {
			return value
		}
	}
	return 0
}

func GetArchType(arch string) int {
	for key, value := range archDict {
		if strings.Contains(strings.ToLower(arch), key) {
			return value
		}
	}
	return 0
}

func GenerateUUID(str string) string {
	return uuid.NewV5(uuid.NamespaceOID, str).String()
}

func GenerateShortUUID() string {
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(letterRunes))]
	}
	return string(b)
}

// 功能：获取用于API调用的IP地址
func GetCURLIP(ip string) string {
	// IPV6地址在API调用时需要增加[]
	if strings.Contains(ip, ":") && !strings.HasPrefix(ip, "[") {
		return "[" + ip + "]"
	}
	return ip
}

// 功能：调用其他模块API并获取返回结果
func CURLPerform(method string, url string, body map[string]interface{}) (*simplejson.Json, error) {
	errResponse, _ := simplejson.NewJson([]byte("{}"))

	// TODO: 通过配置文件获取API超时时间
	client := &http.Client{Timeout: time.Second * 30}

	bodyStr, _ := json.Marshal(&body)
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyStr))
	if err != nil {
		log.Error(err)
		return errResponse, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain")
	req.Header.Set("X-User-Id", "1")
	req.Header.Set("X-User-Type", "1")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("curl (%s) failed, (%v)", url, err)
		return errResponse, err
	} else if resp.StatusCode != http.StatusOK {
		log.Warning("curl (%s) failed, (%v)", url, resp)
		defer resp.Body.Close()
		return errResponse, errors.New(fmt.Sprintf("curl (%s) failed", url))
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("read (%s) body failed, (%v)", url, err)
		return errResponse, err
	}

	response, err := simplejson.NewJson(respBytes)
	if err != nil {
		log.Errorf("parse (%s) body failed, (%v)", url, err)
		return errResponse, err
	}

	optStatus := response.Get("OPT_STATUS").MustString()
	if optStatus != "" && optStatus != SUCCESS {
		description := response.Get("DESCRIPTION").MustString()
		log.Errorf("curl (%s) failed, (%s)", url, description)
		return errResponse, errors.New(description)
	}

	return response, nil
}

// 通过字符串获取UUID
func GetUUID(str string, namespace uuid.UUID) string {
	if str != "" {
		if namespace != uuid.Nil {
			return uuid.NewV5(namespace, str).String()
		}
		return uuid.NewV5(uuid.NamespaceOID, str).String()
	}
	return uuid.NewV4().String()
}

// 功能：判断当前控制器是否为masterController
func IsMasterController() (bool, string, error) {
	// 获取本机hostname
	hostName, err := os.Hostname()
	if err != nil {
		log.Error(err)
		return false, "", err
	}
	// 通过sideCar API获取MasterControllerName
	url := fmt.Sprintf("http://%s:%d", LOCALHOST, MASTER_CONTROLLER_CHECK_PORT)
	response, err := CURLPerform("GET", url, nil)
	if err != nil {
		return false, "", err
	}
	masterControllerName := response.Get("name").MustString()

	// 比较是否相同返回结果
	if hostName != masterControllerName {
		return false, masterControllerName, nil
	}
	return true, masterControllerName, nil
}

func IsValueInSliceString(value string, list []string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}
