package main

import (
	"time"

	"go.uber.org/cadence/workflow"
)

//cadence --address 10.106.130.55:7933 --do SimpleDomain workflow run --tl simpleworker --wt main.IterativeWorkflow --et 60 -i '["input1", "input2", "input3"]'

// --wt main.IterativeWorkflow --et 60 -i '{"a": 5, "b": 10, "iterations": 3}'
// Define the activity function

//cadence --address 10.106.130.55:7933 --do SimpleDomain workflow run --tl simpleworker --wt main Aileflow --et 60 -i '{"ChildWorkflowName": "ChildWorkflow", "ChildWorkflowInput": "{\"Message\":\"hello world\"}"}'

type AileflowInput struct {
	ChildWorkflowName  string
	ChildWorkflowInput string
}

//cadence --address 10.106.130.55:7933 --do SimpleDomain workflow run --tl simpleworker --wt main.Aileflow --et 60 -i '{"ChildWorkflowName": "ChildWorkflow", "ChildWorkflowInput": "{\"Message\":\"hello world\"}"}'

func Aileflow(ctx workflow.Context, input AileflowInput) (string, error) {
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
	err := workflow.ExecuteChildWorkflow(ctx, SimpleWorkflow, SimleWorkflowInput{
		Foo: "tes",
		Bar: 1234,
	}).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return "AAA", nil
}

type ChildWorkflowInput struct {
	Message string
}
