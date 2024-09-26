package action

import (
	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-vmrun-cpi/driver"
)

type DeleteVMMethod struct {
	driverClient driver.Client
	logger       boshlog.Logger
}

func NewDeleteVMMethod(driverClient driver.Client, logger boshlog.Logger) DeleteVMMethod {
	return DeleteVMMethod{
		driverClient: driverClient,
		logger:       logger,
	}
}

func (c DeleteVMMethod) DeleteVM(vmCid apiv1.VMCID) error {
	vmId := "vm-" + vmCid.AsString()
	err := c.driverClient.DestroyVM(vmId)
	if err != nil {
		c.logger.Error("cpi", "deleting vm: %s\n", vmCid)
		return err
	}

	return nil
}
