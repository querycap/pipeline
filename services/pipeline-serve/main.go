package main

import (
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	"github.com/querycap/pipeline/services/pipeline-serve/global"
	"github.com/querycap/pipeline/services/pipeline-serve/routes"
)

func main() {
	ht := &httptransport.HttpTransport{
		Port: 8080,
	}
	ht.SetDefaults()

	if err := global.Pipeline.Start(); err != nil {
		panic(err)
	}

	courier.Run(routes.RootRouter, ht)

	// cleanup
	global.Pipeline.Stop()
}
