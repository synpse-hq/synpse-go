<div align="center">

  <img src="https://github.com/synpse-hq/synpse/blob/main/assets/logo.png" width="200px">
  <br>

  **The easiest way to bootstrap your devices and deploy applications.    
  Synpse manages OTA deployment & updates, provides SSH and network access.**

  ---

  <p align="center">
    <a href="https://synpse.net">Website</a> •
    <a href="https://github.com/synpse-hq/synpse/discussions">Discussions</a> •  
    <a href="https://docs.synpse.net">Docs</a> •  
    <a href="https://discord.gg/dkgN4vVNdm">Discord</a> •
    <a href="https://cloud.synpse.net/">Cloud</a>
  </p>

</div>


## Synpse v1 API client

[![Build Status](https://drone-kr.webrelay.io/api/badges/synpse-hq/synpse-go/status.svg)](https://drone-kr.webrelay.io/synpse-hq/synpse-go)

Synpse provides your device fleet management, application deployment and their configuration. Whole process is simple with very low learning curve.

* 🚧 This API client is still under active development, the methods can be changed 🚧 *

---

- [Synpse v1 API client](#synpse-v1-api-client)
- [Prerequisites](#prerequisites)
- [Install](#install)
- [Authentication](#authentication)
- [Examples](#examples)
	- [Registering New Device](#registering-new-device)
	- [List Devices](#list-devices)
	- [Create Application](#create-application)
	- [Update Application](#update-application)
	- [List Applications](#list-applications)
	- [Delete Applications:](#delete-applications)

## Prerequisites

- [Go](https://golang.org/dl/)
- [Synpse account](https://cloud.synpse.net/) - free up to 5 devices.

## Install

```shell
go get github.com/synpse-hq/synpse-go
```

## Authentication

To authenticate, we recommend using a project level access key that you can generate here: https://cloud.synpse.net/service-accounts.

Alternatively, use [Personal Access Keys](https://cloud.synpse.net/access-keys), however, they will be able to manage multiple projects.


## Examples

Let's start with creating an client API:

```golang
package main

import (
  "os"
  "fmt"

  "github.com/synpse-hq/synpse-go"
)

func main() {
  // Create a new API client with a specified access key. You can get your access key
  // from https://cloud.synpse.net/service-accounts
  apiClient, _ := synpse.NewWithProject(os.Getenv("SYNPSE_PROJECT_ACCESS_KEY"), os.Getenv("SYNPSE_PROJECT_ID"))    
}
```

### Registering New Device

When automating your device fleet operations, you will most likely need to create and manage [device registration tokens](https://docs.synpse.net/synpse-core/devices/provisioning). These tokens can be created with a set of labels and environment variables which will then be inherited by any device that registers using it.

```golang
  // In this example we use user ID but it could be anything else like company name, location identifier, etc.
  var userID = "usr_mkalpxzlab"
  // Optional max registrations. It's a good practice to set these to sane limits. If you expect only one device
  // to register with this token, set it to 1.
  var maxRegistrations = 10

  // Create a registration token
  drt, _ := apiClient.CreateRegistrationToken(ctx, synpse.DeviceRegistrationToken{
    Name:                 "drt-" + userID,
    MaxRegistrations:     &maxRegistrations,                 // optional 
    Labels:               map[string]string{"user": userID}, // optional
  })

  // Use this token together with your project ID:
  // 
  // curl https://downloads.synpse.net/install.sh | \
  //   AGENT_PROJECT={{ PROJECT_ID }} \
  //   AGENT_REGISTRATION_TOKEN={{ DEVICE_REGISTRATION_TOKEN }} \
  //   bash


  // Once registration token is created, you can use device filtering to find it:
  devicesResp, _ := apiClient.ListDevices(context.Background(), &synpse.ListDevicesRequest{
    Labels: map[string]string{
      "user": userID, 
    },
  })

```


### List Devices

```golang
  // List devices
  devicesResp, _ := apiClient.ListDevices(context.Background(), &synpse.ListDevicesRequest{})

  // Print device names
  for _, device := range devicesResp.Devices {
    fmt.Println(device.Name)
  }
```

Here we list an already registered devices. Default page size is 100, if you have more devices, use pagination options and iterate for as long as you have the next page token.

Filtering devices during the query is almost always the preferred solution. You can filter devices by labels:

```golang
  // List devices that have this label
  devicesResp, _ := apiClient.ListDevices(context.Background(), &synpse.ListDevicesRequest{
    Labels: map[string]string{
      "group": "one", 
    },
  })
```

### Create Application

Applications in Synpse can either:
- Run on all devices in the project
- Run on devices with matching labels

To create an application that will run on all devices:

```golang
  // Create an application in 'default' namespace that will be deployed on all devices
  application, err := apiClient.CreateApplication(context.Background(), "default", synpse.Application{
    Name:        "app-name",
    Scheduling: synpse.Scheduling{
      Type: synpse.ScheduleTypeAllDevices,
    },
    Spec: synpse.ApplicationSpec{
      ContainerSpec: []synpse.ContainerSpec{
        {
          Name:  "hello",
          Image: "quay.io/synpse/hello-synpse-go:latest",
          Ports: []string{"8080:8080"},
        },
      },
    },
  })
```

or only on specific devices, based on label selector:

```golang
  // Create an application that will be deployed on devices that have our specified label
  application, err := apiClient.CreateApplication(context.Background(), "default", synpse.Application{
    Name:        "app-name",
    Scheduling: synpse.Scheduling{
      Type: synpse.ScheduleTypeConditional,
      Selector: {
        "location": "power-plant",
      }
    },
    Spec: synpse.ApplicationSpec{
      ContainerSpec: []synpse.ContainerSpec{
        {
          Name:  "hello",
          Image: "quay.io/synpse/hello-synpse-go:latest",
          Ports: []string{"8080:8080"},
        },
      },
    },
  })
```

### Update Application

During the normal lifecycle, you will be updating application many times. For example if you want to update the Docker image or expose additional ports, use the `UpdateApplication` method:

```golang
  // Create an application that will be deployed on devices that have our specified label
  application, err := apiClient.UpdateApplication(context.Background(), "default", synpse.Application{
    ID:          app.ID,
    Name:        "app-name",
    Scheduling: synpse.Scheduling{
      Type: synpse.ScheduleTypeConditional,
      Selector: {
        "location": "power-plant",
      }
    },
    Spec: synpse.ApplicationSpec{
      ContainerSpec: []synpse.ContainerSpec{
        {
          Name:  "hello",
          Image: "quay.io/synpse/hello-synpse-go:new",
          Ports: []string{
            "8080:8080",
            "8888:8888",            
          },
        },
      },
    },
  })
```

### List Applications

To list applications:

```golang
    applications, err := apiClient.ListApplications(
      context.Background(), 
      &synpse.ListApplicationsRequest{Namespace: "default"},
    )
    for _, app := range applications {
      fmt.Println(app.Name)
    }
```

### Delete Applications:

You can remove applications by using name or ID:

```golang
  err := apiClient.DeleteApplication(context.Background(), "default", "app-name")
```

