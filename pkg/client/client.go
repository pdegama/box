package client

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rellitelink/box/pkg/logx"
)

type SMTPClinet struct {
	hostname string

	RcptHost string
	RcptPort int

	From string
	Rcpt string

	data []byte

	Size int
	UTF8 bool

	StartTls     bool
	CheckTlsHost bool
	TlsKey       string
	TlsCert      string

	ChunkSize int

	Timeout  time.Duration
	Response ClientResponse

	Logger *logx.Log
}

func NewClinet() SMTPClinet {
	return SMTPClinet{
		StartTls:     true,
		CheckTlsHost: false,
		hostname:     "localhost",

		RcptHost: "",
		RcptPort: 25,

		From: "",
		Rcpt: "",

		Timeout: time.Minute * 2,

		ChunkSize: 1024 * 1000,
	}
}

// set rcpt port host use for make connection
func (client *SMTPClinet) SetRcptHost(host string) {
	client.RcptHost = host
}

// set rcpt port no use for make connection
func (client *SMTPClinet) SetRcptPort(port int) {
	client.RcptPort = port
}

// set client host name use in hello
func (client *SMTPClinet) SetHostname(host string) {
	client.hostname = host
}

func (client *SMTPClinet) SetFrom(from string) {
	client.From = from
}

func (client *SMTPClinet) SetRcpt(rcpt string) {
	client.Rcpt = rcpt
}

func (client *SMTPClinet) SetData(data []byte) {
	client.data = data
}

func (client *SMTPClinet) SetTimeout(t time.Duration) {
	client.Timeout = t
}

type ClientServerError struct {
	Domain      string
	Error       string
	Code        int // if smtp code avalible
	ServerError bool
}

type ClientResponse struct {
	Time           time.Time
	Errors         []ClientServerError
	Success        bool
	TempError      bool // temp error (4yz)
	AnyClientError bool // any client unknown error
	Status         string
}

func (client *SMTPClinet) SendMail() {
	client.Size = len(client.data) + 5 // add 5 for <crlf>.<crlf>
	client.Response.Time = time.Now().UTC()

	mxRecords := []*net.MX{}
	if client.RcptHost != "" {
		mxRecords = append(mxRecords, &net.MX{
			Host: client.RcptHost,
			Pref: 0,
		})
	} else if client.Rcpt != "" {
		domainName, err := getDomainFromEmail(client.Rcpt)
		if err != nil {
			client.Response.Errors = append(client.Response.Errors, ClientServerError{
				Domain: domainName,
				Error:  fmt.Sprintf("invalid doamin (%s) by client", client.Rcpt),
			})
			client.Logger.Error("invalid doamin (%s) by client", client.Rcpt)
			return
		}

		mxs, err := net.LookupMX(domainName)
		if err != nil {
			client.Response.Errors = append(client.Response.Errors, ClientServerError{
				Domain: domainName,
				Error:  fmt.Sprintf("no any MX records found of %s domain", domainName),
			})
			client.Logger.Error("no any MX records found of %s domain", domainName)
			return
		}
		mxRecords = mxs
	} else {
		client.Response.Errors = append(client.Response.Errors, ClientServerError{
			Domain: "unknown",
			Error:  "no any host and rcpt found",
		})
		client.Logger.Error("no any host and rcpt found")
		return
	}

	if IsSMTPUTF8(client.From) {
		client.UTF8 = true
	}
	if IsSMTPUTF8(client.Rcpt) {
		client.UTF8 = true
	}

	for _, m := range mxRecords {
		if client.Response.Success {
			break
		}

		ips, err := getIPFromString(m.Host)
		if err != nil {
			client.Response.Errors = append(client.Response.Errors, ClientServerError{
				Domain: m.Host,
				Error:  err.Error(),
			})
			client.Logger.Warn(err.Error())
			continue
		}

		for _, rcptIP := range ips {
			address := ""
			if isIPv6(rcptIP) {
				address = fmt.Sprintf("[%s]:%d", rcptIP, client.RcptPort)
			} else {
				address = fmt.Sprintf("%s:%d", rcptIP, client.RcptPort)
			}

			client.RcptHost = m.Host
			conn, err := client.createNewConn(address)
			if err != nil {
				// fail log
				client.Response.Errors = append(client.Response.Errors, ClientServerError{
					Domain: m.Host,
					Error:  err.Error(),
				})
				client.Response.TempError = true
				continue
			}

			client.Logger.Info("client connected to %s - %s server", m.Host, conn.conn.RemoteAddr())

			err = conn.handleConn()
			if err != nil {
				switch e := err.(type) {
				case SMTPServerError:
					if e.GetErrorType() == SMTPErrorTemp {
						client.Response.TempError = true
					}
					client.Response.Errors = append(client.Response.Errors, ClientServerError{
						Domain: m.Host,
						Error:  err.Error(),
						Code:   e.Code,
					})
					client.Logger.Warn("Email %s", err)
					continue
				case net.Error:
					if e.Timeout() {
						client.Response.Errors = append(client.Response.Errors, ClientServerError{
							Domain: m.Host,
							Error:  fmt.Sprintf("client timeout to %s server", address),
						})
						client.Logger.Warn("client timeout to %s server", address)
						client.Response.TempError = true
						continue
					} else {
						client.Response.Errors = append(client.Response.Errors, ClientServerError{
							Domain:      m.Host,
							Error:       err.Error(),
							ServerError: true,
						})
						client.Logger.Error("Server Error: %s", err)
						client.Response.AnyClientError = true
						continue
					}
				default:
					client.Response.Errors = append(client.Response.Errors, ClientServerError{
						Domain:      m.Host,
						Error:       err.Error(),
						ServerError: true,
					})
					client.Logger.Error("Server Error: %s", err)
					client.Response.AnyClientError = true
					continue
				}
			}
			client.Response.Success = true
			client.Logger.Success("email deliver successfull")
			break
		}
	}
}

func (client *SMTPClinet) createNewConn(address string) (ClientConn, error) {
	conn, err := net.DialTimeout("tcp", address, client.Timeout)
	if err != nil {
		switch e := err.(type) {
		case *net.OpError:
			switch e.Op {
			case "dial":
				switch e := e.Err.(type) {
				case *os.SyscallError:
					if e.Syscall == "connect" {
						client.Logger.Warn("connection refused when connect client to %s server", address)
						return ClientConn{}, fmt.Errorf("connection refused when client connect with %s", address)
					}
				case net.Error:
					if e.Timeout() {
						client.Logger.Warn("timeout on connect client to %s server", address)
						return ClientConn{}, fmt.Errorf("connection timeout with %s by server", address)
					}
				}
			}
		case net.Error:
			if e.Timeout() {
				client.Logger.Warn("timeout on connect client to %s server", address)
				return ClientConn{}, fmt.Errorf("connection timeout with %s by server", address)
			}
		}
		return ClientConn{}, fmt.Errorf("internal server error for connect with %s	%v", address, err)
	}

	return ClientConn{
		conn: conn,
		rw:   newTextReaderWriter(conn, client.Timeout),

		smtpClient: client,
	}, nil
}

func (client *SMTPClinet) GetResponse() ClientResponse {
	if client.Response.Success {
		client.Response.Status = "SUCCESS"
		return client.Response
	}

	if client.Response.TempError {
		client.Response.Status = "TRYAGAIN"
	} else {
		client.Response.Status = "FAIL"
	}

	return client.Response
}
