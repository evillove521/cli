// +build V7

package v7

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
)

type cliConnection struct {
	cliServerPort string
}

func NewCliConnection(cliServerPort string) *cliConnection {
	return &cliConnection{
		cliServerPort: cliServerPort,
	}
}

func (c *cliConnection) AccessToken() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.AccessToken", "", &result)
	})

	return result, err
}

func (c *cliConnection) ApiEndpoint() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.ApiEndpoint", "", &result)
	})

	return result, err
}

func (c *cliConnection) CliCommand(args ...string) ([]string, error) {
	var result []string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.CliCommand", args, &result)
	})

	return result, err
}

func (c *cliConnection) GetApp(appName string) (plugin_models.DetailedApplicationSummary, error) {
	var result plugin_models.DetailedApplicationSummary

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetApp", appName, &result)
	})

	return result, err
}

func (c *cliConnection) GetApps() ([]plugin_models.Application, error) {
	var result []plugin_models.Application

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetApps", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetCurrentOrg() (plugin_models.Org, error) {
	var result plugin_models.Org

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetCurrentOrg", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetCurrentSpace() (plugin_models.CurrentSpace, error) {
	var result plugin_models.CurrentSpace

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetCurrentSpace", "", &result)
	})

	return result, err
}

func (c *cliConnection) GetOrg(orgName string) (plugin_models.OrgSummary, error) {
	var result plugin_models.OrgSummary

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetOrg", orgName, &result)
	})

	return result, err
}

func (c *cliConnection) GetSpace(spaceName string) (plugin_models.Space, error) {
	var result plugin_models.Space

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetSpace", spaceName, &result)
	})

	return result, err
}

func (c *cliConnection) GetSpaces() ([]plugin_models.Space, error) {
	var result []plugin_models.Space

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.GetSpaces", "", &result)
	})
	return result, err
}

func (c *cliConnection) IsLoggedIn() (bool, error) {
	var result bool
	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.IsLoggedIn", "", &result)
	})
	return result, err
}

func (c *cliConnection) Username() (string, error) {
	var result string

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.Username", "", &result)
	})

	return result, err
}

func (c *cliConnection) IsSkipSSLValidation() (bool, error) {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.IsSkipSSLValidation", "", &result)
	})

	return result, err
}

func (c *cliConnection) isMinCliVersion(version string) bool {
	var result bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.IsMinCliVersion", version, &result)
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return result
}

func (c *cliConnection) pingCLI() {
	//call back to cf saying we have been setup
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+c.cliServerPort)
		if connErr != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			conn.Close()
			break
		}
	}
	if connErr != nil {
		fmt.Println(connErr)
		os.Exit(1)
	}
}

func (c *cliConnection) sendPluginMetadataToCliServer(metadata PluginMetadata) {
	var success bool

	err := c.withClientDo(func(client *rpc.Client) error {
		return client.Call("CliRpcCmd.SetPluginMetadata", metadata, &success)
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !success {
		os.Exit(1)
	}

	os.Exit(0)
}

func (c *cliConnection) withClientDo(f func(client *rpc.Client) error) error {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+c.cliServerPort)
	if err != nil {
		return err
	}
	defer client.Close()

	return f(client)
}
