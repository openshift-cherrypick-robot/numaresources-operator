/*
 * Copyright 2022 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fixture

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/onsi/ginkgo/v2"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"

	"github.com/openshift-kni/numaresources-operator/internal/objects"
	intwait "github.com/openshift-kni/numaresources-operator/internal/wait"
	e2eclient "github.com/openshift-kni/numaresources-operator/test/utils/clients"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Fixture struct {
	// Client defines the API client to run CRUD operations, that will be used for testing
	Client client.Client
	// K8sClient defines k8s client to run subresource operations, for example you should use it to get pod logs
	K8sClient      *kubernetes.Clientset
	Namespace      corev1.Namespace
	InitialNRTList nrtv1alpha2.NodeResourceTopologyList
	Skipped        bool
}

const (
	defaultTeardownTime = 180 * time.Second
	defaultCooldownTime = 30 * time.Second
)

type Options uint

const (
	OptionNone          = 0
	OptionRandomizeName = 1 << iota
)

var (
	teardownTime time.Duration
	cooldownTime time.Duration
)

func init() {
	cooldownTime = getCooldownTimeFromEnvVar()
	teardownTime = getTeardownTimeFromEnvVar()
}

func SetupWithOptions(name string, nrtList nrtv1alpha2.NodeResourceTopologyList, options Options) (*Fixture, error) {
	if !e2eclient.ClientsEnabled {
		return nil, fmt.Errorf("clients not enabled")
	}
	randomizeName := (options & OptionRandomizeName) == OptionRandomizeName
	ns, err := setupNamespace(e2eclient.Client, name, randomizeName)
	if err != nil {
		klog.Errorf("cannot setup namespace %q: %v", name, err)
		return nil, err
	}
	ginkgo.By(fmt.Sprintf("set up the test namespace %q", ns.Name))
	return &Fixture{
		Client:         e2eclient.Client,
		K8sClient:      e2eclient.K8sClient,
		Namespace:      ns,
		InitialNRTList: nrtList,
	}, nil

}

func Setup(baseName string, nrtList nrtv1alpha2.NodeResourceTopologyList) (*Fixture, error) {
	return SetupWithOptions(baseName, nrtList, OptionRandomizeName)
}

func Teardown(ft *Fixture) error {
	ginkgo.By(fmt.Sprintf("tearing down the test namespace %q", ft.Namespace.Name))
	err := teardownNamespace(ft.Client, ft.Namespace)
	if err != nil {
		klog.Errorf("cannot teardown namespace %q: %s", ft.Namespace.Name, err)
		return err
	}

	if ft.Skipped {
		ft.Skipped = false
		ginkgo.By(fmt.Sprintf("skipped - nothing to cool down"))
		return nil
	}

	Cooldown(ft)
	return nil
}

func Skip(ft *Fixture, message string) {
	ft.Skipped = true
	ginkgo.Skip(message)
}

func Skipf(ft *Fixture, format string, args ...interface{}) {
	Skip(ft, fmt.Sprintf(format, args...))
}

func Cooldown(ft *Fixture) {
	if len(ft.InitialNRTList.Items) > 0 {
		ginkgo.By("cooldown by verifying NRTs data is settled to the initial state")
		intwait.With(ft.Client).Interval(5*time.Second).Timeout(1*time.Minute).ForNodeResourceTopologiesEqualTo(context.TODO(), &ft.InitialNRTList, intwait.NRTIgnoreNothing)
		return
	}
	klog.Warningf("cooling down for %v", cooldownTime)
	time.Sleep(cooldownTime)
}
func setupNamespace(cli client.Client, baseName string, randomize bool) (corev1.Namespace, error) {
	name := baseName
	if randomize {
		// intentionally avoid GenerateName like the k8s e2e framework does
		name = RandomizeName(baseName)
	}
	ns := objects.NewNamespace(name)
	err := cli.Create(context.TODO(), ns)
	if err != nil {
		return *ns, err
	}

	// again we do like the k8s e2e framework does and we try to be robust
	var updatedNs corev1.Namespace
	err = wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
		err := cli.Get(context.TODO(), client.ObjectKeyFromObject(ns), &updatedNs)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return updatedNs, err
}

func teardownNamespace(cli client.Client, ns corev1.Namespace) error {
	err := cli.Delete(context.TODO(), &ns)
	if apierrors.IsNotFound(err) {
		return nil
	}

	klog.Warningf("tearing down up to %v", teardownTime)

	iterations := 0
	updatedNs := corev1.Namespace{}
	return wait.PollImmediate(1*time.Second, teardownTime, func() (bool, error) {
		iterations++
		err := cli.Get(context.TODO(), client.ObjectKeyFromObject(&ns), &updatedNs)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		if iterations%10 == 0 {
			klog.InfoS("tearing down namespace: still not gone", "namespace", ns.Name, "error", err)
		}
		return false, nil
	})
}

func RandomizeName(baseName string) string {
	return fmt.Sprintf("%s-%s", baseName, strconv.Itoa(rand.Intn(10000)))
}

func getCooldownTimeFromEnvVar() time.Duration {
	raw, ok := os.LookupEnv("E2E_NROP_TEST_COOLDOWN")
	if !ok {
		return defaultCooldownTime
	}
	val, err := time.ParseDuration(raw)
	if err != nil {
		klog.Errorf("cannot parse the provided test cooldown time (fallback to default: %v): %v", defaultCooldownTime, err)
		return defaultCooldownTime
	}
	return val
}

func getTeardownTimeFromEnvVar() time.Duration {
	raw, ok := os.LookupEnv("E2E_NROP_TEST_TEARDOWN")
	if !ok {
		return defaultTeardownTime
	}
	val, err := time.ParseDuration(raw)
	if err != nil {
		klog.Errorf("cannot parse the provided test teardown time (fallback to default: %v): %v", defaultTeardownTime, err)
		return defaultTeardownTime
	}
	return val
}
