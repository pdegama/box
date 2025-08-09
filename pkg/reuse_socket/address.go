package reusesocket

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/pdegama/box/pkg/logx"
)

func getSocketAddress(network string, address string) (syscall.Sockaddr, int) {

	switch network {
	case "tcp":
		addr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			logx.LogError("tcp ip resolve faild", err)
			os.Exit(1)
		}

		if addr.IP != nil && addr.IP.To4() == nil && addr.IP.To16() != nil { // if IPv6
			sockAddr := syscall.SockaddrInet6{
				Port: addr.Port,
			}
			copy(sockAddr.Addr[:], addr.IP.To16())

			if addr.Zone != "" {
				iface, err := net.InterfaceByName(addr.Zone)
				if err != nil {
					logx.LogError("ip resolve faild", err)
					os.Exit(1)
				}

				sockAddr.ZoneId = uint32(iface.Index)
			}

			// fmt.Printf("%v %v %v1\n", addr.IP.To4(), addr.IP.To16(), sockAddr)

			return &sockAddr, syscall.AF_INET6

		} else if addr.IP != nil && addr.IP.To4() != nil { // if IPv4
			sockAddr := syscall.SockaddrInet4{
				Port: addr.Port,
			}
			copy(sockAddr.Addr[:], addr.IP.To4())

			// fmt.Printf("%v %v %v1\n", addr.IP.To4(), addr.IP.To16(), sockAddr)

			return &sockAddr, syscall.AF_INET
		} else {
			logx.LogError("ip resolve faild", fmt.Errorf("%s ip resolve faild", network))
			os.Exit(1)
		}

	case "udp":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			logx.LogError("udp ip resolve faild", err)
			os.Exit(1)
		}

		if addr.IP != nil && addr.IP.To4() == nil && addr.IP.To16() != nil { // if IPv6
			sockAddr := syscall.SockaddrInet6{
				Port: addr.Port,
			}
			copy(sockAddr.Addr[:], addr.IP.To16())

			if addr.Zone != "" {
				iface, err := net.InterfaceByName(addr.Zone)
				if err != nil {
					logx.LogError("ip resolve faild", err)
					os.Exit(1)
				}
				sockAddr.ZoneId = uint32(iface.Index)
			}
			// fmt.Printf("%v %v %v1\n", addr.IP.To4(), addr.IP.To16(), sockAddr)

			return &sockAddr, syscall.AF_INET6

		} else if addr.IP != nil && addr.IP.To4() != nil { // if IPv4
			sockAddr := syscall.SockaddrInet4{
				Port: addr.Port,
			}
			copy(sockAddr.Addr[:], addr.IP.To4())
			// fmt.Printf("%v %v %v1\n", addr.IP.To4(), addr.IP.To16(), sockAddr)

			return &sockAddr, syscall.AF_INET
		} else {
			logx.LogError("ip resolve faild", fmt.Errorf("%s ip resolve faild", network))
			os.Exit(1)
		}

	default:
		fmt.Printf("don't allow %s, only allow tcp and udp network, it automatic parse to tcp, tcp4, tcp6, udp, udp4 and udp6 network...\n", network)
		os.Exit(1)
	}

	return nil, 0

}
