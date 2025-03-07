// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package acceptance_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/acceptance"
)

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
			Name:       "extension app",
			App:        "extension",
			MustMatch:  "PASS_EXT",
			MustUse:    []string{composer, composerInstall, phpRuntime, phpWebConfig, utilsNginx},
			MustNotUse: []string{composerGCPBuild, functionFramework, cloudFunctions},
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
