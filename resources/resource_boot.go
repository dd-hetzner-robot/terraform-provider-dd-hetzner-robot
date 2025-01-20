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
		UpdateContext: resourceBootInstallerUpdate,
		DeleteContext: resourceBootInstallerDelete,
		Schema: map[string]*schema.Schema{
			"servers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of servers",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
					},
				},
			},
			"os": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "linux",
				Description: "Operating system for rescue mode (linux, freebsd, etc.).",
			},
			"rescue_os": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "linux",
				Description: "Operating system for rescue mode (e.g. linux, freebsd).",
			},
			"install_os": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Operating system to install (e.g. linux, talos).",
			},
			"install_os_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of the custom OS image to install.",
			},
			"ssh_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of SSH keys to be added during the rescue mode.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:     schema.TypeString,
							Computed: true,
						},
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
	rescueOS := d.Get("rescue_os").(string)
	sshKeysRaw := d.Get("ssh_keys").([]interface{})
	var sshKeys []string
	for _, key := range sshKeysRaw {
		sshKeys = append(sshKeys, key.(string))
	}
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []map[string]interface{}
		diags   diag.Diagnostics
	)
	for _, srv := range servers {
		wg.Add(1)
		go func(srv ServerInput) {
			defer wg.Done()
			serverID, err := strconv.Atoi(srv.ID)
			if err != nil {
				mu.Lock()
				diags = append(diags, diag.Errorf("invalid server ID %s: %v", srv.ID, err)...)
				mu.Unlock()
				return
			}
			_, err = cfg.RenameServer(ctx, serverID, srv.Name)
			if err != nil {
				mu.Lock()
				diags = append(diags, diag.Errorf("failed to rename server %d: %v", serverID, err)...)
				mu.Unlock()
				return
			}
			rescueResp, err := cfg.EnableRescueMode(ctx, serverID, rescueOS, sshKeys)
			if err != nil {
				mu.Lock()
				diags = append(diags, diag.Errorf("failed to enable rescue mode for server %d: %v", serverID, err)...)
				mu.Unlock()
				return
			}
			cfg.ResetServer(ctx, serverID, "hw")
			time.Sleep(10 * time.Second)
			cfg.ResetServer(ctx, serverID, "hw")
			ip := rescueResp.Rescue.ServerIP
			pass := rescueResp.Rescue.Password
			if err := cfg.InstallTalosOS(ctx, ip, pass); err != nil {
				mu.Lock()
				diags = append(diags, diag.Errorf("failed to install Talos OS on server %d: %v", serverID, err)...)
				mu.Unlock()
				return
			}
			mu.Lock()
			results = append(results, map[string]interface{}{
				"id":       srv.ID,
				"ip":       ip,
				"password": pass,
			})
			mu.Unlock()
		}(srv)
	}
	wg.Wait()
	d.SetId("talos-installer")
	d.Set("results", results)
	return diags
}

func resourceBootInstallerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*client.HetznerRobotClient)
	var diags diag.Diagnostics
	if d.HasChange("servers") {
		rawServers := d.Get("servers").([]interface{})
		servers := expandServerList(rawServers)
		for _, srv := range servers {
			serverID, err := strconv.Atoi(srv.ID)
			if err != nil {
				diags = append(diags, diag.Errorf("invalid server ID %s: %v", srv.ID, err)...)
				continue
			}
			serverInfo, err := cfg.FetchServerByID(serverID)
			if err != nil {
				diags = append(diags, diag.Errorf("failed to fetch server %d info: %v", serverID, err)...)
				continue
			}
			if serverInfo.ServerName != srv.Name {
				_, err = cfg.RenameServer(ctx, serverID, srv.Name)
				if err != nil {
					diags = append(diags, diag.Errorf("failed to rename server %d: %v", serverID, err)...)
					continue
				}
			}
		}
	}
	return diags
}

// -------------------- DELETE -------------------- //
func resourceBootInstallerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*client.HetznerRobotClient)
	rawServers := d.Get("servers").([]interface{})
	servers := expandServerList(rawServers)
	var diags diag.Diagnostics
	for _, srv := range servers {
		serverID, err := strconv.Atoi(srv.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("invalid server ID %s: %v", srv.ID, err)...)
			continue
		}
		_, err = cfg.ResetServer(ctx, serverID, "hw")
		if err != nil {
			diags = append(diags, diag.Errorf("failed to reset server %d on delete: %v", serverID, err)...)
		}
	}
	return diags
}
