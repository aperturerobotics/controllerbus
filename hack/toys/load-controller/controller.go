package main

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// ToyController is an example controller.
type ToyController struct {
	// le is the logger
	le *logrus.Entry
	// conf is the config
	conf *ToyControllerConfig
}

// NewToyController constructs a new toy controller.
func NewToyController(le *logrus.Entry, conf *ToyControllerConfig) (*ToyController, error) {
	return &ToyController{le: le, conf: conf}, nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *ToyController) HandleDirective(di directive.Instance) (directive.Resolver, error) {
	// Pass for the example controller
	return nil, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *ToyController) Execute(ctx context.Context) error {
	c.le.Debug("toy controller executed")
	<-ctx.Done()
	return nil
}

// SayHello says hello.
func (c *ToyController) SayHello() {
	c.le.Debugf("Hello %s!", c.conf.GetName())
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *ToyController) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*ToyController)(nil))
