package engine

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/pkg/errors"
)

// SMTPWriter implements LoggerInterface and is used to send emails via given SMTP-server.
type SMTPWriter struct {
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	Host               string   `json:"host"`
	Subject            string   `json:"subject"`
	FromAddress        string   `json:"fromAddress"`
	RecipientAddresses []string `json:"sendTos"`
	Level              int      `json:"level"`
	formatter          LogFormatter
	Formatter          string `json:"formatter"`
}

// NewSMTPWriter creates the smtp writer.
func newSMTPWriter() Logger {
	res := &SMTPWriter{Level: LevelTrace}
	res.formatter = res
	return res
}

// Init SMTP writer with json config.
// config like:
//	{
//		"username":"example@bhojpur.net",
//		"password:"password",
//		"host":"smtp.bhojpur.net:465",
//		"subject":"email title",
//		"fromAddress":"from@bhojpur.net",
//		"sendTos":["email1","email2"],
//		"level":LevelError
//	}
func (s *SMTPWriter) Init(config string) error {
	res := json.Unmarshal([]byte(config), s)
	if res == nil && len(s.Formatter) > 0 {
		fmtr, ok := GetFormatter(s.Formatter)
		if !ok {
			return errors.New(fmt.Sprintf("the formatter with name: %s not found", s.Formatter))
		}
		s.formatter = fmtr
	}
	return res
}

func (s *SMTPWriter) getSMTPAuth(host string) smtp.Auth {
	if len(strings.Trim(s.Username, " ")) == 0 && len(strings.Trim(s.Password, " ")) == 0 {
		return nil
	}
	return smtp.PlainAuth(
		"",
		s.Username,
		s.Password,
		host,
	)
}

func (s *SMTPWriter) SetFormatter(f LogFormatter) {
	s.formatter = f
}

func (s *SMTPWriter) sendMail(hostAddressWithPort string, auth smtp.Auth, fromAddress string, recipients []string, msgContent []byte) error {
	client, err := smtp.Dial(hostAddressWithPort)
	if err != nil {
		return err
	}

	host, _, _ := net.SplitHostPort(hostAddressWithPort)
	tlsConn := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err = client.StartTLS(tlsConn); err != nil {
		return err
	}

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(fromAddress); err != nil {
		return err
	}

	for _, rec := range recipients {
		if err = client.Rcpt(rec); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msgContent)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

func (s *SMTPWriter) Format(lm *LogMsg) string {
	return lm.OldStyleFormat()
}

// WriteMsg writes message in smtp writer.
// Sends an email with subject and only this message.
func (s *SMTPWriter) WriteMsg(lm *LogMsg) error {
	if lm.Level > s.Level {
		return nil
	}

	hp := strings.Split(s.Host, ":")

	// Set up authentication information.
	auth := s.getSMTPAuth(hp[0])

	msg := s.Format(lm)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	contentType := "Content-Type: text/plain" + "; charset=UTF-8"
	mailmsg := []byte("To: " + strings.Join(s.RecipientAddresses, ";") + "\r\nFrom: " + s.FromAddress + "<" + s.FromAddress +
		">\r\nSubject: " + s.Subject + "\r\n" + contentType + "\r\n\r\n" + fmt.Sprintf(".%s", lm.When.Format("2018-03-26 15:04:05")) + msg)

	return s.sendMail(s.Host, auth, s.FromAddress, s.RecipientAddresses, mailmsg)
}

// Flush implementing method. empty.
func (s *SMTPWriter) Flush() {
}

// Destroy implementing method. empty.
func (s *SMTPWriter) Destroy() {
}

func init() {
	Register(AdapterMail, newSMTPWriter)
}
