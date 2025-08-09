package server

import (
	"fmt"
	"io"
	"net"
	"net/textproto"
	"time"

	limitlinereader "github.com/pdegama/box/pkg/limit_line_reader"
)

type TextReaderWriter struct {
	netConn net.Conn
	rwc     *ReadWriteClose
	t       *textproto.Conn
	*limitlinereader.LimitLineReader
}

type ReadWriteClose struct {
	io.Reader
	io.Writer
	io.Closer
}

// get new reader and write form net.Conn or io.ReadWriter
func newTextReaderWriter(conn net.Conn) *TextReaderWriter {
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
	}
}

func (rw *TextReaderWriter) reply(code int, format string, a ...any) {
	// TODO: EnhancedCode
	rw.t.PrintfLine("%d %s", code, fmt.Sprintf(format, a...))
}

func (rw *TextReaderWriter) replyLines(code int, lines []string) {
	// TODO: EnhancedCode
	for i := 0; i < len(lines)-1; i++ {
		rw.t.PrintfLine("%d-%v", code, lines[i])
	}
	rw.t.PrintfLine("%d %v", code, lines[len(lines)-1])
}

func (rw *TextReaderWriter) greet(hostname string) {
	rw.t.PrintfLine("220 %s %s", hostname, config.ClientGreet)
}

func (rw *TextReaderWriter) byyy() {
	rw.t.PrintfLine("221 %s", config.ClientByyy)
}

func (rw *TextReaderWriter) busy(hostname string) {
	rw.t.PrintfLine("421 %s Service not available, max clients exceeded", hostname)
}

func (rw *TextReaderWriter) timeout(hostname string) {
	rw.t.PrintfLine("421 %s Error: timeout exceeded", hostname)
}

func (rw *TextReaderWriter) longLine(hostname string) {
	rw.t.PrintfLine("500 %s Error: too long line", hostname)
}

func (rw *TextReaderWriter) syntaxError(format string, a ...any) {
	rw.t.PrintfLine("501 %s", fmt.Sprintf(format, a...))
}

func (rw *TextReaderWriter) cmdNotRecognized(cmd string) {
	if cmd == "" {
		rw.t.PrintfLine("500 Error: command not recognized")
	} else {
		rw.t.PrintfLine("500 Error: %s command not recognized", cmd)
	}
}

func (rw *TextReaderWriter) cmdNotImplemented(cmd string) {
	rw.t.PrintfLine("502 Error: %s command not implemented", cmd)
}

// set max lines size and return return old size
func (rw *TextReaderWriter) setMaxLineSize(size int) int {
	oldSize := rw.MaxLineSize
	rw.MaxLineSize = size
	return oldSize
}

func (rw *TextReaderWriter) readLine() (string, error) {
	rw.setTimeout(config.Timeout)
	defer rw.clearTimeout()

	return rw.t.ReadLine()
}

func (rw *TextReaderWriter) readData() ([]byte, error) {
	rw.setTimeout(config.Timeout * 5)
	defer rw.clearTimeout()

	return rw.t.ReadDotBytes()
}

// set read time out
func (rw *TextReaderWriter) setTimeout(t time.Duration) {
	rw.netConn.SetReadDeadline(time.Now().Add(t))
}

// clear read time out
func (rw *TextReaderWriter) clearTimeout() {
	rw.netConn.SetReadDeadline(time.Time{})
}
