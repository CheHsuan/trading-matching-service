package main

import (
	"context"
	"flag"
	"log"

	"trading-matching-service/app"
)

const (
	servicePort = "9000"
)

var (
	orderQueueSize  int
	tradeQueueSize  int
	cancelQueueSize int
)

func init() {
	flag.IntVar(&orderQueueSize, "order-q-size", 1000000, "order queue size")
	flag.IntVar(&tradeQueueSize, "trade-q-size", 100000, "trade queue size")
	flag.IntVar(&cancelQueueSize, "cancel-q-size", 100000, "cancel queue size")
}

// @title Trading Matching Service API
// @version 1.0
// @description This is trading matching service.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9000
// @BasePath /api/v1

func main() {
	flag.Parse()

	cfg := app.ApplicationConfig{
		ServicePort:     servicePort,
		OrderQueueSize:  orderQueueSize,
		TradeQueueSize:  tradeQueueSize,
		CancelQueueSize: cancelQueueSize,
	}
	application, err := app.NewApplication(cfg)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("start running application at port %s", servicePort)
	if err := application.Run(context.Background()); err != nil {
		panic(err.Error())
	}
}
