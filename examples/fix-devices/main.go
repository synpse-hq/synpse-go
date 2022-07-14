package main

import (
	"context"
	"log"

	"github.com/synpse-hq/synpse-go"
)

func main() {
	client, err := synpse.NewFromEnv(synpse.WithAPIEndpointURL("https://edge.kaunoziogas.lt/api"))
	if err != nil {
		log.Fatal(err)
	}

	applicationNamespace := "default"
	applicationName := "apollo-fnd-prod-vanguard"

	ctx := context.Background()

	// Getting the app
	app, err := client.GetApplication(ctx, applicationNamespace, applicationName)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Getting a list of device IDs that are scheduled but not running
	var notRunningAppDevices []string

	applications, err := client.ListApplicationAllocations(ctx, &synpse.ListApplicationAllocationsRequest{
		Namespace:   applicationNamespace,
		Application: app.ID,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, application := range applications {
		for _, ws := range application.ApplicationStatuses {
			if ws.State != synpse.StateRunning {
				notRunningAppDevices = append(notRunningAppDevices, application.DeviceID)
				// log.Printf("%s on device %q: %s", application.DeviceName, application.DeviceID, ws.State)
			}

		}
	}

	// 2. Getting a list of devices to correlate which are online and offline. Some devices will have the app
	// scheduled but if they are online, the status will not be known.

	devicesResp, err := client.ListDevices(ctx, &synpse.ListDevicesRequest{
		Labels: app.Scheduling.Selectors, // Targeting only devices where it's scheduled
	})
	if err != nil {
		log.Fatal(err)
	}

	deviceMap := make(map[string]*synpse.Device)
	for _, device := range devicesResp.Devices {
		deviceMap[device.ID] = device
	}

	// 3. Getting a list of devices that are online and have the app scheduled

	for _, deviceID := range notRunningAppDevices {
		device, ok := deviceMap[deviceID]
		if !ok {
			log.Printf("Device %q not found", deviceID)
			continue
		}

		log.Printf("---- %s ----", deviceID)
		log.Printf("Device is %s (Seen at %s)", device.Status, device.LastSeenAt.Local())

		if device.Status == synpse.DeviceStatusOnline {
			// TODO: run a docker ps command on the device
			commandResp, err := client.RunDeviceCommand(ctx, device.ID, "docker ps -a")
			if err != nil {
				log.Printf("[%s] command failed with resp %q (%s)", device.ID, commandResp, err)
				continue
			}
			log.Printf("[%s] %s", device.ID, commandResp)
		}
	}
}
