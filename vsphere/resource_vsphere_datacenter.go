package vsphere

import (
	"fmt"
	"log"
	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

func resourceVsphereDatacenter() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereDatacenterCreate,
		Read:   resourceVsphereDatacenterRead,
		Delete: resourceVsphereDatacenterDelete,

		Schema: map[string]*schema.Schema{
			
			"name": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},			
		},
	}
}

func findDatacenter(d *schema.ResourceData, meta interface{}) (*object.Datacenter, error) {
	
	client := meta.(*govmomi.Client)
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	
	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.Datacenter(context.Background(), d.Id())
	if err != nil {
		return nil, err
	}
	
	return datacenter, nil
}

func resourceVsphereDatacenterRead(d *schema.ResourceData, meta interface{}) error {
		
	_, err := findDatacenter(d, meta)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("datacenter '%s' not found", d.Id())
	}
	
	log.Printf("[DEBUG] Found datacenter: %s", d.Id())
	
	return nil
}

func resourceVsphereDatacenterCreate(d *schema.ResourceData, meta interface{}) error {
	
	client := meta.(*govmomi.Client)
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	finder := find.NewFinder(client.Client, false)
	_, err := finder.Datacenter(context.Background(), d.Get("name").(string))
	if err != nil {
		
		log.Printf("[DEBUG] Creating datacenter: %s", d.Get("name").(string))
		
		rootFolder := object.NewRootFolder(client.Client)
		_, err = rootFolder.CreateDatacenter(context.Background(), d.Get("name").(string))
		if err != nil {
			log.Printf("[ERROR] VMOMI Error creating datacenter: %s", err.Error())
			d.SetId("")
			return fmt.Errorf("error creating datacenter '%s'", d.Id())
		}
	}
	
	d.SetId(d.Get("name").(string))
	return resourceVsphereDatacenterRead(d, meta)
}

func resourceVsphereDatacenterDelete(d *schema.ResourceData, meta interface{}) error {

	datacenter, err := findDatacenter(d, meta)
	if err != nil {
		return fmt.Errorf("datacenter to delete '%s' not found", d.Id())
	}
	
	log.Printf("[DEBUG] Deleting datacenter: %s", d.Id())
	
	task, err := datacenter.Destroy(context.Background())
	if err != nil {
		return err
	}
  	err = task.Wait(context.Background())
	if err != nil {
		return err
	}

	return nil
}
