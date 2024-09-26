package action

import (
	"bosh-vmrun-cpi/driver"
	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
)

type HasDiskMethod struct {
	driverClient driver.Client
}

func NewHasDiskMethod(driverClient driver.Client) HasDiskMethod {
	return HasDiskMethod{
		driverClient: driverClient,
	}
}

func (c HasDiskMethod) HasDisk(diskCid apiv1.DiskCID) (bool, error) {
	diskId := "disk-" + diskCid.AsString()
	diskFound := c.driverClient.HasDisk(diskId)
	return diskFound, nil
}
