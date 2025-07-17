package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		println("error:", err.Error())
	}
}

// run performs the filesystem operations previously done in the test
func run(ctx context.Context) error {
	subCtx, subCtxCancel := context.WithCancel(ctx)
	go func() {
		subCtxCancel()
	}()
	<-subCtx.Done()
	println("context test worked")

	// build a logger
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	le.Debug("hello world from logrus")

	return nil
}
