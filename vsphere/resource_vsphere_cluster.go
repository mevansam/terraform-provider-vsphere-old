package vsphere

import (
	"fmt"
	"log"
	"strconv"
	
	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVsphereCluster() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereClusterCreate,
		Read:   resourceVsphereClusterRead,
		Update: resourceVsphereClusterUpdate,
		Delete: resourceVsphereClusterDelete,

		Schema: map[string]*schema.Schema{
			
			"name": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			
			"datacenter_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			
			"drs": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_vm_automation_override": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
						},
						"default_automation_level": &schema.Schema{
							Type: schema.TypeString, // One of manual, partiallyAutomated or fullyAutomated
							Optional: true,
						},
						"migration_threshold": &schema.Schema{
							Type: schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"ha": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_monitoring": &schema.Schema{
							Type: schema.TypeString, // One of 'enabled' or 'disabled'
							Optional: true,
						},
						"vm_monitoring": &schema.Schema{
							Type: schema.TypeString, // One of vmAndAppMonitoring, vmMonitoringOnly or vmMonitoringDisabled 
							Optional: true,
						},
						"admissionControlEnabled": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
						},
						"admissionControlPolicy": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": &schema.Schema{
										Type: schema.TypeString,
										Required: true,
									},
									"value": &schema.Schema{
										Type: schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func clusterFinder(d *schema.ResourceData, meta interface{}) (*find.Finder, *object.Datacenter, error) {
	
	client := meta.(*govmomi.Client)
	if client == nil {
		return nil, nil, fmt.Errorf("client is nil")
	}
	
	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.Datacenter(context.Background(), d.Get("datacenter_id").(string))
	if err != nil {
		return nil, nil, err
	}
	finder.SetDatacenter(datacenter)
	
	return finder, datacenter, nil
}

func findCluster(d *schema.ResourceData, meta interface{}) (*object.ClusterComputeResource, error) {
	
	finder, _, err := clusterFinder(d, meta)
	if err != nil {
		log.Printf("[ERROR] Unable to create finder for operations on cluster: '%s'", d.Get("name").(string))
		return nil, err
	}
	
	cluster, err := finder.Cluster(context.Background(), d.Id())
	if err != nil {
		log.Printf("[ERROR] Unable find cluster: '%s'", d.Get("name").(string))
		return nil, err
	}
	
	return cluster, nil
}

func resourceVsphereClusterRead(d *schema.ResourceData, meta interface{}) error {
	
//	finder, datacenter, err := clusterFinder(d, meta)
//	if err != nil {
//		log.Printf("[ERROR] Unable to create finder for operations on cluster: '%s'", d.Get("name").(string))
//		d.SetId("")
//		return err
//	}

	return nil
}

func resourceVsphereClusterCreate(d *schema.ResourceData, meta interface{}) error {
	
	finder, datacenter, err := clusterFinder(d, meta)
	if err != nil {
		log.Printf("[ERROR] Unable to create finder for operations on cluster: '%s'", d.Get("name").(string))
		d.SetId("")
		return err
	}
	
	_, err = finder.Cluster(context.Background(), d.Get("name").(string))
	if err != nil {		
		log.Printf("[DEBUG] Creating the cluster: %s", d.Get("name").(string))
		
		var df *object.DatacenterFolders
		
		df, err = datacenter.Folders(context.Background());
		if err != nil {
			return err			
		}

		_, err = df.HostFolder.CreateCluster(context.Background(), d.Get("name").(string), types.ClusterConfigSpecEx {})
		if err != nil {
			log.Printf("[ERROR] VMOMI Error creating cluster: %s", err.Error())
			d.SetId("")
			return err
		}
	}
	
	d.SetId(d.Get("name").(string))
	return resourceVsphereClusterUpdate(d, meta)
}

func resourceVsphereClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	
	cluster, err := findCluster(d, meta)
	if err != nil {
		return err
	}

	spec := types.ClusterConfigSpec {}
	
	drs := d.Get("drs").(map[string]interface {})
	if drs != nil && len(drs) > 0 {
		spec.DrsConfig = &types.ClusterDrsConfigInfo {}
		spec.DrsConfig.Enabled = true
		
		if v := drs["enable_vm_automation_override"]; v != nil {
			spec.DrsConfig.EnableVmBehaviorOverrides, _ = strconv.ParseBool(v.(string))
		}
		if v := drs["migration_threshold"]; v != nil {
			spec.DrsConfig.VmotionRate, _ = strconv.Atoi(v.(string))
		}
		if v := drs["default_automation_level"]; v != nil {
			defaultVmBehavior := types.DrsBehavior(v.(string))
			if defaultVmBehavior != types.DrsBehaviorManual && 
				defaultVmBehavior != types.DrsBehaviorPartiallyAutomated && 
				defaultVmBehavior != types.DrsBehaviorFullyAutomated {
				return fmt.Errorf("invalid automation level. it should be one of manual, partiallyAutomated or fullyAutomated")
			}
			spec.DrsConfig.DefaultVmBehavior = defaultVmBehavior
		}	
	}
	ha := d.Get("ha").(map[string]interface {})
	if ha != nil && len(ha) > 0 {
		spec.DasConfig = &types.ClusterDasConfigInfo {}
		spec.DasConfig.Enabled = true
		
		if v := ha["vm_monitoring"]; v != nil {
			vmMonitoring := v.(string)
			if vmMonitoring != "vmAndAppMonitoring" && 
				vmMonitoring != "vmMonitoringOnly" && 
				vmMonitoring != "vmMonitoringDisabled" {
				return fmt.Errorf("invalid vm monitoring value. it should be one of vmAndAppMonitoring, vmMonitoringOnly or vmMonitoringDisabled")
			}
			spec.DasConfig.VmMonitoring = vmMonitoring
		}
		if v := ha["host_monitoring"]; v != nil {
			hostMonitoring := v.(string)
			if hostMonitoring != "enabled" &&
				hostMonitoring != "disabled" {
				return fmt.Errorf("invalid host monitoring value. it should be one of enabled or disabled")				
			}
			spec.DasConfig.HostMonitoring = hostMonitoring
		}
		if v := drs["admissionControlEnabled"]; v != nil {
			spec.DasConfig.AdmissionControlEnabled, _ = strconv.ParseBool(v.(string))
		}
		if v := drs["admissionControlPolicy"]; v != nil {
			admissionControlPolicy := v.([]map[string]interface {})
			var props []types.DynamicProperty
			for _, p := range admissionControlPolicy {
				props = append(props, types.DynamicProperty {
						Name: p["name"].(string),
						Val: p["value"].(string),
					} )
			}
		}
	}
	
	log.Printf("[DEBUG] Updating cluster: %s", d.Id())

	var task *object.Task
	task, err = cluster.ReconfigureCluster(context.Background(), spec)
	if err != nil {
		return err
	}	
	err = task.Wait(context.Background())
	if err != nil {
		return err
	}
	
	return resourceVsphereClusterRead(d, meta)
}

func resourceVsphereClusterDelete(d *schema.ResourceData, meta interface{}) error {

	cluster, err := findCluster(d, meta)
	if err != nil {
		return err
	}
	
	log.Printf("[DEBUG] Deleting cluster: %s", d.Id())
	
	task, err := cluster.Destroy(context.Background())
	if err != nil {
		return err
	}
  	err = task.Wait(context.Background())
	if err != nil {
		return err
	}
	
	return nil
}
