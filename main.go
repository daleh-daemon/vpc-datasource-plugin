package main

import (
    "fmt"
    "github.com/daleh-daemon/packer-vpc-datasoruce-plugin/datasources"
    "github.com/hashicorp/packer-plugin-sdk/plugin"
    "github.com/hashicorp/packer-plugin-sdk/version"
    "os"
)

func main() {
    pps := plugin.NewSet()
    pps.SetVersion(version.InitializePluginVersion("0.0.1", "dev"))
    pps.RegisterDatasource("aws_vpc", new(datasources.Datasource))
    err := pps.Run()
    if err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
}
