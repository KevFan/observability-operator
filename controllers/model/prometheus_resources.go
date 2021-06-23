package model

import (
	"bytes"
	"fmt"
	"strings"
	t "text/template"

	v1 "github.com/bf2fc6cc711aee1a0c2a/observability-operator/v3/api/v1"
	routev1 "github.com/openshift/api/route/v1"
	coreosv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v13 "k8s.io/api/core/v1"
	v14 "k8s.io/api/rbac/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultPrometheusLabelSelectors = map[string]string{"app": "strimzi"}

func GetPrometheusAuthTokenLifetimes(cr *v1.Observability) *v13.ConfigMap {
	return &v13.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      "observatorium-token-lifetimes",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusOperatorgroup(cr *v1.Observability) *coreosv1.OperatorGroup {
	return &coreosv1.OperatorGroup{
		ObjectMeta: v12.ObjectMeta{
			Name:      "observability-operatorgroup",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusSubscription(cr *v1.Observability) *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		ObjectMeta: v12.ObjectMeta{
			Name:      "prometheus-subscription",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusCatalogSource(cr *v1.Observability) *v1alpha1.CatalogSource {
	return &v1alpha1.CatalogSource{
		ObjectMeta: v12.ObjectMeta{
			Name:      "prometheus-catalogsource",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusProxySecret(cr *v1.Observability) *v13.Secret {
	return &v13.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      "prometheus-proxy",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusTLSSecret(cr *v1.Observability) *v13.Secret {
	return &v13.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      "prometheus-k8s-tls",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusServiceAccount(cr *v1.Observability) *v13.ServiceAccount {
	route := GetPrometheusRoute(cr)
	redirect := fmt.Sprintf("{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"%s\"}}", route.Name)

	return &v13.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:      "kafka-prometheus",
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.primary": redirect,
			},
		},
	}
}

func GetPrometheusService(cr *v1.Observability) *v13.Service {
	return &v13.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:      "kafka-prometheus",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusClusterRole() *v14.ClusterRole {
	return &v14.ClusterRole{
		ObjectMeta: v12.ObjectMeta{
			Name: "kafka-prometheus",
		},
	}
}

func GetPrometheusClusterRoleBinding() *v14.ClusterRoleBinding {
	return &v14.ClusterRoleBinding{
		ObjectMeta: v12.ObjectMeta{
			Name: "kafka-prometheus",
		},
	}
}

func GetPrometheusRoute(cr *v1.Observability) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: v12.ObjectMeta{
			Name:      "kafka-prometheus",
			Namespace: cr.Namespace,
		},
	}
}

func GetFederationConfig(user, pass string, patterns []string) ([]byte, error) {
	const config = `
- job_name: openshift-monitoring-federation
  honor_labels: true
  kubernetes_sd_configs:
    - role: service
      namespaces:
        names:
          - openshift-monitoring
  scrape_interval: 120s
  scrape_timeout: 60s
  metrics_path: /federate
  relabel_configs:
    - action: keep
      source_labels: [ '__meta_kubernetes_service_name' ]
      regex: prometheus-k8s
    - action: keep
      source_labels: [ '__meta_kubernetes_service_port_name' ]
      regex: web
  params:
    match[]: [{{ .Patterns }}]
  scheme: https
  tls_config:
    insecure_skip_verify: true
  basic_auth:
    username: {{ .User }}
    password: {{ .Pass }}
`

	template := t.Must(t.New("template").Parse(config))
	var buffer bytes.Buffer
	err := template.Execute(&buffer, struct {
		User     string
		Pass     string
		Patterns string
	}{
		User:     user,
		Pass:     pass,
		Patterns: strings.Join(patterns, ","),
	})

	return buffer.Bytes(), err
}

func GetPrometheusAdditionalScrapeConfig(cr *v1.Observability) *v13.Secret {
	return &v13.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      "additional-scrape-configs",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheus(cr *v1.Observability) *prometheusv1.Prometheus {
	return &prometheusv1.Prometheus{
		ObjectMeta: v12.ObjectMeta{
			Name:      "kafka-prometheus",
			Namespace: cr.Namespace,
		},
	}
}

func GetPrometheusPodMonitorLabelSelectors(indexes []v1.RepositoryIndex) map[string]string {
	prometheusConfig := getPrometheusRepositoryIndexConfig(indexes)
	if prometheusConfig != nil && prometheusConfig.PodMonitorLabelSelector != nil {
		return prometheusConfig.PodMonitorLabelSelector
	}

	return defaultPrometheusLabelSelectors
}

func GetPrometheusServiceMonitorLabelSelectors(indexes []v1.RepositoryIndex) map[string]string {
	prometheusConfig := getPrometheusRepositoryIndexConfig(indexes)
	if prometheusConfig != nil && prometheusConfig.ServiceMonitorLabelSelector != nil {
		return prometheusConfig.ServiceMonitorLabelSelector
	}

	return defaultPrometheusLabelSelectors
}

func GetPrometheusRuleMonitorLabelSelectors(indexes []v1.RepositoryIndex) map[string]string {
	prometheusConfig := getPrometheusRepositoryIndexConfig(indexes)
	if prometheusConfig != nil && prometheusConfig.RuleLabelSelector != nil {
		return prometheusConfig.RuleLabelSelector
	}

	return defaultPrometheusLabelSelectors
}

// returns the Prometheus configuration from the repository index
func getPrometheusRepositoryIndexConfig(indexes []v1.RepositoryIndex) *v1.PrometheusIndex {
	if len(indexes) > 0 {
		// We should only have one Prometheus CR for the whole cluster. However, we cannot merge
		// all of the label selectors from all of the repository index config as this will result
		// in an AND requirement. Since we do not use multiple repositories on the same cluster just yet,
		// there should only be one index available in the repository index list.
		// This needs to be changed once we start using multiple repository configurations on the same cluster.
		if indexes[0].Config != nil {
			return indexes[0].Config.Prometheus
		}
	}
	return &v1.PrometheusIndex{}
}
