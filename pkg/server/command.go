package server

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pdegama/box/pkg/spf"
)

type HandleCommandStatus int

const (
	HandleCommandOk HandleCommandStatus = iota
	HandleCommandEOF
	HandleCommandClose
)

type BodyType string

const (
	Body7Bit       BodyType = "7BIT"
	Body8BitMIME   BodyType = "8BITMIME"
	BodyBinaryMIME BodyType = "BINARYMIME"
)

func (conn *Connection) handleCommand(cmd string, args string) HandleCommandStatus {
	conn.totalCmd += 1
	switch cmd {
	case "EHLO":
		conn.handleEHello(args)
	case "HELO":
		conn.handleHello(args)
	case "MAIL":
		conn.handleMail(args)
	case "RCPT":
		conn.handleRcpt(args)
	case "DATA":
		return conn.handleData()
	case "BDAT":
		conn.handleBdat(args)
	case "RSET":
		conn.handleReset()
	case "NOOP":
		conn.handleNoop()
	case "QUIT":
		conn.handleQuit()
		return HandleCommandClose
	case "STARTTLS":
		conn.handleStartTls()
	case "SEND", "SOML", "SAML", "EXPN", "HELP", "TURN", "LHLO", "AUTH", "VRFY":
		conn.rw.cmdNotImplemented(cmd)
	default:
		conn.rw.cmdNotRecognized(cmd)
	}

	return HandleCommandOk
}

func (conn *Connection) handleEHello(args string) {
	if !config.ESMTP.Enable {
		conn.rw.reply(502, "Error: ESMTP Disable")
		return
	}

	domain, err := parseHelloArguments(args)
	if err == HelloArgDomainInvalid {
		conn.rw.reply(501, "domain required for HELO")
		return
	}

	conn.domain = domain
	conn.useEsmtp = true
	replyMsg := []string{"Hello " + domain}
	replyMsg = append(replyMsg, greetReplyMessage...)
	conn.rw.replyLines(250, replyMsg)
	conn.passCmd += 1
}

func (conn *Connection) handleStartTls() {
	if !config.ESMTP.Enable || !config.ESMTP.Tls {
		conn.rw.reply(502, "Error: starttls Disable")
		return
	}

	conn.rw.reply(220, "Ready to start TLS")

	tlsConn := tls.Server(conn.conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		conn.rw.reply(550, "TLS Handshake error")
		return
	}

	conn.conn = tlsConn
	conn.useTls = true
	conn.rw = newTextReaderWriter(conn.conn)
	conn.passCmd += 1
}

func (conn *Connection) handleHello(args string) {
	domain, err := parseHelloArguments(args)
	if err == HelloArgDomainInvalid {
		conn.rw.reply(501, "domain required for HELO")
		return
	}

	conn.domain = domain
	conn.useEsmtp = false
	conn.rw.reply(250, "%s ready for you", config.Name)
	conn.passCmd += 1
}

