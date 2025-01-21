package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"hcloud-robot-provider/client"
)

func ResourceFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirewallCreate,
		ReadContext:   resourceFirewallRead,
		UpdateContext: resourceFirewallUpdate,
		DeleteContext: resourceFirewallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFirewallImportState,
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the server to which the firewall will be applied.",
			},
			"active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the firewall is active.",
			},
			"whitelist_hos": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether to whitelist Hetzner services.",
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of the firewall rule.",
						},
						"dst_ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination IP address.",
						},
						"dst_port": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port.",
						},
						"src_ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source IP address.",
						},
						"src_port": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Protocol (e.g., tcp, udp).",
						},
						"tcp_flags": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "TCP flags.",
						},
						"action": {
							Type:             schema.TypeString,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"accept", "discard"}, false)),
							Required:         true,
							Description:      "Action to take (accept or discard).",
						},
					},
				},
			},
		},
	}
}

func resourceFirewallImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return nil, fmt.Errorf("meta is not of type *client.HetznerRobotClient")
	}
	serverID := d.Id()
	serverIDInt, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server_id %s: %w", serverID, err)
	}
	server, err := hClient.FetchServerByID(serverIDInt)
	if err != nil {
		return nil, fmt.Errorf("error fetching server by ID %s: %w", serverID, err)
	}
	serverIP := server.IP
	firewall, err := hClient.GetFirewall(ctx, serverIP)
	if err != nil {
		return nil, fmt.Errorf("could not find firewall for server ID %s: %w", serverID, err)
	}
	active := firewall.Status == "active"
	rules := make([]map[string]interface{}, 0, len(firewall.Rules.Input))
	for _, rule := range firewall.Rules.Input {
		r := map[string]interface{}{
			"name":      rule.Name,
			"src_ip":    rule.SrcIP,
			"src_port":  rule.SrcPort,
			"dst_ip":    rule.DstIP,
			"dst_port":  rule.DstPort,
			"protocol":  rule.Protocol,
			"tcp_flags": rule.TCPFlags,
			"action":    rule.Action,
		}
		rules = append(rules, r)
	}
	d.Set("active", active)
	d.Set("rule", rules)
	d.Set("whitelist_hos", firewall.WhitelistHetznerServices)
	d.Set("server_id", serverID)
	d.SetId(serverID)
	return []*schema.ResourceData{d}, nil
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}
	serverID := d.Get("server_id").(string)
	serverIDInt, err := strconv.Atoi(serverID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server_id %s: %w", serverID, err))
	}
	server, err := hClient.FetchServerByID(serverIDInt)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching server by ID %s: %w", serverID, err))
	}
	serverIP := server.IP
	status := "disabled"
	if d.Get("active").(bool) {
		status = "active"
	}
	rules := buildFirewallRules(d.Get("rule").([]interface{}))
	if err := hClient.SetFirewall(ctx, client.HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: d.Get("whitelist_hos").(bool),
		Status:                   status,
		Rules:                    client.HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(serverID)
	return diags
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}
	serverID := d.Get("server_id").(string)
	serverIDInt, err := strconv.Atoi(serverID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server_id %s: %w", serverID, err))
	}
	server, err := hClient.FetchServerByID(serverIDInt)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching server by ID %s: %w", serverID, err))
	}
	serverIP := server.IP
	firewall, err := hClient.GetFirewall(ctx, serverIP)
	if err != nil {
		return diag.FromErr(err)
	}
	active := firewall.Status == "active"
	rules := make([]map[string]interface{}, 0, len(firewall.Rules.Input))
	for _, rule := range firewall.Rules.Input {
		r := map[string]interface{}{
			"name":      rule.Name,
			"src_ip":    rule.SrcIP,
			"src_port":  rule.SrcPort,
			"dst_ip":    rule.DstIP,
			"dst_port":  rule.DstPort,
			"protocol":  rule.Protocol,
			"tcp_flags": rule.TCPFlags,
			"action":    rule.Action,
		}
		rules = append(rules, r)
	}
	d.Set("active", active)
	d.Set("rule", rules)
	d.Set("whitelist_hos", firewall.WhitelistHetznerServices)
	d.Set("server_id", serverID)
	return diags
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}
	serverID := d.Get("server_id").(string)
	serverIDInt, err := strconv.Atoi(serverID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server_id %s: %w", serverID, err))
	}
	server, err := hClient.FetchServerByID(serverIDInt)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching server by ID %s: %w", serverID, err))
	}
	serverIP := server.IP
	status := "disabled"
	if d.Get("active").(bool) {
		status = "active"
	}
	rules := buildFirewallRules(d.Get("rule").([]interface{}))
	if err := hClient.SetFirewall(ctx, client.HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: d.Get("whitelist_hos").(bool),
		Status:                   status,
		Rules:                    client.HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}
	serverID := d.Get("server_id").(string)
	serverIDInt, err := strconv.Atoi(serverID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server_id %s: %w", serverID, err))
	}
	server, err := hClient.FetchServerByID(serverIDInt)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching server by ID %s: %w", serverID, err))
	}
	serverIP := server.IP
	status := "active"
	rules := []client.HetznerRobotFirewallRule{
		{
			Name:     "Allow all",
			Action:   "accept",
			Protocol: "null",
		},
	}
	if err := hClient.SetFirewall(ctx, client.HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: false,
		Status:                   status,
		Rules:                    client.HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(fmt.Errorf("error updating firewall for server %s: %w", serverID, err))
	}
	d.SetId("")
	return diags
}

func buildFirewallRules(ruleList []interface{}) []client.HetznerRobotFirewallRule {
	rules := make([]client.HetznerRobotFirewallRule, 0, len(ruleList))
	for _, ruleMap := range ruleList {
		ruleProps := ruleMap.(map[string]interface{})
		rules = append(rules, client.HetznerRobotFirewallRule{
			Name:     ruleProps["name"].(string),
			SrcIP:    ruleProps["src_ip"].(string),
			SrcPort:  ruleProps["src_port"].(string),
			DstIP:    ruleProps["dst_ip"].(string),
			DstPort:  ruleProps["dst_port"].(string),
			Protocol: ruleProps["protocol"].(string),
			TCPFlags: ruleProps["tcp_flags"].(string),
			Action:   ruleProps["action"].(string),
		})
	}
	return rules
}
