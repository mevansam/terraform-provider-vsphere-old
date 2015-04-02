package vsphere

import (
//	"fmt"
	
//	"golang.org/x/net/context"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVsphereDatastore() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereDatastoreCreate,
		Read:   resourceVsphereDatastoreRead,
		Update: resourceVsphereDatastoreUpdate,
		Delete: resourceVsphereDatastoreDelete,

		Schema: map[string]*schema.Schema{
			
		},
	}
}

func resourceVsphereDatastoreCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereDatastoreRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereDatastoreUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereDatastoreDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}