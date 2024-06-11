package smtp

import (
	"context"
	"fmt"
	"github.com/JMURv/unona/services/pkg/config"
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
	"io"
)

type EmailServer struct {
	server string
	port   int
	user   string
	pass   string
	admin  string
}

func New(conf *config.EmailConfig) *EmailServer {
	return &EmailServer{
		server: conf.Server,
		port:   conf.Port,
		user:   conf.User,
		pass:   conf.Pass,
		admin:  conf.Admin,
	}
}

func (s *EmailServer) GetMessageBase(subject string, toEmail string) *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("From", s.user)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	return m
}

func (s *EmailServer) Send(m *gomail.Message) error {
	d := gomail.NewDialer(s.server, s.port, s.user, s.pass)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func (s *EmailServer) SendVerificationEmail(_ context.Context, userUUID uuid.UUID, photo []byte) error {
	m := s.GetMessageBase("User Photo Verification", s.admin)

	m.SetBody("text/plain", fmt.Sprintf("User UUID: %v", userUUID.String()))
	m.Attach(
		"user.jpg",
		gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(photo)
			return err
		}),
	)

	return s.Send(m)
}

func (s *EmailServer) SendLoginEmail(_ context.Context, code uint64, toEmail string) error {
	m := s.GetMessageBase("Login Code", toEmail)
	m.SetBody("text/plain", fmt.Sprintf("Login code: %v", code))
	return s.Send(m)
}

func (s *EmailServer) SendActivationCodeEmail(_ context.Context, code uint64, toEmail string) error {
	m := s.GetMessageBase("Activation Code", toEmail)
	m.SetBody("text/plain", fmt.Sprintf("Activation code: %v", code))
	return s.Send(m)
}

func (s *EmailServer) SendForgotPasswordEmail(_ context.Context, token, uid64, toEmail string) error {
	m := s.GetMessageBase("Forgot Password Code", toEmail)

	params := fmt.Sprintf("?uidb64=%v&token=%v", uid64, token)
	resetURL := fmt.Sprintf("/email/password-reset/%v", params)

	m.SetBody("text/plain", fmt.Sprintf("Forgot password URL: %v", resetURL))
	return s.Send(m)
}
