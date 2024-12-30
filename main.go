package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"hcloud-robot-provider/modules"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: modules.Provider,
	})
}
