# hcinstall

**DO NOT USE: WIP**

An **experimental** Go module for downloading or locating HashiCorp binaries, verifying signatures and checksums, and asserting version constraints.

This module is a successor to tfinstall, available in pre-1.0 versions of [terraform-exec](https://github.com/hashicorp/terraform-exec). Current users of tfinstall are advised to move to hcinstall on upgrading terraform-exec to v1.0.0.

## hcinstall is not a package manager

This library is intended for use within Go programs which have some business downloading or otherwise locating HashiCorp binaries.

The included command-line utility, `hcinstall`, is a convenient way of using the library in ad-hoc or CI shell scripting.

hcinstall will not:
 - Install binaries to the appropriate place in your operating system. It does not know whether you think `terraform` should go in `/usr/bin` or `/usr/local/bin`, and does not want to get involved in the discussion.
 - Upgrade existing binaries on your system by overwriting them in place.
 - Add downloaded binaries to your `PATH`.

## API

Loosely inspired by [go-getter](https://github.com/hashicorp/go-getter), the API provides:

 - Simple one-line `Install()` function for locating a product binary of a given, or latest, version, with sensible defaults.
 - Customisable `Client`:
   - Version constraint parsing
   - Tries each `Getter` in turn to locate a binary matching version constraints
   - Verifies downloaded binary signatures and checksums

### Simple

```go
package main

import (
  "fmt"
  
  "github.com/hashicorp/hcinstall")
)

func main() {
  tfPath, err := hcinstall.Install(context.Background(), "", hcinstall.ProductTerraform, "0.13.5", true)
  if err != nil {
    panic(err)
  }
  fmt.Printf("Path to Terraform binary: %s", tfPath)
}
```

### Advanced

```go
package main

import (
  "fmt"
  
  "github.com/hashicorp/hcinstall"
)

func main() {
  v, err := NewVersionConstraints("0.13.5", true)
  if err != nil {
    panic(err)
  }

  client := &hcinstall.Client{
    Product: hcinstall.ProductTerraform,
    InstallDir: "/usr/local/bin",
    Getters: []Getter{hcinstall.LookPath(), hcinstall.Releases()},
    VersionConstraints: v,
  }
  
  tfPath, err := client.Install(context.Background())
  if err != nil {
    panic(err)
  }
  
  fmt.Printf("Path to Terraform binary: %s", tfPath)
}
```
