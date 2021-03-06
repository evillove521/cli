package plugin

import (
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("running plugins", func() {
	Describe("panic handling", func() {
		BeforeEach(func() {
			Eventually(helpers.CF("install-plugin", "-f", panicTestPluginPath)).Should(Exit(0))
		})

		It("will exit 1 if the plugin panics", func() {
			Eventually(helpers.CF("freak-out")).Should(Exit(1))
		})
	})

	Describe("CF_TRACE", func() {
		BeforeEach(func() {
			installTestPlugin()
		})

		AfterEach(func() {
			uninstallTestPlugin()
		})

		When("the plugin is run with CF_TRACE=true", func() {
			It("does not error", func() {
				Eventually(helpers.CustomCF(helpers.CFEnv{
					EnvVars: map[string]string{"CF_TRACE": "true"},
				}, "CoolTest")).Should(Exit(0))
			})
		})
	})

	Describe("when running plugin commands while CF_HOME is set", func() {
		When("CF_PLUGIN_HOME is unset", func() {
			BeforeEach(func() {
				Expect(os.Setenv("CF_PLUGIN_HOME", "")).NotTo(HaveOccurred())
			})

			When("a plugin is installed", func() {
				BeforeEach(func() {
					installTestPlugin()
				})

				AfterEach(func() {
					uninstallTestPlugin()
				})

				It("lists the installed plugins", func() {
					session := helpers.CF("plugins")
					Eventually(session).Should(Say("CoolTest"))
					Eventually(session).Should(Exit(0))
				})

				It("is able to run the CoolTest plugin command", func() { // v7 only
					confirmTestPluginOutput("CoolTest", "I am a test plugin")
				})

				It("is able to run an installed plugin command", func() {
					confirmTestPluginOutput("ApiEndpoint", helpers.GetAPI())
				})
			})
		})

		When("CF_PLUGIN_HOME is set", func() {
			When("a plugin is installed", func() {
				BeforeEach(func() {
					installTestPlugin()
				})

				AfterEach(func() {
					uninstallTestPlugin()
				})

				It("lists the installed plugins", func() {
					session := helpers.CF("plugins")
					Eventually(session).Should(Say("TestPluginCommandWithAlias"))
					Eventually(session).Should(Exit(0))
				})

				It("can call a command by its alias", func() {
					confirmTestPluginOutput("Cool-V7", "You called Test Plugin Command V7 With Alias!")
				})

				When("API is unset", func() {
					var (
						apiURL string
					)

					BeforeEach(func() {
						apiURL = helpers.GetAPI()
						Expect(apiURL).ToNot(BeEmpty())
						helpers.UnsetAPI()
					})

					It("doesn't panic when trying to call a plugin", func() {
						session := helpers.CF("CoolTest")
						Eventually(session).Should(Exit(1))

						Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
						Expect(session).To(Say("FAILED"))
					})
				})
			})
		})
	})
})
