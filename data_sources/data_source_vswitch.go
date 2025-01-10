package data_sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"hcloud-robot-provider/client"
	"hcloud-robot-provider/shared"
)

// DataSourceVSwitches определяет новый источник данных для обработки нескольких VSwitch по их ID.
func DataSourceVSwitches() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSwitchesRead,
		Schema: map[string]*schema.Schema{
			"ids": { // Изменено с "id" на "ids" для отражения множества значений
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"vswitches": { // Новый атрибут для хранения списка VSwitch
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vlan": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cancelled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"servers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"server_ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"server_ipv6_net": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"server_number": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"status": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"subnets": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"mask": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"gateway": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"cloud_networks": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"mask": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"gateway": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceVSwitchesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta is not of type *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)

	// Получаем ids из входных данных
	idsInterface := d.Get("ids").([]interface{})
	var ids []string
	for _, id := range idsInterface {
		switch v := id.(type) {
		case int:
			ids = append(ids, fmt.Sprintf("%d", v)) // Преобразуем int в строку
		case string:
			ids = append(ids, v) // Если это строка, добавляем как есть
		default:
			return diag.Errorf("Invalid type for 'ids': expected int or string, got %T", id)
		}
	}

	var vswitches []client.VSwitch

	for _, id := range ids {
		vswitch, err := api.GetVSwitchByID(id)
		if err != nil {
			if _, ok := err.(*client.NotFoundError); ok {
				diagWarn := diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "VSwitch not found",
					Detail:   fmt.Sprintf("VSwitch with ID %s not found and will be skipped.", id),
				}
				diags = append(diags, diagWarn)
				continue
			}
			return diag.FromErr(fmt.Errorf("error reading VSwitch with ID %s: %w", id, err))
		}
		vswitches = append(vswitches, *vswitch)
	}

	if len(vswitches) == 0 {
		d.SetId("")
		return diags
	}

	d.SetId(fmt.Sprintf("vswitches-%s", strings.Join(ids, "-")))

	if err := d.Set("vswitches", flattenVSwitches(vswitches)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenVSwitches(vswitches []client.VSwitch) []map[string]interface{} {
	var result []map[string]interface{}
	for _, vswitch := range vswitches {
		result = append(result, map[string]interface{}{
			"id":             fmt.Sprintf("%d", vswitch.ID),
			"name":           vswitch.Name,
			"vlan":           vswitch.VLAN,
			"cancelled":      vswitch.Cancelled,
			"servers":        flattenVSwitchServers(vswitch.Servers),
			"subnets":        flattenVSwitchSubnets(vswitch.Subnets),
			"cloud_networks": flattenVSwitchCloudNetworks(vswitch.CloudNetworks),
		})
	}
	return result
}

func flattenVSwitchServers(servers []client.VSwitchServer) []map[string]interface{} {
	var result []map[string]interface{}
	for _, server := range servers {
		result = append(result, map[string]interface{}{
			"server_ip":       server.ServerIP,
			"server_ipv6_net": server.ServerIPv6Net,
			"server_number":   server.ServerNumber,
			"status":          server.Status,
		})
	}
	return result
}

func flattenVSwitchSubnets(subnets []client.VSwitchSubnet) []map[string]interface{} {
	var result []map[string]interface{}
	for _, subnet := range subnets {
		result = append(result, map[string]interface{}{
			"ip":      subnet.IP,
			"mask":    subnet.Mask,
			"gateway": subnet.Gateway,
		})
	}
	return result
}

func flattenVSwitchCloudNetworks(cloudNetworks []client.VSwitchCloudNetwork) []map[string]interface{} {
	var result []map[string]interface{}
	for _, network := range cloudNetworks {
		result = append(result, map[string]interface{}{
			"id":      network.ID,
			"ip":      network.IP,
			"mask":    network.Mask,
			"gateway": network.Gateway,
		})
	}
	return result
}
