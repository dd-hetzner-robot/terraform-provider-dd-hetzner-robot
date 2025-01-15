package resources

import (
	"context"
	"fmt"

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
			"server_ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"whitelist_hos": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dst_ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dst_port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"src_ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"src_port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tcp_flags": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"action": {
							Type: schema.TypeString,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
								"accept",
								"discard",
							}, false)),
							Required: true,
						},
					},
				},
			},
		},
	}
}

// Импорт ресурса
func resourceFirewallImportState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return nil, fmt.Errorf("meta is not of type *client.HetznerRobotClient")
	}

	firewallID := d.Id()

	firewall, err := hClient.GetFirewall(ctx, firewallID)
	if err != nil {
		return nil, fmt.Errorf("could not find firewall with ID %s: %w", firewallID, err)
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
	d.Set("server_ip", firewall.IP)
	d.Set("whitelist_hos", firewall.WhitelistHetznerServices)
	d.SetId(firewall.IP)

	return []*schema.ResourceData{d}, nil
}

// Создание Firewall
func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}

	serverIP := d.Get("server_ip").(string)
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

	d.SetId(serverIP)
	return nil
}

// Чтение Firewall
func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}

	serverIP := d.Id()

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
	d.Set("server_ip", firewall.IP)
	d.Set("whitelist_hos", firewall.WhitelistHetznerServices)

	return nil
}

// Обновление Firewall
func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hClient, ok := m.(*client.HetznerRobotClient)
	if !ok {
		return diag.Errorf("meta is not of type *client.HetznerRobotClient")
	}

	serverIP := d.Get("server_ip").(string)
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
	return nil
}

// Удаление Firewall (здесь можно выключать/сбрасывать правила)
func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Пока пустая логика. При желании можно выключать firewall, если API это позволяет.
	return nil
}

// Преобразуем []interface{} из Terraform в []HetznerRobotFirewallRule
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
