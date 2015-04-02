package vsphere

import (
	"fmt"
	"log"
	"testing"

	"golang.org/x/net/context"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

func TestAccVsphereDatacenter_normal(t *testing.T) {
	
	resource.Test( t, 
		resource.TestCase {
			PreCheck: func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckDatacenterDestroy,
			Steps: []resource.TestStep {
				resource.TestStep {
					Config: testAccDatacenterConfig,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckDatacenterExists("vsphere_datacenter.dc1"),
					),
				},
			},
		} )
}

func testAccCheckDatacenterExists(n string) resource.TestCheckFunc {
	
	return func(s *terraform.State) error {
		
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("datacenter '%s' not found in terraform state", n)
		}
		
		name := rs.Primary.ID
		log.Printf("[DEBUG] Checking if datacenter '%s' exists", name)

		client := testAccProvider.Meta().(*govmomi.Client)
		if client == nil {
			fmt.Errorf("client is nil")
		}
		
		finder := find.NewFinder(client.Client, false)
		_, err := finder.Datacenter(context.Background(), name)
		if err != nil {
			return err
		}
		return nil;
	}
}

func testAccCheckDatacenterDestroy(s *terraform.State) error {

	const n = "vsphere_datacenter.dc1"
	const name = "datacenter1"

	_, ok := s.RootModule().Resources[n]
	if ok {
		return fmt.Errorf("datacenter '%s' still exists in the terraform state", n)
	}
	
	log.Printf("[DEBUG] Checking if datacenter '%s' has been destroyed", name)
	client := testAccProvider.Meta().(*govmomi.Client)
	if client == nil {
		fmt.Errorf("client is nil")
	}
	
	finder := find.NewFinder(client.Client, false)
	_, err := finder.Datacenter(context.Background(), name)
	if err != nil {
		log.Printf("[DEBUG] API response: %s", err.Error())
		return nil
	}
	return fmt.Errorf("datacenter '%s' was not destroyed as expected", n);
}

const testAccDatacenterConfig = `

resource "vsphere_datacenter" "dc1" {
	name = "datacenter1"
}
`
