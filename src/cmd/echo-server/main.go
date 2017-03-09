package main

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

func successCommand(ctx context.Context, peer rinq.Peer, req rinq.Request, res rinq.Response) {
	out := rinq.NewPayload(map[string]interface{}{
		"echo": req.Payload.Value(),
	})
	defer out.Close()

	res.Done(out)
}

func failCommand(ctx context.Context, peer rinq.Peer, req rinq.Request, res rinq.Response) {
	err := rinq.Failure{
		Type:    "echo-failure",
		Message: "Failure requested by client.",
		Payload: req.Payload.Clone(),
	}

	res.Error(err)
}

func errorCommand(ctx context.Context, peer rinq.Peer, req rinq.Request, res rinq.Response) {
	res.Error(errors.New("You done goofed."))
}

func notifyCommand(ctx context.Context, peer rinq.Peer, req rinq.Request, res rinq.Response) {
	out := rinq.NewPayload(map[string]interface{}{
		"echo": req.Payload.Value(),
	})
	defer out.Close()

	sess := peer.Session()
	sess.Notify(
		ctx,
		req.Source.Ref().ID,
		"echo-notification",
		out,
	)

	res.Close()
}

func timeoutCommand(ctx context.Context, peer rinq.Peer, req rinq.Request, res rinq.Response) {
	time.Sleep(10 * time.Second)
	res.Close()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// TODO: this env var will be handled by rinq-go
	// https://github.com/rinq/rinq-go/issues/94
	peer, err := amqp.Dial(os.Getenv("RING_AMQP_DSN"))
	if err != nil {
		panic(err)
	}

	err = peer.Listen(
		"echo.1",
		func(
			ctx context.Context,
			req rinq.Request,
			res rinq.Response,
		) {
			defer req.Payload.Close()

			switch req.Command {
			case "success":
				successCommand(ctx, peer, req, res)
			case "fail":
				failCommand(ctx, peer, req, res)
			case "error":
				errorCommand(ctx, peer, req, res)
			case "notify":
				notifyCommand(ctx, peer, req, res)
			case "timeout":
				timeoutCommand(ctx, peer, req, res)
			default:
				res.Fail("unknown-command", "no such command: %s", req.Command)
			}
		},
	)
	if err != nil {
		panic(err)
	}

	<-peer.Done()

	err = peer.Err()
	if err != nil {
		panic(err)
	}
}
