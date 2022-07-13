package main

import (
	"context"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/synpse-hq/synpse-go"
)

func main() {
	client, err := synpse.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	applications, err := client.ListApplicationAllocations(ctx, &synpse.ListApplicationAllocationsRequest{
		Namespace:   "default",
		Application: "nodered",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, application := range applications {
		spew.Dump(application.ApplicationStatuses)
		for _, ws := range application.ApplicationStatuses {
			log.Printf("%s on device %q: %s", application.DeviceName, application.DeviceID, ws.State)
		}

	}
}
