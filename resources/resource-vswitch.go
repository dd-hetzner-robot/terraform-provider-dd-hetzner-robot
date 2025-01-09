package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"hcloud-robot-provider/client"
	"hcloud-robot-provider/shared"
)

// ResourceVSwitch defines the VSwitch resource for Terraform
func ResourceVSwitch() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSwitchCreate,
		ReadContext:   resourceVSwitchRead,
		UpdateContext: resourceVSwitchUpdate,
		DeleteContext: resourceVSwitchDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Required: true,
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
	}
}

// resourceVSwitchCreate creates a new VSwitch
func resourceVSwitchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta is not of type *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)
	name := d.Get("name").(string)
	vlan := d.Get("vlan").(int)

	vswitch, err := api.CreateVSwitch(name, vlan)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating VSwitch: %w", err))
	}

	d.SetId(fmt.Sprintf("%d", vswitch.ID))
	return resourceVSwitchRead(ctx, d, meta)
}

// resourceVSwitchRead reads the VSwitch information
func resourceVSwitchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta is not of type *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)
	id := d.Id()

	vswitch, err := api.GetVSwitchByID(id)
	if err != nil {
		// If VSwitch not found, remove it from state
		if _, ok := err.(*client.NotFoundError); ok {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading VSwitch: %w", err))
	}

	// Set field values
	if err := d.Set("name", vswitch.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vlan", vswitch.VLAN); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cancelled", vswitch.Cancelled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("servers", flattenVSwitchServers(vswitch.Servers)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("subnets", flattenVSwitchSubnets(vswitch.Subnets)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cloud_networks", flattenVSwitchCloudNetworks(vswitch.CloudNetworks)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceVSwitchUpdate updates an existing VSwitch
func resourceVSwitchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta is not of type *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)
	id := d.Id()
	name := d.Get("name").(string)
	vlan := d.Get("vlan").(int)

	if err := api.UpdateVSwitch(id, name, vlan); err != nil {
		return diag.FromErr(fmt.Errorf("error updating VSwitch: %w", err))
	}

	return resourceVSwitchRead(ctx, d, meta)
}

// resourceVSwitchDelete deletes a VSwitch
func resourceVSwitchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta is not of type *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)
	id := d.Id()

	if err := api.DeleteVSwitch(id); err != nil {
		return diag.FromErr(fmt.Errorf("error deleting VSwitch: %w", err))
	}

	d.SetId("")
	return diag.Diagnostics{}
}

// flattenVSwitchServers converts the list of VSwitch servers to the format suitable for Terraform
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

// flattenVSwitchSubnets converts the list of VSwitch subnets to the format suitable for Terraform
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

// flattenVSwitchCloudNetworks converts the list of VSwitch cloud networks to the format suitable for Terraform
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
