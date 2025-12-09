package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const appName = "cidr"

var version = "undefined"

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

type IP struct {
	net.IP
	inCidr bool
}

func main() {
	v := flag.Bool("v", false, "Invert match")
	versionFlag := flag.Bool("version", false, "Version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <IP...> <CIDR...>\n", os.Args[0])
		os.Exit(1)
	}

	args := flag.Args()
	ipSlice := make([]IP, 0, 1)
	cidrSlice := make([]*net.IPNet, 0, 1)
	for i := range args {
		if strings.Contains(args[i], "/") {
			_, cidr, err := net.ParseCIDR(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid CIDR: %s\n", args[i])
				os.Exit(1)
			}
			cidrSlice = append(cidrSlice, cidr)
		} else {
			ip := net.ParseIP(args[i])
			if ip == nil {
				fmt.Fprintf(os.Stderr, "Invalid IP: %s\n", args[i])
				os.Exit(1)
			}
			ipSlice = append(ipSlice, IP{IP: ip})
		}
	}

	if len(ipSlice) == 0 {
		fmt.Fprintf(os.Stderr, "No IP.\n")
	}
	if len(cidrSlice) == 0 {
		fmt.Fprintf(os.Stderr, "No CIDR.\n")
	}

	for i := range ipSlice {
		for c := range cidrSlice {
			if cidrSlice[c].Contains(ipSlice[i].IP) {
				ipSlice[i].inCidr = true
			}
		}
	}

	result := make([]string, 0)
	for i := range ipSlice {
		if ipSlice[i].inCidr != *v {
			result = append(result, ipSlice[i].String())
		}
	}
	if len(result) == 0 {
		result = append(result, "No result.")
	}
	fmt.Println(strings.Join(result, "\n"))
}
