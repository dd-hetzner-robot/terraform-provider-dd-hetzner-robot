package modules

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTalosInstaller() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTalosInstallerCreate,
		ReadContext:   resourceTalosInstallerRead,
		UpdateContext: resourceTalosInstallerUpdate,
		DeleteContext: resourceTalosInstallerDelete,

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
			"results": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Results of the Talos installation process.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Server ID.",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Server IP address.",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Rescue mode password.",
						},
					},
				},
			},
		},
	}
}

func resourceTalosInstallerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*providerConfig)
	rawServers := d.Get("servers").([]interface{})
	servers := expandServerList(rawServers)

	var results []map[string]interface{}

	for _, srv := range servers {
		serverID, _ := strconv.Atoi(srv.ID)

		rescueResponse, err := EnableRescueMode(cfg, serverID)
		if err != nil {
			return diag.Errorf("failed to enable rescue mode for server %d: %v", serverID, err)
		}

		if err := RebootServer(cfg, serverID); err != nil {
			return diag.Errorf("failed to reboot server %d: %v", serverID, err)
		}

		if err := WaitForHost(rescueResponse.Rescue.ServerIP, 5*time.Minute); err != nil {
			return diag.Errorf("failed to wait for host %s: %v", rescueResponse.Rescue.ServerIP, err)
		}

		isInstalled, err := IsTalosInstalled(rescueResponse.Rescue.ServerIP)
		if err != nil {
			return diag.Errorf("failed to check Talos installation on server %d: %v", serverID, err)
		}

		if !isInstalled {
			sshOutput, err := ExecuteSSHCommand(
				rescueResponse.Rescue.ServerIP,
				"root",
				rescueResponse.Rescue.Password,
				"mkdir -p /123",
			)
			if err != nil {
				return diag.Errorf("failed to create folder on server %d: %v", serverID, err)
			}
			fmt.Printf("SSH Output for server %d: %s\n", serverID, sshOutput)
		} else {
			fmt.Printf("Talos is already installed on server %d\n", serverID)
		}

		if err := RebootServer(cfg, serverID); err != nil {
			return diag.Errorf("failed to reboot server %d: %v", serverID, err)
		}

		results = append(results, map[string]interface{}{
			"id":       strconv.Itoa(serverID),
			"ip":       rescueResponse.Rescue.ServerIP,
			"password": rescueResponse.Rescue.Password,
		})
	}

	d.SetId("talos-installer")
	if err := d.Set("results", results); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTalosInstallerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTalosInstallerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTalosInstallerCreate(ctx, d, meta)
}

func resourceTalosInstallerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
