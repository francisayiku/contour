// Copyright © 2019 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dag

import (
	"testing"

	ingressroutev1 "github.com/heptio/contour/apis/contour/v1beta1"
	projectcontour "github.com/heptio/contour/apis/projectcontour/v1alpha1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesCacheInsert(t *testing.T) {
	tests := map[string]struct {
		pre  []interface{}
		obj  interface{}
		want bool
	}{
		"insert secret": {
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: false,
		},
		"insert secret referenced by ingress": {
			pre: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "www",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							SecretName: "secret",
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret w/ wrong type referenced by ingress": {
			pre: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "www",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							SecretName: "secret",
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: "banana",
			},
			want: false,
		},
		"insert secret referenced by ingress via tls delegation": {
			pre: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "www",
						Namespace: "extra",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							SecretName: "default/secret",
						}},
					},
				},

				&ingressroutev1.TLSCertificateDelegation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "delegation",
						Namespace: "default",
					},
					Spec: ingressroutev1.TLSCertificateDelegationSpec{
						Delegations: []ingressroutev1.CertificateDelegation{{
							SecretName: "secret",
							TargetNamespaces: []string{
								"extra",
							},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret referenced by ingress via wildcard tls delegation": {
			pre: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "www",
						Namespace: "extra",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							SecretName: "default/secret",
						}},
					},
				},

				&ingressroutev1.TLSCertificateDelegation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "delegation",
						Namespace: "default",
					},
					Spec: ingressroutev1.TLSCertificateDelegationSpec{
						Delegations: []ingressroutev1.CertificateDelegation{{
							SecretName: "secret",
							TargetNamespaces: []string{
								"*",
							},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret referenced by ingressroute": {
			pre: []interface{}{
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &ingressroutev1.VirtualHost{
							TLS: &ingressroutev1.TLS{
								SecretName: "secret",
							},
						},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret referenced by ingressroute via tls delegation": {
			pre: []interface{}{
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "extra",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &ingressroutev1.VirtualHost{
							TLS: &ingressroutev1.TLS{
								SecretName: "default/secret",
							},
						},
					},
				},
				&ingressroutev1.TLSCertificateDelegation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "delegation",
						Namespace: "default",
					},
					Spec: ingressroutev1.TLSCertificateDelegationSpec{
						Delegations: []ingressroutev1.CertificateDelegation{{
							SecretName: "secret",
							TargetNamespaces: []string{
								"extra",
							},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret referenced by ingressroute via wildcard tls delegation": {
			pre: []interface{}{
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "extra",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &ingressroutev1.VirtualHost{
							TLS: &ingressroutev1.TLS{
								SecretName: "default/secret",
							},
						},
					},
				},
				&ingressroutev1.TLSCertificateDelegation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "delegation",
						Namespace: "default",
					},
					Spec: ingressroutev1.TLSCertificateDelegationSpec{
						Delegations: []ingressroutev1.CertificateDelegation{{
							SecretName: "secret",
							TargetNamespaces: []string{
								"*",
							},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},

		"insert secret referenced by httploadbalancer": {
			pre: []interface{}{
				&projectcontour.HTTPLoadBalancer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: projectcontour.HTTPLoadBalancerSpec{
						VirtualHost: &projectcontour.VirtualHost{
							TLS: &projectcontour.TLS{
								SecretName: "secret",
							},
						},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret referenced by httploadbalancer via tls delegation": {
			pre: []interface{}{
				&projectcontour.HTTPLoadBalancer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "extra",
					},
					Spec: projectcontour.HTTPLoadBalancerSpec{
						VirtualHost: &projectcontour.VirtualHost{
							TLS: &projectcontour.TLS{
								SecretName: "default/secret",
							},
						},
					},
				},
				&projectcontour.TLSCertificateDelegation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "delegation",
						Namespace: "default",
					},
					Spec: projectcontour.TLSCertificateDelegationSpec{
						Delegations: []projectcontour.CertificateDelegation{{
							SecretName: "secret",
							TargetNamespaces: []string{
								"extra",
							},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert secret referenced by httploadbalancer  via wildcard tls delegation": {
			pre: []interface{}{
				&projectcontour.HTTPLoadBalancer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "extra",
					},
					Spec: projectcontour.HTTPLoadBalancerSpec{
						VirtualHost: &projectcontour.VirtualHost{
							TLS: &projectcontour.TLS{
								SecretName: "default/secret",
							},
						},
					},
				},
				&projectcontour.TLSCertificateDelegation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "delegation",
						Namespace: "default",
					},
					Spec: projectcontour.TLSCertificateDelegationSpec{
						Delegations: []projectcontour.CertificateDelegation{{
							SecretName: "secret",
							TargetNamespaces: []string{
								"*",
							},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"insert certificate secret": {
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ca",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"ca.crt": []byte("ca"),
				},
			},
			// TODO(dfc) this should be false because the CA secret is
			// not referenced, but computing its reference duplicates the
			// work done rebuilding the dag so for the moment assume that
			// any CA secret causes a rebuild.
			want: true,
		},
		"insert certificate secret referenced by ingressroute": {
			pre: []interface{}{
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-com",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &ingressroutev1.VirtualHost{
							Fqdn: "example.com",
						},
						Routes: []ingressroutev1.Route{{
							Match: "/",
							Services: []ingressroutev1.Service{{
								Name: "kuard",
								Port: 8080,
								UpstreamValidation: &ingressroutev1.UpstreamValidation{
									CACertificate: "ca",
									SubjectName:   "example.com",
								},
							}},
						}},
					},
				},
			},
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ca",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"ca.crt": []byte("ca"),
				},
			},
			want: true,
		},
		/*
			"insert certificate secret referenced by httploadbalancer": {
				pre: []interface{}{
					&projectcontour.HTTPLoadBalancer{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-com",
							Namespace: "default",
						},
						Spec: projectcontour.HTTPLoadBalancerSpec{
							VirtualHost: &projectcontour.VirtualHost{
								Fqdn: "example.com",
							},
							Routes: []projectcontour.Route{{
								Match: "/",
								Services: []projectcontour.Service{{
									Name: "kuard",
									Port: 8080,
									UpstreamValidation: &projectcontour.UpstreamValidation{
										CACertificate: "ca",
										SubjectName:   "example.com",
									},
								}},
							}},
						},
					},
				},
				obj: &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ca",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"ca.crt": []byte("ca"),
					},
				},
				want: true,
			},
		*/
		"insert ingress empty ingress class": {
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "incorrect",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert ingress incorrect kubernetes.io/ingress.class": {
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "incorrect",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"insert ingress incorrect contour.heptio.com/ingress.class": {
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "incorrect",
					Namespace: "default",
					Annotations: map[string]string{
						"contour.heptio.com/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"insert ingress explicit kubernetes.io/ingress.class": {
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "incorrect",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": new(KubernetesCache).ingressClass(),
					},
				},
			},
			want: true,
		},
		"insert ingress explicit contour.heptio.com/ingress.class": {
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "incorrect",
					Namespace: "default",
					Annotations: map[string]string{
						"contour.heptio.com/ingress.class": new(KubernetesCache).ingressClass(),
					},
				},
			},
			want: true,
		},
		"insert ingressroute empty ingress annotation": {
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kuard",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert ingressroute incorrect contour.heptio.com/ingress.class": {
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
					Annotations: map[string]string{
						"contour.heptio.com/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"insert ingressroute incorrect kubernetes.io/ingress.class": {
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"insert ingressroute: explicit contour.heptio.com/ingress.class": {
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kuard",
					Namespace: "default",
					Annotations: map[string]string{
						"contour.heptio.com/ingress.class": new(KubernetesCache).ingressClass(),
					},
				},
			},
			want: true,
		},
		"insert ingressroute explicit kubernetes.io/ingress.class": {
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kuard",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": new(KubernetesCache).ingressClass(),
					},
				},
			},
			want: true,
		},
		"insert httploadbalancer empty ingress annotation": {
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kuard",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert httploadbalancer incorrect contour.heptio.com/ingress.class": {
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
					Annotations: map[string]string{
						"contour.heptio.com/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"insert httploadbalancer incorrect kubernetes.io/ingress.class": {
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "simple",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"insert httploadbalancer: explicit contour.heptio.com/ingress.class": {
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kuard",
					Namespace: "default",
					Annotations: map[string]string{
						"contour.heptio.com/ingress.class": new(KubernetesCache).ingressClass(),
					},
				},
			},
			want: true,
		},
		"insert httploadbalancer explicit kubernetes.io/ingress.class": {
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kuard",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": new(KubernetesCache).ingressClass(),
					},
				},
			},
			want: true,
		},
		"insert tls contour/v1beta1.certificate delegation": {
			obj: &ingressroutev1.TLSCertificateDelegation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delegate",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert tls projectcontour/v1alpha1.certificate delegation": {
			obj: &projectcontour.TLSCertificateDelegation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delegate",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert httploadbalancer": {
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httplb",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert unknown": {
			obj:  "not an object",
			want: false,
		},
		"insert service": {
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: false,
		},
		"insert service referenced by ingress backend": {
			pre: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "www",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "service",
						},
					},
				},
			},
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert service in different namespace": {
			pre: []interface{}{
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "www",
						Namespace: "kube-system",
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "service",
						},
					},
				},
			},
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: false,
		},
		"insert service referenced by ingressroute": {
			pre: []interface{}{
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						Routes: []ingressroutev1.Route{{
							Services: []ingressroutev1.Service{{
								Name: "service",
							}},
						}},
					},
				},
			},
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert service referenced by ingressroute tcpproxy": {
			pre: []interface{}{
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						TCPProxy: &ingressroutev1.TCPProxy{
							Services: []ingressroutev1.Service{{
								Name: "service",
							}},
						},
					},
				},
			},
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert service referenced by httploadbalancer": {
			pre: []interface{}{
				&projectcontour.HTTPLoadBalancer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: projectcontour.HTTPLoadBalancerSpec{
						Routes: []projectcontour.Route{{
							Services: []projectcontour.Service{{
								Name: "service",
							}},
						}},
					},
				},
			},
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: true,
		},
		"insert service referenced by httploadbalancer tcpproxy": {
			pre: []interface{}{
				&projectcontour.HTTPLoadBalancer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: projectcontour.HTTPLoadBalancerSpec{
						TCPProxy: &projectcontour.TCPProxy{
							Services: []projectcontour.Service{{
								Name: "service",
							}},
						},
					},
				},
			},
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cache := KubernetesCache{
				FieldLogger: testLogger(t),
			}
			for _, p := range tc.pre {
				cache.Insert(p)
			}
			got := cache.Insert(tc.obj)
			if tc.want != got {
				t.Fatalf("Insert(%v): expected %v, got %v", tc.obj, tc.want, got)
			}
		})
	}
}

