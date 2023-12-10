package main

import (
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi-linode/sdk/v3/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a new linode instance
		instance, err := linode.NewInstance(ctx, "seafileInstance", &linode.InstanceArgs{
			Type:  pulumi.String("g6-standard-1"),
			Image: pulumi.String("linode/ubuntu20.04"),
		})
		if err != nil {
			return err
		}

		// Provision a block storage volume and attach to the instance.
		volume, err := linode.NewVolume(ctx, "seafileVolume", &linode.VolumeArgs{
			Size:     pulumi.Int(50),
			LinodeId: instance.ID(),
		})
		if err != nil {
			return err
		}

		// Define a Docker network to link the Seafile container with the volume.
		network, err := docker.NewNetwork(ctx, "seafileNetwork", &docker.NetworkArgs{})
		if err != nil {
			return err
		}

		// Define a Docker container that will run the Seafile image.
		_, err = docker.NewContainer(ctx, "seafileContainer", &docker.ContainerArgs{
			Image: pulumi.String("docker.io/znd4/seafile-deployment:1234"),

			// Assuming the Seafile application requires these standard ports to be exposed.
			Ports: &docker.ContainerPortArray{
				&docker.ContainerPortArgs{Internal: pulumi.Int(80)},
				&docker.ContainerPortArgs{Internal: pulumi.Int(443)},
			},

			// Configure the volume created above to be attached to this container.
			Volumes: docker.ContainerVolumeArray{
				&docker.ContainerVolumeArgs{
					Source: volume.Label,
					Target: pulumi.String("/path/to/seafile/data"),
				},
			},

			NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
				&docker.ContainerNetworksAdvancedArgs{
					Name: network.Name,
					Aliases: pulumi.StringArray{
						pulumi.String("seafile"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Allocate a domain with an A record pointing to the instance's IP
		_, err = linode.NewDomain(ctx, "seafileDomain", &linode.DomainArgs{
			Type:     pulumi.String("master"),
			Domain:   pulumi.String("znd4.dev"),
			SoaEmail: pulumi.String("admin@znd4.dev"),
		})
		if err != nil {
			return err
		}

		_, err = linode.NewDomainRecord(ctx, "seafileARecord", &linode.DomainRecordArgs{
			DomainId: pulumi.Int(12345), // Replace with the domain ID you're managing
			Type:     pulumi.String("A"),
			Name:     pulumi.String("seafile"),
			Target:   instance.IpAddress,
			TtlSec:   pulumi.Int(300),
		})
		if err != nil {
			return err
		}

		// Export the instance's IP address
		ctx.Export("instanceIpAddress", instance.IpAddress)

		return nil
	})
}
