package vsphere

import (
//	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
//	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVsphereDatastore_normal(t *testing.T) {

	resource.Test( t, 
		resource.TestCase {
			PreCheck: func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckDatastoreDestroy,
			Steps: []resource.TestStep {
				},
		} )
}

func testAccCheckDatastoreDestroy(s *terraform.State) error {
	return nil
}
