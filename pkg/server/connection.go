package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	limitlinereader "github.com/rellitelink/box/pkg/limit_line_reader"
	"github.com/rellitelink/box/pkg/logx"
	"github.com/rellitelink/box/pkg/uid"
)

type MailForwardCode int

const (
	MailForwardFaild   MailForwardCode = iota // is mail not recive perfect or interrupt connection
	MailForwardSuccess                        // mail recive full
	MailForwardIdle                           // if connection close without send data
)

type Connection struct {
	conn net.Conn

	remoteAddress RoLAddress
	localAddress  RoLAddress
	rw            *TextReaderWriter // text protocal for mail

	uid       string
	mailCount int // it is count of how many mail tranfare in this connection

	domain        string
	mailFrom      string
	recipients    []string
	spfFail       bool
	spfStatus     string
	forwardStatus MailForwardCode
	data          []byte
	dataBuffer    *bytes.Buffer // for bdat

	useTls bool

	useEsmtp bool // client use enhanced smtp
	size     int
	put8     bool
	body     BodyType

	logger       *logx.Log
	serverLogger *logx.Log // print server level log

	totalCmd int
	passCmd  int
}

func HandleNewConnection(conn net.Conn, serverLogger *logx.Log) {
	connection := Connection{
		conn:         conn,
		serverLogger: serverLogger,
	}
	clientCount += 1

	ok := connection.init()
	if !ok {
		conn.Close()
		return
	}

	connection.handle()
}

func (conn *Connection) init() bool {
	conn.rw = newTextReaderWriter(conn.conn)

	uid, err := uid.GetNewId()
	if err != nil {
		conn.serverLogger.Error("generate email uid error: %v", err)
		return false
	}

	conn.mailCount = 1
	conn.uid = uid
	conn.logger = conn.serverLogger.GetNewWithPrefix(conn.uid)

	if ok := conn.remoteAddress.SetAddress(conn.conn.RemoteAddr().Network(), conn.conn.RemoteAddr().String()); !ok {
		conn.logger.Warn("no PTR record or faild to find PTR records of %s", conn.remoteAddress.String())
	}

	if ok := conn.localAddress.SetAddress(conn.conn.LocalAddr().Network(), conn.conn.LocalAddr().String()); !ok {
		conn.logger.Warn("no PTR record or faild to find PTR records of local address %s", conn.remoteAddress.String())
	}

	conn.domain = ""
	conn.mailFrom = ""
	conn.recipients = []string{}
	conn.data = nil
	conn.forwardStatus = MailForwardIdle

	conn.logger.Info("client %s[%s]:%d connected", conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)

	return true
}

func (conn *Connection) handle() {

	if config.MaxClients > 0 && clientCount > config.MaxClients {
		conn.rw.busy(conn.localAddress.GetPTR())
		conn.closeForMaxClientsExceeded()
		return
	} else {
		conn.rw.greet(conn.localAddress.GetPTR())
	}

	for {
		line, err := conn.rw.readLine()
		if err == nil {
			cmd, args, err := parseCommand(line)
			if err == CmdParseOk {
				status := conn.handleCommand(cmd, args)
				if status == HandleCommandClose { // if connection close with QUIT...
					break
				} else if status == HandleCommandEOF {
					conn.closeWithFail()
					break
				}
			} else {
				conn.rw.cmdNotRecognized(cmd)
			}
			continue
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			conn.rw.timeout(conn.localAddress.GetPTR())
			conn.closeWithFailAnd("timeout exceeded")
			break
		} else if err == limitlinereader.ErrTooLongLine {
			conn.rw.longLine(conn.localAddress.GetPTR())
			conn.closeWithFailAnd("too long line")
			break
		} else if err == io.ErrUnexpectedEOF { // eof or client connection close
			conn.closeWithFail()
			break
		} else { // if unexpected
			conn.closeWithFail()
			break
		}
	}
}

func (conn *Connection) closeWithFail() {
	conn.forwardStatus = MailForwardFaild
	conn.close()
	conn.logger.Error("disconnected unfortunately client %s[%s]:%d", conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
}

func (conn *Connection) closeWithSuccess() {
	conn.close()
	conn.logger.Info("disconnected client %s[%s]:%d", conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
}

func (conn *Connection) closeForMaxClientsExceeded() {
	conn.close()
	conn.logger.Warn("disconnected client by server for max clients exceeded %s[%s]:%d", conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
}

func (conn *Connection) closeWithFailAnd(resone string) {
	conn.forwardStatus = MailForwardFaild
	conn.close()
	conn.logger.Warn("disconnected client by server for %s %s[%s]:%d", resone, conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
}

func (conn *Connection) reset() {
	conn.forward()

	conn.mailFrom = ""
	conn.recipients = []string{}
	conn.data = nil
	conn.forwardStatus = MailForwardIdle
	conn.dataBuffer = nil
	conn.spfFail = false
	conn.spfStatus = ""
	conn.totalCmd = 0
	conn.passCmd = 0
}

func (conn *Connection) close() {
	conn.forward()

	clientCount -= 1
	err := conn.conn.Close()
	if err != nil {
		if strings.Contains(err.Error(), "but connection was closed anyway") || strings.Contains(err.Error(), "broken pipe") {
			conn.logger.Warn("⚠️⚠️⚠️ conn close error client %s[%s]:%d ⚠️⚠️⚠️ Error: %s", conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port, err.Error())
		} else {
			conn.logger.Error("⚠️⚠️⚠️ conn close error client %s[%s]:%d ⚠️⚠️⚠️ Error: %s", conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port, err.Error())
		}
	}
}

func (conn *Connection) forward() {
	if conn.mailFrom == "" {
		return
	}

	go func(uid string, count int, conn Connection) {

		email := Email{
			Success: false,
			Uid:     fmt.Sprintf("%s_%d", uid, count),
			Time:    time.Now().UTC(),

			Cmds: fmt.Sprintf("%d/%d", conn.passCmd, conn.totalCmd),
			Tls:  conn.useTls,
			PtrIP: ServerClientInfo{
				ServerPtr: conn.localAddress.GetPTR(),
				ServerIP:  conn.localAddress.String(),
				ClinetPtr: conn.remoteAddress.GetPTR(),
				ClientIP:  conn.remoteAddress.String(),
			},

			Domain:    conn.domain,
			PtrMatch:  false,
			SpfFail:   conn.spfFail,
			SpfStatus: conn.spfStatus,

			From:       conn.mailFrom,
			Recipients: conn.recipients,
			UseBdat:    false,
			dataByte:   conn.data,
			// Data:       string(conn.data),
		}

		if conn.remoteAddress.HasPtr(conn.domain) {
			email.PtrMatch = true
		}

		if conn.dataBuffer != nil {
			email.UseBdat = true
		}

		if conn.forwardStatus == MailForwardSuccess {
			email.Success = true
		}

		mailFwd.ForwardMail(email)
	}(conn.uid, conn.mailCount, *conn)
}
