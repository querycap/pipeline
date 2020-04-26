package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/querycap/pipeline/pkg/eventbus"
	"github.com/querycap/pipeline/pkg/operator"
	"github.com/querycap/pipeline/spec"
	"github.com/sirupsen/logrus"
)

func Publish(pipelineController PipelineController, ctx context.Context, topic string, task *Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return pipelineController.Publish(ctx, topic, data)
}

func Subscribe(pipelineController PipelineController, topic string, callback func(ctx context.Context, task *Task)) eventbus.Subscription {
	return pipelineController.Subscribe(topic, func(ctx context.Context, data []byte) {
		task := &Task{}
		if err := json.Unmarshal(data, task); err != nil {
			task.ErrMsg = err.Error()
		}
		callback(ctx, task)
	})
}

func ContextWithTask(ctx context.Context, task *Task) context.Context {
	return context.WithValue(ctx, "pipeline.task", task)
}

func TaskFromContext(ctx context.Context) *Task {
	return ctx.Value("pipeline.task").(*Task)
}

func TaskMetaFromPipeline(spec *spec.Pipeline, pipelineID uint64) (*TaskMeta, error) {
	if spec == nil || spec.Stages == nil || len(spec.Stages) == 0 {
		return nil, fmt.Errorf("pipeline %s missing steps", spec)
	}

	if spec.Starts == "" {
		return nil, fmt.Errorf("pipeline %s unknown starts", spec)
	}

	if spec.Ends == "" {
		return nil, fmt.Errorf("pipeline %s unknown ends", spec)
	}

	taskMeta := TaskMeta{
		Scope:     fmt.Sprintf("pipelines/%s/%d", spec.RefID(), pipelineID),
		Starts:    spec.Starts,
		Ends:      spec.Ends,
		StageDeps: map[string][]string{},
	}

	for name, step := range spec.Stages {
		taskMeta.StageDeps[name] = step.Deps

		for _, dep := range step.Deps {
			if _, ok := spec.Stages[dep]; !ok {
				return nil, fmt.Errorf("pipeline %s step %s has invalid dep %s", spec, step, dep)
			}
		}
	}

	return &taskMeta, nil
}

type TaskMeta struct {
	Scope     string
	Starts    string
	Ends      string
	StageDeps map[string][]string
}

func (taskMeta *TaskMeta) NewTask(taskID uint64) *Task {
	return &Task{
		TaskStage: NewTaskStage("$input", []string{}),
		TaskContext: TaskContext{
			ID:       taskID,
			TaskMeta: *taskMeta,
		},
	}
}

type TaskContext struct {
	ID   uint64
	Meta textproto.MIMEHeader `json:",omitempty"`
	TaskMeta
}

func (c TaskContext) Final() string {
	return strings.Join(append([]string{"tasks", strconv.Itoa(int(c.ID)), "$output"}), "/")
}

func NewTaskStage(stage string, inputs []string) *TaskStage {
	return &TaskStage{
		Stage:  stage,
		Inputs: inputs,
	}
}

type TaskStage struct {
	Upstream *TaskStage `json:",omitempty"`

	Stage  string
	Inputs []string
	ErrMsg string `json:",omitempty"`
}

func (s TaskStage) Next(stage string, inputs []string) *TaskStage {
	if s.Stage == "$input" {
		return &TaskStage{
			Stage:  stage,
			Inputs: inputs,
		}
	}

	return &TaskStage{
		Stage:    stage,
		Inputs:   inputs,
		Upstream: &s,
	}
}

func (s TaskStage) Err(err error) *TaskStage {
	s.ErrMsg = err.Error()
	return &s
}

type Task struct {
	TaskContext

	*TaskStage
}

func (t Task) Next(stage string, inputs []string) *Task {
	return &Task{
		TaskContext: t.TaskContext,
		TaskStage:   t.TaskStage.Next(stage, inputs),
	}
}

func (t Task) Err(err error) *Task {
	return &Task{
		TaskContext: t.TaskContext,
		TaskStage:   t.TaskStage.Err(err),
	}
}

func ServeOperator(pipelineController PipelineController, stage string, operatorHandlerFunc operator.OperatorHandlerFunc) eventbus.Subscription {
	logger := logrus.WithFields(logrus.Fields{
		"pipeline":       pipelineController.Scope(),
		"pipeline/stage": stage,
	})

	wg := sync.WaitGroup{}

	sub := Subscribe(pipelineController, stage, func(ctx context.Context, task *Task) {
		wg.Add(1)
		defer wg.Done()

		if task.ErrMsg != "" {
			return
		}

		l := logger.WithContext(ctx).WithFields(logrus.Fields{
			"taskID": task.ID,
		})

		l.Debugf("%s started.", stage)

		startedAt := time.Now()
		var finalErr error

		defer func() {
			if finalErr != nil {
				l.Warnf("%s failed in %s, err: %s", stage, time.Now().Sub(startedAt), finalErr)

				if err := Publish(pipelineController, ctx, task.Final(), task.Err(finalErr)); err != nil {
					l.Error(err)
				}
			} else {
				l.Debugf("%s done in %s", stage, time.Now().Sub(startedAt))
			}
		}()

		t, err := newTransfer(pipelineController, ContextWithTask(context.Background(), task), task)
		if err != nil {
			finalErr = err
			return
		}

		if err := operatorHandlerFunc(t); err != nil {
			finalErr = err
			return
		}

		if err := t.Send(); err != nil && err != ErrNoInputsForNext {
			finalErr = err
		}
	})

	return eventbus.NewSubscription(func() {
		sub.Unsubscribe()
		wg.Wait()
	})
}
