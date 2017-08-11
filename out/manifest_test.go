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
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("can extract the environment variables", func() {
			envVars := manifest.EnvironmentVariables()
			Ω(envVars["MANIFEST_A"]).Should(Equal("manifest_a"))
			Ω(envVars["MANIFEST_B"]).Should(Equal("manifest_b"))
		})

		Context("when updated", func() {
			var tempFile *os.File

			AfterEach(func() {
				os.Remove(tempFile.Name())
			})

			It("can write out a modified manifest", func() {
				tempFile, err = ioutil.TempFile("assets", "manifest_test.yml_")
				Ω(err).ShouldNot(HaveOccurred())

				manifest.AddEnvironmentVariable("MANIFEST_TEST_A", "manifest_test_a")
				err = manifest.Save(tempFile.Name())
				Ω(err).ShouldNot(HaveOccurred())

				updatedManifest, err := out.NewManifest(tempFile.Name())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(updatedManifest.EnvironmentVariables()["MANIFEST_A"]).Should(Equal("manifest_a"))
				Ω(updatedManifest.EnvironmentVariables()["MANIFEST_B"]).Should(Equal("manifest_b"))
				Ω(updatedManifest.EnvironmentVariables()["MANIFEST_TEST_A"]).Should(Equal("manifest_test_a"))
			})
		})
	})

	Context("invalid manifest path", func() {
		It("returns an error", func() {
			_, err := out.NewManifest("invalid path")
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("invalid manifest YAML", func() {
		It("returns an error", func() {
			_, err := out.NewManifest("invalidManifest.yml")
			Ω(err).Should(HaveOccurred())
		})
	})
})
