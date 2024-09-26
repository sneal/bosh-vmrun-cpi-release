package action

import (
	"bosh-vmrun-cpi/driver"

	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type CreateDiskMethod struct {
	driverClient driver.Client
	uuidGen      boshuuid.Generator
}

func NewCreateDiskMethod(driverClient driver.Client, uuidGen boshuuid.Generator) CreateDiskMethod {
	return CreateDiskMethod{
		driverClient: driverClient,
		uuidGen:      uuidGen,
	}
}

func (c CreateDiskMethod) CreateDisk(sizeMB int,
	cloudProps apiv1.DiskCloudProps, associatedVMCID *apiv1.VMCID) (apiv1.DiskCID, error) {

	diskUuid, _ := c.uuidGen.Generate()
	diskId := "disk-" + diskUuid
	newDiskCID := apiv1.NewDiskCID(diskUuid)

	err := c.driverClient.CreateDisk(diskId, sizeMB)
	if err != nil {
		return newDiskCID, err
	}
	return newDiskCID, nil
}
