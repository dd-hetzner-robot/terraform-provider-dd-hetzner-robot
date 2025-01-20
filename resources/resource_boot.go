package resources

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"hcloud-robot-provider/client"
)

type ServerInput struct {
	ID   string
	Name string
}

func expandServerList(raw []interface{}) []ServerInput {
	servers := make([]ServerInput, len(raw))
	for i, r := range raw {
		m := r.(map[string]interface{})
		servers[i] = ServerInput{
			ID:   m["id"].(string),
			Name: m["name"].(string),
		}
	}
	return servers
}

func ResourceBootInstaller() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBootInstallerCreate,
		ReadContext:   schema.NoopContext,
		UpdateContext: schema.NoopContext,
		DeleteContext: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"servers": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":   {Type: schema.TypeString, Required: true},
						"name": {Type: schema.TypeString, Required: true},
					},
				},
			},
			"os": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "linux",
				Description: "Operating system for rescue mode (linux, freebsd, etc.).",
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":       {Type: schema.TypeString, Computed: true},
						"ip":       {Type: schema.TypeString, Computed: true},
						"password": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func resourceBootInstallerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*client.HetznerRobotClient)
	rawServers := d.Get("servers").([]interface{})
	servers := expandServerList(rawServers)
	osType := d.Get("os").(string)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []map[string]interface{}
	var diags diag.Diagnostics

	for _, srv := range servers {
		wg.Add(1)

		go func(srv ServerInput) {
			defer wg.Done()

			serverID, _ := strconv.Atoi(srv.ID)

			_, err := cfg.RenameServer(ctx, serverID, srv.Name)
			if err != nil {
				mu.Lock()
				diags = append(diags, diag.Errorf("failed to rename server %d: %v", serverID, err)...)
				mu.Unlock()
				return
			}

			rescueResponse, err := cfg.EnableRescueMode(ctx, serverID, osType, nil)
			if err != nil {
				mu.Lock()
				diags = append(diags, diag.Errorf("failed to enable rescue mode for server %d: %v", serverID, err)...)
				mu.Unlock()
				return
			}

			cfg.ResetServer(ctx, serverID, "hw")
			time.Sleep(10 * time.Second)
			cfg.ResetServer(ctx, serverID, "hw")

			mu.Lock()
			results = append(results, map[string]interface{}{
				"id":       srv.ID,
				"ip":       rescueResponse.Rescue.ServerIP,
				"password": rescueResponse.Rescue.Password,
			})
			mu.Unlock()

		}(srv)
	}

	wg.Wait()
	d.SetId("talos-installer")
	d.Set("results", results)

	return diags
}
