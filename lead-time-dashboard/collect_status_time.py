"""
Сбор времени нахождения задач в статусах GitHub Projects.

Использует GitHub GraphQL API для получения истории изменения статусов
(ProjectV2ItemFieldValueChangedEvent) и рассчитывает время, проведённое
задачей в каждом статусе.

Отслеживаемые статусы (из лекции):
  В работе → Ревью → Подготовка к тестированию → На проверку →
  Тестируется → Ожидает приёмку → Приёмка на платформе → Готово → Требуется выгрузка

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


# ─── Статусы в правильном порядке ────────────────────────────────────────────
# Соответствуют легенде на графике из лекции
STATUSES = [
    "В работе",
    "Ревью",
    "Подготовка к тестированию",
    "На проверку",
    "Тестируется",
    "Ожидает приёмку",
    "Приёмка на платформе",
    "Готово",
    "Требуется выгрузка",
]

STATUS_COLORS = {
    "В работе":                  "#3fb950",
    "Ревью":                     "#d29922",
    "Подготовка к тестированию": "#58a6ff",
    "На проверку":               "#e3672a",
    "Тестируется":               "#f85149",
    "Ожидает приёмку":           "#54aeff",
    "Приёмка на платформе":      "#bc8cff",
    "Готово":                    "#8b949e",
    "Требуется выгрузка":        "#39d353",
}


def graphql_request(token: str, query: str, variables: dict) -> dict:
    """Выполняет GraphQL запрос к GitHub API."""
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
        sys.exit(1)

    return data["data"]


# ─── GraphQL запросы ──────────────────────────────────────────────────────────

# Получаем все задачи проекта с их текущим статусом и временем создания
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

# История изменений статусов конкретного элемента
# ProjectV2ItemFieldValueChangedEvent содержит когда и на что изменился статус
QUERY_ITEM_ACTIVITY = """
query($org: String!, $projectNumber: Int!, $itemId: ID!, $cursor: String) {
  organization(login: $org) {
    projectV2(number: $projectNumber) {
      item(id: $itemId) {
        id
        activityItems(first: 100, after: $cursor) {
          pageInfo { hasNextPage endCursor }
          nodes {
            ... on ProjectV2ItemFieldValueChangedEvent {
              createdAt
              previousValue {
                ... on ProjectV2ItemFieldSingleSelectValue {
                  name
                }
              }
              newValue {
                ... on ProjectV2ItemFieldSingleSelectValue {
                  name
                }
              }
            }
          }
        }
      }
    }
  }
}
"""


def fetch_project_items(token: str, org: str, project_number: int) -> tuple[str, list[dict]]:
    """Загружает все задачи из проекта с пагинацией."""
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


def fetch_item_status_history(token: str, org: str, project_number: int, item_id: str) -> list[dict]:
    """
    Получает историю изменений статусов для одной задачи.
    Возвращает список событий {created_at, from_status, to_status}.
    """
    events = []
    cursor = None

    while True:
        data = graphql_request(token, QUERY_ITEM_ACTIVITY, {
            "org": org,
            "projectNumber": project_number,
            "itemId": item_id,
            "cursor": cursor,
        })

        item_data = data["organization"]["projectV2"].get("item")
        if not item_data:
            break

        activity = item_data.get("activityItems", {})

        for node in activity.get("nodes", []):
            # Берём только события изменения статуса (single select field)
            if not node.get("createdAt"):
                continue
            prev = node.get("previousValue", {})
            new = node.get("newValue", {})
            prev_name = prev.get("name") if prev else None
            new_name = new.get("name") if new else None

            if new_name:  # есть новый статус
                events.append({
                    "created_at": node["createdAt"],
                    "from_status": prev_name,
                    "to_status": new_name,
                })

        page_info = activity.get("pageInfo", {})
        if not page_info.get("hasNextPage"):
            break
        cursor = page_info["endCursor"]

    # Сортируем по времени
    events.sort(key=lambda e: e["created_at"])
    return events


def parse_dt(s: Optional[str]) -> Optional[datetime]:
    if not s:
        return None
    return datetime.fromisoformat(s.replace("Z", "+00:00"))


def compute_status_times(item: dict, history: list[dict]) -> dict:
    """
    Рассчитывает время нахождения задачи в каждом статусе (в часах).

    Алгоритм:
    1. Строим временну́ю шкалу переходов статусов
    2. Для каждого статуса считаем суммарное время нахождения в нём
    3. Для последнего статуса используем текущее время (или время закрытия)
    """
    now = datetime.now(timezone.utc)
    content = item.get("content", {})

    # Определяем «конечное» время задачи
    end_time = None
    if content.get("mergedAt"):
        end_time = parse_dt(content["mergedAt"])
    elif content.get("closedAt"):
        end_time = parse_dt(content["closedAt"])
    else:
        end_time = now

    # Определяем текущий статус из fieldValues
    current_status = None
    for fv in item.get("fieldValues", {}).get("nodes", []):
        if fv.get("field", {}).get("name", "").lower() in ("status", "статус"):
            current_status = fv.get("name")

    # Если нет истории — задача может быть в одном статусе с момента создания
    if not history:
        status_hours: dict[str, float] = {}
        if current_status and current_status in STATUSES:
            created = parse_dt(item["createdAt"])
            if created:
                hours = (end_time - created).total_seconds() / 3600
                status_hours[current_status] = round(max(0, hours), 2)
        return status_hours

    # Строим список (время_входа, статус) по событиям
    timeline = []
    item_created = parse_dt(item["createdAt"])

    for event in history:
        event_time = parse_dt(event["created_at"])
        if event_time is None:
            continue

        # Первое событие: до него задача была в from_status с момента создания
        if not timeline and event.get("from_status"):
            timeline.append((item_created or event_time, event["from_status"]))

        timeline.append((event_time, event["to_status"]))

    # Если событий нет — используем created_at + current_status
    if not timeline and current_status:
        timeline.append((item_created or now, current_status))

    # Считаем время в каждом статусе
    status_hours: dict[str, float] = {}

    for i, (enter_time, status) in enumerate(timeline):
        if status not in STATUSES:
            continue
        # Время выхода — вход в следующий статус или конечное время
        if i + 1 < len(timeline):
            exit_time = timeline[i + 1][0]
        else:
            exit_time = end_time

        if enter_time and exit_time and exit_time > enter_time:
            hours = (exit_time - enter_time).total_seconds() / 3600
            status_hours[status] = status_hours.get(status, 0) + hours

    # Округляем
    return {s: round(h, 2) for s, h in status_hours.items()}


def compute_metrics(item: dict, status_hours: dict[str, float]) -> dict:
    """Собирает полную метрику по задаче."""
    content = item.get("content", {})
    now = datetime.now(timezone.utc)

    created_at = parse_dt(item["createdAt"])
    closed_at = parse_dt(content.get("mergedAt") or content.get("closedAt"))
    last_updated = parse_dt(item["updatedAt"])

    # Общий lead time = сумма всех статусов
    total_hours = sum(status_hours.values())

    # Группировка по месяцу последнего изменения
    group_dt = last_updated or created_at or now
    month_key = group_dt.strftime("%Y-%m")

    return {
        "id": item["id"],
        "number": content.get("number"),
        "title": content.get("title", ""),
        "type": "PullRequest" if "mergedAt" in content else "Issue",
        "state": content.get("state", ""),
        "created_at": item["createdAt"],
        "closed_at": (content.get("mergedAt") or content.get("closedAt")),
        "last_updated": item["updatedAt"],
        "month": month_key,
        "status_hours": status_hours,
        "total_lead_time_h": round(total_hours, 2),
    }


def aggregate_by_month(metrics: list[dict]) -> dict:
    """
    Группирует по месяцам и считает среднее время в каждом статусе.
    Это данные для stacked bar chart как на скриншоте из лекции.
    """
    by_month: dict[str, list] = {}
    for m in metrics:
        by_month.setdefault(m["month"], []).append(m)

    result = {}
    for month, items in sorted(by_month.items()):
        count = len(items)
        # Среднее время в каждом статусе
        avg_by_status = {}
        for status in STATUSES:
            hours_list = [i["status_hours"].get(status, 0) for i in items]
            avg = sum(hours_list) / count if count else 0
            avg_by_status[status] = round(avg / 24, 2)  # переводим в дни

        total_items_hours = [i["total_lead_time_h"] for i in items]
        avg_total = sum(total_items_hours) / count if count else 0

        result[month] = {
            "count": count,
            "avg_status_days": avg_by_status,
            "avg_lead_time_days": round(avg_total / 24, 2),
            "median_lead_time_days": round(
                sorted(total_items_hours)[count // 2] / 24, 2
            ) if total_items_hours else 0,
        }

    return result


def main():
    parser = argparse.ArgumentParser(
        description="Сбор времени нахождения задач в статусах GitHub Projects"
    )
    parser.add_argument("--token", default=os.environ.get("GITHUB_TOKEN"))
    parser.add_argument("--org", required=True, help="Организация GitHub")
    parser.add_argument("--project-number", type=int, required=True, help="Номер проекта")
    parser.add_argument("--output", default="metrics.json")
    parser.add_argument(
        "--no-history",
        action="store_true",
        help="Не запрашивать историю событий (быстрее, но менее точно)",
    )
    args = parser.parse_args()

    if not args.token:
        print("Нужен GITHUB_TOKEN", file=sys.stderr)
        sys.exit(1)

    project_title, items = fetch_project_items(args.token, args.org, args.project_number)

    all_metrics = []
    for i, item in enumerate(items):
        item_id = item["id"]
        title = item.get("content", {}).get("title", "")[:50]
        print(f"  [{i+1}/{len(items)}] {title}...")

        if args.no_history:
            history = []
        else:
            history = fetch_item_status_history(args.token, args.org, args.project_number, item_id)

        status_hours = compute_status_times(item, history)
        metric = compute_metrics(item, status_hours)
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