package main

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		n, err := compute.NewNetwork(ctx, "network", &compute.NetworkArgs{})
		if err != nil {
			return err
		}

		// Config is set by the server!
		cidr, ok := ctx.GetConfig("poc:whitelist")
		if !ok {
			return fmt.Errorf("must set whitelist in config")
		}
		_, err = compute.NewFirewall(ctx, "firewall", &compute.FirewallArgs{
			Network: n.ID(),
			SourceRanges: []interface{}{
				cidr,
			},
			Allows: []interface{}{
				map[string]interface{}{
					"protocol": "icmp",
				},
			},
		})
		if err != nil {
			return err
		}

		_, err = compute.NewInstance(ctx, "instance", &compute.InstanceArgs{
			MachineType: "f1-micro",
			Zone:        "us-central1-a",
			BootDisk: map[string]interface{}{
				"initializeParams": map[string]interface{}{
					"image": "debian-cloud/debian-9",
				},
			},
			NetworkInterfaces: []interface{}{map[string]interface{}{
				"network": n.ID(),
				"accessConfigs": []interface{}{
					map[string]interface{}{}},
			}},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
