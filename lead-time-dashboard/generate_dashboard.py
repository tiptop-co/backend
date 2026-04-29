"""
Генерирует HTML-дашборд с визуализацией Lead Time по статусам.

Главный график — stacked horizontal bar chart (как на скриншоте из лекции):
- Ось Y: месяцы (2025-10, 2025-11, ...)
- Ось X: дни
- Сегменты: время в каждом статусе (цвета совпадают с легендой на скриншоте)

Использование:
    python generate_dashboard.py --input metrics.json --output dashboard.html
"""

import argparse
import json
import sys
from datetime import datetime
from string import Template


HTML_TEMPLATE = Template(r"""<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>$title — Lead Time Dashboard</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
<style>
*, *::before, *::after { box-sizing: border-box; }
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background: #0d1117;
  color: #c9d1d9;
}
header {
  background: #161b22;
  border-bottom: 1px solid #30363d;
  padding: 18px 32px;
  display: flex;
  align-items: center;
  gap: 14px;
}
header h1 { margin: 0; font-size: 1.2rem; color: #f0f6fc; }
.tag {
  background: #1f6feb;
  color: #fff;
  font-size: .72rem;
  padding: 2px 10px;
  border-radius: 20px;
  font-weight: 600;
}
header .upd { margin-left: auto; font-size: .78rem; color: #8b949e; }
.container { max-width: 1280px; margin: 0 auto; padding: 28px 24px; }
.kpi-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 14px;
  margin-bottom: 28px;
}
.kpi {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 18px 20px;
}
.kpi .lbl { font-size: .76rem; color: #8b949e; margin-bottom: 6px; }
.kpi .val { font-size: 2rem; font-weight: 700; color: #58a6ff; line-height: 1; }
.kpi .sub { font-size: .76rem; color: #8b949e; margin-top: 4px; }
.panel {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 22px 24px;
  margin-bottom: 22px;
}
.panel h2 {
  margin: 0 0 18px;
  font-size: .95rem;
  color: #f0f6fc;
  padding-bottom: 12px;
  border-bottom: 1px solid #30363d;
}
.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin-top: 14px;
}
.legend-item { display: flex; align-items: center; gap: 6px; font-size: .78rem; }
.legend-dot { width: 12px; height: 12px; border-radius: 2px; flex-shrink: 0; }
.two-col { display: grid; grid-template-columns: 1fr 1fr; gap: 22px; }
@media (max-width: 800px) { .two-col { grid-template-columns: 1fr; } }
.chart-wrap { position: relative; height: 260px; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; font-size: .82rem; }
th {
  background: #0d1117; color: #8b949e; font-weight: 500;
  text-align: left; padding: 9px 12px;
  border-bottom: 1px solid #30363d; white-space: nowrap;
}
td { padding: 9px 12px; border-bottom: 1px solid #21262d; vertical-align: middle; }
tr:hover td { background: #1c2128; }
.chip {
  display: inline-block; border-radius: 4px;
  padding: 1px 8px; font-size: .74rem; white-space: nowrap;
}
footer { text-align: center; padding: 28px; color: #8b949e; font-size: .76rem; }
</style>
</head>
<body>

<header>
  <svg width="22" height="22" viewBox="0 0 16 16" fill="#58a6ff">
    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38
             0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13
             -.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66
             .07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15
             -.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27
             .68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12
             .51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48
             0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
  </svg>
  <h1>$title — Lead Time Dashboard</h1>
  <span class="tag">Скорость разработки</span>
  <span class="upd">Обновлено: $generated_at</span>
</header>

<div class="container">

  <div class="kpi-row">
    <div class="kpi">
      <div class="lbl">Всего задач</div>
      <div class="val">$total_items</div>
    </div>
    <div class="kpi">
      <div class="lbl">Среднее Lead Time</div>
      <div class="val">$avg_lead_days</div>
      <div class="sub">дней</div>
    </div>
    <div class="kpi">
      <div class="lbl">Месяцев данных</div>
      <div class="val">$months_count</div>
    </div>
    <div class="kpi">
      <div class="lbl">Статусов</div>
      <div class="val">$statuses_count</div>
    </div>
  </div>

  <!-- Главный график: Stacked Bar — Lead Time по статусам (как в лекции) -->
  <div class="panel">
    <h2>📊 Lead Time — среднее время в каждом статусе по месяцам (дни)</h2>
    <div id="mainWrap" style="position:relative">
      <canvas id="leadTimeChart"></canvas>
    </div>
    <div class="legend" id="legend"></div>
  </div>

  <div class="two-col">
    <div class="panel">
      <h2>📦 Количество задач по месяцам</h2>
      <div class="chart-wrap">
        <canvas id="countChart"></canvas>
      </div>
    </div>
    <div class="panel">
      <h2>🥧 Суммарное распределение по статусам</h2>
      <div class="chart-wrap">
        <canvas id="pieChart"></canvas>
      </div>
    </div>
  </div>

  <div class="panel">
    <h2>📋 Детали по задачам</h2>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>#</th>
            <th>Название</th>
            <th>Месяц</th>
            <th>Lead Time (дн.)</th>
            $status_th
          </tr>
        </thead>
        <tbody>
          $rows
        </tbody>
      </table>
    </div>
  </div>

</div>

<footer>
  Данные из GitHub Projects · Метрика Lead Time по статусам · Автогенерация
</footer>

<script>
const DATA = $data_json;

const MONTHS = Object.keys(DATA.by_month).sort();
const STATUSES = DATA.statuses;
const COLORS = DATA.status_colors;

// Высота главного графика пропорциональна числу месяцев
const barH = 38;
const mainH = Math.max(200, MONTHS.length * (barH + 8) + 60);
const wrap = document.getElementById("mainWrap");
wrap.style.height = mainH + "px";

// Легенда
const legend = document.getElementById("legend");
STATUSES.forEach(s => {
  const item = document.createElement("div");
  item.className = "legend-item";
  item.innerHTML = '<div class="legend-dot" style="background:' + COLORS[s] + '"></div><span>' + s + '</span>';
  legend.appendChild(item);
});

// Датасеты для stacked bar
const datasets = STATUSES.map(status => ({
  label: status,
  data: MONTHS.map(m => DATA.by_month[m].avg_status_days[status] || 0),
  backgroundColor: COLORS[status],
  borderColor: "transparent",
  borderWidth: 0,
  borderRadius: 2,
}));

// Главный график — горизонтальный stacked bar (как на скриншоте)
new Chart(document.getElementById("leadTimeChart"), {
  type: "bar",
  data: { labels: MONTHS, datasets: datasets },
  options: {
    indexAxis: "y",
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: function(ctx) { return " " + ctx.dataset.label + ": " + ctx.raw.toFixed(1) + " дн."; }
        }
      }
    },
    scales: {
      x: {
        stacked: true,
        ticks: { color: "#8b949e", callback: function(v) { return v + " дн."; } },
        grid: { color: "#21262d" }
      },
      y: {
        stacked: true,
        ticks: { color: "#c9d1d9" },
        grid: { display: false }
      }
    }
  }
});

// Количество задач по месяцам
new Chart(document.getElementById("countChart"), {
  type: "bar",
  data: {
    labels: MONTHS,
    datasets: [{
      label: "Задач",
      data: MONTHS.map(function(m) { return DATA.by_month[m].count; }),
      backgroundColor: "rgba(88,166,255,0.7)",
      borderColor: "#58a6ff",
      borderWidth: 1,
      borderRadius: 4
    }]
  },
  options: {
    responsive: true, maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { ticks: { color: "#8b949e" }, grid: { color: "#21262d" } },
      y: { ticks: { color: "#8b949e" }, grid: { color: "#21262d" } }
    }
  }
});

// Doughnut — суммарное распределение
const pieTotals = STATUSES.map(function(s) {
  return MONTHS.reduce(function(acc, m) { return acc + (DATA.by_month[m].avg_status_days[s] || 0); }, 0);
});
const pieLbls = STATUSES.filter(function(_, i) { return pieTotals[i] > 0; });
const pieVals = pieTotals.filter(function(v) { return v > 0; });

new Chart(document.getElementById("pieChart"), {
  type: "doughnut",
  data: {
    labels: pieLbls,
    datasets: [{
      data: pieVals,
      backgroundColor: pieLbls.map(function(s) { return COLORS[s]; }),
      borderColor: "#161b22",
      borderWidth: 2
    }]
  },
  options: {
    responsive: true, maintainAspectRatio: false,
    plugins: {
      legend: {
        position: "right",
        labels: { color: "#c9d1d9", font: { size: 11 }, boxWidth: 12 }
      },
      tooltip: {
        callbacks: {
          label: function(ctx) { return " " + ctx.label + ": " + ctx.raw.toFixed(1) + " дн."; }
        }
      }
    }
  }
});
</script>
</body>
</html>
""")


