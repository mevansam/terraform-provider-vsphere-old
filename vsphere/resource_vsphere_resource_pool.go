package vsphere

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	
	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/kr/pretty"
)

func resourceVsphereResourcePool() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereResourcePoolCreate,
		Read:   resourceVsphereResourcePoolRead,
		Update: resourceVsphereResourcePoolUpdate,
		Delete: resourceVsphereResourcePoolDelete,

		Schema: map[string]*schema.Schema{
			
			"name": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},			
			"datacenter_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"parent_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"cpu": &schema.Schema{
				Type: schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"shares": &schema.Schema{
							Type: schema.TypeString, // One of low, normal, high or an integer value for custom level
							Optional: true,
						},
						"reservation": &schema.Schema{
							Type: schema.TypeInt,
							Optional: true,
						},
						"expandable_reservation": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
						},
						"limit": &schema.Schema{
							Type: schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"memory": &schema.Schema{
				Type: schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"shares": &schema.Schema{
							Type: schema.TypeString, // One of low, normal, high or an integer value for custom level
							Optional: true,
						},
						"reservation": &schema.Schema{
							Type: schema.TypeInt,
							Optional: true,
						},
						"expandable_reservation": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
							Default: true,
						},
						"limit": &schema.Schema{
							Type: schema.TypeInt,
							Optional: true,
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

func resourceVsphereResourcePoolCreate(d *schema.ResourceData, meta interface{}) error {

	var err error

	resourcePool, err := findResourcePool(d, meta)
	if err != nil {
		
		if strings.Contains(err.Error(), "not found") {
			
			finder, _, err := getFinder(d, meta)
			if err != nil {
				log.Printf("[ERROR] Unable to create finder for operations on resource pool: '%s'", d.Get("name").(string))
				return err
			}
			
			parentResourcePool, err := finder.ResourcePool(
				context.Background(), fmt.Sprintf("/%s/*/%s/Resources", d.Get("datacenter_id").(string), d.Get("parent_id").(string)))
			if err != nil {
				log.Printf("[ERROR] Unable to retrieve default resource pool of parent '%s'", d.Get("parent_id").(string))
				return err
			}
			
			log.Printf("[DEBUG] Creating new resource pool at %s", parentResourcePool.InventoryPath)
			
			spec := types.ResourceConfigSpec{}
			err = getAllocationInfo("cpu", &spec.CpuAllocation, d)
			if err != nil {
				return err
			}
			err = getAllocationInfo("memory", &spec.MemoryAllocation, d)
			if err != nil {
				return err
			}
			
			resourcePool, err = parentResourcePool.Create(context.Background(), d.Get("name").(string), spec)
			if err != nil {
				log.Printf("[ERROR] Unable to create resource pool '%s'", d.Get("name").(string))
				return err				
			}
			
		} else {
			return err
		}
	}
	
	d.SetId(d.Get("name").(string))
	d.Set("object_id", resourcePool.Reference().Value)
	return nil
}

func resourceVsphereResourcePoolRead(d *schema.ResourceData, meta interface{}) error {
	
	resourcePool, err := findResourcePool(d, meta)
	if err != nil {
		d.SetId("")
		return err
	}
	
	log.Printf("[DEBUG] Reading configuration of resource pool at %s", resourcePool.InventoryPath)
	
	var mrp mo.ResourcePool
	
	ps := []string{"config"}
	err = resourcePool.Properties(context.Background(), resourcePool.Reference(), ps, &mrp)
	if err != nil {
		return err
	}
	
	putAllocationInfo("cpu", &mrp.Config.CpuAllocation, d)
	putAllocationInfo("memory", &mrp.Config.CpuAllocation, d)
	
	d.Set("object_id", resourcePool.Reference().Value)
	return nil
}

func resourceVsphereResourcePoolUpdate(d *schema.ResourceData, meta interface{}) error {

	resourcePool, err := findResourcePool(d, meta)
	if err != nil {
		d.SetId("")
		return err
	}
	
	log.Printf("[DEBUG] Updating resource pool at %s", resourcePool.InventoryPath)
	
	spec := types.ResourceConfigSpec{}
	err = getAllocationInfo("cpu", &spec.CpuAllocation, d)
	if err != nil {
		return err
	}
	err = getAllocationInfo("memory", &spec.MemoryAllocation, d)
	if err != nil {
		return err
	}
	
	err = resourcePool.UpdateConfig(context.Background(), d.Get("name").(string), &spec)
	if err != nil {
		log.Printf("[ERROR] Unable to update resource pool '%s'", d.Get("name").(string))
		return err				
	}
	
	return nil
}

func resourceVsphereResourcePoolDelete(d *schema.ResourceData, meta interface{}) error {

	if keep, ok := d.GetOk("keep"); !ok || !keep.(bool) {

		resourcePool, err := findResourcePool(d, meta)
		if err != nil {
			return err
		}
		
		log.Printf("[DEBUG] Deleting resource pool: %s", d.Id())
		
		task, err := resourcePool.Destroy(context.Background())
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

func findResourcePool(d *schema.ResourceData, meta interface{}) (*object.ResourcePool, error) {
	
	var err error
	
	finder, _, err := getFinder(d, meta)
	if err != nil {
		log.Printf("[ERROR] Unable to create finder for operations on resource pool: '%s'", d.Get("name").(string))
		return nil, err
	}
	
	resourcePool, err := getResourcePool(d.Get("name").(string), d.Get("parent_id").(string), finder)
	if err != nil {
		return nil, err
	}
	
	return resourcePool, nil
}

func getResourcePool(name string, parent string, finder *find.Finder) (*object.ResourcePool, error) {
	
	path := fmt.Sprintf("/Resources/%s", name)
	matchPath := fmt.Sprintf("/%s%s", parent, path)
	searchPath := fmt.Sprintf("*%s", path)
	
	resourcePoolList, err := finder.ResourcePoolList(context.Background(), searchPath)
	if err != nil {
		log.Printf("[ERROR] VMOMI error when searching for resource pool at path '%s': %s", searchPath, err.Error())
		return nil, err
	}

	for _, rp := range resourcePoolList {
		
		if strings.HasSuffix(rp.InventoryPath, matchPath) {
			return rp, nil
		}
	}
	
	return nil, fmt.Errorf("resource pool %s was not found", name)
}

func getAllocationInfo(allocType string, allocInfo *types.ResourceAllocationInfo, d *schema.ResourceData) error {

	var err error

	v, ok := d.GetOk(fmt.Sprintf("%s.#", allocType))
	if ok {
		count := v.(int)
		if count > 1 {
			return fmt.Errorf("only 1 %s allocation section permitted", allocType)
		}
		if count == 1 {

			expandableReservation := d.Get(fmt.Sprintf("%s.0.expandable_reservation", allocType)).(bool)
			allocInfo.ExpandableReservation = &expandableReservation
			
			if v, ok := d.GetOk(fmt.Sprintf("%s.0.shares", allocType)); ok {
				shares := types.SharesInfo{}
				level := types.SharesLevel(v.(string))
				if level != types.SharesLevelLow && 
					level != types.SharesLevelNormal && 
					level != types.SharesLevelHigh {
						
					shares.Shares, err = strconv.Atoi(v.(string))
					if err != nil {
						return fmt.Errorf("error converting customer share value to int: %s", err.Error())
					}
					shares.Level = types.SharesLevelCustom
				} else {
					shares.Level = level
				}
				allocInfo.Shares = &shares
			} else {
				allocInfo.Shares = &types.SharesInfo{
					Level: types.SharesLevelNormal,
				}
			}
			if v, ok := d.GetOk(fmt.Sprintf("%s.0.reservation", allocType)); ok {
				allocInfo.Reservation = int64(v.(int))
			}
			if v, ok := d.GetOk(fmt.Sprintf("%s.0.limit", allocType)); ok {
				allocInfo.Limit = int64(v.(int))
			}
		}
	}

	log.Printf("[DEBUG] ResourceAllocationInfo: %# v", pretty.Formatter(allocInfo))
	return nil
}

func putAllocationInfo(allocType string, allocInfo *types.ResourceAllocationInfo, d *schema.ResourceData) {
	
	configState := make(map[string]interface{})
	
	if allocInfo.Reservation != 0 {
		configState["reservation"] = strconv.FormatInt(allocInfo.Reservation, 10)
	}
	if allocInfo.Limit != 0 {
		configState["limit"] = strconv.FormatInt(allocInfo.Limit, 10)
	}
	configState["expandable_reservation"] = strconv.FormatBool(*allocInfo.ExpandableReservation)
	
	if allocInfo.Shares.Level == types.SharesLevelCustom {
		configState["shares"] = strconv.Itoa(allocInfo.Shares.Shares)
	} else {
		configState["shares"] = string(allocInfo.Shares.Level)
	}
	
	d.Set(allocType, append(make([]map[string]interface{}, 0, 1), configState))
}
