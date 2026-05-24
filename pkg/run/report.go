package run

import "github.com/devlikebear/tessera/pkg/queue"

type Report struct {
	RunID       string       `json:"run_id"`
	Closure     ClosureKind  `json:"closure"`
	Succeeded   int          `json:"succeeded"`
	Failed      int          `json:"failed"`
	DeadLetters []queue.Task `json:"dead_letters,omitempty"`
	Events      []Event      `json:"events"`
}

func NewReport(r *Run, tasks, deadLetters []queue.Task) Report {
	report := Report{
		RunID:       r.ID,
		Closure:     r.Closure,
		DeadLetters: cloneTasks(deadLetters),
		Events:      r.Events.Events(),
	}
	for _, task := range tasks {
		switch task.Status {
		case queue.TaskSucceeded:
			report.Succeeded++
		case queue.TaskFailed:
			report.Failed++
		}
	}
	return report
}

func cloneTasks(tasks []queue.Task) []queue.Task {
	out := make([]queue.Task, len(tasks))
	for i := range tasks {
		out[i] = tasks[i]
		out[i].DependsOn = append([]string(nil), tasks[i].DependsOn...)
	}
	return out
}
