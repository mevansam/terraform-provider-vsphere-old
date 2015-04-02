package vsphere

//import (
//	"fmt"
//	"testing"
//
//	"github.com/hashicorp/terraform/helper/resource"
////	"github.com/hashicorp/terraform/helper/schema"
//	"github.com/hashicorp/terraform/terraform"
//)
//
//func TestAccVsphereCluster_normal(t *testing.T) {
//	
//	resource.Test( t, 
//		resource.TestCase {
//			PreCheck: func() { testAccPreCheck(t) },
//			Providers: testAccProviders,
//			CheckDestroy: testAccCheckClusterDestroy,
//			Steps: []resource.TestStep {
//				resource.TestStep {
//					Config: testAccClusterConfig,
//					Check: resource.ComposeTestCheckFunc(
//						testAccCheckClusterExists("vsphere_datacenter.tdc"),
//					),
//				},
//			},
//		} )
//}
//
//func testAccCheckClusterExists(n string) resource.TestCheckFunc {
//	
//	return func(s *terraform.State) error {
//
//		fmt.Println("Checking state...")
//		for _, rs := range s.RootModule().Resources {
//			fmt.Println(" ==> " + rs.Type)
//		}
//		
//		return nil;
//	}
//}
//
//func testAccCheckClusterDestroy(s *terraform.State) error {
//	return nil
//}
//
//const testAccClusterConfig = `
//
//resource "vsphere_datacenter" "tdc" {
//	name = "datacenter"
//}
//
//resource "vsphere_cluster" "tc1" {
//  name = "cluster1"
//  datacenter_id = "${vsphere_datacenter.tdc.id}"
//  
//  drs {
//    default_automation_level = "manual"
//    migration_threshold = 2
//  }
//  ha {
//    host_monitoring = "disabled"
//    vm_monitoring = "vmMonitoringOnly"
//  }
//}
//
//resource "vsphere_cluster" "tc2" {
//  name = "cluster2"
//  datacenter_id = "${vsphere_datacenter.tdc.id}"
//
//  drs {
//    default_automation_level = "partiallyAutomated"
//    migration_threshold = 4
//  }
//  ha {
//    host_monitoring = "enabled"
//    vm_monitoring = "vmAndAppMonitoring"
//  }
//}
//
//`
