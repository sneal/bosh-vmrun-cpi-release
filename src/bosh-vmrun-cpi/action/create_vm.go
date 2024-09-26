package action

import (
	"fmt"

	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/vm"
)

type CreateVMMethod struct {
	driverClient    driver.Client
	agentSettings   vm.AgentSettings
	agentOptions    apiv1.AgentOptions
	agentEnvFactory apiv1.AgentEnvFactory
	uuidGen         boshuuid.Generator
	logger          boshlog.Logger
}

func NewCreateVMMethod(driverClient driver.Client, agentSettings vm.AgentSettings, agentOptions apiv1.AgentOptions, agentEnvFactory apiv1.AgentEnvFactory, uuidGen boshuuid.Generator, logger boshlog.Logger) CreateVMMethod {
	return CreateVMMethod{
		driverClient:    driverClient,
		agentSettings:   agentSettings,
		agentOptions:    agentOptions,
		agentEnvFactory: agentEnvFactory,
		uuidGen:         uuidGen,
		logger:          logger,
	}
}

func (c CreateVMMethod) CreateVMV2(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	associatedDiskCIDs []apiv1.DiskCID, vmEnv apiv1.VMEnv) (apiv1.VMCID, apiv1.Networks, error) {

	vmUuid, _ := c.uuidGen.Generate()
	newVMCID := apiv1.NewVMCID(vmUuid)

	stemcellId := "cs-" + stemcellCID.AsString()
	vmId := "vm-" + vmUuid

	if !c.driverClient.HasVM(stemcellId) {
		return newVMCID, apiv1.Networks{}, fmt.Errorf("stemcell does not exist: %s", stemcellId)
	}

	vmProps, err := vm.NewVMProps(cloudProps)
	if err != nil {
		return newVMCID, apiv1.Networks{}, err
	}

	err = c.driverClient.CloneVM(stemcellId, vmId)
	if err != nil {
		return newVMCID, networks, err
	}

	err = c.driverClient.SetVMResources(vmId, vmProps.CPU, vmProps.RAM)
	if err != nil {
		return newVMCID, networks, err
	}

	for _, network := range networks {
		adapterName, macAddress, err := c.agentSettings.GetNetworkSettings(network)
		if err != nil {
			return newVMCID, networks, err
		}

		network.SetMAC(macAddress)

		err = c.driverClient.SetVMNetworkAdapter(vmId, adapterName, macAddress)
		if err != nil {
			return newVMCID, networks, err
		}
	}

	agentEnv := c.agentEnvFactory.ForVM(agentID, newVMCID, networks, vmEnv, c.agentOptions)

	if vmProps.NeedsBootstrap() {
		err = c.driverClient.BootstrapVM(
			vmId,
			vmProps.Bootstrap.Script_Content,
			vmProps.Bootstrap.Script_Path,
			vmProps.Bootstrap.Interpreter_Path,
			vmProps.Bootstrap.Ready_Process_Name,
			vmProps.Bootstrap.Username,
			vmProps.Bootstrap.Password,
			vmProps.Bootstrap.Min_Wait,
			vmProps.Bootstrap.Max_Wait,
		)
		if err != nil {
			return newVMCID, networks, err
		}
	}

	agentEnv.AttachSystemDisk(apiv1.NewDiskHintFromString("0"))

	if vmProps.Disk > 0 {
		err = c.driverClient.CreateEphemeralDisk(vmId, vmProps.Disk)
		if err != nil {
			return newVMCID, networks, err
		}

		agentEnv.AttachEphemeralDisk(apiv1.NewDiskHintFromString("1"))
	}

	newIsoPath, err := c.agentSettings.GenerateAgentEnvIso(agentEnv)
	defer c.agentSettings.Cleanup()

	if err != nil {
		return newVMCID, networks, err
	}

	err = c.driverClient.UpdateVMIso(vmId, newIsoPath)
	if err != nil {
		return newVMCID, networks, err
	}

	if !c.driverClient.NeedsVMNameChange(vmId) {
		err = c.driverClient.StartVM(vmId)
		if err != nil {
			return newVMCID, networks, err
		}
	}

	return newVMCID, networks, nil
}

func (c CreateVMMethod) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	associatedDiskCIDs []apiv1.DiskCID, vmEnv apiv1.VMEnv) (apiv1.VMCID, error) {

	newVMCID, _, err := c.CreateVMV2(agentID, stemcellCID, cloudProps, networks, associatedDiskCIDs, vmEnv)
	return newVMCID, err
}
