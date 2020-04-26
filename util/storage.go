package util

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-courier/httptransport/client"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func NewStorageWithIPMachineID(s storage.Storage, machineIdentifier pipeline.MachineIdentifier) *StorageWithIPMachineID {
	return &StorageWithIPMachineID{
		Storage:           s,
		MachineIdentifier: machineIdentifier,
	}
}

type StorageWithIPMachineID struct {
	storage.Storage
	pipeline.MachineIdentifier
}

func (f *StorageWithIPMachineID) RemoteMachineID(path string) (string, error) {
	currentMachineID, err := f.MachineID()
	if err != nil {
		return "", err
	}
	id := GetMachineID(path)
	if id == currentMachineID {
		return "", nil
	}
	return id, nil
}

func (f *StorageWithIPMachineID) Del(ctx context.Context, path string) error {
	remoteMachineID, err := f.RemoteMachineID(path)
	if err != nil {
		return err
	}
	if remoteMachineID == "" {
		return f.Storage.Del(ctx, path)
	}
	data, err := f.rpc(ctx, remoteMachineID, path, http.MethodDelete)
	defer data.Close()
	return err
}

func (f *StorageWithIPMachineID) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	remoteMachineID, err := f.RemoteMachineID(path)
	if err != nil {
		return nil, err
	}
	if remoteMachineID == "" {
		return f.Storage.Read(ctx, path)
	}

	return f.rpc(ctx, remoteMachineID, path, http.MethodGet)
}

func (f *StorageWithIPMachineID) rpc(ctx context.Context, machineID string, path string, method string) (io.ReadCloser, error) {
	u := "http://" + machineID + ":777" + "/" + path

	starts := time.Now()
	defer func() {
		logrus.Debugf("%s from %s, cost %s", method, u, time.Now().Sub(starts))
	}()

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, err
	}

	c := client.GetShortConnClient(10 * time.Second)

	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		if errors.Unwrap(err) == context.Canceled {
			return nil, err
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp.Body, nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nil, errors.New(string(data))
}

func (f *StorageWithIPMachineID) Serve() error {
	srv := &http.Server{}

	srv.Addr = ":777"
	srv.Handler = StorageServer(f.Storage)

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logrus.Error(err)
			} else {
				logrus.Fatal(err)
			}
		}
	}()

	<-stopCh

	logrus.Infof("shutdowning in %s", 10*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func GetMachineID(path string) string {
	return pipeline.GetMachineIDFromFilename(filepath.Base(path))
}

func StorageServer(s storage.Storage) *storageSever {
	return &storageSever{s: s}
}

type storageSever struct {
	s storage.Storage
}

func (s *storageSever) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	path := request.RequestURI[1:]

	if request.Method == http.MethodDelete {
		s.s.Del(request.Context(), path)
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	logrus.Debugf("provide data of %s", path)
	file, err := s.s.Read(request.Context(), path)
	if err != nil {
		if err == afero.ErrFileNotFound {
			rw.WriteHeader(http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		rw.Write([]byte(err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, file)
}
