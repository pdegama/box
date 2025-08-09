package server

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/pdegama/box/pkg/logx"
)

var greetReplyMessage []string = []string{} // greet reply message for esmtp/ehlo
var tlsConfig *tls.Config

// smtp server pre-process like EHLO
// greet reply, tls config etc
func smtpServerPreProcess(l *logx.Log) {
	// greet reply
	if config.ESMTP.Enable {
		reply := []string{
			"PIPELINING",
			"8BITMIME",
			// "ENHANCEDSTATUSCODES", // TODO
			"CHUNKING",
		}

		if config.ESMTP.Tls {
			reply = append(reply, "STARTTLS")
		}
		if config.ESMTP.Utf8 {
			reply = append(reply, "SMTPUTF8")
		}
		if config.ESMTP.BinaryMime {
			reply = append(reply, "BINARYMIME")
		}
		// UPDATE: DSN
		// if c.server.EnableDSN {
		// 	reply = append(reply, "DSN")
		// }
		if config.ESMTP.MessageSize > 0 {
			reply = append(reply, fmt.Sprintf("SIZE %v", config.ESMTP.MessageSize))
		} else {
			reply = append(reply, "SIZE")
		}
		if config.MaxRecipients > 0 {
			reply = append(reply, fmt.Sprintf("LIMITS RCPTMAX=%v", config.MaxRecipients))
		}

		greetReplyMessage = reply
	}

	// tls config
	if config.ESMTP.Enable && config.ESMTP.Tls {
		cert, err := tls.LoadX509KeyPair(config.TlsConfig.Cert, config.TlsConfig.Key)
		if err != nil {
			l.Error("Error loading certificates: %v", err)
			os.Exit(1)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS10,
		}
	}
}
