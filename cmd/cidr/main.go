package main

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"
)

type IP struct {
	netip.Addr
	inPrefix bool
}

const appName = "cidr"

var version = "undefined"

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

func SplitPrefixAddr(args []string) (ips []IP, prefixes []netip.Prefix, err error) {
	ips = make([]IP, 0, 1)
	prefixes = make([]netip.Prefix, 0, 1)
	for i := range args {
		if strings.Contains(args[i], "/") {
			prefix, err := netip.ParsePrefix(args[i])
			if err != nil {
				return ips, prefixes, errors.New("Invalid cidr: " + args[i])
			}
			prefixes = append(prefixes, prefix)
		} else {
			addr, err := netip.ParseAddr(args[i])
			if err != nil {
				return ips, prefixes, errors.New("Invalid ip: " + args[i])
			}
			ips = append(ips, IP{Addr: addr})
		}
	}
	return ips, prefixes, nil
}

func main() {
	v := flag.Bool("v", false, "Invert match")
	l := flag.Bool("l", false, "List ips from cidr")
	versionFlag := flag.Bool("version", false, "Version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	if *l {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s -l <cidr...>\n", os.Args[0])
			os.Exit(1)
		}

		_, prefixes, err := SplitPrefixAddr(flag.Args())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if len(prefixes) == 0 {
			fmt.Fprintf(os.Stderr, "No cidr.\n")
			os.Exit(1)
		}

		for p := range prefixes {
			addr := prefixes[p].Addr()
			for addr.IsValid() && prefixes[p].Contains(addr) {
				fmt.Println(addr)
				addr = addr.Next()
			}
		}
		return
	}

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <ip...> <cidr...>\n", os.Args[0])
		os.Exit(1)
	}

	ips, prefixes, err := SplitPrefixAddr(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if len(ips) == 0 {
		fmt.Fprintf(os.Stderr, "No ip.\n")
		os.Exit(1)
	}
	if len(prefixes) == 0 {
		fmt.Fprintf(os.Stderr, "No cidr.\n")
		os.Exit(1)
	}

	for i := range ips {
		for p := range prefixes {
			if prefixes[p].Contains(ips[i].Addr) {
				ips[i].inPrefix = true
			}
		}
	}

	result := make([]string, 0, 1)
	for i := range ips {
		if ips[i].inPrefix != *v {
			result = append(result, ips[i].String())
		}
	}
	if len(result) == 0 {
		result = append(result, "No result.")
	}
	fmt.Println(strings.Join(result, "\n"))
}
