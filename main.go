package main

import (
    "github.com/daleh-daemon/packer-vpc-datasoruce-plugin/datasources"
    "github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
    pps := plugin.NewSet()
    pps.RegisterDatasource("aws_vpc", new(datasources.Datasource))
}
