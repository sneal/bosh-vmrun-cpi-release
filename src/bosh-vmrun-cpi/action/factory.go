package action

import (
	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-vmrun-cpi/config"
	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/stemcell"
	"bosh-vmrun-cpi/vm"
)

type Factory struct {
	driverClient    driver.Client
	stemcellClient  stemcell.StemcellClient
	stemcellStore   stemcell.StemcellStore
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
	config          config.Config
	fs              boshsys.FileSystem
	uuidGen         boshuuid.Generator
	logger          boshlog.Logger
}

type CPI struct {
	CreateStemcellMethod
	DeleteStemcellMethod
	CreateVMMethod
	DeleteVMMethod
	HasVMMethod
	SetVMMetadataMethod
	AttachDiskMethod
	CreateDiskMethod
	DetachDiskMethod
	DeleteDiskMethod
	HasDiskMethod
	InfoMethod
}

var _ apiv1.CPIFactory = Factory{}
var _ apiv1.CPI = CPI{}

func NewFactory(
	driverClient driver.Client,
	stemcellClient stemcell.StemcellClient,
	stemcellStore stemcell.StemcellStore,
	agentSettings vm.AgentSettings,
	agentEnvFactory apiv1.AgentEnvFactory,
	config config.Config,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) Factory {
	return Factory{
		driverClient,
		stemcellClient,
		stemcellStore,
		agentSettings,
		agentEnvFactory,
		config,
		fs,
		uuidGen,
		logger,
	}
}

func (f Factory) New(_ apiv1.CallContext) (apiv1.CPI, error) {
	return CPI{
		NewCreateStemcellMethod(f.driverClient, f.stemcellClient, f.stemcellStore, f.uuidGen, f.fs, f.logger),
		NewDeleteStemcellMethod(f.driverClient, f.logger),
		NewCreateVMMethod(f.driverClient, f.agentSettings, f.config.GetAgentOptions(), f.agentEnvFactory, f.uuidGen, f.logger),
		NewDeleteVMMethod(f.driverClient, f.logger),
		NewHasVMMethod(f.driverClient),
		NewSetVMMetadataMethod(f.driverClient, f.logger),
		NewAttachDiskMethod(f.driverClient, f.agentSettings),
		NewCreateDiskMethod(f.driverClient, f.uuidGen),
		NewDetachDiskMethod(f.driverClient, f.agentSettings),
		NewDeleteDiskMethod(f.driverClient, f.logger),
		NewHasDiskMethod(f.driverClient),
		NewInfoMethod(),
	}, nil
}

func (c CPI) CalculateVMCloudProperties(res apiv1.VMResources) (apiv1.VMCloudProps, error) {
	return apiv1.NewVMCloudPropsFromMap(map[string]interface{}{}), nil
}

func (c CPI) SetDiskMetadata(cid apiv1.DiskCID, metadata apiv1.DiskMeta) error {
	return nil
}

func (c CPI) RebootVM(cid apiv1.VMCID) error {
	return nil
}

func (c CPI) GetDisks(cid apiv1.VMCID) ([]apiv1.DiskCID, error) {
	return []apiv1.DiskCID{}, nil
}

func (c CPI) ResizeDisk(cid apiv1.DiskCID, i int) error {
	return nil
}

func (c CPI) SnapshotDisk(cid apiv1.DiskCID, meta apiv1.DiskMeta) (apiv1.SnapshotCID, error) {
	return apiv1.SnapshotCID{}, nil
}

func (c CPI) DeleteSnapshot(cid apiv1.SnapshotCID) error {
	return nil
}
