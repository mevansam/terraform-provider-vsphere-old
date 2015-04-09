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
	"github.com/kr/pretty"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

var keepDatacenter bool

func TestAccVsphereDatacenter_normal(t *testing.T) {
	
	keepDatacenter = false
	
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
							resource.TestCheckResourceAttr("vsphere_datacenter.dc1", "name", "datacenter1"),
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
		
		log.Printf("[DEBUG] Terraform host: %# v", pretty.Formatter(rs))
		
		attributes := rs.Primary.Attributes
		name := rs.Primary.ID
		
		client := testAccProvider.Meta().(*govmomi.Client)
		if client == nil {
			fmt.Errorf("client is nil")
		}
		
		finder := find.NewFinder(client.Client, false)
		datacenter, err := finder.Datacenter(context.Background(), name)
		if err != nil {
			return err
		}
		
		if datacenter.Reference().Value != attributes["object_id"] {
			return fmt.Errorf("datacenter object id mismatch. expected '%s' but go '%s'", datacenter.Reference().Value, attributes["object_id"])
		}
		
		keepDatacenter = (rs.Primary.Attributes["keep"] == "true")
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
	
	client := testAccProvider.Meta().(*govmomi.Client)
	if client == nil {
		fmt.Errorf("client is nil")
	}
	
	finder := find.NewFinder(client.Client, false)
	_, err := finder.Datacenter(context.Background(), name)
	if err != nil {
		log.Printf("[DEBUG] Datacenter destroyed as expected. API response was: %s", err.Error())
	} else if keepDatacenter {
		log.Printf("[DEBUG] Datacenter not destroyed as expected")
	} else {
		return fmt.Errorf("datacenter '%s' was not destroyed as expected", resource);
	}
	return nil
}

const testAccDatacenterConfig = `

resource "vsphere_datacenter" "dc1" {
	name = "datacenter1"

#	keep = true
}
`
