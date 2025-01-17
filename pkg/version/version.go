/*
Copyright 2021.

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

package version

import (
	"os"
	"path/filepath"
)

const (
	undefinedVersion string = "undefined"
)

// version Must not be const, supposed to be set using ldflags at build time
var version = undefinedVersion

// Get returns the version as a string
func Get() string {
	return version
}

// Undefined returns if version is at it's default value
func Undefined() bool {
	return version == undefinedVersion
}

func ProgramName() string {
	if len(os.Args) == 0 {
		return "undefined"
	}
	return filepath.Base(os.Args[0])
}
