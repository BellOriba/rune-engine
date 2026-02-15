package worker

import (
	"io"
	"log/slog"
	"testing"
)

func TestPool_Submit(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	pool := NewPool(2, 10, log)
	ctx := t.Context()

	pool.Start(ctx)

	processed := make(chan int, 5)
	for range 5 {
		pool.Submit(func() {
			processed <- 1
		})
	}

	count := 0
	for range 5 {
		count += <-processed
	}

	if count != 5 {
		t.Errorf("esperado 5 tarefas processadas, obtido %d", count)
	}
	pool.Shutdown()
}

