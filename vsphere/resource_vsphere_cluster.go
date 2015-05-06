package vsphere

import (
	"fmt"
	"log"
	"strconv"
	
	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
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
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_vm_automation_override": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
							Default: true,
						},
						"default_automation_level": &schema.Schema{
							Type: schema.TypeString, // One of manual, partiallyAutomated or fullyAutomated
							Optional: true,
							Default: "fullyAutomated",
						},
						"migration_threshold": &schema.Schema{
							Type: schema.TypeInt,
							Optional: true,
							Default: 3,
						},
					},
				},
			},
			"ha": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_monitoring": &schema.Schema{
							Type: schema.TypeString, // One of 'enabled' or 'disabled'
							Optional: true,
							Default: "enabled",
						},
						"vm_monitoring": &schema.Schema{
							Type: schema.TypeString, // One of vmAndAppMonitoring, vmMonitoringOnly or vmMonitoringDisabled 
							Optional: true,
							Default: "vmMonitoringDisabled",
						},
						"admissionControlEnabled": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
							Default: true,
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

func resourceVsphereClusterCreate(d *schema.ResourceData, meta interface{}) error {
	
	finder, datacenter, err := getFinder(d, meta)
	if err != nil {
		log.Printf("[ERROR] Unable to create finder for operations on cluster: '%s'", d.Get("name").(string))
		d.SetId("")
		return err
	}
	
	cluster, err := finder.ClusterComputeResource(context.Background(), d.Get("name").(string))
	if err != nil {		
		log.Printf("[DEBUG] Creating the cluster: %s", d.Get("name").(string))
		
		var df *object.DatacenterFolders
		
		df, err = datacenter.Folders(context.Background());
		if err != nil {
			return err			
		}
	
		cluster, err = df.HostFolder.CreateCluster(context.Background(), d.Get("name").(string), types.ClusterConfigSpecEx{})
		if err != nil {
			log.Printf("[ERROR] VMOMI Error creating cluster: %s", err.Error())
			d.SetId("")
			return err
		}
	}
	
	d.SetId(d.Get("name").(string))
	d.Set("object_id", cluster.Reference().Value)
	return resourceVsphereClusterUpdate(d, meta)
}

func resourceVsphereClusterRead(d *schema.ResourceData, meta interface{}) error {
	
	cluster, err := findCluster(d, meta)
	if err != nil {		
		// TODO: reset only if not found. Test for no found in err.Error()
		d.SetId("")
		return err
	}
	
	config, err := getConfiguration(context.Background(), cluster)
	if err != nil {
		log.Printf("[ERROR] Unable read cluster configuration: '%s'", d.Id())
		return err
	}
	
	if *config.DrsConfig.Enabled {
		drs := make(map[string]interface{})
		drs["enable_vm_automation_override"] = strconv.FormatBool(*config.DrsConfig.EnableVmBehaviorOverrides)
		drs["migration_threshold"] = strconv.Itoa(config.DrsConfig.VmotionRate)
		drs["default_automation_level"] = string(config.DrsConfig.DefaultVmBehavior)		
		d.Set("drs", append(make([]map[string]interface{}, 0, 1), drs))
	}
	if *config.DasConfig.Enabled {
		ha := make(map[string]interface{})	
		ha["vm_monitoring"] = config.DasConfig.VmMonitoring
		ha["host_monitoring"] = config.DasConfig.HostMonitoring
		ha["admissionControlEnabled"] = strconv.FormatBool(*config.DasConfig.AdmissionControlEnabled)
		d.Set("ha", append(make([]map[string]interface{}, 0, 1), ha))
 	}
		
	d.Set("object_id", cluster.Reference().Value) 
	return nil
}

func resourceVsphereClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	
	cluster, err := findCluster(d, meta)
	if err != nil {
		return err
	}
	
	spec := types.ClusterConfigSpec {}
	spec.DrsConfig, err = getClusterDrsConfigInfo(d) 
	if err != nil {
		return err
	}
	spec.DasConfig, err = getClusterDasConfigInfo(d)
	if err != nil {
		return err
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
	
	return nil
}

func resourceVsphereClusterDelete(d *schema.ResourceData, meta interface{}) error {

	if keep, ok := d.GetOk("keep"); !ok || !keep.(bool) {
		
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
	}
	
	return nil
}

