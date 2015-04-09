package vsphere

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/kr/pretty"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var keepResourcePool bool

func TestAccVsphereResourcePool_normal(t *testing.T) {
	
	keepResourcePool = false

	_, filename, _, _ := runtime.Caller(0)
	ut := os.Getenv("UNIT_TEST")
	if ut == "" || ut == filepath.Base(filename) {
		
		resource.Test( t, 
			resource.TestCase {
				PreCheck: func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				CheckDestroy: testAccCheckResourcePoolDestroy,
				Steps: []resource.TestStep {
					resource.TestStep {
						Config: testAccResourcePoolConfig,
						Check: resource.ComposeTestCheckFunc(
							testAccCheckResourcePoolExists("vsphere_resource_pool.rp1"),
							resource.TestCheckResourceAttr("vsphere_resource_pool.rp1", "name", "resource_pool1"),
							
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "cpu.0.shares", "low"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "cpu.0.reservation", "512"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "cpu.0.expandable_reservation", "false"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "cpu.0.limit", "1024"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "memory.0.shares", "40960"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "memory.0.reservation", "4096"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "memory.0.expandable_reservation", "true"),
							resource.TestCheckResourceAttr(
								"vsphere_resource_pool.rp1", "memory.0.limit", "8192"),
						),
					},
				},
			} )
	}
}

func testAccCheckResourcePoolExists(resource string) resource.TestCheckFunc {
	
	return func(s *terraform.State) error {
		
		var err error

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("resource pool '%s' not found in terraform state", resource)
		}
		
		log.Printf("[DEBUG] Terraform resource pool: %# v", pretty.Formatter(rs))

		attributes := rs.Primary.Attributes
		resourcePoolName := rs.Primary.ID
		
		datacenterName := attributes["datacenter_id"]
		parentName := attributes["parent_id"]
		
		resourcePool, err := findTestResourcePool(resourcePoolName, parentName, datacenterName)
		if err != nil {
			return err
		}
		
		var mrp mo.ResourcePool
		
		ps := []string{"config"}
		err = resourcePool.Properties(context.Background(), resourcePool.Reference(), ps, &mrp)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Cluster config from read via VMOMI: %# v", pretty.Formatter(mrp.Config))
		
		if err := verifyResourceAllocation("cpu", mrp.Config.CpuAllocation, rs.Primary); err != nil {
			return err		
		}
		if err := verifyResourceAllocation("memory", mrp.Config.MemoryAllocation, rs.Primary); err != nil {
			return err		
		}
		
		keepResourcePool = (attributes["keep"] == "true")
		return nil
	}
}

func testAccCheckResourcePoolDestroy(s *terraform.State) error {

	const rp1 = "vsphere_resource_pool.rp1"
	const datacenter3 = "datacenter3"
	const cluster3 = "cluster3"
	const resourcePool1 = "resource_pool1"

	var(
		ok bool
		err error
	)

	_, ok = s.RootModule().Resources[rp1]
	if ok {
		return fmt.Errorf("resource pool '%s' still exists in the terraform state", rp1)
	}

	resourcePool, err := findTestResourcePool(resourcePool1, cluster3, datacenter3)
	if err != nil {
		log.Printf("[DEBUG] Resource pool '%s' destroyed as expected. API response was: %s", resourcePool1, err.Error())
	} else if keepResourcePool {
		log.Printf("[DEBUG] Resource pool '%s' not destroyed as expected", resourcePool.InventoryPath)
	} else {
		return fmt.Errorf("resource pool '%s' was not destroyed as expected", resourcePool.InventoryPath)
	}
	
	return nil
}

func findTestResourcePool(resourcePoolName, parentName string, datacenterName string) (*object.ResourcePool, error) {
	
	finder, err := getTestFinder(datacenterName)
	if err != nil {
		return nil, err
	}
	
	resourcePool, err := getResourcePool(resourcePoolName, parentName, finder)
	if err != nil {
		return nil, err
	}
	
	return resourcePool, nil
}

func verifyResourceAllocation(allocType string, allocInfo types.ResourceAllocationInfo, instance *terraform.InstanceState) error {
	
	attributes := instance.Attributes
	num, err := strconv.Atoi(attributes[fmt.Sprintf("%s.#", allocType)])
	if err != nil {
		return err
	}
	
	if num > 1 {
		return fmt.Errorf("more than configuration section for resource allocation type for %s was found", allocType)
	}
	if num == 1 {
		
		if strconv.FormatInt(allocInfo.Reservation, 10) != attributes[fmt.Sprintf("%s.0.reservation", allocType)] {
			fmt.Errorf("resource allocation %s reservation value mis-match", allocType)
		}
		if strconv.FormatBool(allocInfo.ExpandableReservation) != attributes[fmt.Sprintf("%s.0.expandable_reservation", allocType)] {
			fmt.Errorf("resource allocation %s expandable reservation value mis-match", allocType)
		}
		if strconv.FormatInt(allocInfo.Limit, 10) != attributes[fmt.Sprintf("%s.0.limit", allocType)] {
			fmt.Errorf("resource allocation %s limit value mis-match", allocType)
		}
		
		v := attributes[fmt.Sprintf("%s.0.shares", allocType)]
		level := types.SharesLevel(v)
		if level != types.SharesLevelLow && 
			level != types.SharesLevelNormal && 
			level != types.SharesLevelHigh &&
			(level != types.SharesLevelCustom || strconv.Itoa(allocInfo.Shares.Shares) != v) {
		
			fmt.Errorf("resource allocation %s shares structure mis-match", allocType)
		}
	}
	
	return nil
} 

const testAccResourcePoolConfig = `

resource "vsphere_datacenter" "dc3" {
	name = "datacenter3"

#	keep = true
}

resource "vsphere_cluster" "c3" {
	name = "cluster3"
	datacenter_id = "${vsphere_datacenter.dc3.id}"
  
	drs {}
	ha {}

#	keep = true
}

resource "vsphere_resource_pool" "rp1" {
	name = "resource_pool1"
	datacenter_id = "${vsphere_datacenter.dc3.id}"
	parent_id = "${vsphere_cluster.c3.id}"
	
	cpu {
		shares = "low"
		reservation = 512
		expandable_reservation = false
		limit = 1024
	}
	
	memory {
		shares = "40960"
		reservation = 4096
		expandable_reservation = true
		limit = 8192
	}
	
#	keep = false
}
`
