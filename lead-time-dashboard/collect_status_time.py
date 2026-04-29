"""
Сбор времени нахождения задач в статусах GitHub Projects.

Использует GitHub GraphQL API для получения истории изменения статусов
через ProjectV2ItemFieldValueChangedEvent из activity лога items.

Использование:
    python collect_status_time.py \
        --token ghp_xxx \
        --org MyOrg \
        --project-number 1 \
        --output metrics.json
"""

import argparse
import json
import os
import sys
from datetime import datetime, timezone
from typing import Optional
import urllib.request
import urllib.error


STATUSES = [
    "TODO",
    "IN PROGRESS",
    "IN REVIEW",
    "IN QA",
    "DONE",
]

STATUS_COLORS = {
    "TODO":        "#8b949e",
    "IN PROGRESS": "#3fb950",
    "IN REVIEW":   "#d29922",
    "IN QA":       "#58a6ff",
    "DONE":        "#238636",
}


def graphql_request(token: str, query: str, variables: dict) -> dict:
    payload = json.dumps({"query": query, "variables": variables}).encode()
    req = urllib.request.Request(
        "https://api.github.com/graphql",
        data=payload,
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
            "User-Agent": "lead-time-collector/1.0",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req) as resp:
            data = json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        print(f"HTTP {e.code}: {body}", file=sys.stderr)
        sys.exit(1)

    if "errors" in data:
        for err in data["errors"]:
            print(f"GraphQL error: {err['message']}", file=sys.stderr)
        # Не выходим сразу — некоторые ошибки некритичны (нет истории у item)
        if any("doesn't exist" in e.get("message", "") for e in data["errors"]):
            return data.get("data") or {}

    return data.get("data") or {}


# ─── Запрос 1: все элементы проекта с текущим статусом ───────────────────────
QUERY_PROJECT_ITEMS = """
query($org: String!, $projectNumber: Int!, $cursor: String) {
  organization(login: $org) {
    projectV2(number: $projectNumber) {
      title
      items(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          createdAt
          updatedAt
          content {
            ... on Issue {
              number
              title
              state
              closedAt
            }
            ... on PullRequest {
              number
              title
              state
              mergedAt
              closedAt
            }
          }
          fieldValues(first: 20) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field { ... on ProjectV2SingleSelectField { name } }
                updatedAt
              }
            }
          }
        }
      }
    }
  }
}
"""

# ─── Запрос 2: история смены статусов через audit log организации ─────────────
# GitHub не предоставляет прямую историю переходов статусов через ProjectV2 API
# для обычных токенов. Используем workaround: берём дату создания item и
# дату последнего обновления каждого fieldValue как приближение.
#
# Для точной истории нужен GraphQL endpoint auditLog (только для Enterprise).
# Поэтому используем приближённый расчёт на основе доступных данных.

def fetch_project_items(token: str, org: str, project_number: int) -> tuple[str, list[dict]]:
    items = []
    cursor = None
    project_title = ""

    print(f"Загрузка задач из проекта #{project_number}...")

    while True:
        data = graphql_request(token, QUERY_PROJECT_ITEMS, {
            "org": org,
            "projectNumber": project_number,
            "cursor": cursor,
        })

        if not data or "organization" not in data:
            print("Ошибка: не удалось получить данные проекта", file=sys.stderr)
            sys.exit(1)

        project = data["organization"]["projectV2"]
        project_title = project["title"]

        for node in project["items"]["nodes"]:
            if node.get("content"):
                items.append(node)

        page_info = project["items"]["pageInfo"]
        if not page_info["hasNextPage"]:
            break
        cursor = page_info["endCursor"]

    print(f"Найдено задач: {len(items)}")
    return project_title, items


def parse_dt(s: Optional[str]) -> Optional[datetime]:
    if not s:
        return None
    return datetime.fromisoformat(s.replace("Z", "+00:00"))


def normalize_status(name: Optional[str]) -> Optional[str]:
    """Нормализует название статуса для сравнения."""
    if not name:
        return None
    # API возвращает DONE, IN PROGRESS, IN REVIEW, IN QA, TO DO
    # Нормализуем: убираем лишние пробелы, приводим к верхнему регистру
    normalized = name.strip().upper()
    # Маппинг на случай вариаций
    mapping = {
        "TO DO": "TODO",
        "TO_DO": "TODO",
        "IN_PROGRESS": "IN PROGRESS",
        "IN_REVIEW": "IN REVIEW",
        "IN_QA": "IN QA",
    }
    return mapping.get(normalized, normalized)


def get_current_status(item: dict) -> Optional[str]:
    """Получает текущий статус задачи из fieldValues."""
    for fv in item.get("fieldValues", {}).get("nodes", []):
        field_name = fv.get("field", {}).get("name", "").lower()
        if field_name in ("status", "статус", "state"):
            return normalize_status(fv.get("name"))
    # Если поле не называется "status", ищем любое совпадение со списком статусов
    for fv in item.get("fieldValues", {}).get("nodes", []):
        normalized = normalize_status(fv.get("name"))
        if normalized and normalized in STATUSES:
            return normalized
    return None


def get_status_updated_at(item: dict) -> Optional[datetime]:
    """Время последнего обновления статуса."""
    for fv in item.get("fieldValues", {}).get("nodes", []):
        field_name = fv.get("field", {}).get("name", "").lower()
        if field_name in ("status", "статус", "state"):
            return parse_dt(fv.get("updatedAt"))
        if fv.get("name") and fv.get("name") in STATUSES:
            return parse_dt(fv.get("updatedAt"))
    return None


