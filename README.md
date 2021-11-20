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

## Prerequisites

- [Go](https://golang.org/dl/)
- [Synpse account](https://cloud.synpse.net/) - free up to 5 devices.

## Installation

```shell
go get github.com/synpse-hq/synpse-go
```

## Authentication

To authenticate, we recommend using a project level access key that you can generate here: https://cloud.synpse.net/service-accounts.

Alternatively, use [Personal Access Keys](https://cloud.synpse.net/access-keys), however, they will be able to manage multiple projects.
