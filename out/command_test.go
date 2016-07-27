package out_test

import (
	"errors"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/out"
	"github.com/concourse/cf-resource/out/fakes"
	"io"
	"io/ioutil"
)

var _ = Describe("Out Command", func() {
	var (
		cloudFoundry *fakes.FakePAAS
		request      out.Request
		command      *out.Command
	)

	BeforeEach(func() {
		cloudFoundry = &fakes.FakePAAS{}
		command = out.NewCommand(cloudFoundry)

		request = out.Request{
			Source: resource.Source{
				API:          "https://api.run.pivotal.io",
				Username:     "awesome@example.com",
				Password:     "hunter2",
				Organization: "secret",
				Space:        "volcano-base",
			},
			Params: out.Params{
				ManifestPath: "assets/manifest.yml",
			},
		}
	})

	Describe("running the command", func() {
		Context("with organization and space only set via params", func() {
			var (
				requestOverwrite out.Request
			)

			BeforeEach(func() {
				requestOverwrite = out.Request{
					Source: resource.Source{
						API:      "https://api.run.pivotal.io",
						Username: "awesome@example.com",
						Password: "hunter2",
					},
					Params: out.Params{
						ManifestPath: "assets/manifest.yml",
						Organization: "new-secret",
						Space:        "sky-base",
					},
				}
			})

			It("pushes an application into cloud foundry", func() {
				response, err := command.Run(requestOverwrite)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response.Version.Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
				Ω(response.Metadata[0]).Should(Equal(
					resource.MetadataPair{
						Name:  "organization",
						Value: "new-secret",
					},
				))
				Ω(response.Metadata[1]).Should(Equal(
					resource.MetadataPair{
						Name:  "space",
						Value: "sky-base",
					},
				))

				By("logging in")
				Ω(cloudFoundry.LoginCallCount()).Should(Equal(1))

				api, username, password, insecure := cloudFoundry.LoginArgsForCall(0)
				Ω(api).Should(Equal("https://api.run.pivotal.io"))
				Ω(username).Should(Equal("awesome@example.com"))
				Ω(password).Should(Equal("hunter2"))
				Ω(insecure).Should(Equal(false))

				By("targetting the organization and space")
				Ω(cloudFoundry.TargetCallCount()).Should(Equal(1))

				org, space := cloudFoundry.TargetArgsForCall(0)
				Ω(org).Should(Equal("new-secret"))
				Ω(space).Should(Equal("sky-base"))

				By("pushing the app")
				Ω(cloudFoundry.PushAppCallCount()).Should(Equal(1))

				manifest, path, currentAppName := cloudFoundry.PushAppArgsForCall(0)
				Ω(manifest).Should(Equal("assets/manifest.yml"))
				Ω(path).Should(Equal(""))
				Ω(currentAppName).Should(Equal(""))
			})
		})

		Context("with overwriting organization and space with params", func() {
			var (
				requestOverwrite out.Request
			)

			BeforeEach(func() {
				requestOverwrite = out.Request{
					Source: resource.Source{
						API:          "https://api.run.pivotal.io",
						Username:     "awesome@example.com",
						Password:     "hunter2",
						Organization: "secret",
						Space:        "volcano-base",
					},
					Params: out.Params{
						ManifestPath: "assets/manifest.yml",
						Organization: "new-secret",
						Space:        "sky-base",
					},
				}
			})

			It("pushes an application into cloud foundry", func() {
				response, err := command.Run(requestOverwrite)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response.Version.Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
				Ω(response.Metadata[0]).Should(Equal(
					resource.MetadataPair{
						Name:  "organization",
						Value: "new-secret",
					},
				))
				Ω(response.Metadata[1]).Should(Equal(
					resource.MetadataPair{
						Name:  "space",
						Value: "sky-base",
					},
				))

				By("logging in")
				Ω(cloudFoundry.LoginCallCount()).Should(Equal(1))

				api, username, password, insecure := cloudFoundry.LoginArgsForCall(0)
				Ω(api).Should(Equal("https://api.run.pivotal.io"))
				Ω(username).Should(Equal("awesome@example.com"))
				Ω(password).Should(Equal("hunter2"))
				Ω(insecure).Should(Equal(false))

				By("targetting the organization and space")
				Ω(cloudFoundry.TargetCallCount()).Should(Equal(1))

				org, space := cloudFoundry.TargetArgsForCall(0)
				Ω(org).Should(Equal("new-secret"))
				Ω(space).Should(Equal("sky-base"))

				By("pushing the app")
				Ω(cloudFoundry.PushAppCallCount()).Should(Equal(1))

				manifest, path, currentAppName := cloudFoundry.PushAppArgsForCall(0)
				Ω(manifest).Should(Equal("assets/manifest.yml"))
				Ω(path).Should(Equal(""))
				Ω(currentAppName).Should(Equal(""))
			})
		})

		It("pushes an application into cloud foundry", func() {
			response, err := command.Run(request)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(response.Version.Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
			Ω(response.Metadata[0]).Should(Equal(
				resource.MetadataPair{
					Name:  "organization",
					Value: "secret",
				},
			))
			Ω(response.Metadata[1]).Should(Equal(
				resource.MetadataPair{
					Name:  "space",
					Value: "volcano-base",
				},
			))

			By("logging in")
			Ω(cloudFoundry.LoginCallCount()).Should(Equal(1))

			api, username, password, insecure := cloudFoundry.LoginArgsForCall(0)
			Ω(api).Should(Equal("https://api.run.pivotal.io"))
			Ω(username).Should(Equal("awesome@example.com"))
			Ω(password).Should(Equal("hunter2"))
			Ω(insecure).Should(Equal(false))

			By("targetting the organization and space")
			Ω(cloudFoundry.TargetCallCount()).Should(Equal(1))

			org, space := cloudFoundry.TargetArgsForCall(0)
			Ω(org).Should(Equal("secret"))
			Ω(space).Should(Equal("volcano-base"))

			By("pushing the app")
			Ω(cloudFoundry.PushAppCallCount()).Should(Equal(1))

			manifest, path, currentAppName := cloudFoundry.PushAppArgsForCall(0)
			Ω(manifest).Should(Equal("assets/manifest.yml"))
			Ω(path).Should(Equal(""))
			Ω(currentAppName).Should(Equal(""))
		})

		Describe("handling any errors", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("it all went wrong")
			})

			It("from logging in", func() {
				cloudFoundry.LoginReturns(expectedError)

				_, err := command.Run(request)
				Ω(err).Should(MatchError(expectedError))
			})

			It("from targetting an org and space", func() {
				cloudFoundry.TargetReturns(expectedError)

				_, err := command.Run(request)
				Ω(err).Should(MatchError(expectedError))
			})

			It("from pushing the application", func() {
				cloudFoundry.PushAppReturns(expectedError)

				_, err := command.Run(request)
				Ω(err).Should(MatchError(expectedError))
			})
		})

		Context("setting environment variables provided as params", func() {
			var err error
			var tempFile *os.File

			BeforeEach(func() {
				sourceFile, err := os.Open("assets/manifest.yml")
				Ω(err).ShouldNot(HaveOccurred())
				defer sourceFile.Close()

				tempFile, err = ioutil.TempFile("assets", "command_test.yml_")
				Ω(err).ShouldNot(HaveOccurred())
				defer tempFile.Close()

				_, err = io.Copy(tempFile, sourceFile)

				request = out.Request{
					Source: resource.Source{
						API:          "https://api.run.pivotal.io",
						Username:     "awesome@example.com",
						Password:     "hunter2",
						Organization: "secret",
						Space:        "volcano-base",
					},
					Params: out.Params{
						ManifestPath: tempFile.Name(),
						EnvironmentVariables: map[string]string{
							"COMMAND_TEST_A": "command_test_a",
							"COMMAND_TEST_B": "command_test_b",
						},
					},
				}
				_, err = command.Run(request)
			})

			AfterEach(func() {
				os.Remove(tempFile.Name())
			})

			It("does not raise an error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("writes the variables into the manifest", func() {
				manifest, _ := out.NewManifest(request.Params.ManifestPath)

				Ω(manifest.EnvironmentVariables()["COMMAND_TEST_A"]).Should(Equal("command_test_a"))
				Ω(manifest.EnvironmentVariables()["COMMAND_TEST_B"]).Should(Equal("command_test_b"))
			})
		})

		Context("no environment variables provided", func() {
			It("doesn't set the environment variables", func() {
				manifest, err := out.NewManifest(request.Params.ManifestPath)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(manifest.EnvironmentVariables()).Should(HaveLen(2))
				Ω(manifest.EnvironmentVariables()).Should(HaveKeyWithValue("MANIFEST_A", "manifest_a"))
				Ω(manifest.EnvironmentVariables()).Should(HaveKeyWithValue("MANIFEST_B", "manifest_b"))
			})
		})

		It("lets people skip the certificate check", func() {
			request = out.Request{
				Source: resource.Source{
					API:           "https://api.run.pivotal.io",
					Username:      "awesome@example.com",
					Password:      "hunter2",
					Organization:  "secret",
					Space:         "volcano-base",
					SkipCertCheck: true,
				},
				Params: out.Params{
					ManifestPath: "a/path/to/a/manifest.yml",
				},
			}

			_, err := command.Run(request)
			Ω(err).ShouldNot(HaveOccurred())

			By("logging in")
			Ω(cloudFoundry.LoginCallCount()).Should(Equal(1))

			_, _, _, insecure := cloudFoundry.LoginArgsForCall(0)
			Ω(insecure).Should(Equal(true))
		})

		It("lets people do a zero downtime deploy", func() {
			request = out.Request{
				Source: resource.Source{
					API:          "https://api.run.pivotal.io",
					Username:     "awesome@example.com",
					Password:     "hunter2",
					Organization: "secret",
					Space:        "volcano-base",
				},
				Params: out.Params{
					ManifestPath:   "a/path/to/a/manifest.yml",
					CurrentAppName: "cool-app-name",
				},
			}

			_, err := command.Run(request)
			Ω(err).ShouldNot(HaveOccurred())

			By("pushing the app")
			Ω(cloudFoundry.PushAppCallCount()).Should(Equal(1))

			_, _, currentAppName := cloudFoundry.PushAppArgsForCall(0)
			Ω(currentAppName).Should(Equal("cool-app-name"))
		})
	})
})
