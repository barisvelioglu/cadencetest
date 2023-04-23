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

	var result string
	for i := 0; i < len(input.Instances); i++ {
		future := workflow.ExecuteActivity(ctx, "ComplexInstanceProcessSimpleActivity", input.Instances[i])
		var value string
		if err := future.Get(ctx, &value); err != nil {
			return "", err
		}

		result += " " + value
	}

	// future2 := workflow.ExecuteActivity(ctx, "ComplexInstanceProcessDifficultActivity", input.Bar)
	// var result2 string
	// if err := future2.Get(ctx, &result2); err != nil {
	// 	return err
	// }

	workflow.GetLogger(ctx).Info("Done", zap.String("result", result))

	return result, nil
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

func ComplexInstanceProcessDifficultActivity(ctx context.Context, value int) (string, error) {

	activity.GetLogger(ctx).Info("ComplexInstanceProcessDifficultActivity called.", zap.Int("Value", value))

	return "Processed: " + string(value), nil
}
