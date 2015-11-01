package main

import (
	"bytes"
	"errors"
	"strings"

	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gopistolet/mta"
	"github.com/gopistolet/gospf"
	"github.com/gopistolet/gospf/dns"
	"github.com/sloonz/go-maildir"
)

var mailDir *maildir.Maildir

func handleMailDir(state *mta.State) {
	err := errors.New("")

	// Open maildir if it's not yet open
	if mailDir == nil {

		// Open a maildir. If it does not exist, create it.
		mailDir, err = maildir.New("./maildir", true)
		if err != nil {
			log.Errorln(err)
		}
	}

	dataReader := bytes.NewReader(state.Data)

	// Save mail in maildir
	filename, err := mailDir.CreateMail(dataReader)
	if err != nil {
		//log.Println(err)
		log.WithFields(log.Fields{
			"SessionId": state.SessionId.String(),
		}).Error(err)
	} else {
		//log.Println("Maildir: mail written to file: " + filename)
		log.WithFields(log.Fields{
			"SessionId": state.SessionId.String(),
		}).Info("Maildir: mail written to file: " + filename)
	}
}

func handleSPF(state *mta.State) {
	// create SPF instance
	spf, err := gospf.NewSPF(state.From.GetDomain(), &dns.GoSPFDNS{})
	if err != nil {
		log.Errorln(err)
		return
	}

	// check the given IP on that instance
	check, err := spf.CheckIP(state.Ip)
	if err != nil {
		log.Errorln(err)
		return
	}

	log.WithFields(log.Fields{
		"Domain": state.From.GetDomain(),
		"Ip":     state.Ip,
	}).Info("SPF returned " + check)

	// write Authentication-Results header
	// TODO: need value from config here...
	//
	// header field is defined in RFC 5451 section 2.2
	// Authentication-Results: receiver.example.org; spf=pass smtp.mailfrom=example.com;
	headerField := "Authentication-Result: receiver.example.org;" + " spf=" + strings.ToLower(check) + " smtp.mailfrom=" + state.From.GetDomain() + ";\r\n"
	state.Data = append([]byte(headerField), state.Data...)

}

func mail(state *mta.State) {
	log.Debugf("From: %s\n", state.From.Address)
	log.Debugf("To: ")
	for i, to := range state.To {
		log.Printf("%s", to.Address)
		if i != len(state.To)-1 {
			log.Printf(",")
		}
	}
	log.Debugf("CONTENT_START:\n")
	log.Debugf("%s\n", string(state.Data))
	log.Debugf("CONTENT_END\n")
}
