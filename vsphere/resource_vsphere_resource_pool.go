package vsphere

import (
//	"fmt"
	
//	"golang.org/x/net/context"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVsphereResourcePool() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereResourcePoolCreate,
		Read:   resourceVsphereResourcePoolRead,
		Update: resourceVsphereResourcePoolUpdate,
		Delete: resourceVsphereResourcePoolDelete,

		Schema: map[string]*schema.Schema{
			
		},
	}
}

func resourceVsphereResourcePoolCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereResourcePoolRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereResourcePoolUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereResourcePoolDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}