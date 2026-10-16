CHART_REPOSITORY := https://victoriametrics.github.io/helm-charts/
VICTORIALOGS_CHART_VERSION := 0.13.10
COLLECTOR_CHART_VERSION := 0.3.8

.PHONY: install render open-ui

install:
	helm upgrade --install victoria-logs victoria-logs-single \
		--repo $(CHART_REPOSITORY) --version $(VICTORIALOGS_CHART_VERSION) \
		--kube-context minikube --namespace logging --create-namespace \
		--values victorialogs-values.yaml --wait --timeout 5m
	helm upgrade --install victoria-logs-collector victoria-logs-collector \
		--repo $(CHART_REPOSITORY) --version $(COLLECTOR_CHART_VERSION) \
		--kube-context minikube --namespace logging \
		--values collector-values.yaml --wait --timeout 5m

render:
	helm template victoria-logs victoria-logs-single \
		--repo $(CHART_REPOSITORY) --version $(VICTORIALOGS_CHART_VERSION) \
		--namespace logging --values victorialogs-values.yaml
	helm template victoria-logs-collector victoria-logs-collector \
		--repo $(CHART_REPOSITORY) --version $(COLLECTOR_CHART_VERSION) \
		--namespace logging --values collector-values.yaml

open-ui:
	kubectl --context=minikube --namespace logging \
		port-forward service/victoria-logs 9428:9428
