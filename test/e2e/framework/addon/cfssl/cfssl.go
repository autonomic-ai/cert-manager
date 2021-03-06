/*
Copyright 2019 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// package cfssl contains an addon that installs CFSSL
package cfssl

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/cert-manager/test/e2e/framework/addon/chart"
	"github.com/jetstack/cert-manager/test/e2e/framework/addon/tiller"
	"github.com/jetstack/cert-manager/test/e2e/framework/config"
)

type Cfssl struct {
	config        *config.Config
	chart         *chart.Chart
	tillerDetails *tiller.Details

	// Tiller is the tiller instance used to deploy the chart
	Tiller *tiller.Tiller

	// Name is a unique name for this Cfssl deployment
	Name string

	// Namespace is the namespace to deploy Cfssl into
	Namespace string

	// AuthKey is the authentication key
	AuthKey string

	details Details
}

type Details struct {
	// Kubectl is the path to kubectl
	Kubectl string

	// Host is the hostname that can be used to connect to Cfssl
	Host string

	// PodName is the name of the cfssl pod
	PodName string

	// Namespace is the namespace cfssl has been deployed into
	Namespace string
}

func (c *Cfssl) Setup(cfg *config.Config) error {
	if c.Name == "" {
		return fmt.Errorf("Name field must be set on Cfssl addon")
	}

	if c.AuthKey == "" {
		return fmt.Errorf("AuthKey field must be set on Cfssl addon")
	}

	if c.Namespace == "" {
		return fmt.Errorf("Namespace name must be specified")
	}

	if c.Tiller == nil {
		return fmt.Errorf("Tiller field must be set on Cfssl addon")
	}

	var err error

	c.tillerDetails, err = c.Tiller.Details()
	if err != nil {
		return err
	}

	c.chart = &chart.Chart{
		Tiller:      c.Tiller,
		ReleaseName: c.ReleaseName(),
		Namespace:   c.Namespace,
		ChartName:   cfg.RepoRoot + "/test/e2e/charts/cfssl",
		// doesn't matter when installing from disk
		ChartVersion: "0",
	}

	err = c.chart.Setup(cfg)
	if err != nil {
		return err
	}

	return nil
}

// Provision will actually deploy this instance of Cfssl to the cluster.
func (c *Cfssl) Provision() error {
	err := c.chart.Provision()
	if err != nil {
		return err
	}

	// otherwise lookup the newly created pods name
	kubeClient := c.Tiller.Base.Details().KubeClient

	retries := 5
	for {
		pods, err := kubeClient.CoreV1().Pods(c.Namespace).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=cfssl,release=%s", c.ReleaseName()),
		})
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			if retries == 0 {
				return fmt.Errorf("failed to create cfssl pod within 10s")
			}
			retries--
			time.Sleep(time.Second * 2)
			continue
		}
		cfsslPod := pods.Items[0]
		// If the cfssl pod exists but is just waiting to be created, we allow
		// it a bit longer.
		if len(cfsslPod.Status.ContainerStatuses) == 0 || !cfsslPod.Status.ContainerStatuses[0].Ready {
			retries--
			time.Sleep(time.Second * 5)
			continue
		}
		c.details.PodName = cfsslPod.Name
		break
	}

	c.details.Namespace = c.Namespace
	c.details.Host = fmt.Sprintf("http://%s.%s:8080", c.ReleaseName(), c.Namespace)

	return nil
}

// Details returns details that can be used to utilise the instance of Cfssl.
func (c *Cfssl) Details() *Details {
	return &c.details
}

// ReleaseName returns the name of the cfssl release
func (c *Cfssl) ReleaseName() string {
	return fmt.Sprintf("chart-cfssl-" + c.Name)
}

// Deprovision will destroy this instance of Cfssl
func (c *Cfssl) Deprovision() error {
	return c.chart.Deprovision()
}

func (c *Cfssl) SupportsGlobal() bool {
	// We don't support global instances of cfssl currently as we need to generate
	// PKI details at deploy time and make them available to tests.
	return false
}

func (c *Cfssl) Logs() (map[string]string, error) {
	return c.chart.Logs()
}

func NewAuthKeySecret(name, authKey string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		StringData: map[string]string{
			"auth-key": authKey,
		},
	}
}
