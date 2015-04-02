package vsphere

import (
//	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
//	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVsphereResourcePool_normal(t *testing.T) {

	resource.Test( t, 
		resource.TestCase {
			PreCheck: func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckResourcePoolDestroy,
			Steps: []resource.TestStep {
				},
		} )
}

func testAccCheckResourcePoolDestroy(s *terraform.State) error {
	return nil
}
