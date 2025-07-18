import os
import json
import yaml
import logging
from datetime import datetime
from typing import List, Dict

import openai
from telegram import Update
from telegram.ext import (
    ApplicationBuilder,
    CommandHandler,
    ContextTypes,
)
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

logging.basicConfig(level=logging.INFO)

# Constants mirroring Go version
DEFAULT_LUNCH_TIME = "13:00"
DEFAULT_BRIEF_TIME = "20:00"
OPENAI_MODEL = os.getenv("OPENAI_MODEL", "gpt-4o")
WHITELIST_FILE = os.getenv("WHITELIST_FILE", "whitelist.json")

LUNCH_PROMPT = (
    "Подавай одну бизнес‑идею + примерный план из 4‑5 пунктов (коротко) + ссылки "
    "на релевантные ресурсы/репо/доки. Стиль панибратский, минимум воды."
)

BRIEF_PROMPT = (
    "Ты говоришь кратко, дерзко, панибратски.\n"
    "Заполни блоки:\n"
    "⚡ Микродействие (одно простое действие на сегодня)\n"
    "🧠 Тема дня (мини‑инсайт/мысль)\n"
    "💰 Что залутать (актив/идея)\n"
    "🏞️ Земля на присмотр (лоты в южном Подмосковье: Бутово, Щербинка, Подольск, Воскресенск), дай 1‑2 лота со ссылками.\n"
    "🪙 Альт дня (актуальная монета, линк CoinGecko)\n"
    "🚀 Пушка с ProductHunt (ссылка)\n"
    "Форматируй одним сообщением, без лишней воды."
)

BASE_PROMPT = ""


def load_whitelist() -> List[int]:
    try:
        with open(WHITELIST_FILE) as f:
            return json.load(f)
    except FileNotFoundError:
        return []
    except json.JSONDecodeError:
        return []


def save_whitelist(ids: List[int]) -> None:
    with open(WHITELIST_FILE, "w") as f:
        json.dump(ids, f)


def load_tasks() -> List[Dict[str, str]]:
    fn = os.getenv("TASKS_FILE")
    txt = os.getenv("TASKS_JSON")
    if fn:
        with open(fn) as f:
            data = yaml.safe_load(f)
        global BASE_PROMPT
        if isinstance(data, dict):
            BASE_PROMPT = data.get("base_prompt", "")
            return data.get("tasks", [])
        return data
    if txt:
        return json.loads(txt)

    lunch = os.getenv("LUNCH_TIME", DEFAULT_LUNCH_TIME)
    brief = os.getenv("BRIEF_TIME", DEFAULT_BRIEF_TIME)
    return [
        {"name": "lunch", "prompt": LUNCH_PROMPT, "time": lunch},
        {"name": "brief", "prompt": BRIEF_PROMPT, "time": brief},
    ]


def apply_template(prompt: str) -> str:
    vars_ = {
        "base_prompt": BASE_PROMPT,
        "date": datetime.utcnow().strftime("%Y-%m-%d"),
        "exchange_api": os.getenv("EXCHANGE_API", ""),
        "chart_path": os.getenv("CHART_PATH", ""),
    }
    for k, v in vars_.items():
        prompt = prompt.replace(f"{{{k}}}", v)
    return prompt


async def chat_completion(message: str) -> str:
    resp = await openai.ChatCompletion.acreate(
        model=OPENAI_MODEL,
        messages=[{"role": "user", "content": message}],
        temperature=0.9,
        max_tokens=600,
    )
    return resp.choices[0].message.content.strip()


async def system_completion(prompt: str) -> str:
    return await chat_completion(apply_template(prompt))


async def start_cmd(update: Update, context: ContextTypes.DEFAULT_TYPE):
    cid = update.effective_chat.id
    ids = load_whitelist()
    if cid not in ids:
        ids.append(cid)
        save_whitelist(ids)
    await update.message.reply_text("chat added")


async def ping_cmd(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await update.message.reply_text("pong")


async def chat_cmd(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not context.args:
        await update.message.reply_text("Usage: /chat <message>")
        return
    q = " ".join(context.args)
    try:
        text = await chat_completion(q)
    except Exception as exc:
        logging.warning("openai error: %s", exc)
        await update.message.reply_text("OpenAI error")
        return
    await update.message.reply_text(text)


async def task_cmd(update: Update, context: ContextTypes.DEFAULT_TYPE):
    name = " ".join(context.args)
    if not name:
        names = [t.get("name") for t in context.bot_data.get("tasks", []) if t.get("name")]
        await update.message.reply_text("\n".join(names) if names else "no tasks")
        return
    for t in context.bot_data.get("tasks", []):
        if t.get("name") == name:
            try:
                text = await system_completion(t["prompt"])
            except Exception as exc:
                logging.warning("openai error: %s", exc)
                await update.message.reply_text("OpenAI error")
                return
            await update.message.reply_text(text)
            return
    await update.message.reply_text("unknown task")


async def lunch_cmd(update: Update, context: ContextTypes.DEFAULT_TYPE):
    text = await system_completion(LUNCH_PROMPT)
    await update.message.reply_text(text)


async def brief_cmd(update: Update, context: ContextTypes.DEFAULT_TYPE):
    text = await system_completion(BRIEF_PROMPT)
    await update.message.reply_text(text)


def schedule_tasks(app, scheduler):
    tasks = context_tasks = load_tasks()
    app.bot_data["tasks"] = tasks
    for t in tasks:
        when = t.get("time", "00:00")
        cron = t.get("cron")
        prompt = t["prompt"]
        if cron:
            trigger = CronTrigger.from_crontab(cron)
        else:
            hh, mm = map(int, when.split(":"))
            trigger = CronTrigger(hour=hh, minute=mm)
        scheduler.add_job(
            lambda p=prompt: app.create_task(broadcast_task(app, p)),
            trigger,
        )


async def broadcast_task(app, prompt):
    try:
        text = await system_completion(prompt)
    except Exception as exc:
        logging.warning("openai error: %s", exc)
        return
    ids = load_whitelist()
    for cid in ids:
        try:
            await app.bot.send_message(cid, text)
        except Exception as exc:
            logging.warning("telegram send: %s", exc)


def main() -> None:
    token = os.getenv("TELEGRAM_TOKEN")
    openai.api_key = os.getenv("OPENAI_API_KEY")
    if not token or not openai.api_key:
        raise RuntimeError("TELEGRAM_TOKEN and OPENAI_API_KEY are required")

    application = ApplicationBuilder().token(token).build()

    scheduler = AsyncIOScheduler()
    scheduler.start()

    schedule_tasks(application, scheduler)

    application.add_handler(CommandHandler("start", start_cmd))
    application.add_handler(CommandHandler("ping", ping_cmd))
    application.add_handler(CommandHandler("chat", chat_cmd))
    application.add_handler(CommandHandler("task", task_cmd))
    application.add_handler(CommandHandler("lunch", lunch_cmd))
    application.add_handler(CommandHandler("brief", brief_cmd))

    application.run_polling()


if __name__ == "__main__":
    main()
