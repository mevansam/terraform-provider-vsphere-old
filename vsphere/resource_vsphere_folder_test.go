package vsphere

import (
//	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
//	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVsphereFolder_normal(t *testing.T) {

	resource.Test( t, 
		resource.TestCase {
			PreCheck: func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckFolderDestroy,
			Steps: []resource.TestStep {
				},
		} )
}

func testAccCheckFolderDestroy(s *terraform.State) error {
	return nil
}
