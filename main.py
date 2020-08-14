import json
import os
import glob
import schedule
import threading
import time
import telebot

import health
from telebot.types import Message
from telebot import apihelper

if os.environ.__contains__("TG_PROXY"):
    apihelper.proxy = {"https": os.environ["TG_PROXY"]}

os.chdir(os.path.dirname(os.path.realpath(__file__)))
bot = telebot.TeleBot(os.environ["BOT_TOKEN"], parse_mode=None, threaded=True)
user_dict = {}


def load_config():
    for filename in glob.glob("./accounts/*.json"):
        data = json.load(open(filename))
        user_dict[data["chat_id"]] = data


def save_user_config(chat_id):
    info = user_dict[chat_id]
    json.dump(info, open("./accounts/{}.json".format(chat_id), "w"))


@bot.message_handler(commands=["info"])
def info(message: Message):
    chat_id = message.chat.id
    if chat_id not in user_dict:
        bot.reply_to(message, "无用户信息！")
    else:
        user = user_dict[chat_id]
        bot.reply_to(message, "用户名：{}\n"
                              "密码：[隐藏]\n"
                              "校区：{}\n"
                              "寝室楼号：{}\n"
                              "寝室号：{}".format(user["username"], user["fields"]["fieldSQxq"],
                                              user["fields"]["fieldSQgyl"], user["fields"]["fieldSQqsh"]))


@bot.message_handler(commands=["start"])
def start(message: Message):
    msg = bot.reply_to(message, "欢迎使用本科生每日打卡 Bot。\n"
                                "为正常使用该 Bot,请按照提示的步骤进行信息填写。\n"
                                "\n"
                                "请输入用户名：")
    bot.register_next_step_handler(msg, step_username)


def step_username(message: Message):
    try:
        chat_id = message.chat.id
        user_dict[chat_id] = {
            "chat_id": chat_id,
            "username": message.text,
            "password": "",
            "fields": {
                "fieldSQxq": "",
                "fieldSQgyl": "",
                "fieldSQqsh": "",
            }
        }
        msg = bot.reply_to(message, "请输入密码：")
        bot.register_next_step_handler(msg, step_password)
    except Exception as e:
        bot.reply_to(message, e.__str__())


def step_password(message: Message):
    try:
        chat_id = message.chat.id
        user_dict[chat_id]["password"] = message.text

        msg = bot.reply_to(message, "请输入校区号（1 为中心校区）：")

        # delete password message immediately
        bot.delete_message(chat_id, message.message_id)
        bot.register_next_step_handler(msg, step_district)
    except Exception as e:
        bot.reply_to(message, e.__str__())
        bot.delete_message(chat_id, message.message_id)


def step_district(message: Message):
    try:
        chat_id = message.chat.id
        user_dict[chat_id]["fields"]["fieldSQxq"] = message.text

        msg = bot.reply_to(message, "请输入寝室楼号（1 为北苑一公寓）：")
        bot.register_next_step_handler(msg, step_dormitory)
    except Exception as e:
        bot.reply_to(message, e.__str__())


def step_dormitory(message: Message):
    try:
        chat_id = message.chat.id
        user_dict[chat_id]["fields"]["fieldSQgyl"] = message.text

        msg = bot.reply_to(message, "请输入寝室号：")
        bot.register_next_step_handler(msg, step_room)
    except Exception as e:
        bot.reply_to(message, e.__str__())


def step_room(message: Message):
    try:
        chat_id = message.chat.id
        user_dict[chat_id]["fields"]["fieldSQqsh"] = message.text

        save_user_config(chat_id)
        bot.send_message(chat_id, "信息填写完成！请使用 /info 查看信息。\n"
                                  "如有信息错误，请使用 /start 重新开始本步骤。")
    except Exception as e:
        bot.reply_to(message, e.__str__())


def rpt(name, type: int):
    def report():
        for chat_id in user_dict:
            health.report(bot, chat_id, user_dict[chat_id], name, type)

    return report


schedule.every().day.at("07:20").do(rpt("早打卡", 0))
schedule.every().day.at("11:20").do(rpt("午打卡", 1))
schedule.every().day.at("18:20").do(rpt("晚打卡", 2))
schedule.every().day.at("21:20").do(rpt("晚点名", 2))


@bot.message_handler(commands=["trigger", "trigger0", "trigger1", "trigger2", "trigger3"])
def trigger(message: Message):
    if message.text == "/trigger 0" or message.text == "/trigger0":
        rpt("早打卡", 0)()
    elif message.text == "/trigger 1" or message.text == "/trigger1":
        rpt("午打卡", 1)()
    elif message.text == "/trigger 2" or message.text == "/trigger2":
        rpt("晚打卡", 2)()
    elif message.text == "/trigger 3" or message.text == "/trigger3":
        rpt("晚点名", 2)()
    else:
        bot.reply_to(message, "/trigger 命令使用格式\n"
                              "1. trigger 0：进行早打卡\n"
                              "2. trigger 1：进行午打卡\n"
                              "3. trigger 2 进行晚打卡\n"
                              "4. trigger 3：进行晚点名")


def schedule_continuous_run(interval=1):
    cease_continuous_run = threading.Event()

    class ScheduleThread(threading.Thread):
        @classmethod
        def run(cls):
            # debug
            # time.sleep(2)
            # rpt("测试打卡")()
            while not cease_continuous_run.is_set():
                schedule.run_pending()
                time.sleep(interval)

    continuous_thread = ScheduleThread()
    continuous_thread.start()


if __name__ == '__main__':
    load_config()
    schedule_continuous_run(10)
    bot.polling()
