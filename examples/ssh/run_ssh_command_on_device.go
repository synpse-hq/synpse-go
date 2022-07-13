package main

import (
	"context"
	"log"
	"os"

	"github.com/synpse-hq/synpse-go"
)

// A simple example that will run a 'uname -a' command on the first online device
// and exit the program.
// Example output:
//
// go run run_ssh_command_on_device.go
// 2022/07/13 21:05:29 jetson-nano: Linux synpse-desktop 4.9.253-tegra #1 SMP PREEMPT Mon Jul 26 12:13:06 PDT 2021 aarch64 aarch64 aarch64 GNU/Linux

func main() {
	client, err := synpse.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List devices
	devicesResp, err := client.ListDevices(ctx, &synpse.ListDevicesRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devicesResp.Devices {
		if device.Status != synpse.DeviceStatusOnline {
			log.Printf("Skipping device %q as it's not online", device.Name)
			continue
		}

		// Run SSH command on device, response will be returned as a string
		commandResp, err := client.RunDeviceCommand(ctx, device.ID, "uname -a")
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s: %s", device.Name, commandResp)

		os.Exit(0)
	}

	log.Printf("no online devices found, have you connected at least one?")
	os.Exit(1)
}
