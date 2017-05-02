package gbosh

import (
	"fmt"
	"github.com/cloudfoundry/bosh-utils/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"strings"
)

type Director struct {
	uuidgen uuid.Generator
}

func NewDirectorFromEnv() Director {
	return Director{
		uuidgen: uuid.NewGenerator(),
	}
}

func (d Director) Start(args ...string) *gexec.Session {
	GinkgoWriter.Write([]byte(fmt.Sprintf("+ bosh %s", strings.Join(args, " "))))

	cmd := exec.Command("bosh", args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func (d Director) NewDeployment() Deployment {
	name, err := d.uuidgen.Generate()
	Expect(err).ToNot(HaveOccurred())

	return Deployment{
		director: d,
		name:     fmt.Sprintf("gbosh-%s", name),
	}
}
