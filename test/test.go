package main

import "fmt"
import "oscar"

type Bot struct {
	*oscar.Client
}

func NewBot() *Bot {
	bot := &Bot{}
	bot.Client = oscar.New("93366018", "rvRktYpb", bot)
	return bot
}

func (b *Bot) Message(username, text string) {
	println("username =", username, "text =", text)
	b.SendMessage(username, "и тебе")

	/* b.RequestUserInfo(username,
		func(info oscar.UserInfo) {
			fmt.Println("username =", username, ", info =", info)
		}) */
	b.UpdateUserInfo(map[string]string{"nickname": "Bots - Test", "firstname": ""},
		func(f bool) {
			fmt.Println("UpdateUserInfo: f =", f)
		})
}

func (b *Bot) Subscription(username string) {
	println("username =", username)
}

func main() {
	NewBot().Run()
}
