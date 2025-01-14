package provider

import (
	"context"
	"hcloud-robot-provider/data_sources"
	"hcloud-robot-provider/resources"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"hcloud-robot-provider/client"
	"hcloud-robot-provider/shared"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_URL", "https://robot-ws.your-server.de"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hetznerrobot_vswitch": resources.ResourceVSwitch(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hetznerrobot_server":  data_sources.DataSourceServers(),
			"hetznerrobot_vswitch": data_sources.DataSourceVSwitches(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	url := d.Get("url").(string)

	if username == "" || password == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing credentials",
			Detail:   "Both username and password must be provided.",
		})
		return nil, diags
	}

	config := &shared.ProviderConfig{
		Username: username,
		Password: password,
		BaseURL:  url,
	}

	client := client.NewHetznerRobotClient(config)
	return client, diags
}
