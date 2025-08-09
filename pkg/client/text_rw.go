package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	limitlinereader "github.com/pdegama/box/pkg/limit_line_reader"
)

type TextReaderWriter struct {
	rwc *ReadWriteClose
	t   *textproto.Conn
	*limitlinereader.LimitLineReader
	netConn net.Conn
	Timeout time.Duration
}

type ReadWriteClose struct {
	io.Reader
	io.Writer
	io.Closer
}

type SMTPErrorType int

const (
	SMTPErrorTemp SMTPErrorType = iota // 4yz  Transient Negative Completion reply
	SMTPErrorFail                      // 5yz  Permanent Negative Completion reply
)

type EnhancedCode [3]int
type SMTPServerError struct {
	Code         int
	EnhancedCode EnhancedCode
	Message      string
}

func (err SMTPServerError) Error() string {
	s := fmt.Sprintf("Delivery Error: %03d", err.Code)
	if err.Message != "" {
		s += " - " + err.Message
	}
	return s
}

func (err SMTPServerError) GetErrorType() SMTPErrorType {
	if err.Code >= 400 && err.Code < 500 {
		return SMTPErrorTemp
	}
	return SMTPErrorFail
}

func newTextReaderWriter(conn net.Conn, timeout time.Duration) *TextReaderWriter {
	textReader := &limitlinereader.LimitLineReader{
		Reader:      conn,
		MaxLineSize: 2000, // Doubled maximum line length per RFC 5321 (Section 4.5.3.1.6)
	}

	rwc := ReadWriteClose{
		Reader: textReader,
		Writer: conn,
		Closer: conn,
	}

	text := textproto.NewConn(rwc)
	return &TextReaderWriter{
		netConn:         conn,
		rwc:             &rwc,
		t:               text,
		LimitLineReader: textReader,
		Timeout:         timeout,
	}
}

func (rw *TextReaderWriter) readResponse(expectCode int) (int, string, error) {
	rw.netConn.SetReadDeadline(time.Now().Add(rw.Timeout))
	defer rw.netConn.SetReadDeadline(time.Time{})

	code, msg, err := rw.t.ReadResponse(expectCode)
	if protoErr, ok := err.(*textproto.Error); ok {
		err = toSMTPServerErr(protoErr)
	}
	return code, msg, err
}

func (rw *TextReaderWriter) cmd(expectCode int, format string, args ...interface{}) (int, string, error) {
	id, err := rw.t.Cmd(format, args...)
	if err != nil {
		if protoErr, ok := err.(*textproto.Error); ok {
			err = toSMTPServerErr(protoErr)
		}
		return 0, "", err
	}

	rw.t.StartResponse(id)
	defer rw.t.EndResponse(id)

	return rw.readResponse(expectCode)
}

func (rw *TextReaderWriter) bdat(r io.Reader, n int, last bool) (int, string, error) {
	f := fmt.Sprintf("BDAT %d", n)
	if last {
		f += " LAST"
	}

	id, err := rw.t.Cmd(f)
	if err != nil {
		if protoErr, ok := err.(*textproto.Error); ok {
			err = toSMTPServerErr(protoErr)
		}
		return 0, "", err
	}

	lr := io.LimitReader(r, int64(n))
	io.Copy(rw.t.W, lr)

	rw.t.W.Flush()

	rw.t.StartResponse(id)
	defer rw.t.EndResponse(id)

	return rw.readResponse(250)
}

func (rw *TextReaderWriter) data(data []byte) (int, string, error) {
	dataReader := bytes.NewReader(data)
	dataWriter := rw.t.DotWriter()

	_, err := io.Copy(dataWriter, dataReader)
	if err != nil {
		return 0, "", err
	}

	err = dataWriter.Close()
	if err != nil {
		if protoErr, ok := err.(*textproto.Error); ok {
			err = toSMTPServerErr(protoErr)
		}
		return 0, "", err
	}

	return rw.readResponse(250)
}

func toSMTPServerErr(protoErr *textproto.Error) SMTPServerError {
	smtpErr := SMTPServerError{
		Code:    protoErr.Code,
		Message: protoErr.Msg,
	}

	parts := strings.SplitN(protoErr.Msg, " ", 2)
	if len(parts) != 2 {
		return smtpErr
	}

	enchCode, err := parseEnhancedCode(parts[0])
	if err != nil {
		return smtpErr
	}

	msg := parts[1]

	// Per RFC 2034, enhanced code should be prepended to each line.
	msg = strings.ReplaceAll(msg, "\n"+parts[0]+" ", "\n")

	smtpErr.EnhancedCode = enchCode
	smtpErr.Message = msg
	return smtpErr
}

func parseEnhancedCode(s string) (EnhancedCode, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return EnhancedCode{}, fmt.Errorf("wrong amount of enhanced code parts")
	}

	code := EnhancedCode{}
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return code, err
		}
		code[i] = num
	}
	return code, nil
}
