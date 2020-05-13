package pipeline

import (
	"context"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strconv"
)

var (
	ErrNoInputsForNext = errors.New("no inputs for next")
)

func newTransfer(pipelineController PipelineController, ctx context.Context, task *Task) (*transfer, error) {
	return &transfer{ctx: ctx, pipelineController: pipelineController, task: task}, nil
}

type transfer struct {
	ctx context.Context

	pipelineController PipelineController

	task         *Task
	inputScanIdx int

	outputs []string
}

func (t *transfer) Context() context.Context {
	if t.ctx == nil {
		return context.Background()
	}
	return t.ctx
}

func (t *transfer) Scan() bool {
	return t.inputScanIdx < len(t.task.Inputs)
}

func (t *transfer) Next() (io.ReadCloser, error) {
	if !t.Scan() {
		return nil, errors.New("no more inputs")
	}

	inputFile := t.task.Inputs[t.inputScanIdx]

	file, err := t.pipelineController.Read(t.Context(), inputFile)
	if err != nil {
		return nil, err
	}
	t.inputScanIdx++
	return file, nil
}

func (t *transfer) Put(writerTo io.WriterTo) error {
	machineID, err := t.pipelineController.MachineID()
	if err != nil {
		return err
	}

	fileID, err := t.pipelineController.ID()
	if err != nil {
		return err
	}

	filename := FilenameWithMachineID(machineID, strconv.FormatUint(fileID, 10))

	if contentTypeDescriber, ok := writerTo.(ContentTypeDescriber); ok {
		contentType := contentTypeDescriber.ContentType()
		if contentType != "" {
			ext, err := mime.ExtensionsByType(contentType)
			if err == nil && len(ext) > 0 {
				filename = filename + ext[0]
			}
		}
	}

	filename = filepath.Join(
		"tasks", strconv.FormatUint(t.task.ID, 10),
		"stages", t.task.Stage,
		"results", filename,
	)

	if err := t.pipelineController.Put(t.Context(), filename, writerTo); err != nil {
		return err
	}

	t.outputs = append(t.outputs, filename)

	return nil
}

func (t *transfer) Send() error {
	if len(t.outputs) == 0 {
		return ErrNoInputsForNext
	}

	nextStages := make([]string, 0)

	switch t.task.Stage {
	case "$input":
		nextStages = append(nextStages, t.task.Starts)
	case t.task.Ends:
		nextStages = append(nextStages, t.task.Final())
	default:
		for step, deps := range t.task.StageDeps {
			for _, dep := range deps {
				if t.task.Stage == dep {
					nextStages = append(nextStages, step)
				}
			}
		}
	}

	for _, next := range nextStages {
		if err := Publish(t.pipelineController, t.Context(), next, t.task.Next(next, t.outputs)); err != nil {
			return err
		}
	}

	t.outputs = []string{}

	return nil
}
