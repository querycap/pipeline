package tests_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-courier/snowflakeid"
	. "github.com/onsi/gomega"
	memeventbus "github.com/querycap/pipeline/pkg/eventbus/mem"
	memoperator "github.com/querycap/pipeline/pkg/operator/mem"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/storage/fs"
	fetchImage "github.com/querycap/pipeline/services/fetch-image/handler"
	grayifyImage "github.com/querycap/pipeline/services/grayify-image/handler"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestPipeline(t *testing.T) {
	pipelineSpec, err := pipeline.PipelineFromYAML("../pipeline.yml")
	NewWithT(t).Expect(err).To(BeNil())

	id, _ := snowflakeid.NewSnowflake(1)

	pipelineController := pipeline.NewPipelineController(
		memeventbus.NewMemEventBus(),
		fs.NewFsStorage(afero.NewMemMapFs()),
		id,
		&MID{},
	)

	t.Run("simple pipeline", func(t *testing.T) {
		operatorMgr := memoperator.NewMemOperatorMgr(pipelineController)

		operatorFetchImage, _ := pipeline.OperatorFromYAML("../../fetch-image/operator.yml")
		operatorGrayifyImage, _ := pipeline.OperatorFromYAML("../../grayify-image/operator.yml")

		operatorMgr.Register(operatorFetchImage, fetchImage.Handler)
		operatorMgr.Register(operatorGrayifyImage, grayifyImage.Handler)

		pipelineMgr := pipeline.NewPipelineMgr(operatorMgr, pipelineController)

		p, err := pipelineMgr.NewPipeline(pipelineSpec)
		NewWithT(t).Expect(err).To(BeNil())

		err = p.Start()
		NewWithT(t).Expect(err).To(BeNil())

		defer p.Stop()

		for i := 0; i < 5; i++ {
			t.Run("next", func(t *testing.T) {
				ret, err := p.Next(context.Background(), bytes.NewBuffer([]byte(`{"uri":"https://www.baidu.com/img/bd_logo1.png"}`)))
				NewWithT(t).Expect(err).To(BeNil())

				<-ret.Done()
				NewWithT(t).Expect(ret.Err()).To(BeNil())

				for ret.Scan() {
					err := pipeline.ReadNext(ret, func(r io.Reader) error {
						return writeToFile(fmt.Sprintf("./results/img_%d.png", i), r)
					})
					NewWithT(t).Expect(err).To(BeNil())
				}
			})
		}
	})
}

func writeToFile(filename string, r io.Reader) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	_, e := io.Copy(file, r)
	return e
}

type MID struct {
}

func (MID) MachineID() (string, error) {
	return "1", nil
}