func findCluster(d *schema.ResourceData, meta interface{}) (*object.ClusterComputeResource, error) {
	
	finder, _, err := getFinder(d, meta)
	if err != nil {
		log.Printf("[ERROR] Unable to create finder for operations on cluster: '%s'", d.Get("name").(string))
		return nil, err
	}
	
	cluster, err := finder.ClusterComputeResource(context.Background(), d.Id())
	if err != nil {
		log.Printf("[ERROR] Unable find cluster: '%s'", d.Get("name").(string))
		return nil, err
	}
	
	return cluster, nil
}

func getClusterDrsConfigInfo(d *schema.ResourceData) (*types.ClusterDrsConfigInfo, error) {
	
	drsCount := d.Get("drs.#").(int)
	if drsCount > 1 {
		return nil, fmt.Errorf("only 1 drs configuration section permitted")
	}
	if drsCount == 1 {
		
		drsConfigEnabled := true
		
		drsConfig := &types.ClusterDrsConfigInfo {}
		drsConfig.Enabled = &drsConfigEnabled
		
		if v, ok := d.GetOk("drs.0.enable_vm_automation_override"); ok {
			enableVmBehaviorOverrides := v.(bool)
			drsConfig.EnableVmBehaviorOverrides = &enableVmBehaviorOverrides
		}
		if v, ok := d.GetOk("drs.0.migration_threshold"); ok {
			drsConfig.VmotionRate = v.(int)
		}
		if v, ok := d.GetOk("drs.0.default_automation_level"); ok {
			defaultVmBehavior := types.DrsBehavior(v.(string))
			if defaultVmBehavior != types.DrsBehaviorManual && 
				defaultVmBehavior != types.DrsBehaviorPartiallyAutomated && 
				defaultVmBehavior != types.DrsBehaviorFullyAutomated {
				return nil, fmt.Errorf("invalid automation level. it should be one of manual, partiallyAutomated or fullyAutomated")
			}
			drsConfig.DefaultVmBehavior = defaultVmBehavior
		}
		return drsConfig, nil
	}
	
	return nil, nil
}

func getClusterDasConfigInfo(d *schema.ResourceData) (*types.ClusterDasConfigInfo, error) {
	
	haCount := d.Get("ha.#").(int)
	if haCount > 1 {
		return nil, fmt.Errorf("only 1 ha configuration section permitted")
	}
	if haCount == 1 {
		
		dasConfigEnabled := true
		
		dasConfig := &types.ClusterDasConfigInfo {}
		dasConfig.Enabled = &dasConfigEnabled
		
		if v, ok := d.GetOk("ha.0.vm_monitoring"); ok {
			vmMonitoring := v.(string)
			if vmMonitoring != "vmAndAppMonitoring" && 
				vmMonitoring != "vmMonitoringOnly" && 
				vmMonitoring != "vmMonitoringDisabled" {
				return nil, fmt.Errorf("invalid vm monitoring value. it should be one of vmAndAppMonitoring, vmMonitoringOnly or vmMonitoringDisabled")
			}
			dasConfig.VmMonitoring = vmMonitoring
		}
		if v, ok := d.GetOk("ha.0.host_monitoring"); ok {
			hostMonitoring := v.(string)
			if hostMonitoring != "enabled" &&
				hostMonitoring != "disabled" {
				return nil, fmt.Errorf("invalid host monitoring value. it should be one of enabled or disabled")				
			}
			dasConfig.HostMonitoring = hostMonitoring
		}
		if v, ok := d.GetOk("ha.0.admissionControlEnabled"); ok {
			admissionControlEnabled := v.(bool)
			dasConfig.AdmissionControlEnabled = &admissionControlEnabled
		}
		if d.Get("ha.0.admissionControlPolicy.#").(int) > 0 {
			var props []types.DynamicProperty
			v, ok := d.GetOk("ha.0.admissionControlPolicy")
			if ok {
				for _, p := range v.([]map[string]interface{}) {
					if p != nil {
						props = append(props, types.DynamicProperty {
								Name: p["name"].(string),
								Val: p["value"].(string),
							} )
					}
				}
			}
		}
		
		return dasConfig, nil
	}
	
	return nil, nil
}

func getConfiguration(ctx context.Context, cluster *object.ClusterComputeResource) (*types.ClusterConfigInfo, error) {	
	var mccr mo.ClusterComputeResource
	
	ps := []string{"configuration"}
	err := cluster.Properties(ctx, cluster.Reference(), ps, &mccr)
	if err != nil {
		return nil, err
	}
	
	return &mccr.Configuration, nil
}
