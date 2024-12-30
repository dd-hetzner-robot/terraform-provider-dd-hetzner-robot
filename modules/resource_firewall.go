package modules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirewallCreate,
		ReadContext:   resourceFirewallRead,
		UpdateContext: resourceFirewallUpdate,
		DeleteContext: resourceFirewallDelete,
		Schema: map[string]*schema.Schema{
			"server": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the server.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the server.",
						},
					},
				},
				Description: "List of servers to configure firewall.",
			},
			"whitelist_hos": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable or disable whitelist of Hetzner services.",
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Rule name.",
						},
						"dst_port": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Protocol (e.g., tcp, udp).",
						},
						"action": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Action to apply (e.g., accept, discard).",
						},
					},
				},
				Description: "List of firewall rules.",
			},
		},
	}
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*HetznerRobotClient)

	servers := expandServers(d.Get("server").([]interface{}))
	rules := expandRules(d.Get("rule").([]interface{}))
	whitelistHOS := d.Get("whitelist_hos").(bool)

	for _, server := range servers {
		err := client.setFirewall(ctx, HetznerRobotFirewall{
			IP:                       server.ID,
			WhitelistHetznerServices: whitelistHOS,
			Status:                   "active",
			Rules:                    HetznerRobotFirewallRules{Input: rules},
		})
		if err != nil {
			return diag.Errorf("failed to configure firewall for server %s: %v", server.ID, err)
		}
	}

	d.SetId("firewall-configuration")

	return resourceFirewallRead(ctx, d, meta)
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*HetznerRobotClient)
	servers := expandServers(d.Get("server").([]interface{}))

	var diags diag.Diagnostics

	for _, server := range servers {
		firewall, err := client.getFirewall(ctx, server.ID)
		if err != nil {
			return diag.Errorf("failed to retrieve firewall configuration for server %s: %v", server.ID, err)
		}

		_ = d.Set("whitelist_hos", firewall.WhitelistHetznerServices)
	}

	return diags
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceFirewallCreate(ctx, d, meta)
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func expandServers(raw []interface{}) []ServerInput {
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

func expandRules(raw []interface{}) []HetznerRobotFirewallRule {
	rules := make([]HetznerRobotFirewallRule, len(raw))
	for i, r := range raw {
		m := r.(map[string]interface{})
		rules[i] = HetznerRobotFirewallRule{
			Name:     m["name"].(string),
			DstPort:  m["dst_port"].(string),
			Protocol: m["protocol"].(string),
			Action:   m["action"].(string),
		}
	}
	return rules
}
