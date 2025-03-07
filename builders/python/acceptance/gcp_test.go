// Copyright 2020 Google LLC
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
package acceptance

import (
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/acceptance"
)

const (
	entrypoint    = "google.config.entrypoint"
	pythonFF      = "google.python.functions-framework"
	pythonPIP     = "google.python.pip"
	pythonRuntime = "google.python.runtime"
)

func init() {
	acceptance.DefineFlags()
}

func TestAcceptancePython(t *testing.T) {
	builderImage, runImage, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.Test{
		{
			Name:            "entrypoint from procfile web",
			App:             "simple",
			MustUse:         []string{pythonRuntime, pythonPIP, entrypoint},
			EnableCacheTest: true,
		},
		{
			Name:       "entrypoint from procfile custom",
			App:        "simple",
			Path:       "/custom",
			Entrypoint: "custom", // Must match the non-web process in Procfile.
			MustUse:    []string{pythonRuntime, pythonPIP, entrypoint},
		},
		{
			Name:    "entrypoint from env",
			App:     "simple",
			Path:    "/custom",
			Env:     []string{"GOOGLE_ENTRYPOINT=gunicorn -b :8080 custom:app"},
			MustUse: []string{pythonRuntime, pythonPIP, entrypoint},
		},
		{
			Name:    "entrypoint with env var",
			App:     "simple",
			Path:    "/env?want=bar",
			Env:     []string{"GOOGLE_ENTRYPOINT=FOO=bar gunicorn -b :8080 main:app"},
			MustUse: []string{pythonRuntime, pythonPIP, entrypoint},
		},
		// TODO (b/236138363): convert this one to a unit test
		{
			Name:                       "runtime version from .python-version",
			VersionInclusionConstraint: "3.8.12", // version is set in the .python-version file
			App:                        "python_version_file",
			Path:                       "/version?want=3.8.12",
			MustUse:                    []string{pythonRuntime, pythonPIP, entrypoint},
		},
		{
			Name:    "python with client-side scripts correctly builds as a python app",
			App:     "scripts",
			Env:     []string{"GOOGLE_ENTRYPOINT=gunicorn -b :8080 main:app"},
			MustUse: []string{pythonRuntime, pythonPIP, entrypoint},
		},
	}

	// Tests for specific versions of Python available on dl.google.com.
	for _, v := range []string{"3.7.12", "3.8.12", "3.9.10", "3.10.2"} {
		testCases = append(testCases, acceptance.Test{
			Name:    "dl.google.com runtime version " + v,
			App:     "version",
			Path:    "/version?want=" + v,
			Env:     []string{"GOOGLE_PYTHON_VERSION=" + v},
			MustUse: []string{pythonRuntime, pythonPIP, entrypoint},
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			acceptance.TestApp(t, builderImage, runImage, tc)
		})
	}
}

func TestFailuresPython(t *testing.T) {
	builderImage, runImage, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.FailureTest{
		{
			Name:      "bad runtime version",
			App:       "simple",
			Env:       []string{"GOOGLE_RUNTIME_VERSION=BAD_NEWS_BEARS", "GOOGLE_ENTRYPOINT=gunicorn -b :8080 main:app"},
			MustMatch: "invalid Python version specified",
		},
		// TODO (b/236138363): convert this one to a unit test
		{
			Name:      "python-version empty",
			App:       "empty_version",
			MustMatch: ".python-version exists but does not specify a version",
		},
		{
			Name:      "missing entrypoint",
			App:       "missing_entrypoint",
			MustMatch: `for Python, an entrypoint must be manually set, either with "GOOGLE_ENTRYPOINT" env var or by creating a "Procfile" file`,
		},
		{
			Name:      "mismatch runtime version between .python-version and env var",
			App:       "python_version_file", // set to python 3.8.12 via .python-version
			Env:       []string{"GOOGLE_PYTHON_VERSION=3.8.11"},
			MustMatch: "are inconsistent, pick one of them or set them to the same value",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			acceptance.TestBuildFailure(t, builderImage, runImage, tc)
		})
	}
}
