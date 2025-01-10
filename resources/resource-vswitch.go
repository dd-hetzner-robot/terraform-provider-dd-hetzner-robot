package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"hcloud-robot-provider/client"
	"hcloud-robot-provider/shared"
)

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

func resourceVSwitchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta не является *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)

	name := d.Get("name").(string)
	vlan := d.Get("vlan").(int)

	vswitch, err := api.CreateVSwitch(name, vlan)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка при создании VSwitch: %w", err))
	}

	d.SetId(fmt.Sprintf("%d", vswitch.ID))

	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta не является *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)

	id := d.Id()

	vswitch, err := api.GetVSwitchByID(id)
	if err != nil {
		if _, ok := err.(*client.NotFoundError); ok {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("ошибка при чтении VSwitch: %w", err))
	}

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

func resourceVSwitchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta не является *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)

	id := d.Id()

	name := d.Get("name").(string)
	vlan := d.Get("vlan").(int)

	if err := api.UpdateVSwitch(id, name, vlan); err != nil {
		return diag.FromErr(fmt.Errorf("ошибка при обновлении VSwitch: %w", err))
	}

	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config, ok := meta.(*shared.ProviderConfig)
	if !ok {
		return diag.Errorf("meta не является *shared.ProviderConfig")
	}

	api := client.NewHetznerRobotClient(config)
	id := d.Id()

	if err := api.DeleteVSwitch(id); err != nil {
		return diag.FromErr(fmt.Errorf("ошибка при удалении VSwitch: %w", err))
	}

	d.SetId("") // Удаляем ресурс из состояния Terraform

	return diag.Diagnostics{}
}
