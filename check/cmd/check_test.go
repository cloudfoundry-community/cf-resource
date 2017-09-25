package cmd_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Check", func() {
	It("outputs an empty JSON array so that it satisfies the resource interface", func() {
		var (
			err error
			bin string
		)
		if _, err = os.Stat("/opt/resource/check"); err == nil {
			bin = "/opt/resource/check"
		} else {
			bin, err = Build("github.com/concourse/cf-resource/check/cmd/check")
			Expect(err).NotTo(HaveOccurred())
		}
		cmd := exec.Command(bin)
		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(Exit(0))
		Expect(session.Out).To(Say(`\[\]`))
	})
})
