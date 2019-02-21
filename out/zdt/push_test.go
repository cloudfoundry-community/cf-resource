package zdt

import (
	"errors"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("BuildActions", func() {
	stdout := gbytes.NewBuffer()

	cf := func(args ...string) *exec.Cmd {
		cmd := exec.Command("assets/cf", args...)
		cmd.Stdout = stdout
		return cmd
	}

	It("pushes an app with zero downtime", func() {
		pushFunction := func() error { return cf("push", "my-app").Run() }
		err := Push(cf, "my-app", pushFunction, false)

		Expect(err).NotTo(HaveOccurred())
		Expect(stdout).To(gbytes.Say("cf rename my-app my-app-venerable"))
		Expect(stdout).To(gbytes.Say("cf push my-app"))
		Expect(stdout).To(gbytes.Say("cf delete -f my-app-venerable"))
	})

	It("rolls back on failed push", func() {
		pushErr := errors.New("push failed")
		pushFunction := func() error {
			_ = cf("push", "my-app").Run()
			return pushErr
		}
		err := Push(cf, "my-app", pushFunction, false)

		Expect(err).To(Equal(pushErr))
		Expect(stdout).To(gbytes.Say("cf rename my-app my-app-venerable"))
		Expect(stdout).To(gbytes.Say("cf push my-app"))
		Expect(stdout).ToNot(gbytes.Say("cf logs"))
		Expect(stdout).To(gbytes.Say("cf delete -f my-app"))
		Expect(stdout).To(gbytes.Say("cf rename my-app-venerable my-app"))
	})

	It("shows logs on failure when flag is set", func() {
		pushFunction := func() error {
			_ = cf("push", "my-app").Run()
			return errors.New("push failed")
		}
		err := Push(cf, "my-app", pushFunction, true)

		Expect(err).To(HaveOccurred())
		Expect(stdout).To(gbytes.Say("cf rename my-app my-app-venerable"))
		Expect(stdout).To(gbytes.Say("cf push my-app"))
		Expect(stdout).To(gbytes.Say("cf logs my-app --recent"))
		Expect(stdout).To(gbytes.Say("cf delete -f my-app"))
		Expect(stdout).To(gbytes.Say("cf rename my-app-venerable my-app"))
	})
})
