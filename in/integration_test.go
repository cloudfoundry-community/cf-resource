package in_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/in"
)

var _ = Describe("In", func() {
	var (
		tmpDir   string
		request  in.Request
		response in.Response
	)

	JustBeforeEach(func() {
		var (
			binPath string
			err     error
		)

		if _, err = os.Stat("/opt/resource/in"); err == nil {
			binPath = "/opt/resource/in"
		} else {
			binPath, err = gexec.Build("github.com/concourse/cf-resource/in/cmd/in")
			Expect(err).NotTo(HaveOccurred())
		}

		tmpDir, err = ioutil.TempDir("", "cf_resource_in")

		stdin := &bytes.Buffer{}
		err = json.NewEncoder(stdin).Encode(request)
		Expect(err).NotTo(HaveOccurred())

		cmd := exec.Command(binPath, tmpDir)
		cmd.Stdin = stdin
		cmd.Dir = tmpDir

		session, err := gexec.Start(
			cmd,
			GinkgoWriter,
			GinkgoWriter,
		)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		err = json.Unmarshal(session.Out.Contents(), &response)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when a version is given to the executable", func() {
		BeforeEach(func() {
			request = in.Request{
				Source: resource.Source{
					API:           "https://api.run.pivotal.io",
					Username:      "awesome@example.com",
					Password:      "hunter2",
					Organization:  "org",
					Space:         "space",
					SkipCertCheck: true,
				},
				Version: resource.Version{
					Timestamp: time.Now().Add(332 * time.Hour),
				},
			}
		})

		It("outputs that version", func() {
			Expect(response.Version.Timestamp).To(BeTemporally("~", request.Version.Timestamp, time.Second))
		})
	})

	Context("when a version is not given to the executable", func() {
		BeforeEach(func() {
			request = in.Request{
				Source: resource.Source{
					API:           "https://api.run.pivotal.io",
					Username:      "awesome@example.com",
					Password:      "hunter2",
					Organization:  "org",
					Space:         "space",
					SkipCertCheck: true,
				},
			}
		})

		It("generates a 'fake' current version", func() {
			Expect(response.Version.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
		})
	})
})
