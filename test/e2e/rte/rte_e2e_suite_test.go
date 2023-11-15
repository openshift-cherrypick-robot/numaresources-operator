/*
 * Copyright 2021 Red Hat, Inc.
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

package rte

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/k8stopologyawareschedwg/resource-topology-exporter/test/e2e/rte"
	_ "github.com/k8stopologyawareschedwg/resource-topology-exporter/test/e2e/topology_updater"
)

var (
	randomSeed int64
)

func TestRTE(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RTE Test Suite")
}

var _ = BeforeSuite(func() {
	By(fmt.Sprintf("Using random seed %v", randomSeed))
})
