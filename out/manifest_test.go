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
			var tempFileVar *os.File
			AfterEach(func() {
				os.Remove(tempFile.Name())
				os.Remove(tempFileVar.Name())
			})

			It("can write out a modified manifest", func() {
				tempFile, err = ioutil.TempFile("assets", "manifest_test.yml_")
				Ω(err).ShouldNot(HaveOccurred())

				tempFileVar, err = ioutil.TempFile("assets", "var_env_file")
				Ω(err).ShouldNot(HaveOccurred())
				err = ioutil.WriteFile(tempFileVar.Name(), []byte(string("VAR-VALUE-INSIDE-FILE")), os.FileMode(int(0755)))
				Ω(err).ShouldNot(HaveOccurred())
				manifest.AddEnvironmentVariableFromFile("ENV_FROM_FILE", tempFileVar.Name())

				manifest.AddEnvironmentVariable("MANIFEST_TEST_A", "manifest_test_a")
				err = manifest.Save(tempFile.Name())
				Ω(err).ShouldNot(HaveOccurred())

				updatedManifest, err := out.NewManifest(tempFile.Name())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(updatedManifest.EnvironmentVariables()["MANIFEST_A"]).Should(Equal("manifest_a"))
				Ω(updatedManifest.EnvironmentVariables()["MANIFEST_B"]).Should(Equal("manifest_b"))
				Ω(updatedManifest.EnvironmentVariables()["MANIFEST_TEST_A"]).Should(Equal("manifest_test_a"))
				Ω(updatedManifest.EnvironmentVariables()["ENV_FROM_FILE"]).Should(Equal("VAR-VALUE-INSIDE-FILE"))
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
