package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

func main() {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "192.168.1.21",
		Port: 6334,
	})

	if err != nil {
		fmt.Errorf("error conectando qdrant: %s", err.Error())
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	health, err := client.HealthCheck(ctx)
	if err != nil {
		fmt.Errorf("error chequeando salud: %s", err.Error())
	}

	fmt.Printf("Conexión exitosa. Qdrant versión: %s\n", health.GetVersion())
}
