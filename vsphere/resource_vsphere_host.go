package vsphere

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	
	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVsphereHost() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereHostCreate,
		Read:   resourceVsphereHostRead,
		Delete: resourceVsphereHostDelete,


		Schema: map[string]*schema.Schema{
			
			"host": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"datacenter_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"cluster_id": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
			},
			"user": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"license": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
			},
			"ssl_no_verify": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
			},
			"keep": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
			},
			"object_id": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},			
		},
	}
}

func resourceVsphereHostCreate(d *schema.ResourceData, meta interface{}) error {
		
	var (
		err error
		lic *string
		task *object.Task
	)
	
	datacenterName := d.Get("datacenter_id").(string)
	hostName := d.Get("host").(string)

	hostSystem, err := findHost(d, meta)
	if err != nil {
		
		if strings.HasPrefix(err.Error(), "found host") {
			log.Printf("[ERROR] Host '%s' already exists at path '%s'", hostName, hostSystem.InventoryPath)
			return err
		}
		
		finder, datacenter, err := getFinder(d, meta)
		if err != nil {
			log.Printf("[ERROR] Unable to create finder to create host in: '%s'", hostName)
			return err
		}

		spec := types.HostConnectSpec{
			Force: true,
			HostName: hostName,
			UserName: d.Get("user").(string),
			Password: d.Get("password").(string),
		}
		if v, ok := d.GetOk("license"); ok {
			license := v.(string)
			lic = &license 
		}
		
		v, ok := d.GetOk("cluster_id")
		if ok {
			
			clusterName := v.(string)
			
			cluster, err := finder.Cluster(context.Background(), clusterName)
			if err != nil {
				log.Printf("[ERROR] Cluster '%s' to which host '%s' should be added was not found", clusterName, hostName)
				return err
			}
			
			log.Printf("[DEBUG] Adding host '%s' to cluster '%s'", hostName, clusterName)
			
			task, err = cluster.AddHost(context.Background(), spec, true, lic, nil)
			if err != nil {
				return err
			}
		  	err = task.Wait(context.Background())
			if err != nil {
				var t mo.Task
				
				if v, ok := d.GetOk("ssl_no_verify"); ok && v.(bool) &&
					task.Properties(context.Background(), task.Reference(), []string{"info"}, &t) == nil &&
					reflect.TypeOf(t.Info.Error.Fault).String() == "*types.SSLVerifyFault" {
							
					spec.SslThumbprint = t.Info.Error.Fault.(*types.SSLVerifyFault).Thumbprint
					task, err = cluster.AddHost(context.Background(), spec, true, lic, nil)
					if err == nil {
						err = task.Wait(context.Background())
					}
				}
				if err != nil {
					return err
				}
			}

		} else {
			
			var df *object.DatacenterFolders
			
			df, err = datacenter.Folders(context.Background());
			if err != nil {
				return err
			}
						
			log.Printf("[DEBUG] Adding standalone host '%s' to datacenter '%s'", hostName, datacenterName)
						
			task, err = df.HostFolder.AddStandaloneHost(context.Background(), spec, true, lic, nil)
			if err != nil {
				return err
			}
		  	err = task.Wait(context.Background())
			if err != nil {
				var t mo.Task
				
				if v, ok := d.GetOk("ssl_no_verify"); ok && v.(bool) &&
					task.Properties(context.Background(), task.Reference(), []string{"info"}, &t) == nil &&
					reflect.TypeOf(t.Info.Error.Fault).String() == "*types.SSLVerifyFault" {
							
					spec.SslThumbprint = t.Info.Error.Fault.(*types.SSLVerifyFault).Thumbprint
					task, err = df.HostFolder.AddStandaloneHost(context.Background(), spec, true, lic, nil)
					if err == nil {
						err = task.Wait(context.Background())
					}
				}
				if err != nil {
					return err
				}
			}
		}
	}
	
	d.SetId(d.Get("host").(string))
	return resourceVsphereHostRead(d, meta)
}

func resourceVsphereHostRead(d *schema.ResourceData, meta interface{}) error {

	hostSystem, err := findHost(d, meta)
	if err != nil {
		d.SetId("")		
		return err
	}	

	d.Set("object_id", hostSystem.Reference().Value) 
	return nil
}

func resourceVsphereHostDelete(d *schema.ResourceData, meta interface{}) error {

	if keep, ok := d.GetOk("keep"); !ok || !keep.(bool) {

		_, err := findHost(d, meta)
		if err != nil {
			return err
		}
		
		v, ok := d.GetOk("cluster_id")
		if !ok {
			
			_, datacenter, err := getFinder(d, meta)
			if err != nil {
				return err
			}
			
			df, err := datacenter.Folders(context.Background());
			if err != nil {
				return err
			}
			
			var computeResource *object.ComputeResource = nil
			
			references, _ := df.HostFolder.Children(context.Background())
			for _, r := range references {
			
				switch r.(type) {
					case *object.ComputeResource:
					
						computeResource = r.(*object.ComputeResource)
						hosts, err := computeResource.Hosts(context.Background());
						if err != nil {
							return err
						}
						if len(hosts) > 0 && hosts[0].Value == d.Get("object_id").(string) {
							break
						}
				}
			}
			if computeResource != nil {	
				log.Printf("[DEBUG] Removing standalone host: %s", d.Get("host"))
				task, err := computeResource.Destroy(context.Background())
				if err != nil {
					return err
				}
			  	err = task.Wait(context.Background())
				if err != nil {
					return err
				}				
			}

		} else {
			log.Printf("[WARN] To remove host '%s' from cluster '%s' delete the cluster", d.Get("host"), v.(string))
		}
	}
	return nil
}

func findHost(d *schema.ResourceData, meta interface{}) (*object.HostSystem, error) {
	
	var (
		clusterName *string
		err error
	)
	
	finder, _, err := getFinder(d, meta)
	if err != nil {
		log.Printf("[ERROR] Unable to create finder for operations on host: '%s'", d.Id())
		return nil, err
	}
	
	v, ok := d.GetOk("cluster_id")
	if ok {
		cluster := v.(string)
		clusterName = &cluster
	} else {
		clusterName = nil
	}

	hostSystem, err :=  getHost(d.Get("host").(string), d.Get("datacenter_id").(string), clusterName, finder)
	if err != nil {
		return nil, err
	}
	
	return hostSystem, nil
}

func getHost(hostName string, datacenterName string, clusterName *string, finder *find.Finder) (*object.HostSystem, error) {

	var path string

	if clusterName == nil {
		path = fmt.Sprintf("/%s/host/%s/%s", datacenterName, hostName, hostName)
	} else {
		path = fmt.Sprintf("/%s/host/%s/%s", datacenterName, *clusterName, hostName)
	}
	
 	hostSystem, err := finder.HostSystem(context.Background(), fmt.Sprintf("*/%s", hostName))
	if err != nil {
		log.Printf("[ERROR] Unable find host: '%s'", hostName)
		return nil, err
	}
	
	if hostSystem.InventoryPath != path {
		return hostSystem, fmt.Errorf(
			"found host at path '%s' which does not match the expected path of '%s'", 
			hostSystem.InventoryPath, path)
	}
	
	return hostSystem, nil
}