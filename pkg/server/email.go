package server

import (
	"bytes"
	"log"
	"time"

	"github.com/mnako/letters"
	"gopkg.in/yaml.v3"
)

// recive mail structure
type Email struct {
	Uid     string    `yaml:"uid"`
	Time    time.Time `yaml:"time"`
	Success bool      `yaml:"success"`

	Cmds  string           `yaml:"cmds"`
	Tls   bool             `yaml:"tls"`
	PtrIP ServerClientInfo `yaml:"ptr_ip"`

	Domain    string `yaml:"domain"`
	PtrMatch  bool   `yaml:"ptr_match"`
	SpfFail   bool   `yaml:"spf_fail"`
	SpfStatus string `yaml:"spf_status"`

	From       string   `yaml:"from"`
	Recipients []string `yaml:"recipients"`
	UseBdat    bool     `yaml:"use_bdat"`
	Data       string   `yaml:"data"`
	dataByte   []byte
}

type ServerClientInfo struct {
	ServerPtr string `yaml:"server_ptr"`
	ServerIP  string `yaml:"server_ip"`
	ClinetPtr string `yaml:"client_ptr"`
	ClientIP  string `yaml:"client_ip"`
}

func (e *Email) ToDocument() ([]byte, error) {
	e.Data = string(e.dataByte)
	data, err := yaml.Marshal(&e)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (e *Email) GetBytes() []byte {
	return e.dataByte
}

func (e *Email) ParseMail() (letters.Email, bool) {
	emailReader := bytes.NewReader([]byte(e.Data))
	email, err := letters.ParseEmail(emailReader)
	if err != nil {
		if config.Dev {
			log.Println(err)
		}
		return letters.Email{}, false
	}

	return email, true
}
