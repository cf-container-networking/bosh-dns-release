package gbosh

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"time"
)

type Deployment struct {
	director Director
	name     string
}

func (d Deployment) Director() Director {
	return d.director
}

func (d Deployment) Name() string {
	return d.name
}

func (d Deployment) ExecuteDelete() *gexec.Session {
	session := d.Start("-n", "delete-deployment")
	Eventually(session, 300*time.Second).Should(gexec.Exit(0))

	return session
}

func (d Deployment) ExecuteDeploy(manifestPath string, opsPaths []string, vars map[string]string) *gexec.Session {
	args := []string{"-n", "deploy"}

	args = append(args, manifestPath)

	for _, opsPath := range opsPaths {
		args = append(args, "-o", opsPath)
	}

	for k, v := range vars {
		args = append(args, "-v", fmt.Sprintf("%s=%s", k, v))
	}

	// give the deployment a random name
	overwritenameops, err := ioutil.TempFile("", "gbosh-overwrite-name-ops-")
	Expect(err).ToNot(HaveOccurred())

	defer os.Remove(overwritenameops.Name())

	err = ioutil.WriteFile(overwritenameops.Name(), []byte(`- path: /name
  type: replace
  value: ((gbosh_deployment_name))`), 0644)
	Expect(err).ToNot(HaveOccurred())

	args = append(args, "-o", overwritenameops.Name(), "-v", fmt.Sprintf("gbosh_deployment_name=%s", d.name))

	session := d.Start(args...)
	Eventually(session, 300*time.Second).Should(gexec.Exit(0))

	return session
}

func (d Deployment) Start(args ...string) *gexec.Session {
	return d.director.Start(append([]string{"-d", d.name}, args...)...)
}

func (d Deployment) StartSSH(args ...string) *gexec.Session {
	return d.Start(append([]string{"ssh"}, args...)...)
}
