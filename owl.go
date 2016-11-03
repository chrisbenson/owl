// Inspired by the owls that carry letters in the world of Harry Potter.
package owl

import (
  "crypto/tls"
  "errors"
  "fmt"
  "net"
  "net/mail"
  "net/smtp"
  "github.com/sourcegraph/go-ses"
)

type provider string

const (
    SMTP provider = "" // FQDN + port
    AWSSMTP provider = ""
    AWSSES provider = ""
)

type Message struct {
	From     string    `json:"from, omitempty"`
	To       []string  `json:"to, omitempty"`
	CC       []string  `json:"cc, omitempty"`
	BCC      []string  `json:"bcc, omitempty"`
	Subject  string    `json:"subject, omitempty"`
	Body     string    `json:"body, omitempty"`
  HTML    bool `json:"html"`
	Error    error     `json:"error, omitempty"`
}

type Params struct {
  Provider  provider `json:"provider"`
  ID  string `json:"id"`
  Password string `json:"password"`
  Server string `json:"server, omitempty"`
}

type Sender interface {
    Send(p *Params) error
}

func (m *Message) Send(p *Params) error {
  // var err error
  // provider := p.Provider
  // err := sendToAWSSES(m, p)

  // for testing/debugging
  err := sendToConsole(m, p)

  // switch provider {
  // case SMTP:
  //     //err = sendToSMTP(m, p)
  //     err = test(m, p)
  //   case AWSSMTP:
  //     err = sendToAWSSMTP(m, p)
  //   case AWSSES:
  //     err = sendToAWSSES(m, p)
  //   default:
  //     err = errors.New("No email provider was provided.")
  // }
  if err != nil {
    return errors.New("switch p.Provider | " + err.Error())
  }

  return nil
}

/*
Provider-specific implementations below.
*/

// SMTP
func sendToSMTP(m *Message, p *Params) error {
  var err error
  auth := smtp.PlainAuth(
      "",
      p.ID,
      p.Password,
      p.Server,
  )
  err = smtp.SendMail(
      p.Server,
      auth,
      m.From,
      m.To,
      []byte(m.Body),
  )
  return errors.New("smtp.SendMail | " + err.Error())
}

// AWS-SMTP
func sendToAWSSMTP(m *Message, p *Params) error {
  return errors.New("AWS SMTP not implemented yet.")
}

// AWS-SES
func sendToAWSSES(m *Message, p *Params) error {

  var provider ses.Config
  var result string
  var err error

  if p.Provider == "" || p.ID == "" || p.Password == "" {
    provider = ses.EnvConfig
  } else {
    provider = ses.Config{
      Endpoint: string(p.Provider),
      AccessKeyID: p.ID,
      SecretAccessKey: p.Password}
  }

  if m.HTML {
    result, err = provider.SendEmailHTML(m.From, m.To[0], m.Subject, m.Body, m.Body)
  } else {
    result, err = provider.SendEmail(m.From, m.To[0], m.Subject, m.Body)
  }

  fmt.Println(result)

  if err != nil {
    return err
  }

  return nil

}

func sendToConsole(m *Message, p *Params) error {
  fmt.Println("From: " + m.From)
  fmt.Println("To: " + m.To[0])
  fmt.Println("Subject: " + m.Subject)
  fmt.Println(m.Body)
  return nil
}

func test(m *Message, p *Params) error {

  to := mail.Address{"", m.To[0]}
  from := mail.Address{"", m.From}

  // Setup headers
  headers := make(map[string]string)
  headers["From"] = from.String()
  headers["To"] = to.String()
  headers["Subject"] = m.Subject

  fmt.Println("From: " + headers["From"])
  fmt.Println("To: " + headers["To"])
  fmt.Println("Subject: " + headers["Subject"])

  // Setup message
  message := ""
  for k,v := range headers {
      message += fmt.Sprintf("%s: %s\r\n", k, v)
  }
  message += "\r\n" + m.Body

  // Connect to the SMTP Server
  servername := p.Server

  host, _, _ := net.SplitHostPort(servername)

  auth := smtp.PlainAuth("", p.ID, p.Password, host)

  // TLS config
  tlsconfig := &tls.Config {
      InsecureSkipVerify: true,
      ServerName: host,
  }

  // Here is the key, you need to call tls.Dial instead of smtp.Dial
  // for smtp servers running on 465 that require an ssl connection
  // from the very beginning (no starttls)
  conn, err := tls.Dial("tcp", servername, tlsconfig)
  if err != nil {
      return errors.New("tls.Dial | " + err.Error())
  }

  c, err := smtp.NewClient(conn, host)
  if err != nil {
      return errors.New("smtp.NewClient | " + err.Error())
  }

  // Auth
  if err = c.Auth(auth); err != nil {
      return errors.New("c.Auth | " + err.Error())
  }

  // To && From
  if err = c.Mail(m.From); err != nil {
      return errors.New("c.Mail | " + err.Error())
  }

  if err = c.Rcpt(m.To[0]); err != nil {
      return errors.New("c.Rcpt | " + err.Error())
  }

  // Data
  w, err := c.Data()
  if err != nil {
      return errors.New("c.Data | " + err.Error())
  }

  _, err = w.Write([]byte(m.Body))
  if err != nil {
      return errors.New("w.Write | " + err.Error())
  }

  err = w.Close()
  if err != nil {
      return errors.New("w.Close | " + err.Error())
  }

  c.Quit()

  return err

}
