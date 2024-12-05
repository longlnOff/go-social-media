package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)


type SendGridMailer struct {
	fromEmail string
	apiKey string
	client *sendgrid.Client
}

func NewSendGridMailer(fromEmail string, apiKey string) *SendGridMailer {
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey: apiKey,
		client: sendgrid.NewSendClient(apiKey),
	}
}

func (m *SendGridMailer) Send(templateFile string, username string, email string, data any, isSandbox bool) (int, error) {
	fromEmail := mail.NewEmail(FromName, m.fromEmail)
	toEmail := mail.NewEmail(username, email)

	// template parsing and building
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}
	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}


	message := mail.NewSingleEmail(fromEmail, subject.String(), toEmail, "", body.String())

	// Check sandbox
	message.SetMailSettings(
		&mail.MailSettings{
			SandboxMode: &mail.Setting{
				Enable: &isSandbox,
			},
		},
	)


	// var retryError error
    var responseStruct ErrorResponse

	for i := 0; i < MaxRetries; i++ {
		response, retryError := m.client.Send(message)
		_ = json.Unmarshal([]byte(response.Body), &responseStruct)

		if retryError != nil || response.StatusCode != 202 {
			// exponential backoff
			time.Sleep(time.Millisecond * time.Duration((1+i)*100))
			continue
		}
		
		return response.StatusCode, nil
	}
	
	return -1, fmt.Errorf("failed to send email after %d attempts, error: %v", MaxRetries, errors.New(responseStruct.Errors[0].Message))
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Message string      `json:"message"`
	Field   *string    `json:"field"`    // using pointer since it can be null
	Help    *string    `json:"help"`     // using pointer since it can be null
}
