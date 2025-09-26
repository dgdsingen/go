package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	v := flag.Bool("v", false, "Invert match")
	flag.Parse()

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <IP...> <CIDR...>\n", os.Args[0])
		os.Exit(1)
	}

	ipSlice := []net.IP{}
	cidrSlice := []*net.IPNet{}
	for _, arg := range flag.Args() {
		if strings.Contains(arg, "/") {
			_, cidr, err := net.ParseCIDR(arg)
			if err != nil {
				fmt.Printf("Invalid CIDR: %s\n", arg)
				os.Exit(1)
			}
			cidrSlice = append(cidrSlice, cidr)
		} else {
			ip := net.ParseIP(arg)
			if ip == nil {
				fmt.Printf("Invalid IP: %s\n", arg)
				os.Exit(1)
			}
			ipSlice = append(ipSlice, ip)
		}
	}

	if len(ipSlice) == 0 {
		fmt.Println("No IP.")
	}

	if len(cidrSlice) == 0 {
		fmt.Println("No CIDR.")
	}

	for _, ip := range ipSlice {
		for _, cidr := range cidrSlice {
			if cidr.Contains(ip) != *v {
				fmt.Println(ip)
			}
		}
	}
}
