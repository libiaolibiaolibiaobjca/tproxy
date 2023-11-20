package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"strings"
)

var s Settings

func main() {

	var (
		localHost   = flag.String("l", getDefaultIp(), "Local address to listen on")
		delay       = flag.Duration("d", 0, "the delay to relay packets")
		protocol    = flag.String("t", "", "The type of protocol, currently support http2, grpc, redis and mongodb")
		mapping     = flag.String("m", "", "13306>192.168.3.66:3306[,……]")
		mappingFile = flag.String("mf", "", "支持从文件中读映射（每一行 1 条映射）. tproxy.mapping")
		stat        = flag.Bool("s", false, "Enable statistics")
		quiet       = flag.Bool("q", false, "Quiet mode, only prints connection open/close and stats, default false")
		upLimit     = flag.Int64("up", 0, "Upward speed limit(bytes/second)")
		downLimit   = flag.Int64("down", 0, "Downward speed limit(bytes/second)")
	)

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	flag.Parse()

	saveSettings(*localHost, *mapping, *mappingFile, *delay, *protocol, *stat, *quiet, *upLimit, *downLimit)

	if s.Mapping == nil || len(s.Mapping) == 0 {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Remote target required"))
		flag.PrintDefaults()
		os.Exit(1)
	} else {
		fmt.Printf("共 %d 条 mapping\n", len(s.Mapping))
		for k, v := range s.Mapping {
			fmt.Printf("[%d] => [%s]\n", k, v)
		}
	}

	if err := startListener(); err != nil {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Failed to start listener: %v", err))
		os.Exit(1)
	}
}

func getDefaultIp() string {
	defaultIp := "localhost"
	m, _ := GetClientIp("")
	for k, v := range m {
		if k == "en0" || k == "eth0" {
			defaultIp = v
		}
	}
	return defaultIp
}

func GetClientIp(filter string) (map[string]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	nameIPS := map[string]string{}
	for _, i := range interfaces {
		name := i.Name
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		// handle err
		for _, addr := range addrs {
			var (
				ip net.IP
			)
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if filter == "" || strings.Contains(name, filter) {
				nameIPS[name] = ip.String()
			}
		}
	}
	if len(nameIPS) == 0 {
		return nil, errors.New("can not find the client ip address ")
	}
	fmt.Printf("GetClientIP ips:%v\n", nameIPS)
	return nameIPS, nil
}
