package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <IP> <CIDR>\n", os.Args[0])
		os.Exit(1)
	}

	ipStr := os.Args[1]
	cidrStr := os.Args[2]

	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Printf("Invalid IP: %s\n", ipStr)
		os.Exit(1)
	}

	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		fmt.Printf("Invalid CIDR: %s\n", cidrStr)
		os.Exit(1)
	}

	if cidr.Contains(ip) {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
