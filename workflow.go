package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

//cadence --address 10.106.130.55:7933 --do SimpleDomain workflow run --tl simpleworker --wt main.SimpleWorkflow --et 60 -i '{"foo": "bar", "bar": 6161616161}'

func SimpleWorkflow(ctx workflow.Context, input *SimleWorkflowInput) error {
	ao := workflow.ActivityOptions{
		TaskList:               "simpleworker",
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 30,
		WaitForCancellation:    false,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	future := workflow.ExecuteActivity(ctx, "SimpleActivity", input.Foo)
	var result string
	if err := future.Get(ctx, &result); err != nil {
		return err
	}

	future2 := workflow.ExecuteActivity(ctx, "SimpleActivity2", input.Bar)
	var result2 string
	if err := future2.Get(ctx, &result2); err != nil {
		return err
	}

	workflow.GetLogger(ctx).Info("Done", zap.String("result", result))

	return nil
}

type SimleWorkflowInput struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

// SimpleActivity is a sample Cadence activity function that takes one parameter and
// returns a string containing the parameter value.
func SimpleActivity(ctx context.Context, value string) (string, error) {

	activity.GetLogger(ctx).Info("SimpleActivity called.", zap.String("Value", value))

	return "Processed: " + string(value), nil
}

func SimpleActivity2(ctx context.Context, value int) (string, error) {

	activity.GetLogger(ctx).Info("SimpleActivity2 called.", zap.Int("Value", value))

	return "Processed: " + string(value), nil
}

//FlowScheduledWorkflow

// ComplexInstanceChildWorkflow
//==> InstanceWorkflow
//==> InstanceWorkflow

// InstanceWorkflow
// InputWorkflow
