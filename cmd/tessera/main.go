package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	novelteam "github.com/devlikebear/tessera/examples/novel-team"
	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/queue"
	"github.com/devlikebear/tessera/pkg/run"
)

func main() {
	if err := runMain(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("tessera", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	example := fs.String("example", "novel-team", "example template to run")
	goal := fs.String("goal", "Draft a hopeful climate fiction opening", "goal text")
	approvedBy := fs.String("approve-by", "operator", "mandate approver")
	workers := fs.Int("workers", 2, "worker count")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *example != "novel-team" {
		return fmt.Errorf("unsupported example: %s", *example)
	}

	built, err := novelteam.Build(ctx, *goal, *approvedBy, time.Now().UTC())
	if err != nil {
		return err
	}
	result, err := run.ExecuteTaskGraph(ctx, run.ExecutionConfig{
		RunID:       "novel-team-run",
		Mandate:     built.Mandate,
		Graph:       built.Graph,
		Queue:       queue.NewInMemory(),
		Workers:     *workers,
		MaxAttempts: 2,
		RoleLimits: map[string]int{
			"leader":     1,
			"researcher": 1,
			"writer":     2,
			"editor":     1,
		},
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			if task.Payload == "" {
				return executor.Result{}, errors.New("empty task payload")
			}
			return executor.Result{Output: "completed " + task.ID}, nil
		}),
	})
	if err != nil {
		return err
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result.Report)
}
