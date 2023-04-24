package main

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

//cadence --address 10.106.130.55:7933 --do SimpleDomain workflow run --tl simpleworker --wt main.ComplexInstanceProcessWorkflow --et 60 -i '{"foo": "bar", "bar": 6161616161}'

func ComplexInstanceProcessWorkflow(ctx workflow.Context, input *ComplexInstanceProcessWorkflowInput) (string, error) {
	ao := workflow.ActivityOptions{
		TaskList:               "simpleworker",
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 30,
		WaitForCancellation:    false,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second * 1,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Second * 10,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Use a workflow.Group to execute all activities in parallel
	waitGroup := workflow.NewWaitGroup(ctx)
	for _, instance := range input.Instances {
		activityName := "ComplexInstanceProcessInputActivityNats"
		activityArgs := instance
		workflow.Go(ctx, func(ctx workflow.Context) {

			// Execute each activity in its own coroutine
			future := workflow.ExecuteActivity(ctx, activityName, activityArgs)
			var value string
			if err := future.Get(ctx, &value); err != nil {
				return
			}
			workflow.GetLogger(ctx).Info("Activity completed", zap.String("value", value))
		})
	}

	// Wait for all activities to complete
	waitGroup.Wait(ctx)

	workflow.GetLogger(ctx).Info("All activities completed")

	// Execute the remaining activity
	future2 := workflow.ExecuteActivity(ctx, "ComplexInstanceProcessOutputActivity", input.Foo)
	var result2 string
	if err := future2.Get(ctx, &result2); err != nil {
		return "", err
	}

	workflow.GetLogger(ctx).Info("Done", zap.String("result", result2))

	return result2, nil
}

type Instance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ComplexInstanceProcessWorkflowInput struct {
	Foo       string     `json:"foo"`
	Bar       int        `json:"bar"`
	Instances []Instance `json:"instances"`
}

// SimpleActivity is a sample Cadence activity function that takes one parameter and
// returns a string containing the parameter value.
func ComplexInstanceProcessInputActivity(ctx context.Context, value Instance) (string, error) {

	activity.GetLogger(ctx).Info("ComplexInstanceProcessInputActivity called.", zap.String("Value", value.Name))

	return "Processed: " + value.Name, nil
}

// SimpleActivity is a sample Cadence activity function that takes one parameter and
// returns a string containing the parameter value.
func ComplexInstanceProcessInputActivityNats(ctx context.Context, value Instance) (string, error) {

	msg, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	reply, err := natsConnection.Request("ComplexInstanceProcessInputActivityNats", msg, time.Second*60)
	if err != nil {
		return "", err
	}

	var result string
	if err := json.Unmarshal(reply.Data, &result); err != nil {
		return "", err
	}

	activity.GetLogger(ctx).Info("ComplexInstanceProcessInputActivityNats called.", zap.String("Value", result))

	return "Processed: " + result, nil
}

func ComplexInstanceProcessOutputActivity(ctx context.Context, value string) (string, error) {

	activity.GetLogger(ctx).Info("ComplexInstanceProcessOutputActivity called.", zap.String("Value", value))

	return "Processed 61: " + value, nil
}
