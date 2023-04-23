package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	nats "github.com/nats-io/nats.go"
	"github.com/uber-go/tally"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
)

var HostPort = "127.0.0.1:7933"
var Domain = "SimpleDomain"
var TaskListName = "simpleworker"
var ClientName = "simpleworker"
var CadenceService = "cadence-frontend"

var natsConnection *nats.Conn

func main() {
	nc, err := nats.Connect(nats.DefaultURL, nats.UserInfo("nats_client", "UT2asYgXi6"))
	if err != nil {
		log.Panicf("Failed to connect to NATS %v", err)
	}
	defer nc.Close()

	natsConnection = nc

	workflow.Register(InputProcessWorkflow)
	activity.RegisterWithOptions(InputProcessSimpleActivity, activity.RegisterOptions{Name: "InputProcessSimpleActivity"})
	activity.RegisterWithOptions(InputProcessDifficultActivity, activity.RegisterOptions{Name: "InputProcessDifficultActivity"})

	workflow.Register(InstanceProcessWorkflow)
	activity.RegisterWithOptions(InstanceProcessSimpleActivity, activity.RegisterOptions{Name: "InstanceProcessSimpleActivity"})
	activity.RegisterWithOptions(InstanceProcessDifficultActivity, activity.RegisterOptions{Name: "InstanceProcessDifficultActivity"})

	workflow.Register(ComplexInstanceProcessWorkflow)
	activity.RegisterWithOptions(ComplexInstanceProcessInputActivity, activity.RegisterOptions{Name: "ComplexInstanceProcessInputActivity"})
	activity.RegisterWithOptions(ComplexInstanceProcessOutputActivity, activity.RegisterOptions{Name: "ComplexInstanceProcessOutputActivity"})
	activity.RegisterWithOptions(ComplexInstanceProcessInputActivityNats, activity.RegisterOptions{Name: "ComplexInstanceProcessInputActivityNats"})

	workflow.Register(FlowScheduledWorkflow)

	// Subscribe to the "ComplexInstanceProcessSimpleActivity" subject
	_, err = nc.Subscribe("ComplexInstanceProcessSimpleActivity", func(m *nats.Msg) {
		handleNatsActivityMessage(m, ComplexInstanceProcessInputActivity)
	})

	go startWorker(buildLogger(), buildCadenceClient())
	//go startWorker2(buildLogger(), buildCadenceClient())

	select {}
}

func buildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger")
	}

	return logger
}

func buildCadenceClient() workflowserviceclient.Interface {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(ClientName))
	if err != nil {
		panic("Failed to setup tchannel")
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: ClientName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: ch.NewSingleOutbound(HostPort)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	return workflowserviceclient.New(dispatcher.ClientConfig(CadenceService))
}

func handleNatsActivityMessage(m *nats.Msg, activityFunc func(context.Context, Instance) (string, error)) {
	var input Instance
	if err := json.Unmarshal(m.Data, &input); err != nil {
		return
	}
	ctx := context.Background()
	result, err := activityFunc(ctx, input)
	if err != nil {
		return
	}
	fmt.Printf("Activity completed: %s\n", result)

	// Send the result back as a reply
	replyData, err := json.Marshal(result)
	if err != nil {
		return
	}
	m.Respond(replyData)
}

func startWorker(logger *zap.Logger, service workflowserviceclient.Interface) {
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, map[string]string{}),
	}

	worker := worker.New(
		service,
		Domain,
		TaskListName,
		workerOptions)
	err := worker.Start()
	if err != nil {
		panic("Failed to start worker")
	}

	logger.Info("Started Worker.", zap.String("worker", TaskListName))
}

// func startWorker2(logger *zap.Logger, service workflowserviceclient.Interface) {
// 	workerOptions := worker.Options{
// 		Logger:                       logger,
// 		MetricsScope:                 tally.NewTestScope(TaskListName, map[string]string{}),
// 		StickyScheduleToStartTimeout: time.Minute,
// 	}

// 	worker := worker.New(
// 		service,
// 		Domain,
// 		TaskListName,
// 		workerOptions)

// 	worker.RegisterWorkflow(ParentWorkflow)
// 	err := worker.Start()
// 	if err != nil {
// 		panic("Failed to start worker")
// 	}

// 	logger.Info("Started Worker.", zap.String("worker", TaskListName))
// }
