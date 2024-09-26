package action_test

import (
	"encoding/json"

	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedriver "bosh-vmrun-cpi/driver/fakes"
	fakestemcell "bosh-vmrun-cpi/stemcell/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-vmrun-cpi/action"
)

var _ = Describe("CreateStemcell", func() {
	var (
		err error

		fs             *fakesys.FakeFileSystem
		stemcellStore  *fakestemcell.FakeStemcellStore
		stemcellClient *fakestemcell.FakeStemcellClient
		driverClient   *fakedriver.FakeClient
		m              action.CreateStemcellMethod
	)

	BeforeEach(func() {
		driverClient = &fakedriver.FakeClient{}
		stemcellClient = &fakestemcell.FakeStemcellClient{}
		stemcellStore = &fakestemcell.FakeStemcellStore{}
		logger := &fakelogger.FakeLogger{}
		fs = fakesys.NewFakeFileSystem()
		uuidGen := &fakeuuid.FakeGenerator{}
		m = action.NewCreateStemcellMethod(driverClient, stemcellClient, stemcellStore, uuidGen, fs, logger)
	})

	It("uses the supplied image", func() {
		err = fs.MkdirAll("image-path", 0777)
		Expect(err).ToNot(HaveOccurred())

		localImagePath := "/path/to/image"

		stemcellStore.GetByImagePathMappingReturns(localImagePath, nil)
		stemcellClient.ExtractOvfReturns("extracted-path", nil)

		var resourceCloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{}`), &resourceCloudProps)

		cid, err := m.CreateStemcell(localImagePath, resourceCloudProps)
		Expect(err).ToNot(HaveOccurred())

		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		clientImportOvfPath, clientImportOvfVmId := driverClient.ImportOvfArgsForCall(0)
		Expect(clientImportOvfPath).To(Equal("extracted-path"))
		Expect(clientImportOvfVmId).To(Equal("cs-fake-uuid-0"))

		Expect(stemcellClient.ExtractOvfArgsForCall(0)).To(Equal(localImagePath))
		Expect(stemcellClient.CleanupCallCount()).To(Equal(1))
	})

	It("uses the stemcell store if supplied image path does not exist", func() {
		var err error

		err = fs.MkdirAll("image-path", 0777)
		Expect(err).ToNot(HaveOccurred())

		storeImagePath := "/path/to/store/image"

		stemcellStore.GetByImagePathMappingReturns("", nil)
		stemcellStore.GetByMetadataReturns(storeImagePath, nil)
		stemcellClient.ExtractOvfReturns("extracted-path", nil)

		var resourceCloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{}`), &resourceCloudProps)

		cid, err := m.CreateStemcell("local-image-does-not-exist", resourceCloudProps)
		Expect(err).ToNot(HaveOccurred())

		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		clientImportOvfPath, clientImportOvfVmId := driverClient.ImportOvfArgsForCall(0)
		Expect(clientImportOvfPath).To(Equal("extracted-path"))
		Expect(clientImportOvfVmId).To(Equal("cs-fake-uuid-0"))

		Expect(stemcellClient.ExtractOvfArgsForCall(0)).To(Equal(storeImagePath))
		Expect(stemcellClient.CleanupCallCount()).To(Equal(1))
	})
})
