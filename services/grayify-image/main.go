package main

import (
	"github.com/querycap/pipeline/services/grayify-image/handler"
	"github.com/querycap/pipeline/util"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	util.ServeOperator(handler.Handler)
}
