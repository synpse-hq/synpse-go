package main

import (
	"context"
	"fmt"
	"os"

	synpse "github.com/synpse-hq/synpse-go"
)

func main() {
	fmt.Println("Synpse-GO simple client example")
	ctx := context.Background()

	accessKey := os.Getenv("SYNPSE_ACCESS_KEY")

	// Create a new client
	client, err := synpse.New(accessKey)
	if err != nil {
		panic(err)
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		panic(err)
	}

	for _, p := range projects {
		fmt.Printf("Project: %s\n", p.Name)
	}

}
