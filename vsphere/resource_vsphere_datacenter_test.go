package vsphere

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/net/context"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

func TestAccVsphereDatacenter_normal(t *testing.T) {
	
	_, filename, _, _ := runtime.Caller(0)
	ut := os.Getenv("UNIT_TEST")
	if ut == "" || ut == filepath.Base(filename) {
		
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
}

func testAccCheckDatacenterExists(resource string) resource.TestCheckFunc {
	
	return func(s *terraform.State) error {
	
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("datacenter '%s' not found in terraform state", resource)
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

	const resource = "vsphere_datacenter.dc1"
	const name = "datacenter1"

	_, ok := s.RootModule().Resources[resource]
	if ok {
		return fmt.Errorf("datacenter '%s' still exists in the terraform state", resource)
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
	return fmt.Errorf("datacenter '%s' was not destroyed as expected", resource);
}

const testAccDatacenterConfig = `

resource "vsphere_datacenter" "dc1" {
	name = "datacenter1"
}
`
