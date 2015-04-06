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
			"vsphere_host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_HOST", nil),
			},
			"vsphere_username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_USERNAME", nil),
			},
			"vsphere_password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSPHERE_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vsphere_datacenter": resourceVsphereDatacenter(),
			"vsphere_cluster": resourceVsphereCluster(),
			"vsphere_resource_pool": resourceVsphereResourcePool(),
			"vsphere_folder": resourceVsphereFolder(),
			"vsphere_datastore": resourceVsphereDatastore(),
			"vsphere_vm": resourceVsphereVM(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	
	config := Config{
		Host: d.Get("vsphere_host").(string),
		Username: d.Get("vsphere_username").(string),
		Password: d.Get("vsphere_password").(string),
	}
	return config.Client()
}

func GetFinder(d *schema.ResourceData, meta interface{}) (*find.Finder, *object.Datacenter, error) {
	
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

func FindHost(d *schema.ResourceData, meta interface{}) (*object.HostSystem, error) {
	
	finder, _, err := GetFinder(d, meta)
	if err != nil {
		return nil, err
	}
	
	hostSystem, err := finder.HostSystem(context.Background(), d.Get("host_id").(string))
	if err != nil {
		return nil, err
	}
	
	return hostSystem, nil
}

func FindCluster(d *schema.ResourceData, meta interface{}) (*object.ClusterComputeResource, error) {
	
	finder, _, err := GetFinder(d, meta)
	if err != nil {
		return nil, err
	}
	
	cluster, err := finder.Cluster(context.Background(), d.Get("cluster_id").(string))
	if err != nil {
		return nil, err
	}
	
	return cluster, nil
}

func FindResourcePool(d *schema.ResourceData, meta interface{}) (*object.ResourcePool, error) {
	
	finder, _, err := GetFinder(d, meta)
	if err != nil {
		return nil, err
	}
	
	resourcePool, err := finder.ResourcePool(context.Background(), d.Get("resource_pool_id").(string))
	if err != nil {
		return nil, err
	}
	
	return resourcePool, nil
}
