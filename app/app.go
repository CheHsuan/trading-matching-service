package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	httpswagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"

	// import swagger docs
	_ "trading-matching-service/docs"
	"trading-matching-service/pkg/api"
	"trading-matching-service/pkg/engine"
	cancelsvc "trading-matching-service/pkg/service/cancel"
	msgsvc "trading-matching-service/pkg/service/message"
	ordersvc "trading-matching-service/pkg/service/order"
	tradesvc "trading-matching-service/pkg/service/trade"
)

var (
	qNameOrder  = "order"
	qNameTrade  = "trade"
	qNameCancel = "cancel"
)

// ApplicationConfig defines application config struct.
type ApplicationConfig struct {
	ServicePort string

	OrderQueueSize  int
	TradeQueueSize  int
	CancelQueueSize int
}

// Application is a collection of applications including http server or any other apps.
type Application struct {
	ApplicationConfig
	handler      http.Handler
	matchEngine  engine.Engine
	tradeEngine  engine.Engine
	cancelEngine engine.Engine
}

// NewApplication creates a application.
func NewApplication(config ApplicationConfig) (*Application, error) {
	queues := getQueues(config)
	store := ordersvc.NewMemoryStore()

	h, err := getHTTPHandler(config, queues, store)
	if err != nil {
		return nil, err
	}

	me := getMatchEngine(queues, store)
	te := getTradeEngine(queues)
	ce := getCancelEngine(queues)

	return &Application{
		ApplicationConfig: config,
		handler:           h,
		matchEngine:       me,
		tradeEngine:       te,
		cancelEngine:      ce,
	}, nil
}

// Run runs the application.
func (a *Application) Run(ctx context.Context) error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return a.tradeEngine.Run(ctx)
	})
	eg.Go(func() error {
		return a.cancelEngine.Run(ctx)
	})
	eg.Go(func() error {
		return a.matchEngine.Run(ctx)
	})
	eg.Go(func() error {
		return http.ListenAndServe(fmt.Sprintf(":%s", a.ServicePort), a.handler)
	})

	if err := eg.Wait(); err != nil {
		return errors.Errorf("application got an error: %v", err)
	}

	return nil
}

func getQueues(config ApplicationConfig) map[string]msgsvc.Queue {
	m := map[string]msgsvc.Queue{
		qNameOrder:  msgsvc.NewQueue(config.OrderQueueSize),
		qNameTrade:  msgsvc.NewQueue(config.TradeQueueSize),
		qNameCancel: msgsvc.NewQueue(config.CancelQueueSize),
	}
	return m
}

func getHTTPHandler(config ApplicationConfig, queues map[string]msgsvc.Queue, orderStore ordersvc.Store) (http.Handler, error) {
	router, err := getRouter(queues, orderStore)
	if err != nil {
		return nil, errors.Errorf("failed to get router: %v", err)
	}

	headersOk := handlers.AllowedHeaders([]string{"Origin", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{fmt.Sprintf("http://localhost:%s", config.ServicePort), fmt.Sprintf("http://127.0.0.1:%s", config.ServicePort)})
	methodsOk := handlers.AllowedMethods([]string{"POST", "DELETE", "OPTIONS"})

	return handlers.CORS(headersOk, originsOk, methodsOk)(router), nil
}

func getRouter(queues map[string]msgsvc.Queue, orderStore ordersvc.Store) (*mux.Router, error) {
	controller, err := getController(queues, orderStore)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.PathPrefix("/swagger-ui/").Handler(httpswagger.WrapHandler)
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.HandleFunc("/orders", controller.PlaceOrder).Methods(http.MethodPost)
	apiV1.HandleFunc("/orders/{oid}", controller.CancelOrder).Methods(http.MethodDelete)
	return r, nil
}

func getController(queues map[string]msgsvc.Queue, orderStore ordersvc.Store) (*api.Controller, error) {
	return api.NewController(queues[qNameOrder], orderStore), nil
}

func getMatchEngine(queues map[string]msgsvc.Queue, orderStore ordersvc.Store) engine.Engine {
	return engine.NewMatchEngine(orderStore, queues[qNameOrder], queues[qNameTrade], queues[qNameCancel])
}

func getTradeEngine(queues map[string]msgsvc.Queue) engine.Engine {
	return engine.NewTradeEngine(queues[qNameTrade], tradesvc.NewStdoutRecorder())
}

func getCancelEngine(queues map[string]msgsvc.Queue) engine.Engine {
	return engine.NewCancelEngine(queues[qNameCancel], cancelsvc.NewStdoutRecorder())
}
