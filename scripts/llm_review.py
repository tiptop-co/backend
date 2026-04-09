import openai
import argparse
import requests
import os

parser = argparse.ArgumentParser()
parser.add_argument("--commit-messages")
parser.add_argument("--branch")
parser.add_argument("--repo")
args = parser.parse_args()

# Читаем коммиты
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

# Запрос к GPT
response = openai.ChatCompletion.create(
    model="gpt-4.1-mini",
    messages=[{"role": "user", "content": prompt}],
    temperature=0.2
)

review_text = response['choices'][0]['message']['content']

# Публикуем комментарий в PR (для теста можно создать новый Issue с названием ветки)
url = f"https://api.github.com/repos/{args.repo}/issues"
res = requests.post(url,
    headers={
        "Authorization": f"token {os.environ.get('GITHUB_TOKEN')}",
        "Accept": "application/vnd.github+json"
    },
    json={
        "title": f"LLM commit review for branch {args.branch}",
        "body": review_text
    }
)

if res.status_code == 201:
    print("Commit review posted successfully")
else:
    print("Failed to post review:", res.text)