package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	novelteam "github.com/devlikebear/tessera/examples/novel-team"
	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/observe"
	"github.com/devlikebear/tessera/pkg/queue"
	"github.com/devlikebear/tessera/pkg/run"
	"github.com/devlikebear/tessera/pkg/visualize"
)

func main() {
	if err := runMain(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "visualize" {
		return runVisualize(args[1:])
	}

	fs := flag.NewFlagSet("tessera", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	example := fs.String("example", "novel-team", "example template to run")
	goal := fs.String("goal", "Draft a hopeful climate fiction opening", "goal text")
	approvedBy := fs.String("approve-by", "operator", "mandate approver")
	workers := fs.Int("workers", 2, "worker count")
	eventsOut := fs.String("events-out", "", "write execution events as JSONL")
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
	var sink run.EventSink
	var eventsFile *observe.JSONLinesSink
	if *eventsOut != "" {
		eventsFile, err = observe.NewJSONLinesFile(*eventsOut)
		if err != nil {
			return err
		}
		sink = eventsFile
	}
	result, err := run.ExecuteTaskGraph(ctx, run.ExecutionConfig{
		RunID:       "novel-team-run",
		Mandate:     built.Mandate,
		Graph:       built.Graph,
		Queue:       queue.NewInMemory(),
		EventSink:   sink,
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
	if closeErr := closeEventSink(eventsFile); closeErr != nil {
		return closeErr
	}
	if err != nil {
		return err
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result.Report)
}

func runVisualize(args []string) error {
	eventsPath, outPath, err := parseVisualizeArgs(args)
	if err != nil {
		return err
	}
	return visualize.WriteHTMLReportFile(eventsPath, outPath)
}

func parseVisualizeArgs(args []string) (string, string, error) {
	var eventsPath, outPath string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--out":
			i++
			if i >= len(args) {
				return "", "", errors.New("missing value for --out")
			}
			outPath = args[i]
		default:
			if strings.HasPrefix(arg, "-") {
				return "", "", fmt.Errorf("unsupported visualize flag: %s", arg)
			}
			if eventsPath != "" {
				return "", "", fmt.Errorf("unexpected visualize argument: %s", arg)
			}
			eventsPath = arg
		}
	}
	if eventsPath == "" {
		return "", "", errors.New("visualize requires an events JSONL path")
	}
	if outPath == "" {
		return "", "", errors.New("visualize requires --out <report.html>")
	}
	return eventsPath, outPath, nil
}

func closeEventSink(sink *observe.JSONLinesSink) error {
	if sink == nil {
		return nil
	}
	return sink.Close()
}