func (conn *Connection) handleMail(args string) {
	if conn.domain == "" {
		conn.rw.reply(503, "Error: send HELO/EHLO first")
		return
	}

	p := newParser(args)
	if ok := p.cutPrefix("FROM:"); !ok {
		conn.rw.syntaxError("Syntax: MAIL FROM:<address>")
		return
	}

	p.trim()
	from, err := p.parseReversePath()
	if err != nil {
		conn.rw.syntaxError("invalid address")
		return
	}
	conn.mailFrom = from

	if !conn.useEsmtp {
		conn.rw.reply(250, "Ok")
		return
	}

	mailArgs, err := parseArgs(p.s)
	if err != nil {
		conn.rw.reply(501, "Unable to parse MAIL ESMTP parameters")
		return
	}

	for key, value := range mailArgs {
		switch key {
		case "SIZE":
			size, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				conn.rw.reply(501, "Unable to parse SIZE as an integer")
				return
			}
			if config.ESMTP.MessageSize > 0 && int(size) > config.ESMTP.MessageSize {
				conn.rw.reply(552, "Max message size exceeded")
				return
			}

			conn.size = int(size)
		case "SMTPUTF8":
			if !config.ESMTP.Utf8 {
				conn.rw.reply(504, "SMTPUTF8 is not implemented")
				return
			}
			conn.put8 = true
		case "BODY":
			value = strings.ToUpper(value)
			switch BodyType(value) {
			case BodyBinaryMIME:
				if !config.ESMTP.BinaryMime {
					conn.rw.reply(504, "BINARYMIME is not implemented")
					return
				}
			case Body7Bit, Body8BitMIME:
				// This space is intentionally left blank
			default:
				conn.rw.reply(501, "Unknown BODY value")
				return
			}
			conn.body = BodyType(value)
		case "RET":
			// UPDATE: // DSN
			// if !c.server.EnableDSN {
			// 	c.writeResponse(504, EnhancedCode{5, 5, 4}, "RET is not implemented")
			// 	return
			// }
			// value = strings.ToUpper(value)
			// switch DSNReturn(value) {
			// case DSNReturnFull, DSNReturnHeaders:
			// 	// This space is intentionally left blank
			// default:
			// 	c.writeResponse(501, EnhancedCode{5, 5, 4}, "Unknown RET value")
			// 	return
			// }
			// opts.Return = DSNReturn(value)
		case "ENVID":
			// UPDATE: // ENVID
			// if !c.server.EnableDSN {
			// 	c.writeResponse(504, EnhancedCode{5, 5, 4}, "ENVID is not implemented")
			// 	return
			// }
			// value, err := decodeXtext(value)
			// if err != nil || value == "" || !isPrintableASCII(value) {
			// 	c.writeResponse(501, EnhancedCode{5, 5, 4}, "Malformed ENVID parameter value")
			// 	return
			// }
		case "AUTH":
			// value, err := decodeXtext(value)
			// if err != nil || value == "" {
			// 	c.writeResponse(500, EnhancedCode{5, 5, 4}, "Malformed AUTH parameter value")
			// 	return
			// }
			// if value == "<>" {
			// 	value = ""
			// } else {
			// 	p := parser{s: value}
			// 	value, err = p.parseMailbox()
			// 	if err != nil || p.s != "" {
			// 		c.writeResponse(500, EnhancedCode{5, 5, 4}, "Malformed AUTH parameter mailbox")
			// 		return
			// 	}
			// }
		default:
			conn.rw.reply(500, "Unknown MAIL FROM argument")
			return
		}
	}

	conn.rw.reply(250, "Ok")
	conn.passCmd += 1
}

func (conn *Connection) handleRcpt(args string) {
	if conn.mailFrom == "" {
		conn.rw.reply(503, "Error: send MAIL first")
		return
	}

	if len(conn.recipients) == config.MaxRecipients {
		conn.rw.reply(452, "Maximum limit of %v recipients reached", config.MaxRecipients)
		return
	}

	p := newParser(args)
	if ok := p.cutPrefix("TO:"); !ok {
		conn.rw.syntaxError("Syntax: RCPT TO:<address>")
		return
	}

	p.trim()
	rcpt, err := p.parseReversePath()
	if err != nil {
		conn.rw.syntaxError("invalid address")
		return
	}

	if config.CheckMailBoxExist {
		if !mailFwd.ExistMailBox(rcpt) {
			conn.rw.reply(550, "mailbox unavailable")
			return
		}
	}

	// UPDATE: parse args

	conn.recipients = append(conn.recipients, rcpt)
	conn.rw.reply(250, "Ok")
	conn.passCmd += 1
}

func (conn *Connection) handleNoop() {
	conn.rw.reply(250, "Yeop")
	conn.passCmd += 1
}

func (conn *Connection) handleQuit() {
	conn.rw.byyy()
	conn.passCmd += 1
	conn.closeWithSuccess()
}

func (conn *Connection) handleReset() {
	conn.passCmd += 1
	conn.reset()
	conn.rw.reply(250, "Flushed")
}

