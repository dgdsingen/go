package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"
)

type IP struct {
	netip.Addr
	inPrefix bool
}

var (
	appName = "cidr"
	version = "undefined"
	// -l 로 나열할 수 있는 주소 개수 상한 (same as IPv4 /8)
	// 상한이 없으면 IPv6 /64 (2^64개) 같은 입력에서 사실상 종료되지 않음
	maxListAddrs      uint64 = 1 << 24
	listBufLength            = 64 << 10 // 64KB
	maxAddrTextLength        = 64       // IPv6 최대 표기 45자 + zone(ParseAddr 한정) 여유분
)

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
				return ips, prefixes, errors.New("invalid cidr: " + args[i])
			}
			prefixes = append(prefixes, prefix)
		} else {
			addr, err := netip.ParseAddr(args[i])
			if err != nil {
				return ips, prefixes, errors.New("invalid ip: " + args[i])
			}
			ips = append(ips, IP{Addr: addr})
		}
	}
	return ips, prefixes, nil
}

// ip count in prefix <= maxListAddrs
func PrefixAddrCount(p netip.Prefix) (uint64, bool) {
	hostBits := p.Addr().BitLen() - p.Bits()
	if hostBits >= 64 {
		return 0, false
	}
	count := uint64(1) << hostBits
	return count, count <= maxListAddrs
}

func ListAddrs(w io.Writer, prefixes []netip.Prefix) error {
	// validation: print all or nothing
	for p := range prefixes {
		if _, ok := PrefixAddrCount(prefixes[p]); !ok {
			return fmt.Errorf("too many addresses to list: %s (limit %d)", prefixes[p], maxListAddrs)
		}
	}

	// ip 마다 Print 하면 매번 write syscall + 문자열 할당이 발생하므로 bufio 처리
	bw := bufio.NewWriterSize(w, listBufLength)
	buf := make([]byte, 0, maxAddrTextLength)
	for p := range prefixes {
		// ParsePrefix는 10.0.0.5/24 처럼 masking 안된 주소도 그대로 받으므로
		// Masked() 로 network addr 부터 시작해 PrefixAddrCount와 일치시킴
		prefix := prefixes[p].Masked()
		addr := prefix.Addr()
		for addr.IsValid() && prefix.Contains(addr) {
			buf = append(addr.AppendTo(buf[:0]), '\n')
			if _, err := bw.Write(buf); err != nil {
				return err
			}
			addr = addr.Next()
		}
	}
	return bw.Flush()
}

// is ip in prefix?
func MatchIPs(ips []IP, prefixes []netip.Prefix) {
	for i := range ips {
		for p := range prefixes {
			if prefixes[p].Contains(ips[i].Addr) {
				ips[i].inPrefix = true
				break
			}
		}
	}
}

func WriteMatched(w io.Writer, ips []IP, invert bool) error {
	// 출력이 IP 개수만큼뿐이라 ListAddrs 만큼 큰 버퍼는 필요없음
	bw := bufio.NewWriterSize(w, min((len(ips)+1)*maxAddrTextLength, listBufLength))
	buf := make([]byte, 0, maxAddrTextLength)
	found := false
	// 짧은 루프이므로 개별 Write 에러는 보지 않고 Flush에서 한 번에 받는다
	// (bufio.Writer는 첫 에러를 물고 있다가 Flush에서 반환한다)
	for i := range ips {
		if ips[i].inPrefix == invert {
			continue
		}
		found = true
		buf = append(ips[i].AppendTo(buf[:0]), '\n')
		bw.Write(buf)
	}
	if !found {
		bw.WriteString("No result.\n")
	}
	return bw.Flush()
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
		// len(os.Args)는 flag까지 세므로 실제 인자 개수인 flag.NArg()로 판단
		if flag.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "Usage: %s -l <cidr...>\n", os.Args[0])
			os.Exit(1)
		}

		ips, prefixes, err := SplitPrefixAddr(flag.Args())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if len(prefixes) == 0 {
			fmt.Fprintf(os.Stderr, "No cidr.\n")
			os.Exit(1)
		}
		if len(ips) > 0 {
			fmt.Fprintf(os.Stderr, "warning: -l ignores %d ip argument(s)\n", len(ips))
		}
		if *v {
			fmt.Fprintf(os.Stderr, "warning: -l ignores -v\n")
		}

		if err := ListAddrs(os.Stdout, prefixes); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return
	}

	if flag.NArg() < 2 {
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

	MatchIPs(ips, prefixes)

	if err := WriteMatched(os.Stdout, ips, *v); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
