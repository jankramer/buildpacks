// Copyright 2021 Google LLC
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
			Name: "function without deps",
			App:  "no_deps",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path: "/Func",
		},
		{
			Name:       "vendored function without dependencies",
			App:        "no_framework_vendored",
			Env:        []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path:       "/Func",
			MustOutput: []string{"Found function with vendored dependencies excluding functions-framework"},
		},
		{
			Name:            "function without framework",
			App:             "no_framework",
			Env:             []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path:            "/Func",
			MustOutput:      []string{"go.sum not found, generating"},
			EnableCacheTest: true,
		},
		{
			Name:          "function with go.sum",
			App:           "no_framework",
			Env:           []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Setup:         goSumSetup,
			Path:          "/Func",
			MustNotOutput: []string{"go.sum not found, generating"},
		},
		{
			Name:  "vendored function with framework",
			App:   "with_framework",
			Env:   []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path:  "/Func",
			Setup: vendorSetup,
		},
		{
			Name: "function with old framework",
			App:  "with_framework_old_version",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path: "/Func",
		},
		{
			Name:  "vendored function with old framework",
			App:   "with_framework_old_version",
			Env:   []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path:  "/Func",
			Setup: vendorSetup,
		},
		{
			Name: "function at /*",
			App:  "no_framework",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Path: "/",
		},
		{
			Name: "function with subdirectories",
			App:  "with_subdir",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
		},
		{
			Name: "declarative http function",
			App:  "declarative_http",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
		},
		{
			Name: "declarative http anonymous function",
			App:  "declarative_anonymous",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
		},
		{
			Name:        "declarative cloudevent function",
			App:         "declarative_cloud_event",
			RequestType: acceptance.CloudEventType,
			MustMatch:   "",
			Env:         []string{"GOOGLE_FUNCTION_TARGET=Func"},
		},
		{
			Name:        "non declarative cloudevent function",
			App:         "non_declarative_cloud_event",
			RequestType: acceptance.CloudEventType,
			MustMatch:   "",
			Env:         []string{"GOOGLE_FUNCTION_TARGET=Func", "GOOGLE_FUNCTION_SIGNATURE_TYPE=cloudevent"},
		},
		{
			Name: "declarative and non declarative registration",
			App:  "declarative_old_and_new",
			Env:  []string{"GOOGLE_FUNCTION_TARGET=Func"},
		},
		{
			Name:                "no auto registration in main.go if declarative detected",
			App:                 "declarative_cloud_event",
			RequestType:         acceptance.CloudEventType,
			MustMatchStatusCode: 404,
			MustMatch:           "404 page not found",
			// If the buildpack detects the declarative functions package, then
			// functions must be explicitly registered. The main.go written out
			// by the buildpack will NOT use the GOOGLE_FUNCTION_TARGET env var
			// to register a non-declarative function.
			Env: []string{"GOOGLE_FUNCTION_TARGET=NonDeclarativeFunc", "GOOGLE_FUNCTION_SIGNATURE_TYPE=cloudevent"},
		},
		{
			Name:                "declarative function signature but wrong target",
			App:                 "declarative_http",
			Env:                 []string{"GOOGLE_FUNCTION_TARGET=ThisDoesntExist"},
			MustMatchStatusCode: 404,
			MustMatch:           "404 page not found",
		},
		{
			Name:        "background function",
			App:         "background_function",
			RequestType: acceptance.BackgroundEventType,
			Env:         []string{"GOOGLE_FUNCTION_TARGET=Func", "GOOGLE_FUNCTION_SIGNATURE_TYPE=event"},
		},
	}

	for _, tc := range testCases {
		tc.Env = append(tc.Env, "X_GOOGLE_TARGET_PLATFORM=gcf")
		tc.FilesMustExist = append(tc.FilesMustExist,
			"/layers/google.utils.archive-source/src/source-code.tar.gz",
			"/workspace/.googlebuild/source-code.tar.gz",
		)
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if tc.Setup != nil {
				t.Skip("TODO: The setup functions require go to be pre-installed which is not true for the unified builder")
			}
			acceptance.TestApp(t, builderImage, runImage, tc)
		})
	}
}

func TestFailures(t *testing.T) {
	builderImage, runImage, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.FailureTest{
		{
			App:       "no_framework_relative",
			Env:       []string{"GOOGLE_FUNCTION_TARGET=Func"},
			MustMatch: "the module path in the function's go.mod must contain a dot in the first path element before a slash, e.g. example.com/module, found: func",
		},
		{
			App:       "no_framework",
			Env:       []string{"GOOGLE_FUNCTION_TARGET=Func"},
			Setup:     vendorSetup,
			MustMatch: "vendored dependencies must include \"github.com/GoogleCloudPlatform/functions-framework-go\"; if your function does not depend on the module, please add a blank import: `_ \"github.com/GoogleCloudPlatform/functions-framework-go/funcframework\"`",
		},
	}

	for _, tc := range testCases {
		tc.Env = append(tc.Env, "X_GOOGLE_TARGET_PLATFORM=gcf")
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if tc.Setup != nil {
				t.Skip("TODO: The setup functions require go to be pre-installed which is not true for the unified builder")
			}
			acceptance.TestBuildFailure(t, builderImage, runImage, tc)
		})
	}
}
