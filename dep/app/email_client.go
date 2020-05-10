package app

import (
	"fmt"
	"net/smtp"
)

type IEmailClient interface {
	Send(from string, to []string, message []byte) (err error)
}

type EmailClient struct {
	Username string
	Password string
	Server   string
	Port     string
}

func NewEmailClient(username string, password string, server string, port string) IEmailClient {
	return &EmailClient{
		Username: username,
		Password: password,
		Server:   server,
		Port:     port,
	}
}

func (e *EmailClient) Send(from string, to []string, message []byte) (err error) {
	auth := smtp.PlainAuth("", e.Username, e.Password, e.Server)
	err = smtp.SendMail(fmt.Sprintf("%s:%s", e.Server, e.Port), auth, from, to, message)
	return err
}
