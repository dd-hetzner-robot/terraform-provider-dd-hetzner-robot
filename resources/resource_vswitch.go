package resources

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"hcloud-robot-provider/client"
)

func ResourceVSwitch() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSwitchCreate,
		ReadContext:   resourceVSwitchRead,
		UpdateContext: resourceVSwitchUpdate,
		DeleteContext: resourceVSwitchDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vSwitch.",
			},
			"vlan": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The VLAN ID for the vSwitch.",
			},
			"servers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of server IDs to connect to the vSwitch.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"is_cancelled": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The cancellation date for the vSwitch. If not provided, defaults to 'now'.",
			},
		},
	}
}

func resourceVSwitchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.HetznerRobotClient)

	name := d.Get("name").(string)
	vlan := d.Get("vlan").(int)

	vswitch, err := client.CreateVSwitch(ctx, name, vlan)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating vSwitch: %w", err))
	}

	if servers, ok := d.GetOk("servers"); ok {
		serverIDs := parseServerIDs(servers.([]interface{}))
		serverObjects := parseServerIDsToVSwitchServers(serverIDs)
		if len(serverObjects) > 0 {
			if err := client.AddVSwitchServers(ctx, strconv.Itoa(vswitch.ID), serverObjects); err != nil {
				return diag.FromErr(fmt.Errorf("error adding servers to vSwitch: %w", err))
			}
		}
	}

	d.SetId(fmt.Sprintf("%d", vswitch.ID))
	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.HetznerRobotClient)

	id := d.Id()
	vswitch, err := client.FetchVSwitchByIDWithContext(ctx, id)
	if err != nil {
		if client.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading vSwitch: %w", err))
	}

	d.Set("name", vswitch.Name)
	d.Set("vlan", vswitch.VLAN)
	d.Set("is_cancelled", vswitch.Cancelled)

	serverIDs := flattenServers(vswitch.Servers)
	if err := d.Set("servers", serverIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting servers for vSwitch: %w", err))
	}

	return nil
}

func resourceVSwitchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.HetznerRobotClient)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("vlan") {
		name := d.Get("name").(string)
		vlan := d.Get("vlan").(int)
		if err := client.UpdateVSwitch(ctx, id, name, vlan); err != nil {
			return diag.FromErr(fmt.Errorf("error updating vSwitch: %w", err))
		}
	}

	if d.HasChange("is_cancelled") {
		cancellationDate := d.Get("is_cancelled").(string)
		if cancellationDate == "" {
			cancellationDate = time.Now().Format("2006-01-02") // Default to current date
		}
		if err := client.SetVSwitchCancellation(ctx, id, cancellationDate); err != nil {
			return diag.FromErr(fmt.Errorf("error updating cancellation date: %w", err))
		}
	}

	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.HetznerRobotClient)

	id := d.Id()

	cancellationDate := d.Get("is_cancelled").(string)
	if cancellationDate == "" {
		cancellationDate = "now"
	}

	if err := client.DeleteVSwitch(ctx, id, cancellationDate); err != nil {
		return diag.FromErr(fmt.Errorf("error deleting vSwitch: %w", err))
	}

	d.SetId("")
	return nil
}

func parseServerIDs(servers []interface{}) []int {
	var result []int
	for _, s := range servers {
		result = append(result, s.(int))
	}
	return result
}

func parseServerIDsToVSwitchServers(serverIDs []int) []client.VSwitchServer {
	var servers []client.VSwitchServer
	for _, id := range serverIDs {
		servers = append(servers, client.VSwitchServer{ServerNumber: id})
	}
	return servers
}

func flattenServers(servers []client.VSwitchServer) []int {
	var result []int
	for _, server := range servers {
		result = append(result, server.ServerNumber)
	}
	return result
}
