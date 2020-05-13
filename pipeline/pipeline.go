package pipeline

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/querycap/pipeline/spec"
)

func NewPipelineMgr(operatorMgr OperatorMgr, c PipelineController) *PipelineMgr {
	return &PipelineMgr{operatorMgr: operatorMgr, pipelineController: c}
}

type PipelineMgr struct {
	operatorMgr        OperatorMgr
	pipelineController PipelineController
}

func (p *PipelineMgr) NewPipeline(spec *spec.Pipeline) (*Pipeline, error) {
	id, err := p.pipelineController.ID()
	if err != nil {
		return nil, err
	}

	taskMeta, err := TaskMetaFromPipeline(spec, id)
	if err != nil {
		return nil, err
	}

	return &Pipeline{
		id:       id,
		taskMeta: taskMeta,
		spec:     spec,
		mgr: &PipelineMgr{
			operatorMgr:        p.operatorMgr,
			pipelineController: p.pipelineController.WithScope(taskMeta.Scope),
		},
	}, nil
}

type Pipeline struct {
	id       uint64
	taskMeta *TaskMeta
	spec     *spec.Pipeline
	mgr      *PipelineMgr
	results  sync.Map
}

func (p *Pipeline) Start() error {
	for name, step := range p.spec.Stages {
		for i := 0; i < 3; i++ {
			if err := p.mgr.operatorMgr.Start(p.taskMeta.Scope, name, step); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Pipeline) Stop() error {
	for name := range p.spec.Stages {
		if err := p.mgr.operatorMgr.Stop(p.taskMeta.Scope, name); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) Next(ctx context.Context, input io.Reader) (Result, error) {
	task, err := p.newTask()
	if err != nil {
		return nil, err
	}

	go func() {
		sub := Subscribe(p.mgr.pipelineController, task.Final(), p.finish)
		defer sub.Unsubscribe()

		<-ctx.Done()
		p.finish(ctx, task.Err(ctx.Err()))
	}()

	t, err := newTransfer(p.mgr.pipelineController, ContextWithTask(ctx, task), task)
	if err != nil {
		return nil, err
	}

	if err := SendByReader(t, input); err != nil {
		return nil, err
	}

	return p.register(task), nil
}

func (p *Pipeline) newTask() (*Task, error) {
	taskID, err := p.mgr.pipelineController.ID()
	if err != nil {
		return nil, err
	}
	return p.taskMeta.NewTask(taskID), nil
}

func (p *Pipeline) finish(ctx context.Context, t *Task) {
	r := p.getResult(t)

	if r == nil {
		return
	}

	defer p.results.Delete(t.ID)

	if t.ErrMsg != "" {
		r.finish(nil, fmt.Errorf("[%s]%s: %s", p.spec.RefID(), t.Stage, t.ErrMsg))
		return
	}

	r.finish(newTransfer(p.mgr.pipelineController, ContextWithTask(ctx, t), t))
}

func (p *Pipeline) register(task *Task) *result {
	r := newResult()
	p.results.Store(task.ID, r)
	return r
}

func (p *Pipeline) getResult(task *Task) *result {
	v, ok := p.results.Load(task.ID)
	if ok {
		return v.(*result)
	}
	return nil
}
