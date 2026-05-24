package visualize

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/devlikebear/tessera/pkg/run"
)

type Projection struct {
	RunID      string
	Closure    run.ClosureKind
	EventCount int
	Tasks      []TaskProjection
	Roles      []RoleProjection
	Timeline   []run.Event
}

type TaskProjection struct {
	ID          string
	Role        string
	Status      string
	LastMessage string
	Events      int
	FirstSeq    int
	LastSeq     int
}

type RoleProjection struct {
	Role      string
	Tasks     int
	Queued    int
	Running   int
	Succeeded int
	Failed    int
}

func ReadEventsFile(path string) ([]run.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ReadEventsJSONL(file)
}

func ReadEventsJSONL(r io.Reader) ([]run.Event, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var events []run.Event
	line := 0
	for scanner.Scan() {
		line++
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var event run.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("parse events jsonl line %d: %w", line, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func Project(events []run.Event) Projection {
	projection := Projection{
		EventCount: len(events),
		Timeline:   append([]run.Event(nil), events...),
	}
	tasksByID := make(map[string]*TaskProjection)

	for _, event := range events {
		if projection.RunID == "" {
			projection.RunID = event.RunID
		}
		if event.Type == run.EventClosure {
			projection.Closure = run.ClosureKind(event.To)
		}
		if event.Type != run.EventTaskTransition || event.TaskID == "" {
			continue
		}

		task, ok := tasksByID[event.TaskID]
		if !ok {
			task = &TaskProjection{ID: event.TaskID, FirstSeq: event.Seq}
			tasksByID[event.TaskID] = task
		}
		if task.FirstSeq == 0 || event.Seq < task.FirstSeq {
			task.FirstSeq = event.Seq
		}
		if event.Seq > task.LastSeq {
			task.LastSeq = event.Seq
		}
		if event.Role != "" {
			task.Role = event.Role
		}
		if event.To != "" {
			task.Status = event.To
		}
		if event.Message != "" {
			task.LastMessage = event.Message
		}
		task.Events++
	}

	for _, task := range tasksByID {
		projection.Tasks = append(projection.Tasks, *task)
	}
	sort.Slice(projection.Tasks, func(i, j int) bool {
		if projection.Tasks[i].FirstSeq == projection.Tasks[j].FirstSeq {
			return projection.Tasks[i].ID < projection.Tasks[j].ID
		}
		return projection.Tasks[i].FirstSeq < projection.Tasks[j].FirstSeq
	})
	projection.Roles = summarizeRoles(projection.Tasks)
	return projection
}

func summarizeRoles(tasks []TaskProjection) []RoleProjection {
	rolesByName := make(map[string]*RoleProjection)
	for _, task := range tasks {
		roleName := task.Role
		if roleName == "" {
			roleName = "unknown"
		}
		role, ok := rolesByName[roleName]
		if !ok {
			role = &RoleProjection{Role: roleName}
			rolesByName[roleName] = role
		}
		role.Tasks++
		switch task.Status {
		case "queued":
			role.Queued++
		case "running":
			role.Running++
		case "succeeded":
			role.Succeeded++
		case "failed":
			role.Failed++
		}
	}

	roles := make([]RoleProjection, 0, len(rolesByName))
	for _, role := range rolesByName {
		roles = append(roles, *role)
	}
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Role < roles[j].Role
	})
	return roles
}
