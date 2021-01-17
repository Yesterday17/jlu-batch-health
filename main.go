package main

import (
	"flag"
	"log"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	var token, proxy, accountsPath string
	var owner int64
	var maxUsers uint
	flag.StringVar(&token, "token", "", "Telegram Bot Token")
	flag.StringVar(&proxy, "proxy", "", "http 代理地址")
	flag.Int64Var(&owner, "owner", 0, "Bot 拥有者 ChatID")
	flag.StringVar(&accountsPath, "accounts", "./accounts.json", "存储用户帐户的路径")
	flag.UintVar(&maxUsers, "max-users", 100, "最大用户数量")
	flag.Parse()

	LoadConfig(token, proxy, owner, accountsPath, maxUsers)
	err := LoadUsers()
	if err != nil {
		log.Fatal(err)
	}

	b, err := tb.NewBot(tb.Settings{
		Token:  Config.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	InitCronJobs(b)

	b.Handle("/report", func(m *tb.Message) {
		user, ok := Users.Load(m.Payload)
		if !ok {
			return
		}

		u := user.(*User)
		t := Config.Mode.GetReportTime()
		if t == ReportTimeUnknown {
			_, _ = b.Reply(m, "不在打卡时间段内。")
			return
		}

		go Report(b, t, u)
	})

	b.Start()
}
