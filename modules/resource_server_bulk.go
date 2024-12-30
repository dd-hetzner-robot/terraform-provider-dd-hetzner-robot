package modules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTPServerBulk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTPServerBulkCreate,
		ReadContext:   resourceTPServerBulkRead,
		UpdateContext: resourceTPServerBulkUpdate,
		DeleteContext: resourceTPServerBulkDelete,

		Schema: map[string]*schema.Schema{
			"servers": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				Description: "List of servers as array of objects. Example: servers = [{ id = \"2551828\", name = \"foo\" }]",
			},
		},
	}
}

func resourceTPServerBulkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*providerConfig)
	rawServers := d.Get("servers").([]interface{})
	servers := expandServerList(rawServers)

	for _, srv := range servers {
		serverID, err := strconv.Atoi(srv.ID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("invalid server ID %s: must be an integer", srv.ID))
		}

		if _, err := fetchServerByID(cfg, serverID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to fetch server with ID %d: %v", serverID, err))
		}

		if srv.Name != "" {
			if err := renameServer(cfg, serverID, srv.Name); err != nil {
				return diag.FromErr(fmt.Errorf("failed to rename server with ID %d: %v", serverID, err))
			}
		}
	}

	d.SetId("tp-server-bulk")
	return resourceTPServerBulkRead(ctx, d, meta)
}

func resourceTPServerBulkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*providerConfig)
	rawServers := d.Get("servers").([]interface{})
	servers := expandServerList(rawServers)

	currentState := make([]map[string]string, 0)

	for _, srv := range servers {
		serverID, err := strconv.Atoi(srv.ID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("invalid server ID %s: must be an integer", srv.ID))
		}

		server, err := fetchServerByID(cfg, serverID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to fetch server with ID %d: %v", serverID, err))
		}

		currentState = append(currentState, map[string]string{
			"id":   strconv.Itoa(serverID),
			"name": server.Name,
		})
	}

	// Compare current state with desired state
	if err := d.Set("servers", currentState); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set servers state: %v", err))
	}

	return nil
}

func resourceTPServerBulkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("servers") {
		return resourceTPServerBulkCreate(ctx, d, meta)
	}
	return nil
}

func resourceTPServerBulkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
