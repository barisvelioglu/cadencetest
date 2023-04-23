package main

import (
	"context"
	"time"

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
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Use a workflow.Group to execute all activities in parallel
	waitGroup := workflow.NewWaitGroup(ctx)
	for _, instance := range input.Instances {
		activityName := "ComplexInstanceProcessSimpleActivity"
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
	future2 := workflow.ExecuteActivity(ctx, "ComplexInstanceProcessDifficultActivity", input.Foo)
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
func ComplexInstanceProcessSimpleActivity(ctx context.Context, value Instance) (string, error) {

	activity.GetLogger(ctx).Info("ComplexInstanceProcessSimpleActivity called.", zap.String("Value", value.Name))

	return "Processed: " + value.Name, nil
}

func ComplexInstanceProcessDifficultActivity(ctx context.Context, value string) (string, error) {

	activity.GetLogger(ctx).Info("ComplexInstanceProcessDifficultActivity called.", zap.String("Value", value))

	return "Processed 61: " + value, nil
}
