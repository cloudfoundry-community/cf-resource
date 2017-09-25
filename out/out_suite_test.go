package out_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestOut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Suite")
}

var binPath string

var _ = BeforeSuite(func() {
	var err error

	if _, err = os.Stat("/opt/resource/out"); err == nil {
		binPath = "/opt/resource/out"
	} else {
		binPath, err = gexec.Build("github.com/concourse/cf-resource/out/cmd/out")
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
