import os
import argparse
import requests
from anthropic import Anthropic, HUMAN_PROMPT, AI_PROMPT

parser = argparse.ArgumentParser()
parser.add_argument("--commit-messages")
parser.add_argument("--branch")
parser.add_argument("--repo")
args = parser.parse_args()
args = parser.parse_args()

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

client = Anthropic(api_key=os.environ.get("CLAUDE_API_KEY"))

response = client.completions.create(
    model="claude-2.0",
    prompt=f"{HUMAN_PROMPT}{prompt}{AI_PROMPT}",
    max_tokens_to_sample=1000,
    stop_sequences=[HUMAN_PROMPT]
)

review_text = response["completion"].strip()

url = f"https://api.github.com/repos/{args.repo}/issues"
res = requests.post(
    url,
    headers={
        "Authorization": f"token {os.environ.get('GITHUB_TOKEN')}",
        "Accept": "application/vnd.github+json"
    },
    json={
        "title": f"Claude commit review for branch {args.branch}",
        "body": review_text
    }
)

if res.status_code == 201:
    print("Commit review posted successfully")
else:
    print("Failed to post review:", res.text)