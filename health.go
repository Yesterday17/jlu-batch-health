package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type ReportTime uint8

func ReportAll(b *tb.Bot, m ReportMode, msg string) {
	Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		if m == Config.Mode {
			if msg != "" {
				_, _ = b.Send(tb.ChatID(Config.Owner), msg)
			}
			Report(b, m.GetReportTime(), user)
			time.Sleep(3 * time.Minute)
		}
		return true
	})
}

func Report(bot *tb.Bot, t ReportTime, u *User) {
	var fields = GetReportFields(t)
	var msg *tb.Message

	for i := 0; i < Config.MaxRetry; i++ {
		if i == 0 {
			_, _ = bot.Send(tb.ChatID(Config.Owner), fmt.Sprintf("%s 开始进行%s……", u.Username, t.Name()))
		} else {
			_, _ = bot.Send(tb.ChatID(Config.Owner), fmt.Sprintf("%s 开始进行第 %d/%d 次%s重试……", u.Username, i, Config.MaxRetry-1, t.Name()))
		}

		msg, _ = bot.Send(tb.ChatID(Config.Owner), "登录中……")
		err := u.Login()
		if msg != nil {
			_ = bot.Delete(msg)
		}
		if err != nil {
			_, _ = bot.Send(tb.ChatID(Config.Owner), fmt.Sprintf("登录失败：%s", err.Error()))
			time.Sleep(10 * time.Second)
			continue
		}

		msg, _ = bot.Send(tb.ChatID(Config.Owner), fmt.Sprintf("登录成功！正在获取表单……"))
		csrf, body, form, err := u.GetForm()
		if msg != nil {
			_ = bot.Delete(msg)
		}
		if err != nil {
			_, _ = bot.Send(tb.ChatID(Config.Owner), fmt.Sprintf("表单获取失败：%s", err.Error()))
			if errors.Is(err, EhallSystemError) {
				break
			}

			time.Sleep(10 * time.Second)
			continue
		}

		// Merge common fields
		u.MergeTo(&form.Data)
		fields.MergeTo(&form.Data)

		// Suggest fields if needed
		if form.Data["fieldSQbj"] == "" || form.Data["fieldSQbj_Name"] == "null" {
			// merge Suggest fields
			cNj, err := u.GetBotField("nj")
			if err != nil {
				_, _ = bot.Send(tb.ChatID(Config.Owner), err.Error())
				break
			}
			err = u.SuggestField(form, form.Fields["fieldSQnj"], cNj, csrf)
			if err != nil {
				_, _ = bot.Send(tb.ChatID(Config.Owner), err.Error())
				break
			}

			cBj, err := u.GetBotField("bj")
			if err != nil {
				_, _ = bot.Send(tb.ChatID(Config.Owner), err.Error())
				break
			}
			err = u.SuggestField(form, form.Fields["fieldSQbj"], cBj, csrf)
			if err != nil {
				_, _ = bot.Send(tb.ChatID(Config.Owner), err.Error())
				break
			}
		}

		formJson, _ := json.Marshal(form.Data)
		body.Set("formData", string(formJson))

		msg, _ = bot.Send(tb.ChatID(Config.Owner), "表单获取成功，正在打卡……")
		err = u.DoReport(body)
		if msg != nil {
			_ = bot.Delete(msg)
		}
		if err != nil {
			_, _ = bot.Send(tb.ChatID(Config.Owner), fmt.Sprintf("打卡失败：%s", err.Error()))
			if errors.Is(err, EhallSystemError) {
				break
			}

			time.Sleep(10 * time.Second)
			continue
		}

		// success
		_, _ = bot.Send(tb.ChatID(Config.Owner), "打卡成功！")
		break
	}
	if msg != nil {
		_ = bot.Delete(msg)
	}
}
