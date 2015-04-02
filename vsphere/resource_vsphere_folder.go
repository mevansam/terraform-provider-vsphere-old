package vsphere

import (
//	"fmt"
	
//	"golang.org/x/net/context"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVsphereFolder() *schema.Resource {
	
	return &schema.Resource{
		
		Create: resourceVsphereFolderCreate,
		Read:   resourceVsphereFolderRead,
		Update: resourceVsphereFolderUpdate,
		Delete: resourceVsphereFolderDelete,

		Schema: map[string]*schema.Schema{
			
		},
	}
}

func resourceVsphereFolderCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereFolderRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereFolderUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVsphereFolderDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}