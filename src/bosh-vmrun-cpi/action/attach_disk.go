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

func (c AttachDiskMethod) AttachDiskV2(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) (apiv1.DiskHint, error) {
	var err error
	var agentEnv apiv1.AgentEnv
	vmId := "vm-" + vmCID.AsString()
	diskId := "disk-" + diskCID.AsString()

	if !c.driverClient.HasDisk(diskId) {
		return apiv1.DiskHint{}, fmt.Errorf("disk does not exist: %s", diskId)
	}

	err = c.driverClient.StopVM(vmId)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	err = c.driverClient.AttachDisk(vmId, diskId)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	currentIsoPath := c.driverClient.GetVMIsoPath(vmId)
	agentEnv, err = c.agentSettings.GetIsoAgentEnv(currentIsoPath)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	diskHint := apiv1.NewDiskHintFromMap(
		map[string]interface{}{
			"path":      "/dev/sdc", //can be removed?
			"volume_id": "2",        //should be 3?
			"lun":       "0",
		})

	agentEnv.AttachPersistentDisk(diskCID, diskHint)

	envIsoPath, err := c.agentSettings.GenerateAgentEnvIso(agentEnv)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	err = c.driverClient.UpdateVMIso(vmId, envIsoPath)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	c.agentSettings.Cleanup()

	err = c.driverClient.StartVM(vmId)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	return diskHint, nil
}

func (c AttachDiskMethod) AttachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	_, err := c.AttachDiskV2(vmCID, diskCID)
	return err
}
