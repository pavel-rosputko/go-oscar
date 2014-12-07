package main

import (
	"io/ioutil"
	"strings"
	"time"

	"g/oscar"
)

var nickname = "Bots' Test Account"

var ch = make(chan bool)

type Bot struct {
	*oscar.Client
}

func NewBot(username, password string) *Bot {
	bot := &Bot{}
	bot.Client = oscar.New(username, password, bot)
	return bot
}

func (b *Bot) Ready() {
	println("username =", b.Username(), "ready")
	b.UpdateUserInfo(map[string]string{"nickname": nickname},
		func(f bool) {
			println("username =", b.Username(), "f =", f)
			ch <- true
		})
}

func (b *Bot) Message(username, message string) {
}

func (b *Bot) Subscription(username string) {
}

var ips []string

func main() {
	ipsBytes, _ := ioutil.ReadFile("etc/ips")
	ips = strings.Split(string(ipsBytes), "\n", -1)

	accountsBytes, _ := ioutil.ReadFile("etc/accounts")
	accounts := strings.Split(string(accountsBytes), "\n", -1)

	for _, account := range accounts {
		list := strings.Split(account, ";", 2)
		username, password := list[0], list[1]

		println("username =", username, "password =", password)

		go NewBot(username, password).Run()

		time.Sleep(7 * 1000000000 / int64(len(ips)))
		defer func() { <-ch }()
	}
}
