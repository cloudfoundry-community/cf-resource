package zdt_test

import (
	"errors"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/concourse/cf-resource/out/zdt"
)

var stdout *gbytes.Buffer

func cli(path string) func(args ...string) *exec.Cmd {
	return func(args ...string) *exec.Cmd {
		cmd := exec.Command(path, args...)
		cmd.Stdout = stdout
		return cmd
	}
}

var _ = Describe("CanPush", func() {
	cf := cli("assets/cf")
	errCf := cli("assets/erroringCf")

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
	})

	It("needs a currentAppName", func() {
		Expect(zdt.CanPush(cf, "")).To(BeFalse())
		Expect(stdout.Contents()).To(BeEmpty())
	})

	It("needs the app to exist", func() {
		Expect(zdt.CanPush(errCf, "my-app")).To(BeFalse())
		Expect(stdout).To(gbytes.Say("cf app my-app"))
	})

	It("is ok when app exists", func() {
		Expect(zdt.CanPush(cf, "my-app")).To(BeTrue())
		Expect(stdout).To(gbytes.Say("cf app my-app"))
	})
})

var _ = Describe("Push", func() {
	cf := cli("assets/cf")

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
	})

	It("pushes an app with zero downtime", func() {
		pushFunction := func() error { return cf("push", "my-app").Run() }
		err := zdt.Push(cf, "my-app", pushFunction, false)

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
		err := zdt.Push(cf, "my-app", pushFunction, false)

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
		err := zdt.Push(cf, "my-app", pushFunction, true)

		Expect(err).To(HaveOccurred())
		Expect(stdout).To(gbytes.Say("cf rename my-app my-app-venerable"))
		Expect(stdout).To(gbytes.Say("cf push my-app"))
		Expect(stdout).To(gbytes.Say("cf logs my-app --recent"))
		Expect(stdout).To(gbytes.Say("cf delete -f my-app"))
		Expect(stdout).To(gbytes.Say("cf rename my-app-venerable my-app"))
	})
})
