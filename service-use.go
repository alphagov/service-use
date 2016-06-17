package main

import (
	"fmt"
	"github.com/cdelashmutt-pivotal/service-use/apihelper"
	"github.com/cloudfoundry/cli/plugin"
	"os"
)

type ServiceUsePlugin struct {
	apiHelper apihelper.CFAPIHelper
	cli       plugin.CliConnection
}

func (cmd *ServiceUsePlugin) ServiceUseCommand(args []string) {
	fmt.Println("---Getting service instances")

	if nil == cmd.cli {
		fmt.Println("ERROR: CLI Connection is nil!")
		os.Exit(1)
	}

	services, _ := cmd.getServices()

	for _, service := range services {
		fmt.Printf("Service %s was found.\n", service.label)
		for _, serviceplan := range service.plans {
			fmt.Printf("-Plan: %s\n", serviceplan.name)
			for _, serviceinstance := range serviceplan.serviceinstances {
				fmt.Printf("--Instance: %s\n", serviceinstance.name)
				fmt.Printf("---Org: %s\n", serviceinstance.space.organization.name)
				fmt.Printf("---Space: %s\n", serviceinstance.space.name)
			}
		}
		fmt.Printf("\n")
	}
}

type service struct {
	label string
	plans []serviceplan
}

func (cmd *ServiceUsePlugin) getServices() ([]service, error) {
	rawServices, err := cmd.apiHelper.GetServices(cmd.cli)
	if nil != err {
		return nil, err
	}

	var services = []service{}

	for _, s := range rawServices {

		serviceplans, err := cmd.getServicePlans(s.ServicePlansURL)
		if nil != err {
			return nil, err
		}

		services = append(services, service{
			label: s.Label,
			plans: serviceplans,
		})
	}
	return services, nil
}

type serviceplan struct {
	name             string
	serviceinstances []serviceinstance
}

func (cmd *ServiceUsePlugin) getServicePlans(servicePlansURL string) ([]serviceplan, error) {
	rawServicePlans, err := cmd.apiHelper.GetServicePlans(cmd.cli, servicePlansURL)
	if nil != err {
		return nil, err
	}

	var serviceplans = []serviceplan{}

	for _, sp := range rawServicePlans {

		serviceinstances, err := cmd.getServiceInstances(sp.ServiceInstancesURL)
		if nil != err {
			return nil, err
		}
		serviceplans = append(serviceplans, serviceplan{
			name:             sp.Name,
			serviceinstances: serviceinstances,
		})
	}
	return serviceplans, nil
}

type serviceinstance struct {
	name  string
	space space
}

func (cmd *ServiceUsePlugin) getServiceInstances(serviceInstancesURL string) ([]serviceinstance, error) {
	rawServiceInstances, err := cmd.apiHelper.GetServiceInstances(cmd.cli, serviceInstancesURL)
	if nil != err {
		return nil, err
	}

	var serviceinstances = []serviceinstance{}

	for _, si := range rawServiceInstances {

		space, err := cmd.getSpace(si.SpaceURL)
		if nil != err {
			return nil, err
		}

		serviceinstances = append(serviceinstances, serviceinstance{
			name:  si.Name,
			space: space,
		})
	}
	return serviceinstances, nil
}

type space struct {
	name         string
	organization organization
}

func (cmd *ServiceUsePlugin) getSpace(spaceURL string) (space, error) {

	rawSpace, err := cmd.apiHelper.GetSpace(cmd.cli, spaceURL)
	if nil != err {
		return space{}, err
	}

	organization, err := cmd.getOrganization(rawSpace.OrganizationURL)
	if nil != err {
		return space{}, err
	}

	space := space{
		name:         rawSpace.Name,
		organization: organization,
	}
	return space, nil
}

type organization struct {
	name string
}

var orgCache map[string]organization = make(map[string]organization)

func (cmd *ServiceUsePlugin) getOrganization(organizationURL string) (organization, error) {
	if retOrg, present := orgCache[organizationURL]; !present {
		retOrg, _ = cmd.actualGetOrganization(organizationURL)
		orgCache[organizationURL] = retOrg
	}
	return orgCache[organizationURL], nil
}

func (cmd *ServiceUsePlugin) actualGetOrganization(organizationURL string) (organization, error) {
	rawOrg, err := cmd.apiHelper.GetOrganization(cmd.cli, organizationURL)
	if nil != err {
		return organization{}, err
	}

	organization := organization{
		name: rawOrg.Name,
	}
	return organization, nil
}

func (cmd *ServiceUsePlugin) Run(cli plugin.CliConnection, args []string) {

	if args[0] == "service-use" {
		cmd.apiHelper = &apihelper.APIHelper{}
		cmd.cli = cli
		cmd.ServiceUseCommand(args)
	}
}

func (cmd *ServiceUsePlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "ServiceUsePlugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "service-use",
				HelpText: "Infomation about service instances and bound apps",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "service-use\n   cf service-use",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(ServiceUsePlugin))
}
