package main

import (
	"flag"
	"fmt"
	"github.com/stvp/go-toml-config"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var (
	mailAddrsStr = flag.String("mail-address", "", "the mail addresses send to")
	smtpStr      string
	savedIp      string
	exitChan     chan int
	smtp_usr     = config.String("smtp_usr", "")
	smtp_pwd     = config.String("smtp_pwd", "")
	smtp_host    = config.String("smtp_host", "")
	mailListFile = "./mail_list.ini"
	savedFile    = "./saved_ip.ini"
	configFile   = "./cfg.ini"
)

func main() {
	flag.Parse()
	config.Parse(configFile)
	if strings.EqualFold(*mailAddrsStr, "") {
		log.Println("")
	}
	log.Println("mailAddrs = ", *mailAddrsStr)

	var mailArray []string
	if len(*mailAddrsStr) > 0 {
		mailArray = strings.Split(*mailAddrsStr, ",")
		log.Printf("use mail args")
		mailArray = strings.Split(*mailAddrsStr, ",")
	} else {
		mailArray, _ = ReadLines(mailListFile)
	}

	addrsStr := ""

	log.Println("len = ", len(mailArray))
	for i, addr := range mailArray {
		if strings.EqualFold("", strings.Trim(addr, "")) {
			continue
		}
		if i != 0 {
			addrsStr = addrsStr + ","
		}
		log.Printf("%d = %s", i, addr)
		addrsStr = addrsStr + `"` + strings.Trim(addr, "") + `"`
	}

	log.Println("addrsStr = ", addrsStr)

	smtpStr = fmt.Sprintf(`{"username":"%s","password":"%s","host":"%s","sendTos":[%s]}`,
		*smtp_usr, *smtp_pwd, *smtp_host, addrsStr)

	log.Printf("username = %s\n", *smtp_usr)
	log.Printf("sendTos = [%s]", addrsStr)
	ip, err := ioutil.ReadFile(savedFile)
	if err == nil {
		savedIp = string(ip)
	}

	log.Println("savedIp = ", savedIp)
	exitChan = make(chan int)
	go sendMailIfNeed()
	<-exitChan
}

func sendMailIfNeed() {
	for {
		var err error
		log.Println(time.Now(), ": detech if ip had changed...")
		newIp := GetIp()
		if !strings.EqualFold(newIp, "") && !strings.EqualFold(savedIp, newIp) {
			savedIp = newIp
			err = ioutil.WriteFile(savedFile, []byte(newIp), 0644)
			if err != nil {
				log.Println("WriteFile error = ", err.Error())
			} else {
				log.Println("new ip was saved to file ", savedFile)
			}
			smtp, _ := NewSmtpWriter(smtpStr)
			err = smtp.WriteMsg(newIp)
			if err != nil {
				log.Println("smtp write msg error = ", err.Error())
			} else {
				log.Printf("smtp write msg %s ok", newIp)
			}
		} else {
			log.Printf("ip has no changed: newIp = %s, savedIp = %s\n", newIp, savedIp)
		}
		time.Sleep(60 * 1000 * time.Millisecond)
	}
}
