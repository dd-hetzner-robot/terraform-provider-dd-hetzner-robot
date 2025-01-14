package data_sources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"hcloud-robot-provider/client"
	"hcloud-robot-provider/shared"
)

func DataSourceVSwitches() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSwitchesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"vswitches": {
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
					},
				},
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
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

	idsInterface := d.Get("ids").([]interface{})
	var ids []string
	for _, id := range idsInterface {
		ids = append(ids, id.(string))
	}

	var (
		foundIDs  []string
		vswitches []client.VSwitch
	)

	for _, id := range ids {
		vswitch, err := api.FetchVSwitchByIDWithContext(ctx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			return diag.FromErr(fmt.Errorf("error reading VSwitch with ID %s: %w", id, err))
		}
		vswitches = append(vswitches, *vswitch)
		foundIDs = append(foundIDs, id)
	}

	if len(vswitches) == 0 {
		d.SetId("No found")
		placeholder := []map[string]interface{}{
			{
				"id":        "",
				"name":      "vswitches not found",
				"vlan":      0,
				"cancelled": false,
			},
		}
		if err := d.Set("vswitches", placeholder); err != nil {
			return diag.FromErr(err)
		}
		return diags
	}

	d.SetId(fmt.Sprintf("vswitches-%s", strings.Join(foundIDs, "-")))
	if err := d.Set("vswitches", flattenVSwitches(vswitches)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenVSwitches(vswitches []client.VSwitch) []map[string]interface{} {
	var result []map[string]interface{}
	for _, vswitch := range vswitches {
		result = append(result, map[string]interface{}{
			"id":        strconv.Itoa(vswitch.ID),
			"name":      vswitch.Name,
			"vlan":      vswitch.VLAN,
			"cancelled": vswitch.Cancelled,
		})
	}
	return result
}
