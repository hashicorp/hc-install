# hcinstall

**DO NOT USE: WIP**

An **experimental** module for downloading or locating HashiCorp binaries, verifying signatures and checksums, and asserting version constraints.

This module is a successor to tfinstall, available in pre-1.0 versions of [terraform-exec](https://github.com/hashicorp/terraform-exec). Current users of tfinstall are advised to move to hcinstall on upgrading terraform-exec to v1.0.0.

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
