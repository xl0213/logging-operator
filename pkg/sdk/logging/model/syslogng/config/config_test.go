// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"strings"
	"testing"

	"github.com/banzaicloud/logging-operator/pkg/sdk/logging/api/v1beta1"
	"github.com/banzaicloud/logging-operator/pkg/sdk/logging/model/syslogng/filter"
	"github.com/banzaicloud/logging-operator/pkg/sdk/logging/model/syslogng/output"
	"github.com/banzaicloud/operator-tools/pkg/secret"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRenderConfigInto(t *testing.T) {
	testCases := map[string]struct {
		input   Input
		wantOut string
		wantErr any
	}{
		"empty input": {
			input: Input{
				SecretLoaderFactory: &secretLoaderFactory{},
			},
			wantErr: true,
		},
		"no syslog-ng spec": {
			input: Input{
				Logging: v1beta1.Logging{
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: nil,
					},
				},
				SecretLoaderFactory: &secretLoaderFactory{},
			},
			wantErr: true,
		},
		"single flow with single output": {
			input: Input{
				SourcePort: 601,
				Logging: v1beta1.Logging{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "config-test",
						Name:      "test",
					},
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: &v1beta1.SyslogNGSpec{},
					},
				},
				Outputs: []v1beta1.SyslogNGOutput{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test-syslog-out",
						},
						Spec: v1beta1.SyslogNGOutputSpec{
							Syslog: &output.SyslogOutput{
								Host:      "test.local",
								Transport: "tcp",
							},
						},
					},
				},
				Flows: []v1beta1.SyslogNGFlow{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test-flow",
						},
						Spec: v1beta1.SyslogNGFlowSpec{
							Match: &v1beta1.SyslogNGMatch{
								Regexp: &filter.RegexpMatchExpr{
									Pattern: "nginx",
									Value:   "kubernetes.labels.app",
								},
							},
							Filters: []v1beta1.SyslogNGFilter{
								{
									Rewrite: []filter.RewriteConfig{
										{
											Set: &filter.SetConfig{
												FieldName: "cluster",
												Value:     "test-cluster",
											},
										},
									},
								},
							},
							LocalOutputRefs: []string{"test-syslog-out"},
						},
					},
				},
				SecretLoaderFactory: &secretLoaderFactory{},
			},
			wantOut: untab(`@version: 3.37

@include "scl.conf"

source "main_input" {
    channel {
        source {
            network(flags("no-parse") port(601) transport("tcp"));
        };
        parser {
            json-parser(prefix("json."));
        };
    };
};

destination "output_default_test-syslog-out" {
	syslog("test.local" transport("tcp") persist_name("output_default_test-syslog-out"));
};

filter "flow_default_test-flow_match" {
	match("nginx" value("kubernetes.labels.app"));
};
rewrite "flow_default_test-flow_filters_0" {
	set("test-cluster" value("cluster"));
};
log {
	source("main_input");
	filter {
		match("default" value("json.kubernetes.namespace_name") type("string"));
	};
	filter("flow_default_test-flow_match");
	rewrite("flow_default_test-flow_filters_0");
	destination("output_default_test-syslog-out");
};
`),
		},
		"global options": {
			input: Input{
				Logging: v1beta1.Logging{
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: &v1beta1.SyslogNGSpec{
							GlobalOptions: &v1beta1.GlobalOptions{
								StatsLevel: amp(3),
								StatsFreq:  amp(0),
							},
						},
					},
				},
				SourcePort:          601,
				SecretLoaderFactory: &secretLoaderFactory{},
			},
			wantOut: `@version: 3.37

@include "scl.conf"

options {
    stats_level(3);
    stats_freq(10);
};

source "main_input" {
    channel {
        source {
            network(flags("no-parse") port(601) transport("tcp"));
        };
        parser {
            json-parser(prefix("json."));
        };
    };
};
`,
		},
		"rewrite condition": {
			input: Input{
				Logging: v1beta1.Logging{
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: &v1beta1.SyslogNGSpec{},
					},
				},
				SourcePort:          601,
				SecretLoaderFactory: &secretLoaderFactory{},
				Flows: []v1beta1.SyslogNGFlow{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test-flow",
						},
						Spec: v1beta1.SyslogNGFlowSpec{
							Filters: []v1beta1.SyslogNGFilter{
								{
									Rewrite: []filter.RewriteConfig{
										{
											Unset: &filter.UnsetConfig{
												FieldName: "MESSAGE",
												Condition: &filter.MatchExpr{
													Not: &filter.MatchExpr{
														Regexp: &filter.RegexpMatchExpr{
															Pattern: "foo",
															Value:   "MESSAGE",
															Type:    "string",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantOut: `@version: 3.37

@include "scl.conf"

source "main_input" {
    channel {
        source {
            network(flags("no-parse") port(601) transport("tcp"));
        };
        parser {
            json-parser(prefix("json."));
        };
    };
};

rewrite "flow_default_test-flow_filters_0" {
    unset(value("MESSAGE") condition((not match("foo" value("MESSAGE") type("string")))));
};
log {
    source("main_input");
    filter {
        match("default" value("json.kubernetes.namespace_name") type("string"));
    };
    rewrite("flow_default_test-flow_filters_0");
};
`,
		},
		"output with secret": {
			input: Input{
				Logging: v1beta1.Logging{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "logging",
						Name:      "test",
					},
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: &v1beta1.SyslogNGSpec{},
					},
				},
				Outputs: []v1beta1.SyslogNGOutput{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "my-output",
						},
						Spec: v1beta1.SyslogNGOutputSpec{
							Syslog: &output.SyslogOutput{
								Host: "127.0.0.1",
								TLS: &output.TLS{
									CaFile: &secret.Secret{
										MountFrom: &secret.ValueFrom{
											SecretKeyRef: &corev1.SecretKeySelector{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: "my-secret",
												},
												Key: "tls.crt",
											},
										},
									},
								},
							},
						},
					},
				},
				SecretLoaderFactory: &secretLoaderFactory{
					reader: secretReader{
						secrets: []corev1.Secret{
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "default",
									Name:      "my-secret",
								},
								Data: map[string][]byte{
									"tls.crt": []byte("asdf"),
								},
							},
						},
					},
					mountPath: "/etc/syslog-ng/secret",
				},
				SourcePort: 601,
			},
			wantOut: `@version: 3.37

@include "scl.conf"

source "main_input" {
    channel {
        source {
            network(flags("no-parse") port(601) transport("tcp"));
        };
        parser {
            json-parser(prefix("json."));
        };
    };
};

destination "output_default_my-output" {
    syslog("127.0.0.1" tls(ca_file("/etc/syslog-ng/secret/default-my-secret-tls.crt")) persist_name("output_default_my-output"));
};
`,
		},
		"parser": {
			input: Input{
				Logging: v1beta1.Logging{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "logging",
						Name:      "test",
					},
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: &v1beta1.SyslogNGSpec{},
					},
				},
				Flows: []v1beta1.SyslogNGFlow{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test-flow",
						},
						Spec: v1beta1.SyslogNGFlowSpec{
							Filters: []v1beta1.SyslogNGFilter{
								{
									Parser: &filter.ParserConfig{
										Regexp: &filter.RegexpParser{
											Patterns: []string{
												".*test_field -> (?<test_field>.*)$",
											},
											Prefix: ".regexp.",
										},
									},
								},
							},
						},
					},
				},
				SecretLoaderFactory: &secretLoaderFactory{},
				SourcePort:          601,
			},
			wantOut: `@version: 3.37

@include "scl.conf"

source "main_input" {
    channel {
        source {
            network(flags("no-parse") port(601) transport("tcp"));
        };
        parser {
            json-parser(prefix("json."));
        };
    };
};

parser "flow_default_test-flow_filters_0" {
    regexp-parser(patterns(".*test_field -> (?<test_field>.*)$") prefix(".regexp."));
};
log {
    source("main_input");
    filter {
        match("default" value("json.kubernetes.namespace_name") type("string"));
    };
    parser("flow_default_test-flow_filters_0");
};
`,
		},
		"filter with name": {
			input: Input{
				Logging: v1beta1.Logging{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "logging",
						Name:      "test",
					},
					Spec: v1beta1.LoggingSpec{
						SyslogNGSpec: &v1beta1.SyslogNGSpec{},
					},
				},
				Flows: []v1beta1.SyslogNGFlow{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "test-flow",
						},
						Spec: v1beta1.SyslogNGFlowSpec{
							Filters: []v1beta1.SyslogNGFilter{
								{
									ID: "remove message",
									Rewrite: []filter.RewriteConfig{
										{
											Unset: &filter.UnsetConfig{
												FieldName: "MESSAGE",
											},
										},
									},
								},
							},
						},
					},
				},
				SecretLoaderFactory: &secretLoaderFactory{},
				SourcePort:          601,
			},
			wantOut: `@version: 3.37

@include "scl.conf"

source "main_input" {
    channel {
        source {
            network(flags("no-parse") port(601) transport("tcp"));
        };
        parser {
            json-parser(prefix("json."));
        };
    };
};

rewrite "flow_default_test-flow_filters_remove message" {
    unset(value("MESSAGE"));
};
log {
    source("main_input");
    filter {
        match("default" value("json.kubernetes.namespace_name") type("string"));
    };
    rewrite("flow_default_test-flow_filters_remove message");
};
`,
		},
	}
	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			var buf strings.Builder
			err := RenderConfigInto(testCase.input, &buf)
			checkError(t, testCase.wantErr, err)
			require.Equal(t, testCase.wantOut, buf.String())
		})
	}
}

type secretLoaderFactory struct {
	reader    client.Reader
	mountPath string
	secrets   secret.MountSecrets
}

func (f *secretLoaderFactory) SecretLoaderForNamespace(ns string) secret.SecretLoader {
	return secret.NewSecretLoader(f.reader, ns, f.mountPath, &f.secrets)
}

type secretReader struct {
	secrets []corev1.Secret
}

func (r secretReader) Get(_ context.Context, key client.ObjectKey, obj client.Object) error {
	if secret, ok := obj.(*corev1.Secret); ok {
		if secret == nil {
			return nil
		}
		for _, s := range r.secrets {
			if s.Namespace == key.Namespace && s.Name == key.Name {
				*secret = s
				return nil
			}
		}
		return apierrors.NewNotFound(corev1.Resource("secret"), key.String())
	}
	return apierrors.NewNotFound(schema.GroupResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Resource: strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind),
	}, key.String())
}

func (r secretReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	panic("not implemented")
}

var _ client.Reader = (*secretReader)(nil)