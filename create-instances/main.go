package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/googleapi"

	"github.com/otiai10/debug"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

var (
	project string
	zone    string
	name    string
)

func init() {
	flag.StringVar(&project, "project", "otiai10-sandbox", "Project name on GCP")
	flag.StringVar(&zone, "zone", "asia-northeast1-a", "Zone of GCP")
	flag.StringVar(&name, "name", "test-instance", "Instance name to create")
	flag.Parse()
}

func main() {

	ctx := context.Background()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		debug.Fatalln(err)
	}

	service, err := compute.New(client)
	if err != nil {
		debug.Fatalln(err)
	}

	// Read
	instance, err := service.Instances.Get(project, zone, name).Do()

	if err == nil && instance != nil {
		// Delete
		log.Printf("Instance found, trying to delete.")
		_, err := service.Instances.Delete(project, zone, name).Do()
		if err != nil {
			debug.Fatalln(err)
		}
		for count := 0; ; count++ {
			fmt.Print(".")
			_, err := service.Instances.Get(project, zone, name).Do()
			if apierror, ok := err.(*googleapi.Error); ok && apierror.Code == 404 {
				fmt.Print("\n")
				break
			}
			time.Sleep(2 * time.Second)
			if count > 15 {
				debug.Fatalln("Couldn't wait for instance deletion ;(")
			}
		}
		log.Printf("Deleted: %v", name)
	} else if apierror, ok := err.(*googleapi.Error); !ok || apierror.Code != 404 {
		debug.Fatalln(err)
	}

	// Create
	instance = &compute.Instance{
		Description: "This is test",
		Name:        name,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/n1-standard-1", zone),
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				Network: fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/default", project),
			},
		},
		Disks: []*compute.AttachedDisk{
			&compute.AttachedDisk{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/debian-cloud/global/images/debian-9-stretch-v20180806",
					DiskSizeGb:  10,
				},
			},
		},
	}
	_, err = service.Instances.Insert(project, zone, instance).Do()
	if err != nil {
		debug.Fatalln(err)
	}

	log.Println("Created:", instance.Name)
}
