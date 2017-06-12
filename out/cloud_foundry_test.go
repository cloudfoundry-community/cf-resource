package out_test

import (
	"github.com/concourse/cf-resource/out"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("CloudFoundry", func() {
	Context("happy path", func() {
		var cf *out.CloudFoundry
		env := os.Environ()
		baseExpectedEnvVariableCount := len(env) + 1

		BeforeEach(func() {
			cf = out.NewCloudFoundry()
		})

		It("default command environment should contain CF_COLOR=true", func() {
			cfEnv := cf.CommandEnvironment().Environment()
			Ω(cfEnv).Should(HaveLen(baseExpectedEnvVariableCount))
			Ω(cfEnv).Should(ContainElement("CF_COLOR=true"))
		})
	})
})
