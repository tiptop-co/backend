import argparse
import requests
import os

# --- Аргументы ---
parser = argparse.ArgumentParser()
parser.add_argument("--commit-messages")
parser.add_argument("--branch")
parser.add_argument("--repo")
args = parser.parse_args()

print("Parsed args:", args)

with open(args.commit_messages) as f:
    commits = f.read()

prompt = f"""
Ты LLM-ревьюер коммитов.
Прочитай последние коммиты в ветке {args.branch} и дай краткое ревью:
- Логика изменений
- Стиль и понятность сообщений
- Возможные проблемы

Коммиты:
{commits}

Сделай ревью в короткой, структурированной форме.
"""

MISTRAL_KEY = os.environ.get("MISTRAL_KEY")
if not MISTRAL_KEY:
    raise ValueError("MISTRAL_KEY не найден в переменных окружения")

MISTRAL_URL = "https://api.mistral.ai/v1/chat/completions"

headers = {
    "Authorization": f"Bearer {MISTRAL_KEY}",
    "Content-Type": "application/json"
}

data = {
    "model": "mistral-small-latest",
    "messages": [
        {"role": "user", "content": prompt}
    ],
    "temperature": 0.2
}

resp = requests.post(MISTRAL_URL, headers=headers, json=data)
resp.raise_for_status()
review_text = resp.json()["choices"][0]["message"]["content"]

GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN")
if not GITHUB_TOKEN:
    raise ValueError("GITHUB_TOKEN не найден в переменных окружения")

url = f"https://api.github.com/repos/{args.repo}/issues"
res = requests.post(
    url,
    headers={
        "Authorization": f"token {GITHUB_TOKEN}",
        "Accept": "application/vnd.github+json"
    },
    json={
        "title": f"Commit review for branch {args.branch}",
        "body": review_text
    }
)

if res.status_code == 201:
    print("Commit review posted successfully")
else:
    print("Failed to post review:", res.text)