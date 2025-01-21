package resources

import (
	"context"
	"fmt"
	"math/rand"
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
				Optional:    true,
				Description: "The VLAN ID for the vSwitch. If not provided, one will be chosen randomly from [4000..4091].",
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
	c := meta.(*client.HetznerRobotClient)

	name := d.Get("name").(string)

	vlan, vlanProvided := d.GetOk("vlan")
	var chosenVLAN int

	if vlanProvided {
		chosenVLAN = vlan.(int)
	} else {
		freeVLAN, err := pickRandomFreeVLAN(ctx, c)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to pick random free VLAN: %w", err))
		}
		chosenVLAN = freeVLAN

		_ = d.Set("vlan", chosenVLAN)
	}

	vsw, err := c.CreateVSwitch(ctx, name, chosenVLAN)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating vSwitch: %w", err))
	}

	if servers, ok := d.GetOk("servers"); ok {
		serverIDs := parseServerIDs(servers.([]interface{}))
		serverObjects := parseServerIDsToVSwitchServers(serverIDs)
		if len(serverObjects) > 0 {
			if err := c.AddVSwitchServers(ctx, strconv.Itoa(vsw.ID), serverObjects); err != nil {
				return diag.FromErr(fmt.Errorf("error adding servers to vSwitch: %w", err))
			}
		}
	}

	d.SetId(strconv.Itoa(vsw.ID))

	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.HetznerRobotClient)
	id := d.Id()

	vsw, err := c.FetchVSwitchByIDWithContext(ctx, id)
	if err != nil {
		if c.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading vSwitch: %w", err))
	}

	storedVLAN, vlanExists := d.GetOk("vlan")
	if vlanExists {
		if storedVLAN.(int) != vsw.VLAN {
			return diag.FromErr(fmt.Errorf(
				"VLAN mismatch: state VLAN = %d, robot VLAN = %d", storedVLAN, vsw.VLAN,
			))
		}
	}

	_ = d.Set("name", vsw.Name)
	_ = d.Set("vlan", vsw.VLAN)
	_ = d.Set("is_cancelled", vsw.Cancelled)
	_ = d.Set("servers", flattenServers(vsw.Servers))

	return nil
}

func resourceVSwitchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.HetznerRobotClient)
	id := d.Id()

	if d.HasChange("name") || d.HasChange("vlan") {
		name := d.Get("name").(string)
		vlan := d.Get("vlan").(int)

		if err := c.UpdateVSwitch(ctx, id, name, vlan); err != nil {
			return diag.FromErr(fmt.Errorf("error updating vSwitch: %w", err))
		}
	}

	if d.HasChange("is_cancelled") {
		cancellationDate := d.Get("is_cancelled").(string)
		if cancellationDate == "" {
			cancellationDate = time.Now().Format("2006-01-02")
		}
		if err := c.SetVSwitchCancellation(ctx, id, cancellationDate); err != nil {
			return diag.FromErr(fmt.Errorf("error updating cancellation date: %w", err))
		}
	}

	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.HetznerRobotClient)
	id := d.Id()

	cancellationDate := d.Get("is_cancelled").(string)
	if cancellationDate == "" {
		cancellationDate = "now"
	}

	if err := c.DeleteVSwitch(ctx, id, cancellationDate); err != nil {
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
	for _, s := range servers {
		result = append(result, s.ServerNumber)
	}
	return result
}

func pickRandomFreeVLAN(ctx context.Context, c *client.HetznerRobotClient) (int, error) {
	vswitches, err := c.FetchAllVSwitches(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch all vswitches error: %w", err)
	}

	used := make(map[int]bool)
	for _, v := range vswitches {
		used[v.VLAN] = true
	}

	var free []int
	for vlan := 4000; vlan <= 4091; vlan++ {
		if !used[vlan] {
			free = append(free, vlan)
		}
	}
	if len(free) == 0 {
		return 0, fmt.Errorf("no free VLAN in [4000..4091], all are taken")
	}

	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(free))
	return free[idx], nil
}
