package out_test

import (
	"github.com/concourse/cf-resource/out"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var oneEnvironmentPair = map[string]interface{}{"ENV_ONE": "env_one"}

// json from config is unmarshalled in to map[string]interface{}
// keys are always strings, but values can be anything
var multipleEnvironmentPairs = map[string]interface{}{
	"ENV_ONE":   "env_one",
	"ENV_TWO":   2,
	"ENV_THREE": true,
}
var fiveEnvironmentPair = map[string]interface{}{"ENV_FIVE": "env_five"}

var _ = Describe("utility functions", func() {
	Context("happy path", func() {

		osLikeKVArray := []string{
			"OS_ONE=one",
			"OS_TWO=two",
			"OS_THREE=three",
		}

		simpleKVString := "SIMPLE=pair"
		keyWithValueContainingEqualsString := "EQUAL_VALUE=val_key=val_val"

		It("simple k=v string splits correctly", func() {
			key, value := out.SplitKeyValueString(simpleKVString)

			Ω(key).Should(Equal("SIMPLE"))
			Ω(value).Should(Equal("pair"))
		})

		It("value containing equal is parsed correctly (needs `SplitN()`)", func() {
			key, value := out.SplitKeyValueString(keyWithValueContainingEqualsString)

			Ω(key).Should(Equal("EQUAL_VALUE"))
			Ω(value).Should(Equal("val_key=val_val"))
		})

		It("array of k=v strings split correctly in to map", func() {
			kvMap := out.SplitKeyValueStringArrayInToMap(osLikeKVArray)

			Ω(kvMap).Should(HaveLen(3))
			Ω(kvMap).Should(HaveKeyWithValue("OS_ONE", "one"))
			Ω(kvMap).Should(HaveKeyWithValue("OS_TWO", "two"))
			Ω(kvMap).Should(HaveKeyWithValue("OS_THREE", "three"))
		})
	})
})

var _ = Describe("CfEnvironment from Empty", func() {
	Context("happy path", func() {
		var cfEnvironment *out.CfEnvironment

		BeforeEach(func() {
			cfEnvironment = out.NewCfEnvironment()
		})

		It("default command environment should ONLY contain CF_COLOR=true", func() {
			cfEnv := cfEnvironment.Environment()
			Ω(cfEnv).Should(HaveLen(1))
			Ω(cfEnv).Should(ContainElement("CF_COLOR=true"))
		})

		It("added environment switch ends up in environment", func() {

			cfEnvironment.AddEnvironmentVariable(oneEnvironmentPair)
			cfEnv := cfEnvironment.Environment()

			Ω(cfEnv).Should(HaveLen(2))
			Ω(cfEnv).Should(ContainElement("ENV_ONE=env_one"))
		})

		It("multiple environment switches all end up in environment", func() {

			cfEnvironment.AddEnvironmentVariable(multipleEnvironmentPairs)
			cfEnv := cfEnvironment.Environment()

			Ω(cfEnv).Should(HaveLen(4))
			Ω(cfEnv).Should(ContainElement("ENV_ONE=env_one"))
			Ω(cfEnv).Should(ContainElement("ENV_TWO=2"))
			Ω(cfEnv).Should(ContainElement("ENV_THREE=true"))
		})

		It("multiple adds to environment retains all additions", func() {
			cfEnvironment.AddEnvironmentVariable(multipleEnvironmentPairs)
			cfEnvironment.AddEnvironmentVariable(fiveEnvironmentPair)
			cfEnv := cfEnvironment.Environment()

			Ω(cfEnv).Should(HaveLen(5))
			Ω(cfEnv).Should(ContainElement("ENV_ONE=env_one"))
			Ω(cfEnv).Should(ContainElement("ENV_TWO=2"))
			Ω(cfEnv).Should(ContainElement("ENV_THREE=true"))

			Ω(cfEnv).Should(ContainElement("ENV_FIVE=env_five"))
		})

	})
})

var _ = Describe("CfEnvironment from OS", func() {
	Context("happy path", func() {
		var cfEnvironment *out.CfEnvironment
		env := os.Environ()
		baseExpectedEnvVariables := len(env) + 1

		BeforeEach(func() {
			cfEnvironment = out.NewCfEnvironmentFromOS()
		})

		It("default command environment should contain CF_COLOR=true", func() {
			cfEnv := cfEnvironment.Environment()
			Ω(cfEnv).Should(HaveLen(baseExpectedEnvVariables))
			Ω(cfEnv).Should(ContainElement("CF_COLOR=true"))
		})
	})
})