def make_status_th(statuses):
    return "".join(f"<th>{s}</th>" for s in statuses)


def make_rows(items, statuses, status_colors):
    rows = []
    for item in sorted(items, key=lambda i: i.get("month", ""), reverse=True)[:300]:
        number = item.get("number", "")
        title = (item.get("title", "")[:60]).replace("<", "&lt;").replace(">", "&gt;")
        month = item.get("month", "")
        lead = round(item.get("total_lead_time_h", 0) / 24, 1)

        cells = ""
        for s in statuses:
            h = item.get("status_hours", {}).get(s, 0)
            d = round(h / 24, 1)
            if d > 0:
                color = status_colors.get(s, "#555")
                cells += f'<td><span class="chip" style="background:{color}22;color:{color}">{d}д</span></td>'
            else:
                cells += "<td style='color:#333'>—</td>"

        rows.append(f"""<tr>
          <td style="color:#8b949e">#{number}</td>
          <td style="max-width:260px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="{title}">{title}</td>
          <td style="color:#8b949e">{month}</td>
          <td style="color:#58a6ff;font-weight:600">{lead}</td>
          {cells}
        </tr>""")
    return "\n".join(rows)


def main():
    parser = argparse.ArgumentParser(description="Генерация Lead Time дашборда")
    parser.add_argument("--input", default="metrics.json")
    parser.add_argument("--output", default="dashboard.html")
    args = parser.parse_args()

    try:
        with open(args.input, encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"Файл {args.input} не найден. Сначала запустите collect_status_time.py", file=sys.stderr)
        sys.exit(1)

    statuses = data["statuses"]
    status_colors = data["status_colors"]
    by_month = data["by_month"]
    items = data["items"]

    total_items = len(items)
    months_count = len(by_month)
    all_lead = [i["total_lead_time_h"] for i in items if i.get("total_lead_time_h", 0) > 0]
    avg_lead_days = round(sum(all_lead) / len(all_lead) / 24, 1) if all_lead else 0

    generated_dt = datetime.fromisoformat(data["generated_at"].replace("Z", "+00:00"))
    generated_str = generated_dt.strftime("%d.%m.%Y %H:%M UTC")

    html = HTML_TEMPLATE.substitute(
        title=data["project_title"],
        generated_at=generated_str,
        total_items=total_items,
        avg_lead_days=avg_lead_days,
        months_count=months_count,
        statuses_count=len(statuses),
        status_th=make_status_th(statuses),
        rows=make_rows(items, statuses, status_colors),
        data_json=json.dumps(data, ensure_ascii=False),
    )

    with open(args.output, "w", encoding="utf-8") as f:
        f.write(html)

    print(f"Дашборд сохранён: {args.output}")


if __name__ == "__main__":
    main()