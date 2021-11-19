package http

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

func ExternalIP() (string, error) {
	iFaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iFace := range iFaces {
		if iFace.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iFace.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iFace.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func RequestServerModel(url, jsonStr, serverId, serviceKey string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(jsonStr))
	if err != nil {
		return "", err
	}

	dataBytes := []byte(jsonStr + "_" + serviceKey)
	digestBytes := md5.Sum(dataBytes)
	md5Str := fmt.Sprintf("%x", digestBytes)

	req.Header.Set("HSB-OPENAPI-CALLERSERVICEID", serverId)
	req.Header.Set("HSB-OPENAPI-SIGNATURE", md5Str)
	req.Header.Set("content-type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", errors.New("StatusCode != 200")
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
