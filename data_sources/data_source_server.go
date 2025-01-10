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

func DataSourceServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServersRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip":     {Type: schema.TypeString, Computed: true},
						"name":   {Type: schema.TypeString, Computed: true},
						"number": {Type: schema.TypeInt, Computed: true},
						"status": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceServersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta is not of type *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)

	idsInterface := d.Get("ids").([]interface{})
	var ids []int
	for _, id := range idsInterface {
		ids = append(ids, id.(int))
	}

	servers, err := api.FetchServersByIDs(ids)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch servers: %w", err))
	}

	var serverList []map[string]interface{}
	for _, server := range servers {
		serverList = append(serverList, map[string]interface{}{
			"ip":     server.IP,
			"name":   server.Name,
			"number": server.Number,
			"status": server.Status,
		})
	}

	if err := d.Set("servers", serverList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("servers-%s", strings.Join(intSliceToStringSlice(ids), "-")))

	return diag.Diagnostics{}
}

func intSliceToStringSlice(intSlice []int) []string {
	stringSlice := make([]string, len(intSlice))
	for i, val := range intSlice {
		stringSlice[i] = fmt.Sprintf("%d", val)
	}
	return stringSlice
}