func (conn *Connection) handleData() HandleCommandStatus {
	if len(conn.recipients) == 0 {
		conn.rw.reply(503, "Error: send RCPT first")
		return HandleCommandOk
	}

	// UPDATE: through error when body type is binerymime

	conn.rw.reply(354, "Start mail input end with <CRLF>.<CRLF>")

	data, err := conn.rw.readData()
	if err == io.ErrUnexpectedEOF || err != nil {
		return HandleCommandEOF
	}

	conn.data = data

	ok, status := conn.checkSpf()
	conn.spfStatus = status
	if !ok {
		conn.spfFail = true
		conn.forwardStatus = MailForwardFaild
		conn.reset()

		conn.logger.Warn("%d email received successfully but spf status faild from %s[%s]:%d", conn.mailCount, conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
		conn.mailCount += 1
		return HandleCommandOk
	}

	conn.rw.reply(250, "Ok")
	conn.forwardStatus = MailForwardSuccess
	conn.passCmd += 1
	conn.reset()

	conn.logger.Success("%d email received successfully from %s[%s]:%d", conn.mailCount, conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
	conn.mailCount += 1

	return HandleCommandOk
}

func (conn *Connection) handleBdat(arg string) {
	if !config.ESMTP.Enable {
		conn.rw.reply(502, "Error: ESMTP Disable")
		return
	}

	if len(conn.recipients) == 0 {
		conn.rw.reply(503, "Error: send RCPT first")
		return
	}
	args := strings.Fields(arg)
	if len(args) == 0 {
		conn.rw.reply(501, "Missing chunk size argument")
		return
	}
	if len(args) > 2 {
		conn.rw.reply(501, "Too many arguments")
		return
	}

	last := false
	if len(args) == 2 {
		if !strings.EqualFold(args[1], "LAST") {
			conn.rw.reply(501, "Unknown BDAT argument")
			return
		}
		last = true
	}

	size, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		conn.rw.reply(501, "Malformed size argument")
		return
	}

	if conn.dataBuffer == nil {
		conn.dataBuffer = new(bytes.Buffer)
	}

	if config.ESMTP.MessageSize > 0 && conn.dataBuffer.Len()+int(size) > config.ESMTP.MessageSize {
		conn.rw.reply(552, "Max message size exceeded")
		return
	}

	oldSize := conn.rw.setMaxLineSize(0)
	defer conn.rw.setMaxLineSize(oldSize)

	lr := io.LimitReader(conn.rw.t.R, int64(size))
	n, err := io.Copy(conn.dataBuffer, lr)
	if err != nil || n != int64(size) {
		conn.rw.reply(554, "Error: Transaction failed, data reading error.")
		conn.reset()
		return
	}

	if last {
		conn.data = conn.dataBuffer.Bytes()

		ok, status := conn.checkSpf()
		conn.spfStatus = status
		if !ok {
			conn.spfFail = true
			conn.forwardStatus = MailForwardFaild
			conn.reset()

			conn.logger.Warn("%d email received successfully but spf status faild from %s[%s]:%d", conn.mailCount, conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
			conn.mailCount += 1
			return
		}

		conn.rw.reply(250, "Ok, last %d octets received, total %d", size, conn.dataBuffer.Len())

		conn.forwardStatus = MailForwardSuccess
		conn.passCmd += 1
		conn.reset()

		conn.logger.Success("%d email received successfully from %s[%s]:%d", conn.mailCount, conn.remoteAddress.GetPTR(), conn.remoteAddress.ip.String(), conn.remoteAddress.port)
		conn.mailCount += 1
	} else {
		conn.rw.reply(250, "%d octets received, total %d", size, conn.dataBuffer.Len())
		conn.passCmd += 1
	}
}

func (conn *Connection) checkSpf() (bool, string) {
	if !config.SpfCheck {
		return true, "PASS"
	}
	domain, err := getDomainFromEmail(conn.mailFrom)
	if err != nil {
		conn.rw.replyLines(550, []string{
			"email doesn't delivered because sender",
			fmt.Sprintf("[%s] domain name", conn.mailFrom),
			"is invalid.",
		})
		return false, "FAIL"
	}

	a := spf.CheckHost(conn.remoteAddress.ip, domain, conn.mailFrom, "")
	if a != "PASS" && a != "SOFTFAIL" {
		conn.rw.replyLines(550, []string{
			"email doesn't delivered because sender",
			fmt.Sprintf("domain [%s] does not", domain),
			fmt.Sprintf("designate %s as", conn.remoteAddress.ip.String()),
			"permitted sender.",
		})
		return false, string(a)
	}

	return true, string(a)
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	emailRegex := regexp.MustCompile(`^[\p{L}\p{N}\p{M}\p{S}\p{P}._%+\-]+@[\p{L}\p{N}.\-]+\.[\p{L}]{2,}$`)
	return emailRegex.MatchString(email)
}

func getDomainFromEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if !isValidEmail(email) {
		return "", fmt.Errorf("invalid email address")
	}
	parts := strings.Split(email, "@")
	return parts[1], nil
}
