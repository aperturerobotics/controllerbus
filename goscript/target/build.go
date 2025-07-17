package main

import (
	"context"
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

	return nil
}
