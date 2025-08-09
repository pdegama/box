package server

import (
	"fmt"
	"net"
	"os"

	"github.com/pdegama/box/pkg/logx"
)

func (s *SMTPServer) getHostAddress() string {

	if s.Host == "" {
		if s.IsIPv6 {
			s.Host = "::1"
		} else {
			s.Host = "127.0.0.1"
		}

		return s.getHostAddress()
	}

	if s.IsIPv6 {
		if isIPv6(s.Host) {
			return fmt.Sprintf("[%s]:%d", s.Host, s.Port)
		} else {
			ipv6s, err := getIPv6FromDomain(s.Host)
			if err != nil {
				logx.LogError(fmt.Sprintf("invalid hostname for IPv6 %s.", s.Host), fmt.Errorf("invalid hostname for IPv6 %s", s.Host))
				os.Exit(1)
			}

			if len(ipv6s) < 1 {
				logx.LogError(fmt.Sprintf("invalid hostname for IPv6 %s.", s.Host), fmt.Errorf("invalid hostname for IPv6 %s", s.Host))
				os.Exit(1)
			}

			s.Host = ipv6s[0]

			return s.getHostAddress()
		}
	} else {
		if isIPv4(s.Host) {
			return fmt.Sprintf("%s:%d", s.Host, s.Port)
		} else {
			ipv4s, err := getIPv4FromDomain(s.Host)
			if err != nil {
				logx.LogError(fmt.Sprintf("invalid hostname for IPv4 %s.", s.Host), fmt.Errorf("invalid hostname for IPv4 %s", s.Host))
				os.Exit(1)
			}

			if len(ipv4s) < 1 {
				logx.LogError(fmt.Sprintf("invalid hostname for IPv4 %s.", s.Host), fmt.Errorf("invalid hostname for IPv4 %s", s.Host))
				os.Exit(1)
			}

			s.Host = ipv4s[0]

			return s.getHostAddress()
		}
	}
}

// isIPv4 checks if the given string is an IPv4 address.
func isIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// isIPv6 checks if the given string is an IPv6 address.
func isIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil && parsedIP.To16() != nil
}

// get IPv4 from Domains
func getIPv4FromDomain(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var ipv4s []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			ipv4s = append(ipv4s, ipv4.String())
		}
	}
	return ipv4s, nil
}

func getIPv6FromDomain(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var ipv6s []string
	for _, ip := range ips {
		if ipv6 := ip.To16(); ipv6 != nil && ip.To4() == nil {
			ipv6s = append(ipv6s, ipv6.String())
		}
	}
	return ipv6s, nil
}
