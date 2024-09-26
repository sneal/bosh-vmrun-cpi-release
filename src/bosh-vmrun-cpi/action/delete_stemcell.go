package action

import (
	"bosh-vmrun-cpi/driver"
	"fmt"

	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type DeleteStemcellMethod struct {
	driverClient driver.Client
	logger       boshlog.Logger
}

func NewDeleteStemcellMethod(driverClient driver.Client, logger boshlog.Logger) DeleteStemcellMethod {
	return DeleteStemcellMethod{
		driverClient: driverClient,
		logger:       logger,
	}
}

func (c DeleteStemcellMethod) DeleteStemcell(stemcellCid apiv1.StemcellCID) error {
	stemcellId := "cs-" + stemcellCid.AsString()
	err := c.driverClient.DestroyVM(stemcellId)
	if err != nil {
		c.logger.Error("delete-stemcell", fmt.Sprintf("failed to delete stemcell. cid: %s", stemcellCid))
		return err
	}

	return nil
}
