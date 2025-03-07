// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package acceptance

import (
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/acceptance"
)

func init() {
	acceptance.DefineFlags()
}

func TestAcceptance(t *testing.T) {
	builderImage, runImage, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.Test{
		{
			Name:      "simple path",
			App:       "simple",
			MustMatch: "PASS_INDEX",
			MustUse:   []string{phpRuntime, composerInstall, composer, phpWebConfig, utilsNginx},
		},
		{
			Name:      "custom path",
			App:       "simple",
			Path:      "/custom",
			MustMatch: "PASS_CUSTOM",
			MustUse:   []string{phpRuntime, composerInstall, composer, phpWebConfig, utilsNginx},
		},
		{
			Name:      "php ini config",
			App:       "php_ini_config",
			MustMatch: "PASS_PHP_INI",
			MustUse:   []string{phpRuntime, phpWebConfig, utilsNginx},
		},
		{
			Name:      "runtime version 7.4.27",
			App:       "simple",
			Path:      "/version?want=7.4.27",
			Env:       []string{"GOOGLE_RUNTIME_VERSION=7.4.27"},
			MustMatch: "PASS_VERSION",
			MustUse:   []string{phpRuntime, phpWebConfig, utilsNginx},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			acceptance.TestApp(t, builderImage, runImage, tc)
		})

	}
}
