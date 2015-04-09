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

type EsxHost struct {
	IP string
	User string
	Password string
	License string
}

var testEsxHost EsxHost

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
	
	testEsxHost.IP = os.Getenv("ESX_HOST")
	testEsxHost.User = os.Getenv("ESX_USER")
	testEsxHost.Password = os.Getenv("ESX_PASSWORD")
	testEsxHost.License = os.Getenv("ESX_LICENSE")	
	
	if username == "" || password == "" || host == "" || testEsxHost.IP == "" || testEsxHost.User == "" || testEsxHost.Password =="" {
		t.Fatal("VSPHERE_USERNAME, VSPHERE_PASSWORD, VSPHERE_HOST, ESX_HOST, and ESX_USER and ESX_PASSWORD must be set for acceptance tests to work.")
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
