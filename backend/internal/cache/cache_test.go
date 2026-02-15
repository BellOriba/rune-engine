package cache

import (
	"context"
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	c, err := NewCache()
	if err != nil {
		t.Skip("Redis não disponível para teste: ", err)
	}

	ctx := context.Background()
	key := "test_key"
	value := "ascii_art_data"

	err = c.Set(ctx, key, value, 10*time.Second)
	if err != nil {
		t.Fatalf("Erro ao salvar no cache: %v", err)
	}

	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("Erro ao buscar no cache: %v", err)
	}

	if got != value {
		t.Errorf("esperado %s, obtido %s", value, got)
	}
}

