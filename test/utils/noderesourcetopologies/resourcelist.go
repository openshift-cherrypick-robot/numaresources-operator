/*
Copyright 2022 The Kubernetes Authors.

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

package noderesourcetopologies

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func ResourceListToString(res corev1.ResourceList) string {
	items := []string{}
	for resName, resQty := range res {
		items = append(items, fmt.Sprintf("%s=%s", string(resName), resQty.String()))
	}
	return strings.Join(items, ", ")
}
