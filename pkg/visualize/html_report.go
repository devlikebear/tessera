package visualize

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
)

func WriteHTMLReportFile(eventsPath, outPath string) error {
	events, err := ReadEventsFile(eventsPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return WriteHTMLReport(file, Project(events))
}

func WriteHTMLReport(w io.Writer, projection Projection) error {
	return reportTemplate.Execute(w, projection)
}

var reportTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Tessera Run Report</title>
<style>
:root {
  color-scheme: light;
  --ink: #17202a;
  --muted: #64748b;
  --line: #d7dee8;
  --panel: #f8fafc;
  --accent: #0f766e;
  --accent-soft: #ccfbf1;
  --danger: #b91c1c;
  --ok: #15803d;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  color: var(--ink);
  background: #ffffff;
}
main {
  max-width: 1120px;
  margin: 0 auto;
  padding: 32px 20px 48px;
}
header {
  border-bottom: 1px solid var(--line);
  padding-bottom: 20px;
  margin-bottom: 24px;
}
h1 {
  font-size: 28px;
  line-height: 1.2;
  margin: 0 0 8px;
  letter-spacing: 0;
}
h2 {
  font-size: 17px;
  line-height: 1.3;
  margin: 28px 0 12px;
  letter-spacing: 0;
}
.summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin: 18px 0 0;
}
.metric {
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 12px;
  background: var(--panel);
}
.label {
  color: var(--muted);
  font-size: 12px;
  text-transform: uppercase;
}
.value {
  margin-top: 6px;
  font-size: 20px;
  font-weight: 650;
}
table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid var(--line);
}
th, td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--line);
  text-align: left;
  vertical-align: top;
  font-size: 14px;
}
th {
  background: var(--panel);
  color: #334155;
  font-weight: 650;
}
tr:last-child td { border-bottom: 0; }
.status {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 650;
}
.status.failed {
  background: #fee2e2;
  color: var(--danger);
}
.status.succeeded {
  background: #dcfce7;
  color: var(--ok);
}
.timeline {
  display: grid;
  gap: 8px;
}
.event {
  border-left: 3px solid var(--accent);
  padding: 8px 10px;
  background: var(--panel);
}
.event-meta {
  color: var(--muted);
  font-size: 12px;
}
</style>
</head>
<body>
<main>
  <header>
    <h1>Tessera Run Report</h1>
    <div>{{.RunID}}</div>
    <div class="summary">
      <div class="metric"><div class="label">Closure</div><div class="value">{{.Closure}}</div></div>
      <div class="metric"><div class="label">Events</div><div class="value">{{.EventCount}}</div></div>
      <div class="metric"><div class="label">Tasks</div><div class="value">{{len .Tasks}}</div></div>
      <div class="metric"><div class="label">Roles</div><div class="value">{{len .Roles}}</div></div>
    </div>
  </header>

  <section>
    <h2>Task Graph</h2>
    <table>
      <thead><tr><th>Task</th><th>Role</th><th>Status</th><th>Events</th><th>Last Message</th></tr></thead>
      <tbody>
      {{range .Tasks}}
        <tr>
          <td>{{.ID}}</td>
          <td>{{.Role}}</td>
          <td><span class="status {{.Status}}">{{.Status}}</span></td>
          <td>{{.Events}}</td>
          <td>{{.LastMessage}}</td>
        </tr>
      {{else}}
        <tr><td colspan="5">No task events found.</td></tr>
      {{end}}
      </tbody>
    </table>
  </section>

  <section>
    <h2>Roles</h2>
    <table>
      <thead><tr><th>Role</th><th>Tasks</th><th>Queued</th><th>Running</th><th>Succeeded</th><th>Failed</th></tr></thead>
      <tbody>
      {{range .Roles}}
        <tr><td>{{.Role}}</td><td>{{.Tasks}}</td><td>{{.Queued}}</td><td>{{.Running}}</td><td>{{.Succeeded}}</td><td>{{.Failed}}</td></tr>
      {{else}}
        <tr><td colspan="6">No role data found.</td></tr>
      {{end}}
      </tbody>
    </table>
  </section>

  <section>
    <h2>Timeline</h2>
    <div class="timeline">
      {{range .Timeline}}
        <div class="event">
          <div><strong>{{.Type}}</strong> {{.From}} -> {{.To}}</div>
          <div class="event-meta">#{{.Seq}} {{.At}} {{.TaskID}} {{.Role}}</div>
          <div>{{.Message}}</div>
        </div>
      {{else}}
        <div class="event">No events found.</div>
      {{end}}
    </div>
  </section>
</main>
</body>
</html>
`))
