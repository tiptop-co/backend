import requests
import os

with open("commit_messages.txt") as f:
    commits = f.read()

with open("pr_description.txt") as f:
    pr_text = f.read()

with open ("pr_diff.txt") as f:
    diffs = f.read()

prompt = f"""
Ты LLM-ревьюер pull request.

Описание PR:
{pr_text}

Коммиты:
{commits}

Изменения:
{diffs}

Сделай краткое ревью:
- соответствует ли описание изменениям
- логика изменений
- проблемы
- что стоит проверить
"""

MISTRAL_KEY = os.environ.get("MISTRAL_KEY")
GITHUB_TOKEN = os.environ.get("TOKEN")
PR_NUMBER = os.environ.get("PR_NUMBER")

headers = {
    "Authorization": f"Bearer {MISTRAL_KEY}",
    "Content-Type": "application/json"
}

data = {
    "model": "mistral-small-latest",
    "messages": [{"role": "user", "content": prompt}],
    "temperature": 0.2
}

resp = requests.post(
    "https://api.mistral.ai/v1/chat/completions",
    headers=headers,
    json=data
)

resp.raise_for_status()
review_text = resp.json()["choices"][0]["message"]["content"]

repo = os.environ.get("GITHUB_REPOSITORY")

comment_url = f"https://api.github.com/repos/{repo}/issues/{PR_NUMBER}/comments"

res = requests.post(
    comment_url,
    headers={
        "Authorization": f"token {GITHUB_TOKEN}",
        "Accept": "application/vnd.github+json"
    },
    json={"body": f"Mistral PR Review\n\n{review_text}"}
)

print(res.status_code, res.text)