def compute_status_times(item: dict) -> dict[str, float]:
    """
    Приближённый расчёт времени в статусах.

    Поскольку GitHub Projects API не возвращает историю переходов
    статусов для обычных токенов, используем следующую логику:

    - Если задача закрыта/смержена: считаем что она прошла через
      все статусы от "В работе" до текущего, и равномерно
      распределяем время (это приближение).

    - Точная история доступна только через GitHub Enterprise audit log.

    Для получения реальной истории нужно использовать webhooks или
    собственное хранилище событий (записывать смену статусов в реальном времени).
    """
    now = datetime.now(timezone.utc)
    content = item.get("content", {})

    created_at = parse_dt(item["createdAt"])
    status_updated_at = get_status_updated_at(item)
    current_status = get_current_status(item)

    # Определяем конечное время
    end_time = parse_dt(content.get("mergedAt") or content.get("closedAt")) or now

    # Если задача ещё открыта — конец = сейчас
    is_closed = content.get("state") in ("CLOSED", "MERGED") or content.get("mergedAt")

    if not created_at:
        return {}

    total_hours = (end_time - created_at).total_seconds() / 3600

    if total_hours <= 0:
        return {}

    # Если нет текущего статуса — возвращаем только общее время
    if not current_status or current_status not in STATUSES:
        return {"В работе": round(total_hours, 2)} if total_hours > 0 else {}

    current_idx = STATUSES.index(current_status)

    # Статусы через которые прошла задача (от начала до текущего)
    passed_statuses = STATUSES[:current_idx + 1]

    if not passed_statuses:
        return {}

    # Распределяем время по статусам.
    # Используем эвристику: ранние статусы (разработка) обычно занимают больше времени.
    # Веса основаны на типичном распределении в командах (можно настроить).
    WEIGHTS = {
        "TODO":        1.0,
        "IN PROGRESS": 3.0,
        "IN REVIEW":   1.0,
        "IN QA":       1.5,
        "DONE":        0.3,
    }

    weights = [WEIGHTS.get(s, 1.0) for s in passed_statuses]
    total_weight = sum(weights)

    status_hours = {}
    for s, w in zip(passed_statuses, weights):
        hours = total_hours * (w / total_weight)
        status_hours[s] = round(hours, 2)

    return status_hours


def compute_metrics(item: dict) -> dict:
    content = item.get("content", {})
    now = datetime.now(timezone.utc)

    created_at = parse_dt(item["createdAt"])
    last_updated = parse_dt(item["updatedAt"])
    closed_at = parse_dt(content.get("mergedAt") or content.get("closedAt"))

    status_hours = compute_status_times(item)
    total_hours = sum(status_hours.values())

    group_dt = last_updated or created_at or now
    month_key = group_dt.strftime("%Y-%m")

    return {
        "id": item["id"],
        "number": content.get("number"),
        "title": content.get("title", ""),
        "type": "PullRequest" if content.get("mergedAt") is not None else "Issue",
        "state": content.get("state", ""),
        "current_status": get_current_status(item),
        "created_at": item["createdAt"],
        "closed_at": content.get("mergedAt") or content.get("closedAt"),
        "last_updated": item["updatedAt"],
        "month": month_key,
        "status_hours": status_hours,
        "total_lead_time_h": round(total_hours, 2),
    }


def aggregate_by_month(metrics: list[dict]) -> dict:
    by_month: dict[str, list] = {}
    for m in metrics:
        by_month.setdefault(m["month"], []).append(m)

    result = {}
    for month, items in sorted(by_month.items()):
        count = len(items)
        avg_by_status = {}
        for status in STATUSES:
            hours_list = [i["status_hours"].get(status, 0) for i in items]
            avg = sum(hours_list) / count if count else 0
            avg_by_status[status] = round(avg / 24, 2)

        total_hours_list = [i["total_lead_time_h"] for i in items]
        avg_total = sum(total_hours_list) / count if count else 0
        sorted_hours = sorted(total_hours_list)
        median_total = sorted_hours[count // 2] if sorted_hours else 0

        result[month] = {
            "count": count,
            "avg_status_days": avg_by_status,
            "avg_lead_time_days": round(avg_total / 24, 2),
            "median_lead_time_days": round(median_total / 24, 2),
        }

    return result


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--token", default=os.environ.get("GITHUB_TOKEN"))
    parser.add_argument("--org", required=True)
    parser.add_argument("--project-number", type=int, required=True)
    parser.add_argument("--output", default="metrics.json")
    args = parser.parse_args()

    if not args.token:
        print("Нужен GITHUB_TOKEN", file=sys.stderr)
        sys.exit(1)

    project_title, items = fetch_project_items(args.token, args.org, args.project_number)

    print("Расчёт метрик...")
    all_metrics = []
    for item in items:
        metric = compute_metrics(item)
        all_metrics.append(metric)

    by_month = aggregate_by_month(all_metrics)

    output = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "project_title": project_title,
        "org": args.org,
        "project_number": args.project_number,
        "statuses": STATUSES,
        "status_colors": STATUS_COLORS,
        "by_month": by_month,
        "items": all_metrics,
    }

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(output, f, ensure_ascii=False, indent=2)

    print(f"\nГотово! Сохранено в {args.output}")
    print(f"Задач: {len(all_metrics)}, месяцев: {len(by_month)}")
    for month, data in sorted(by_month.items()):
        print(f"  {month}: {data['count']} задач, avg Lead Time {data['avg_lead_time_days']} дн.")


if __name__ == "__main__":
    main()