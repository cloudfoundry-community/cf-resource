package out_test

import (
	"github.com/concourse/cf-resource/out"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("Manifest", func() {
	Context("happy path", func() {
		var manifest out.Manifest
		var err error

		BeforeEach(func() {
			manifest, err = out.NewManifest("assets/manifest.yml")
		})

		It("can parse a manifest", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("can extract the environment variables", func() {
			envVars := manifest.EnvironmentVariables()
			Expect(envVars["MANIFEST_A"]).To(Equal("manifest_a"))
			Expect(envVars["MANIFEST_B"]).To(Equal("manifest_b"))
		})

		Context("when updated", func() {
			var tempFile *os.File

			AfterEach(func() {
				os.Remove(tempFile.Name())
			})

			It("can write out a modified manifest", func() {
				tempFile, err = ioutil.TempFile("assets", "manifest_test.yml_")
				Expect(err).NotTo(HaveOccurred())

				manifest.AddEnvironmentVariable("MANIFEST_TEST_A", "manifest_test_a")
				err = manifest.Save(tempFile.Name())
				Expect(err).NotTo(HaveOccurred())

				updatedManifest, err := out.NewManifest(tempFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedManifest.EnvironmentVariables()["MANIFEST_A"]).To(Equal("manifest_a"))
				Expect(updatedManifest.EnvironmentVariables()["MANIFEST_B"]).To(Equal("manifest_b"))
				Expect(updatedManifest.EnvironmentVariables()["MANIFEST_TEST_A"]).To(Equal("manifest_test_a"))
			})
		})
	})

	Context("invalid manifest path", func() {
		It("returns an error", func() {
			_, err := out.NewManifest("invalid path")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("invalid manifest YAML", func() {
		It("returns an error", func() {
			_, err := out.NewManifest("invalidManifest.yml")
			Expect(err).To(HaveOccurred())
		})
	})
})
