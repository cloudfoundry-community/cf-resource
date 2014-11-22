package cmd_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Check", func() {
	It("outputs an empty JSON array so that it satisfies the resource interface", func() {
		bin, err := Build("github.com/concourse/cf-resource/check/cmd/check")
		Ω(err).ShouldNot(HaveOccurred())

		cmd := exec.Command(bin)
		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(Exit(0))
		Ω(session.Out).Should(Say(`\[\]`))
	})
})