func TestKubernetesCacheRemove(t *testing.T) {
	cache := func(objs ...interface{}) *KubernetesCache {
		cache := KubernetesCache{
			FieldLogger: testLogger(t),
		}
		for _, o := range objs {
			cache.Insert(o)
		}
		return &cache
	}

	tests := map[string]struct {
		cache *KubernetesCache
		obj   interface{}
		want  bool
	}{
		"remove secret": {
			cache: cache(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			}),
			obj: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Type: v1.SecretTypeTLS,
			},
			want: true,
		},
		"remove service": {
			cache: cache(&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			}),
			obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
				},
			},
			want: true,
		},
		"remove ingress": {
			cache: cache(&v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingress",
					Namespace: "default",
				},
			}),
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingress",
					Namespace: "default",
				},
			},
			want: true,
		},
		"remove ingress incorrect ingressclass": {
			cache: cache(&v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingress",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			}),
			obj: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingress",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"remove ingressroute": {
			cache: cache(&ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
				},
			}),
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
				},
			},
			want: true,
		},
		"remove ingressroute incorrect ingressclass": {
			cache: cache(&ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			}),
			obj: &ingressroutev1.IngressRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"remove httploadbalancer": {
			cache: cache(&projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
				},
			}),
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
				},
			},
			want: true,
		},
		"remove httploadbalancer incorrect ingressclass": {
			cache: cache(&projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			}),
			obj: &projectcontour.HTTPLoadBalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute",
					Namespace: "default",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
			},
			want: false,
		},
		"remove unknown": {
			cache: cache("not an object"),
			obj:   "not an object",
			want:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.cache.Remove(tc.obj)
			if tc.want != got {
				t.Fatalf("Remove(%v): expected %v, got %v", tc.obj, tc.want, got)
			}
		})
	}
}

func testLogger(t *testing.T) logrus.FieldLogger {
	log := logrus.New()
	log.Out = &testWriter{t}
	return log
}

type testWriter struct {
	*testing.T
}

func (t *testWriter) Write(buf []byte) (int, error) {
	t.Logf("%s", buf)
	return len(buf), nil
}