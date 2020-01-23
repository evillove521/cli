// +build V7

package rpc

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"sync"

	"code.cloudfoundry.org/cli/command"
	plugin "code.cloudfoundry.org/cli/plugin/v7"
	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

type CliRpcCmd struct {
	PluginMetadata       *plugin.PluginMetadata
	MetadataMutex        *sync.RWMutex
	Config               command.Config
	PluginActor          PluginActor
	outputCapture        OutputCapture
	terminalOutputSwitch TerminalOutputSwitch
	outputBucket         *bytes.Buffer
	stdout               io.Writer
}

func (cmd *CliRpcCmd) ApiEndpoint(args string, retVal *string) error {
	*retVal = cmd.Config.Target()

	return nil
}

func (cmd *CliRpcCmd) IsMinCliVersion(passedVersion string, retVal *bool) error {
	if version.VersionString() == version.DefaultVersion {
		*retVal = true
		return nil
	}

	actualVersion, err := semver.Make(version.VersionString())
	if err != nil {
		return err
	}

	requiredVersion, err := semver.Make(passedVersion)
	if err != nil {
		return err
	}

	*retVal = actualVersion.GTE(requiredVersion)

	return nil
}

func (cmd *CliRpcCmd) SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool) error {
	cmd.MetadataMutex.Lock()
	defer cmd.MetadataMutex.Unlock()

	cmd.PluginMetadata = &pluginMetadata
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) DisableTerminalOutput(disable bool, retVal *bool) error {
	cmd.terminalOutputSwitch.DisableTerminalOutput(disable)
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) CliCommand(args []string, retVal *[]string) error {
	outBuffer := bytes.Buffer{}
	errBuffer := bytes.Buffer{}

	p, err := command_parser.NewCommandParserForPlugins(&outBuffer, &errBuffer)
	if err != nil {
		return err
	}
	exitCode := p.ParseCommandFromArgs(args)

	if exitCode == command_parser.UnknownCommandCode {
		return errors.New("UnknownCommandCode")
	} else if exitCode > 0 {
		return errors.New("Some other error")
	}

	*retVal = []string{outBuffer.String(), errBuffer.String()}

	return nil
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	return errors.New("unimplemented")
}

func (cmd *CliRpcCmd) GetApp(appName string, retVal *plugin_models.DetailedApplicationSummary) error {
	if !cmd.Config.HasTargetedOrganization() {
		return errors.New("no organization targeted")
	}
	spaceGUID := cmd.Config.TargetedSpace().GUID
	if spaceGUID == "" {
		return errors.New("no space targeted")
	}

	app, _, err := cmd.PluginActor.GetDetailedAppSummary(appName, spaceGUID, true)

	if err != nil {
		return err
	}
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err = enc.Encode(app)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(retVal)
	if err != nil {
		panic(err)
	}
	return nil
}

func (cmd *CliRpcCmd) GetApps(appName string, retVal *[]plugin_models.Application) error {
	spaceGUID := cmd.Config.TargetedSpace().GUID
	if spaceGUID == "" {
		return errors.New("no space targeted")
	}
	applications, warnings, err := cmd.PluginActor.GetApplicationsBySpace(spaceGUID)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		fmt.Printf("GetApps warning: %s\n", warning)
	}
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err = enc.Encode(applications)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(retVal)
	if err != nil {
		panic(err)
	}
	return nil
}

func (cmd *CliRpcCmd) GetCurrentOrg(args string, retVal *plugin_models.Org) error {
	org := cmd.Config.TargetedOrganization()
	if org.Name == "" {
		return errors.New("no organization targeted")
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err := enc.Encode(org)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(retVal)
	if err != nil {
		panic(err)
	}
	return nil
}

func (cmd *CliRpcCmd) GetCurrentSpace(args string, retVal *plugin_models.Space) error {
	orgGUID := cmd.Config.TargetedOrganization().GUID
	if orgGUID == "" {
		return errors.New("no organization targeted")
	}
	spaceName := cmd.Config.TargetedSpace().Name
	if spaceName == "" {
		return errors.New("no space targeted")
	}
	space, _, err := cmd.PluginActor.GetSpaceByNameAndOrganization(spaceName, orgGUID)

	if err != nil {
		return err
	}
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err = enc.Encode(space)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(retVal)
	if err != nil {
		panic(err)
	}
	return nil
}

func (cmd *CliRpcCmd) GetOrg(orgName string, retVal *plugin_models.OrgSummary) error {
	org, _, err := cmd.PluginActor.GetOrganizationByName(orgName)
	if err != nil {
		return err
	}

	spaces, _, err := cmd.PluginActor.GetOrganizationSpaces(org.GUID)
	if err != nil {
		return err
	}

	domains, _, err := cmd.PluginActor.GetOrganizationDomains(org.GUID, "")
	if err != nil {
		return err
	}

	for _, space := range spaces {
		newSpace := plugin_models.Space{
			GUID: space.GUID,
			Name: space.Name,
		}
		retVal.Spaces = append(retVal.Spaces, newSpace)
	}

	for _, domain := range domains {
		newDomain := plugin_models.Domain{GUID: domain.GUID, Name: domain.Name}
		retVal.Domains = append(retVal.Domains, newDomain)
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err = enc.Encode(org.Metadata)
	if err != nil {
		panic(err)
	}

	var metadata plugin_models.Metadata

	err = dec.Decode(&metadata)
	if err != nil {
		panic(err)
	}

	retVal.Name = org.Name
	retVal.GUID = org.GUID
	retVal.Metadata = metadata

	return nil
}

func (cmd *CliRpcCmd) GetSpace(spaceName string, retVal *plugin_models.Space) error {
	orgGUID := cmd.Config.TargetedOrganization().GUID
	if orgGUID == "" {
		return errors.New("no organization targeted")
	}
	space, _, err := cmd.PluginActor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err = enc.Encode(space)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(retVal)
	if err != nil {
		panic(err)
	}

	return nil
}

func (cmd *CliRpcCmd) GetSpaces(not_used string, retVal *[]plugin_models.Space) error {
	orgGUID := cmd.Config.TargetedOrganization().GUID
	if orgGUID == "" {
		return errors.New("no organization targeted")
	}
	spaces, _, err := cmd.PluginActor.GetOrganizationSpaces(orgGUID)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	err = enc.Encode(spaces)
	if err != nil {
		return err
	}

	err = dec.Decode(retVal)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *CliRpcCmd) IsLoggedIn(args string, retVal *bool) error {
	*retVal = cmd.Config.AccessToken() != ""
	return nil
}

func (cmd *CliRpcCmd) IsSkipSSLValidation(args string, retVal *bool) error {
	*retVal = cmd.Config.SkipSSLValidation()
	return nil
}

func (cmd *CliRpcCmd) AccessToken(args string, retVal *string) error {
	accessToken, err := cmd.PluginActor.RefreshAccessToken()
	if err != nil {
		panic(err)
	}
	*retVal = accessToken
	return nil
}

func (cmd *CliRpcCmd) Username(args string, retVal *string) error {
	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return fmt.Errorf("error processing config: %s", err.Error())
	}
	if username == "" {
		return errors.New("not logged in")
	}
	*retVal = username

	return nil
}
