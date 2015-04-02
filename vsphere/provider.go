package vsphere

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
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
