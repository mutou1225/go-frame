package toolkit

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
)

func GetHostName() string {
	name, err := os.Hostname()
	if err == nil {
		return name
	}

	ip, err := GetLocalIp()
	if err == nil {
		return ip.String()
	}

	return "127.0.0.1"
}

func Uniq(data sort.Interface) (size int) {
	p, l := 0, data.Len()
	if l <= 1 {
		return l
	}
	for i := 1; i < l; i++ {
		if !data.Less(p, i) {
			continue
		}
		p++
		if p < i {
			data.Swap(p, i)
		}
	}
	return p + 1
}

func GetLocalIp() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}

//uniq Id contains ms timestamp + local ip addr + interface name
//use md5 get 32bytes id

func GetUniqId(apiName string) (string, string, error) {
	ip, err := GetLocalIp()
	if err != nil {
		ip = net.IP{}
	}
	digestStr := fmt.Sprintf("%s%d%s", ip.String(), GetMSTimeStamp(), apiName)
	md5Str := Md5Digest(digestStr)
	return digestStr, md5Str, nil
}
