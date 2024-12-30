package modules

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServersRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of server IDs to fetch. If not provided, all servers will be fetched.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"servers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of servers with their details.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the server.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the server.",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address of the server.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the server.",
						},
					},
				},
			},
		},
	}
}

func dataSourceServersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*providerConfig)

	// Get optional IDs from configuration
	rawIDs := d.Get("id").([]interface{})
	serverIDs := make([]int, len(rawIDs))
	for i, rawID := range rawIDs {
		serverIDs[i] = rawID.(int)
	}

	// Fetch servers
	var servers []Server
	var err error

	if len(serverIDs) == 0 {
		// Fetch all servers
		servers, err = fetchAllServers(cfg)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to fetch all servers: %v", err))
		}
	} else {
		// Fetch specific servers by ID
		for _, serverID := range serverIDs {
			server, err := fetchServerByID(cfg, serverID)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to fetch server with ID %d: %v", serverID, err))
			}
			servers = append(servers, server)
		}
	}

	// Set the servers in the Terraform state
	serverList := make([]map[string]interface{}, len(servers))
	for i, server := range servers {
		serverList[i] = map[string]interface{}{
			"id":     server.Number,
			"name":   server.Name,
			"ip":     server.IP,
			"status": server.Status,
		}
	}
	if err := d.Set("servers", serverList); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set servers state: %v", err))
	}

	// Set resource ID (use a static value as this is a data source)
	d.SetId("hetzner-servers")

	return nil
}
