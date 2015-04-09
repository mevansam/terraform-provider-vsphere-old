package vsphere

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi/object"
	"github.com/kr/pretty"
)

var keepHost bool

func TestAccVsphereStandaloneHost_normal(t *testing.T) {
	
	keepHost = false

	_, filename, _, _ := runtime.Caller(0)
	ut := os.Getenv("UNIT_TEST")
	if ut == "" || ut == filepath.Base(filename) {
		
		resource.Test( t, 
			resource.TestCase {
				PreCheck: func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				CheckDestroy: testAccCheckStandaloneHostDestroy,
				Steps: []resource.TestStep {
						resource.TestStep {
							Config: fmt.Sprintf( testAccStandaloneHostConfig, 
								testEsxHost.IP,
								testEsxHost.User,
								testEsxHost.Password,
								testEsxHost.License,
							),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckHostExists("vsphere_host.h4"),
								
								resource.TestCheckResourceAttr(
									"vsphere_host.h4", "host", testEsxHost.IP),
								resource.TestCheckResourceAttr(
									"vsphere_host.h4", "user", testEsxHost.User),
								resource.TestCheckResourceAttr(
									"vsphere_host.h4", "password", testEsxHost.Password),
								resource.TestCheckResourceAttr(
									"vsphere_host.h4", "license", testEsxHost.License),
							),
						},
					},
			} )
	}
}

func TestAccVsphereClusteredHost_normal(t *testing.T) {
	
	keepHost = false

	_, filename, _, _ := runtime.Caller(0)
	ut := os.Getenv("UNIT_TEST")
	if ut == "" || ut == filepath.Base(filename) {
		
		resource.Test( t, 
			resource.TestCase {
				PreCheck: func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				CheckDestroy: testAccCheckStandaloneHostDestroy,
				Steps: []resource.TestStep {
						resource.TestStep {
							Config: fmt.Sprintf( testAccClusteredHostConfig, 
								testEsxHost.IP,
								testEsxHost.User,
								testEsxHost.Password,
								testEsxHost.License,
							),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckHostExists("vsphere_host.h4"),
							),
						},
					},
			} )
	}
}

func testAccCheckHostExists(resource string) resource.TestCheckFunc {
	
	return func(s *terraform.State) error {
		
		var err error

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("host '%s' not found in terraform state", resource)
		}
		
		log.Printf("[DEBUG] Terraform host: %# v", pretty.Formatter(rs))
		
		attributes := rs.Primary.Attributes
		hostId := rs.Primary.ID
		
		hostName := attributes["host"]
		datacenterName := attributes["datacenter_id"]
		clusterName := attributes["cluster_id"]
		
		hostSystem, err := findTestHost(hostName, datacenterName, clusterName)
		if err != nil {
			return err
		}
		
		if hostSystem.Reference().Value != attributes["object_id"] {
			return fmt.Errorf("host object id mismatch. expected '%s' but go '%s'", hostSystem.Reference().Value, attributes["object_id"])
		}
		
		log.Printf("[DEBUG] Found host '%s' with id '%s' at path '%s'", hostName, hostId, hostSystem.InventoryPath)
		
		keepHost = (attributes["keep"] == "true")
		return nil
	}
}

func testAccCheckStandaloneHostDestroy(s *terraform.State) error {

	const h4 = "vsphere_host.h4"
	const datacenter4 = "datacenter4"

	_, ok := s.RootModule().Resources[h4]
	if ok {
		return fmt.Errorf("host '%s' still exists in the terraform state", h4)
	}

	err := testCheckHostDestroy(testEsxHost.IP, datacenter4, "")
	if err != nil {
		return err
	}
	
	return nil
}

func findTestHost(hostName, datacenterName string, clusterName string) (*object.HostSystem, error) {
	
	finder, err := getTestFinder(datacenterName)
	if err != nil {
		return nil, err
	}
	
	var cluster *string
	if clusterName == "" {
		cluster = nil
	} else {
		cluster = &clusterName
	}
	
	hostSystem, err := getHost(hostName, datacenterName, cluster, finder)
	if err != nil {
		return nil, err
	}
	
	return hostSystem, nil
}

func testCheckHostDestroy(hostName string, datacenterName string, clusterName string) error {
	
	hostSystem, err := findTestHost(hostName, datacenterName, clusterName)
	if err != nil {
		log.Printf("[DEBUG] Host '%s' destroyed as expected: %s", 
			hostName, err.Error())
	} else if keepHost {
		log.Printf("[DEBUG] Host '%s' with id '%s' not destroyed as expected", 
			hostName, hostSystem.Reference().Value)
	} else {
		return fmt.Errorf("host '%s' was not destroyed as expected", hostName)
	}
	
	return nil
}

const testAccStandaloneHostConfig = `

resource "vsphere_datacenter" "dc4" {
	name = "datacenter4"

#	keep = true
}

resource "vsphere_host" "h4" {
	host = "%s"
	datacenter_id = "${vsphere_datacenter.dc4.id}"
	
	user = "%s"
	password = "%s"
	license = "%s"
	
	ssl_no_verify = true
#	keep = true
}
`

const testAccClusteredHostConfig = `

resource "vsphere_datacenter" "dc4" {
	name = "datacenter4"

#	keep = true
}

resource "vsphere_cluster" "c4" {
	name = "cluster4"
	datacenter_id = "${vsphere_datacenter.dc4.id}"
  
	drs {}
	ha {}

#	keep = true
}

resource "vsphere_host" "h4" {
	host = "%s"
	datacenter_id = "${vsphere_datacenter.dc4.id}"
	cluster_id = "${vsphere_cluster.c4.id}"
	
	user = "%s"
	password = "%s"
	license = "%s"
	
	ssl_no_verify = true
#	keep = true
}
`