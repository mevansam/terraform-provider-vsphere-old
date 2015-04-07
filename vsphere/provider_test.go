package vsphere

import (
	"fmt"
	"log"
	"os"
	"testing"

	"golang.org/x/net/context"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"vsphere": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	host := os.Getenv("VSPHERE_HOST")
	username := os.Getenv("VSPHERE_USERNAME")
	password := os.Getenv("VSPHERE_PASSWORD")
	if username == "" || password == "" || host == "" {
		t.Fatal("VSPHERE_USERNAME, VSPHERE_PASSWORD and VSPHERE_HOST must be set for acceptance tests to work.")
	}
}

func getTestFinder(datacenterName string) (*find.Finder, error) {
	
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
	
	return finder, nil
}

func testFindResource(s *terraform.State, id string, dependencies []string) (*terraform.ResourceState, error) {

	for _, d := range dependencies {
		
		rs := s.RootModule().Resources[d]
		if id == rs.Primary.ID {
			return rs, nil
		}
	}
	
	return nil, fmt.Errorf("unable to find resource with id '%s' in terraform state.", id)
}
