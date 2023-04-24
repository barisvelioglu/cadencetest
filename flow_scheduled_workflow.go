package main

import (
	"time"

	"go.uber.org/cadence/workflow"
)

// --wt main.IterativeWorkflow --et 60 -i '{"a": 5, "b": 10, "iterations": 3}'
// Define the activity function

//cadence --address 10.98.213.224:7933 --do SimpleDomain workflow run --tl simpleworker --wt main FlowScheduledWorkflow --et 60 -i '{"ChildWorkflowName": "ChildWorkflow", "ChildWorkflowInput": "{\"Message\":\"hello world\"}"}'

type FlowScheduledWorkflowInput struct {
	ChildWorkflowName  string
	ChildWorkflowInput string
}

//cadence --address 10.106.130.55:7933 --do SimpleDomain workflow run --tl simpleworker --wt main.FlowScheduledWorkflow --et 60 -i '{"ChildWorkflowName": "ChildWorkflow", "ChildWorkflowInput": "{\"Message\":\"hello world\"}"}'

func FlowScheduledWorkflow(ctx workflow.Context, input FlowScheduledWorkflowInput) (string, error) {
	ao := workflow.ActivityOptions{
		TaskList:               "simpleworker",
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 30,
		WaitForCancellation:    false,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	cwo := workflow.ChildWorkflowOptions{
		ExecutionStartToCloseTimeout: 10 * time.Minute,
		TaskStartToCloseTimeout:      time.Minute,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	// Start child workflow

	var childFlowResult string
	err := workflow.ExecuteChildWorkflow(ctx, ComplexInstanceProcessWorkflow, ComplexInstanceProcessWorkflowInput{
		Foo: "tes",
		Bar: 1234,
		Instances: []Instance{
			{
				ID:   "12345",
				Name: "Instance1",
			},
			{
				ID:   "123456",
				Name: "Instance2",
			},
			{
				ID:   "1234567",
				Name: "Instance3",
			},
			{
				ID:   "12345678",
				Name: "Instance4",
			},
			{
				ID:   "12345678",
				Name: "Instance5",
			},
			{
				ID:   "12345678",
				Name: "Instance6",
			},
			{
				ID:   "12345678",
				Name: "Instance7",
			},
			{
				ID:   "12345678",
				Name: "Instance8",
			},
			{
				ID:   "12345678",
				Name: "Instance9",
			},
		},
	}).Get(ctx, &childFlowResult)
	if err != nil {
		return "", err
	}

	return childFlowResult, nil
}
