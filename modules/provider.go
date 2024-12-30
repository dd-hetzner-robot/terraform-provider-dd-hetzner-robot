package modules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type providerConfig struct {
	username string
	password string
	url      string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_PASSWORD", nil),
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_URL", "https://robot-ws.your-server.de"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hetzner-robot_server_bulk":     resourceTPServerBulk(),
			"hetzner-robot_talos_installer": resourceTalosInstaller(),
			"hetzner-robot_firewall":        resourceFirewall(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hetzner-robot_servers": dataSourceServers(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	url := d.Get("url").(string)

	var diags diag.Diagnostics

	if username == "" || password == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing required configuration",
			Detail:   "Both username and password must be provided.",
		})
		return nil, diags
	}

	return &providerConfig{
		username: username,
		password: password,
		url:      url,
	}, diags
}
