CHART_REPOSITORY := https://victoriametrics.github.io/helm-charts/
VICTORIAMETRICS_CHART_VERSION := 0.49.0
VMAGENT_CHART_VERSION := 0.49.0
VMALERT_CHART_VERSION := 0.52.1
VMALERT_TOOL_IMAGE := victoriametrics/vmalert-tool:v1.155.0

.PHONY: install render test-rules open-ui open-agent open-alert

install:
	helm upgrade --install vmsingle victoria-metrics-single \
		--repo $(CHART_REPOSITORY) --version $(VICTORIAMETRICS_CHART_VERSION) \
		--kube-context minikube --namespace metrics --create-namespace \
		--values victoriametrics-values.yaml --wait --timeout 5m
	kubectl --context=minikube apply -f blackbox
	kubectl --context=minikube --namespace metrics \
		rollout restart deployment/blackbox-exporter
	kubectl --context=minikube --namespace metrics \
		rollout status deployment/blackbox-exporter --timeout=120s
	helm upgrade --install vmagent victoria-metrics-agent \
		--repo $(CHART_REPOSITORY) --version $(VMAGENT_CHART_VERSION) \
		--kube-context minikube --namespace metrics \
		--values vmagent-values.yaml --wait --timeout 5m
	kubectl --context=minikube --namespace metrics create configmap vmalert-rules \
		--from-file=alert-rules.yaml=rules/alert-rules.yaml --dry-run=client -o yaml \
		| kubectl --context=minikube apply -f -
	helm upgrade --install vmalert victoria-metrics-alert \
		--repo $(CHART_REPOSITORY) --version $(VMALERT_CHART_VERSION) \
		--kube-context minikube --namespace metrics \
		--values vmalert-values.yaml --wait --timeout 5m
	kubectl --context=minikube --namespace metrics \
		rollout restart deployment/vmalert-victoria-metrics-alert-server
	kubectl --context=minikube --namespace metrics \
		rollout status deployment/vmalert-victoria-metrics-alert-server --timeout=120s

render:
	@for manifest in blackbox/*.yaml; do cat "$$manifest"; printf '\n---\n'; done
	helm template vmsingle victoria-metrics-single \
		--repo $(CHART_REPOSITORY) --version $(VICTORIAMETRICS_CHART_VERSION) \
		--namespace metrics --values victoriametrics-values.yaml
	helm template vmagent victoria-metrics-agent \
		--repo $(CHART_REPOSITORY) --version $(VMAGENT_CHART_VERSION) \
		--namespace metrics --values vmagent-values.yaml
	helm template vmalert victoria-metrics-alert \
		--repo $(CHART_REPOSITORY) --version $(VMALERT_CHART_VERSION) \
		--namespace metrics --values vmalert-values.yaml

test-rules:
	docker run --rm --volume "$(CURDIR)/rules:/rules" --workdir /rules \
		$(VMALERT_TOOL_IMAGE) unittest --files=alert-rules-test.yaml

open-ui:
	kubectl --context=minikube --namespace metrics \
		port-forward service/vmsingle-victoria-metrics-single-server 8428:8428

open-agent:
	kubectl --context=minikube --namespace metrics \
		port-forward deployment/vmagent-victoria-metrics-agent 8429:8429

open-alert:
	kubectl --context=minikube --namespace metrics \
		port-forward service/vmalert-victoria-metrics-alert-server 8880:8880
