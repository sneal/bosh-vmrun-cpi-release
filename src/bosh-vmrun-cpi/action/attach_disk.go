package action

import (
	"fmt"

	"github.com/cloudfoundry/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/vm"
)

type AttachDiskMethod struct {
	driverClient    driver.Client
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
}

func NewAttachDiskMethod(driverClient driver.Client, agentSettings vm.AgentSettings) AttachDiskMethod {
	return AttachDiskMethod{
		driverClient:  driverClient,
		agentSettings: agentSettings,
	}
}

func (c AttachDiskMethod) AttachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	var err error
	var agentEnv apiv1.AgentEnv
	vmId := "vm-" + vmCID.AsString()
	diskId := "disk-" + diskCID.AsString()

	if !c.driverClient.HasDisk(diskId) {
		return fmt.Errorf("disk does not exist: %s", diskId)
	}

	err = c.driverClient.StopVM(vmId)
	if err != nil {
		return err
	}

	err = c.driverClient.AttachDisk(vmId, diskId)
	if err != nil {
		return err
	}

	currentIsoPath := c.driverClient.GetVMIsoPath(vmId)
	agentEnv, err = c.agentSettings.GetIsoAgentEnv(currentIsoPath)
	if err != nil {
		return err
	}

	agentEnv.AttachPersistentDisk(diskCID, apiv1.NewDiskHintFromMap(
		map[string]interface{}{
			"path":      "/dev/sdc", //can be removed?
			"volume_id": "2",        //should be 3?
			"lun":       "0",
		}))

	envIsoPath, err := c.agentSettings.GenerateAgentEnvIso(agentEnv)
	if err != nil {
		return err
	}

	err = c.driverClient.UpdateVMIso(vmId, envIsoPath)
	if err != nil {
		return err
	}

	c.agentSettings.Cleanup()

	err = c.driverClient.StartVM(vmId)
	if err != nil {
		return err
	}

	return nil
}
