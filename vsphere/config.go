package vsphere

import (
	"fmt"
	"net/url"
	
	"golang.org/x/net/context"
	"github.com/vmware/govmomi"
)

type Config struct {
	Host string
	Username string
	Password string
}

func (c *Config) Client() (*govmomi.Client, error) {
	
	sdkURL, err := url.Parse(
		fmt.Sprintf(
			"https://%s:%s@%s/sdk",
			c.Username,
			c.Password,
			c.Host ) )

	if err != nil {
		return nil, err
	}

	client, err := govmomi.NewClient(context.Background(), sdkURL, true)
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}
