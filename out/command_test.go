
package out_test

import (
	"errors"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io"
	"io/ioutil"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/out"
	"github.com/concourse/cf-resource/out/outfakes"
)

var _ = Describe("Out Command", func() {
	var (
		cloudFoundry *outfakes.FakePAAS
		request      out.Request
		command      *out.Command
	)

	BeforeEach(func() {
		cloudFoundry = &outfakes.FakePAAS{}
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
				Vars:         map[string]interface{}{"foo": "bar"},
				VarsFiles:    []string{"vars.yml"},
			},
		}
	})

	Describe("running the command", func() {
		Context("when requesting rolling deployments", func() {
			BeforeEach(func() {
				request.Params.UseRollingAppDeployment = true
			})
			It("pushes an application using cf v3-zdt-push", func() {
				response, err := command.Run(request)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Version.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
				Expect(response.Metadata[0]).To(Equal(
					resource.MetadataPair{
						Name:  "organization",
						Value: "secret",
					},
				))
				Expect(response.Metadata[1]).To(Equal(
					resource.MetadataPair{
						Name:  "space",
						Value: "volcano-base",
					},
				))

				By("logging in")
				Expect(cloudFoundry.LoginCallCount()).To(Equal(1))

				api, username, password, clientID, clientSecret, insecure := cloudFoundry.LoginArgsForCall(0)
				Expect(api).To(Equal("https://api.run.pivotal.io"))
				Expect(username).To(Equal("awesome@example.com"))
				Expect(password).To(Equal("hunter2"))
				Expect(clientID).To(Equal(""))
				Expect(clientSecret).To(Equal(""))
				Expect(insecure).To(Equal(false))

				By("targeting the organization and space")
				Expect(cloudFoundry.TargetCallCount()).To(Equal(1))

				org, space := cloudFoundry.TargetArgsForCall(0)
				Expect(org).To(Equal("secret"))
				Expect(space).To(Equal("volcano-base"))

				By("pushing the app")
				Expect(cloudFoundry.PushAppCallCount()).To(Equal(0))
				Expect(cloudFoundry.PushAppWithRollingDeploymentCallCount()).To(Equal(1))

				path, currentAppName, dockerUser, showAppLog, noStart, manifest := cloudFoundry.PushAppWithRollingDeploymentArgsForCall(0)
				Expect(path).To(Equal(""))
				Expect(currentAppName).To(Equal(""))
				Expect(dockerUser).To(Equal(""))
				Expect(showAppLog).To(Equal(false))
				Expect(noStart).To(Equal(false))
				Expect(manifest).To(Equal("assets/manifest.yml"))
			})
		})
		It("pushes an application into cloud foundry", func() {
			response, err := command.Run(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Version.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
			Expect(response.Metadata[0]).To(Equal(
				resource.MetadataPair{
					Name:  "organization",
					Value: "secret",
				},
			))
			Expect(response.Metadata[1]).To(Equal(
				resource.MetadataPair{
					Name:  "space",
					Value: "volcano-base",
				},
			))

			By("logging in")
			Expect(cloudFoundry.LoginCallCount()).To(Equal(1))

			api, username, password, clientID, clientSecret, insecure := cloudFoundry.LoginArgsForCall(0)
			Expect(api).To(Equal("https://api.run.pivotal.io"))
			Expect(username).To(Equal("awesome@example.com"))
			Expect(password).To(Equal("hunter2"))
			Expect(clientID).To(Equal(""))
			Expect(clientSecret).To(Equal(""))
			Expect(insecure).To(Equal(false))

			By("targetting the organization and space")
			Expect(cloudFoundry.TargetCallCount()).To(Equal(1))

			org, space := cloudFoundry.TargetArgsForCall(0)
			Expect(org).To(Equal("secret"))
			Expect(space).To(Equal("volcano-base"))

			By("pushing the app")
			Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))
			Expect(cloudFoundry.PushAppWithRollingDeploymentCallCount()).To(Equal(0))

			manifest, path, currentAppName, vars, varsFiles, dockerUser, showAppLog, noStart := cloudFoundry.PushAppArgsForCall(0)
			Expect(manifest).To(Equal(request.Params.ManifestPath))
			Expect(path).To(Equal(""))
			Expect(currentAppName).To(Equal(""))
			Expect(vars).To(Equal(map[string]interface{}{"foo": "bar"}))
			Expect(varsFiles).To(Equal([]string{"vars.yml"}))
			Expect(dockerUser).To(Equal(""))
			Expect(showAppLog).To(Equal(false))
			Expect(noStart).To(Equal(false))
		})

		Describe("handling any errors", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("it all went wrong")
			})

			It("from logging in", func() {
				cloudFoundry.LoginReturns(expectedError)

				_, err := command.Run(request)
				Expect(err).To(MatchError(expectedError))
			})

			It("from targetting an org and space", func() {
				cloudFoundry.TargetReturns(expectedError)

				_, err := command.Run(request)
				Expect(err).To(MatchError(expectedError))
			})

			It("from pushing the application", func() {
				cloudFoundry.PushAppReturns(expectedError)

				_, err := command.Run(request)
				Expect(err).To(MatchError(expectedError))
			})
		})

		Describe("no_start handling", func() {
			Context("when no_start is specified", func() {
				BeforeEach(func() {
					request.Params.NoStart = true
				})

				It("sets noStart to true", func() {
					_, err := command.Run(request)
					Expect(err).NotTo(HaveOccurred())

					_, _, _, _, _, _, _, noStart := cloudFoundry.PushAppArgsForCall(0)
					Expect(noStart).To(Equal(true))
				})
			})
		})

		Context("setting environment variables provided as params", func() {
			var err error
			var tempFile *os.File

			BeforeEach(func() {
				sourceFile, err := os.Open("assets/manifest.yml")
				Expect(err).NotTo(HaveOccurred())
				defer sourceFile.Close()

				tempFile, err = ioutil.TempFile("assets", "command_test.yml_")
				Expect(err).NotTo(HaveOccurred())
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
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes the variables into the manifest", func() {
				manifest, _ := out.NewManifest(request.Params.ManifestPath)

				Expect(manifest.EnvironmentVariables()[0]["COMMAND_TEST_A"]).To(Equal("command_test_a"))
				Expect(manifest.EnvironmentVariables()[0]["COMMAND_TEST_B"]).To(Equal("command_test_b"))
				Expect(manifest.EnvironmentVariables()[1]["COMMAND_TEST_A"]).To(Equal("command_test_a"))
				Expect(manifest.EnvironmentVariables()[1]["COMMAND_TEST_B"]).To(Equal("command_test_b"))
			})
		})

		Context("no environment variables provided", func() {
			It("doesn't set the environment variables", func() {
				manifest, err := out.NewManifest(request.Params.ManifestPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(manifest.EnvironmentVariables()).To(HaveLen(2))
				Expect(manifest.EnvironmentVariables()[0]).To(HaveLen(2))
				Expect(manifest.EnvironmentVariables()[0]).To(HaveKeyWithValue("MANIFEST_A", "manifest_a"))
				Expect(manifest.EnvironmentVariables()[0]).To(HaveKeyWithValue("MANIFEST_B", "manifest_b"))
				Expect(manifest.EnvironmentVariables()[1]).To(HaveLen(0))
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
			Expect(err).NotTo(HaveOccurred())

			By("logging in")
			Expect(cloudFoundry.LoginCallCount()).To(Equal(1))

			_, _, _, _, _, insecure := cloudFoundry.LoginArgsForCall(0)
			Expect(insecure).To(Equal(true))
		})

		It("lets users authenticate with client credentials", func() {
			request = out.Request{
				Source: resource.Source{
					API:          "https://api.run.pivotal.io",
					ClientID:     "awesome",
					ClientSecret: "hunter2",
					Organization: "secret",
					Space:        "volcano-base",
				},
				Params: out.Params{
					ManifestPath: "a/path/to/a/manifest.yml",
				},
			}

			_, err := command.Run(request)
			Expect(err).NotTo(HaveOccurred())

			By("logging in")
			Expect(cloudFoundry.LoginCallCount()).To(Equal(1))

			_, username, password, clientID, clientSecret, _ := cloudFoundry.LoginArgsForCall(0)
			Expect(username).To(Equal(""))
			Expect(password).To(Equal(""))
			Expect(clientID).To(Equal("awesome"))
			Expect(clientSecret).To(Equal("hunter2"))
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
			Expect(err).NotTo(HaveOccurred())

			By("pushing the app")
			Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))

			_, _, currentAppName, _, _, _, _, _ := cloudFoundry.PushAppArgsForCall(0)
			Expect(currentAppName).To(Equal("cool-app-name"))
		})

		It("lets people define a user for connecting to a docker registry", func() {
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
					DockerUsername: "DOCKER_USER",
				},
			}

			_, err := command.Run(request)
			Expect(err).NotTo(HaveOccurred())

			By("pushing the app")
			Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))

			_, _, _, _, _, dockerUser, _, _ := cloudFoundry.PushAppArgsForCall(0)
			Expect(dockerUser).To(Equal("DOCKER_USER"))
		})

		Context("using a docker registry which requires authentication", func() {

			var savedVar string

			BeforeEach(func() {
				savedVar = saveCurrentVariable()
			})

			AfterEach(func() {
				restoreOldVariable(savedVar)
			})

			Context("docker password provided", func() {
				It("sets the system environment variable", func() {
					request = out.Request{
						Source: resource.Source{
							API:          "https://api.run.pivotal.io",
							Username:     "awesome@example.com",
							Password:     "hunter2",
							Organization: "secret",
							Space:        "volcano-base",
						},
						Params: out.Params{
							ManifestPath:   "a/path/to/a/manifest/using/docker.yml",
							DockerPassword: "mySuperSecretPassword",
						},
					}
					_, err := command.Run(request)
					Expect(err).NotTo(HaveOccurred())

					By("pushing the app")
					Expect(os.Getenv(out.CfDockerPassword)).To(Equal("mySuperSecretPassword"))
					Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))
				})
			})

			Context("no docker password provided", func() {
				It("doesn't set the system environment variable", func() {
					request = out.Request{
						Source: resource.Source{
							API:          "https://api.run.pivotal.io",
							Username:     "awesome@example.com",
							Password:     "hunter2",
							Organization: "secret",
							Space:        "volcano-base",
						},
						Params: out.Params{
							ManifestPath: "a/path/to/a/manifest/using/docker.yml",
						},
					}
					os.Setenv(out.CfDockerPassword, "MyOwnUntouchedVariable")

					_, err := command.Run(request)
					Expect(err).NotTo(HaveOccurred())

					By("pushing the app")
					Expect(os.Getenv(out.CfDockerPassword)).To(Equal("MyOwnUntouchedVariable"))
					Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))
				})
			})
		})
	})
})

func restoreOldVariable(currentEnvironmentVariable string) error {
	return os.Setenv(out.CfDockerPassword, currentEnvironmentVariable)
}

func saveCurrentVariable() string {
	return os.Getenv(out.CfDockerPassword)
}
