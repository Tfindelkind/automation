package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/Tfindelkind/automation/NTNX-AVM/sendEmail/email"

	"flag"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
)

const (
	appVersion = "0.9-beta"
	gmail      = "gmail"
	other      = "other"
)

// SMTPConfig ...
type SMTPConfig struct {
	Recipient string
	Subject   string
	Message   string
	File      string
	User      string
	Password  string
	Server    string
	Port      string
}

var (
	recipient    *string
	subject      *string
	message      *string
	file         *string
	provider     *string
	user         *string
	password     *string
	server       *string
	port         *string
	listprovider *bool
	debug        *bool
	help         *bool
	version      *bool

	gmailConfig = SMTPConfig{Server: "smtp.gmail.com", Port: "587"}
)

func init() {
	recipient = flag.String("recipient", "", "a string")
	subject = flag.String("subject", "", "a string")
	message = flag.String("message", "", "a string")
	file = flag.String("file", "", "a string")
	provider = flag.String("provider", "", "a string")
	user = flag.String("user", "", "a string")
	password = flag.String("password", "", "a string")
	server = flag.String("server", "", "a string")
	port = flag.String("port", "", "a string")
	listprovider = flag.Bool("listprovider", false, "a bool")
	debug = flag.Bool("debug", false, "a bool")
	help = flag.Bool("help", false, "a bool")
	version = flag.Bool("version", false, "a bool")
}

func printHelp() {

	fmt.Println("Usage: sendEmail [OPTIONS]")
	fmt.Println("sendEmail [ --help | --version ]")
	fmt.Println("")
	fmt.Println("Send email via defined providers")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("")
	fmt.Println("--recipient        Specify recipient of the email")
	fmt.Println("--subject          Specify subject of the email")
	fmt.Println("--message          Optional specify message of the email")
	fmt.Println("--file             Optional specify file which will be attached")
	fmt.Println("--provider         Specify email provider for sending")
	fmt.Println("--user             Specify user for provider")
	fmt.Println("--password         Specify password for provider")
	fmt.Println("--server           Optional when provider=other is used")
	fmt.Println("--port             Optional when provider=other is used")
	fmt.Println("--listprovider     List all providers")
	fmt.Println("--debug            Enables debug mode")
	fmt.Println("--help             List this help")
	fmt.Println("--version          Shows the sendEmail version")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("")
	fmt.Println("sendEmail --recipient=webmaster@thomas-findelkind.de --subject=GetABigger --message=dog --user=MyUser@gmail.com --password=12345")
	fmt.Println("")
}

func evaluateFlags() {

	//help
	if *help {
		printHelp()
		os.Exit(0)
	}

	//listprovider
	if *listprovider {
		fmt.Println(gmail)
		fmt.Println(other)
		os.Exit(0)
	}

	//version
	if *version {
		fmt.Println("Version: " + appVersion)
		os.Exit(0)
	}

	//version
	if *version {
		fmt.Println("Version: " + appVersion)
		os.Exit(0)
	}

	//debug
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	//recipient
	if *recipient == "" {
		log.Warn("mandatory option '--recipient=' is not set")
		os.Exit(0)
	}

	//subject
	if *subject == "" {
		log.Warn("mandatory option '--subject=' is not set")
		os.Exit(0)
	}

	//file
	if *file != "" {
		if _, err := os.Stat(*file); os.IsNotExist(err) {
			log.Error("file: " + *file + " does not exist")
			os.Exit(1)
		}
	}

	//provider
	if *provider == "" {
		log.Warn("option '--provider=' is not set  Default: gmail is used")
		*provider = gmail
	} else {
		if *provider != other {
			log.Fatal("provider: " + *provider + " is unknown. Use --listprovider to list available provider.")
		}
	}

	//username
	if *user == "" {
		log.Warn("mandatory option '--user=' is not set")
		os.Exit(0)
	}

	//password
	if *password == "" {
		log.Warn("ndatory option '--password=' is not set")
		os.Exit(0)
	}
	return
}

//SendGmail ...
func SendGmail(smtpConfig SMTPConfig) {

	m := email.NewMessage(smtpConfig.Subject, smtpConfig.Message)
	m.From = mail.Address{Name: "From", Address: smtpConfig.User}
	m.To = []string{smtpConfig.Recipient}

	// add attachments
	if smtpConfig.File != "" {
		if err := m.Attach(smtpConfig.File); err != nil {
			log.Fatal(err)
		}
	}

	auth := smtp.PlainAuth("", smtpConfig.User, smtpConfig.Password, smtpConfig.Server)
	err := smtp.SendMail(smtpConfig.Server+":"+smtpConfig.Port, auth, m.From.Address, m.Tolist(), m.Bytes())

	if err != nil {
		log.Fatalf("smtp error: %s", err)
		return
	}

	log.Info("Email sent")
}

//SendOther ...
func SendOther(smtpConfig SMTPConfig) {

	m := email.NewMessage(smtpConfig.Subject, smtpConfig.Message)
	m.From = mail.Address{Name: "From", Address: smtpConfig.User}
	m.To = []string{smtpConfig.Recipient}

	// add attachments
	if smtpConfig.File != "" {
		if err := m.Attach(smtpConfig.File); err != nil {
			log.Fatal(err)
		}
	}

	auth := smtp.PlainAuth("", smtpConfig.User, smtpConfig.Password, smtpConfig.Server)
	err := smtp.SendMail(smtpConfig.Server+":"+smtpConfig.Port, auth, m.From.Address, m.Tolist(), m.Bytes())

	if err != nil {
		log.Fatalf("smtp error: %s", err)
		return
	}

	log.Info("Email sent")
}

func main() {

	flag.Usage = printHelp
	flag.Parse()

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	evaluateFlags()

	var smtpCon SMTPConfig

	switch *provider {
	case gmail:
		smtpCon = gmailConfig

		smtpCon.Recipient = *recipient
		smtpCon.Subject = *subject
		smtpCon.Message = *message
		smtpCon.File = *file
		smtpCon.User = *user
		smtpCon.Password = *password

		SendGmail(smtpCon)
	case other:
		smtpCon.Recipient = *recipient
		smtpCon.Subject = *subject
		smtpCon.Message = *message
		smtpCon.File = *file
		smtpCon.User = *user
		smtpCon.Password = *password
		smtpCon.Server = *server
		smtpCon.Port = *port

		SendOther(smtpCon)
	}

}
