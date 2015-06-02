package vsphere

import (
	"fmt"
	
	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

func Provider() terraform.ResourceProvider {
	
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_HOST", nil),
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_USERNAME", nil),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vsphere_datacenter": resourceVsphereDatacenter(),
			"vsphere_cluster": resourceVsphereCluster(),
			"vsphere_resource_pool": resourceVsphereResourcePool(),
			"vsphere_host": resourceVsphereHost(),
			"vsphere_folder": resourceVsphereFolder(),
			"vsphere_datastore": resourceVsphereDatastore(),
			"vsphere_vm": resourceVsphereVM(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	
	config := Config{
		Host: d.Get("host").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}
	return config.Client()
}

func getFinder(d *schema.ResourceData, meta interface{}) (*find.Finder, *object.Datacenter, error) {
	
	client := meta.(*govmomi.Client)
	if client == nil {
		return nil, nil, fmt.Errorf("client is nil")
	}
	
	var (
		datacenter *object.Datacenter
		err error
	)
	
	finder := find.NewFinder(client.Client, false)
	v, ok := d.GetOk("datacenter_id")
	if ok {	
		datacenter, err = finder.Datacenter(context.Background(), v.(string))
	} else {
		datacenter, err = finder.DefaultDatacenter(context.Background())
	}
	if err != nil {
		return nil, nil, err
	}
	finder.SetDatacenter(datacenter)
	
	return finder, datacenter, nil
}
