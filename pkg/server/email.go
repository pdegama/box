// U2SMTP - smtp server
//
// Licensed under the MIT License for individual use or a commercial
// license for enterprise use. See LICENSE.txt and COMMERCIAL_LICENSE.txt.

package server

import (
	"bytes"
	"log"

	"github.com/mnako/letters"
	"gopkg.in/yaml.v3"
)

type Email struct {
	Uid string `yaml:"uid"`

	Domain     string   `yaml:"domain"`
	From       string   `yaml:"from"`
	Recipients []string `yaml:"recipients"`
	Data       string   `yaml:"data"`
}

func (e *Email) ToDocument() ([]byte, error) {
	data, err := yaml.Marshal(&e)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (e *Email) ParseMail() (letters.Email, bool) {
	emailReader := bytes.NewReader([]byte(e.Data))
	email, err := letters.ParseEmail(emailReader)
	log.Println("lol")
	log.Println(err)
	log.Println("lol")
	log.Println(email)
	log.Println("lol")
	if err != nil {
		return  letters.Email{}, true
	}

	return email, false
}
