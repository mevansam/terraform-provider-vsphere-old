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
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/kr/pretty"
)

var keepClusters bool // keep must have the same value for both test clusters. otherwise the last value wins.

func TestAccVsphereCluster_normal(t *testing.T) {
	
	keepClusters = false
	
	_, filename, _, _ := runtime.Caller(0)
	ut := os.Getenv("UNIT_TEST")
	if ut == "" || ut == filepath.Base(filename) {
		
		resource.Test( t, 
			resource.TestCase {
				PreCheck: func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				CheckDestroy: testAccCheckClusterDestroy,
				Steps: []resource.TestStep {
					resource.TestStep {
						Config: testAccClusterConfig,
						Check: resource.ComposeTestCheckFunc(
							testAccCheckClusterExists("vsphere_cluster.c1"),
							testAccCheckClusterExists("vsphere_cluster.c2"),
							resource.TestCheckResourceAttr("vsphere_cluster.c1", "name", "cluster1"),
							resource.TestCheckResourceAttr("vsphere_cluster.c1", "drs.0.default_automation_level", "manual"),
							resource.TestCheckResourceAttr("vsphere_cluster.c1", "drs.0.migration_threshold", "2"),
							resource.TestCheckResourceAttr("vsphere_cluster.c1", "ha.0.host_monitoring", "disabled"),
							resource.TestCheckResourceAttr("vsphere_cluster.c1", "ha.0.vm_monitoring", "vmMonitoringOnly"),
							resource.TestCheckResourceAttr("vsphere_cluster.c2", "name", "cluster2"),
							resource.TestCheckResourceAttr("vsphere_cluster.c2", "drs.0.default_automation_level", "partiallyAutomated"),
							resource.TestCheckResourceAttr("vsphere_cluster.c2", "drs.0.migration_threshold", "4"),
							resource.TestCheckResourceAttr("vsphere_cluster.c2", "ha.0.host_monitoring", "enabled"),
							resource.TestCheckResourceAttr("vsphere_cluster.c2", "ha.0.vm_monitoring", "vmAndAppMonitoring"),
						),
					},
				},
			} )
	}
}

func findTestCluster(datacenterName string, clusterName string) (*object.ClusterComputeResource, error) {
	
	client := testAccProvider.Meta().(*govmomi.Client)
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	
	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.Datacenter(context.Background(), datacenterName)
	if err != nil {
		log.Printf("[ERROR] Unable find datacenter: '%s'", datacenterName)
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	cluster, err := finder.Cluster(context.Background(), clusterName)
	if err != nil {
		log.Printf("[ERROR] Unable find cluster: '%s'", clusterName)
		return nil, err
	}
	
	return cluster, nil
}

func testAccCheckClusterExists(resource string) resource.TestCheckFunc {
	
	return func(s *terraform.State) error {
		
		var err error

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("cluster '%s' not found in terraform state", resource)
		}

		attributes := rs.Primary.Attributes

		clusterName := rs.Primary.ID
		datacenterName := attributes["datacenter_id"]

		cluster, err := findTestCluster(datacenterName, clusterName)
		if err != nil {
			return err
		}		
		config, err := cluster.Configuration(context.Background())
		if err != nil {
			log.Printf("[ERROR] Unable read configuration for cluster '%s'.", clusterName)
			return err
		}

		log.Printf("[DEBUG] Cluster state: %# v", pretty.Formatter(rs.Primary))
		log.Printf("[DEBUG] Cluster config from read via VMOMI: %# v", pretty.Formatter(config))
		
		if v, _ := strconv.Atoi(attributes["drs.#"]); !config.DrsConfig.Enabled || v != 1 {
			fmt.Errorf("dynamice resource scheduler enabled but not reflected in terraform state")
		}
		if string(config.DrsConfig.DefaultVmBehavior) != attributes["drs.0.default_automation_level"] {
			fmt.Errorf("high-availability vm monitoring attribute mis-match")
		}
		if strconv.Itoa(config.DrsConfig.VmotionRate) != attributes["drs.0.migration_threshold"] {
			fmt.Errorf("high-availability vm monitoring attribute mis-match")
		}
		if v, _ := strconv.Atoi(attributes["ha.#"]); !config.DasConfig.Enabled || v != 1 {
			fmt.Errorf("high-availability enabled but not reflected in terraform state")
		}
		if config.DasConfig.VmMonitoring != attributes["ha.0.vm_monitoring"] {
			fmt.Errorf("high-availability vm monitoring attribute mis-match")
		}
		if config.DasConfig.HostMonitoring != attributes["ha.0.host_monitoring"] {
			fmt.Errorf("high-availability host monitoring attribute mis-match")
		}
		if strconv.FormatBool(config.DasConfig.AdmissionControlEnabled) != attributes["ha.0.admissionControlEnabled"] {
			fmt.Errorf("high-availability adminission control attribute mis-match")
		}
		
		keepClusters = (attributes["keep"] == "true")
		return nil;
	}
}

func testAccCheckClusterDestroy(s *terraform.State) error {

	const datacenter = "datacenter2"
	const resource1 = "vsphere_cluster.c1"
	const resource2 = "vsphere_cluster.c2"
	const cluster1 = "cluster1"
	const cluster2 = "cluster2"
	
	var(
		ok bool
		err error
	)

	_, ok = s.RootModule().Resources[resource1]
	if ok {
		return fmt.Errorf("cluster '%s' still exists in the terraform state", cluster1)
	}
	_, ok = s.RootModule().Resources[resource2]
	if ok {
		return fmt.Errorf("cluster '%s' still exists in the terraform state", cluster2)
	}
	
	_, err = findTestCluster(datacenter, cluster1)
	if err != nil {
		log.Printf("[DEBUG] Cluster '%s' destroyed as expected. API response was: %s", cluster1, err.Error())
	} else if keepClusters {
		log.Printf("[DEBUG] Cluster '%s' not destroyed as expected.", cluster1)
	} else {
		return fmt.Errorf("datacenter '%s' and cluster '%s' was not destroyed as expected", datacenter, cluster1);		
	} 
	
	_, err = findTestCluster(datacenter, cluster2)
	if err != nil {
		log.Printf("[DEBUG] Cluster '%s' destroyed as expected. API response was: %s", cluster2, err.Error())
	} else if keepClusters {
		log.Printf("[DEBUG] Cluster '%s' not destroyed as expected.", cluster2)
	} else {
		return fmt.Errorf("datacenter '%s' and cluster '%s' was not destroyed as expected", datacenter, cluster2);		
	} 

	return nil
}

const testAccClusterConfig = `

resource "vsphere_datacenter" "dc2" {
	name = "datacenter2"

#	keep = true
}

resource "vsphere_cluster" "c1" {
	name = "cluster1"
	datacenter_id = "${vsphere_datacenter.dc2.id}"
  
	drs {
		default_automation_level = "manual"
		migration_threshold = 2
	}
	ha {
		host_monitoring = "disabled"
		vm_monitoring = "vmMonitoringOnly"
	}

#	keep = true
}

resource "vsphere_cluster" "c2" {
	name = "cluster2"
	datacenter_id = "${vsphere_datacenter.dc2.id}"

	drs {
		default_automation_level = "partiallyAutomated"
		migration_threshold = 4
	}
	ha {
		host_monitoring = "enabled"
		vm_monitoring = "vmAndAppMonitoring"
	}

#	keep = true
}
`
