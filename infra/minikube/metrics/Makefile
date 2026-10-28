CHART_REPOSITORY := https://victoriametrics.github.io/helm-charts/
VICTORIAMETRICS_CHART_VERSION := 0.49.0
VMAGENT_CHART_VERSION := 0.49.0

.PHONY: install render test open-ui open-agent

install:
	helm upgrade --install vmsingle victoria-metrics-single \
		--repo $(CHART_REPOSITORY) --version $(VICTORIAMETRICS_CHART_VERSION) \
		--kube-context minikube --namespace metrics --create-namespace \
		--values victoriametrics-values.yaml --wait --timeout 5m
	kubectl --context=minikube apply -k blackbox
	kubectl --context=minikube --namespace metrics \
		rollout status deployment/blackbox-exporter --timeout=120s
	helm upgrade --install vmagent victoria-metrics-agent \
		--repo $(CHART_REPOSITORY) --version $(VMAGENT_CHART_VERSION) \
		--kube-context minikube --namespace metrics \
		--values vmagent-values.yaml --wait --timeout 5m

render:
	kubectl kustomize blackbox
	helm template vmsingle victoria-metrics-single \
		--repo $(CHART_REPOSITORY) --version $(VICTORIAMETRICS_CHART_VERSION) \
		--namespace metrics --values victoriametrics-values.yaml
	helm template vmagent victoria-metrics-agent \
		--repo $(CHART_REPOSITORY) --version $(VMAGENT_CHART_VERSION) \
		--namespace metrics --values vmagent-values.yaml

test:
	PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s . -p test_blackbox.py

open-ui:
	kubectl --context=minikube --namespace metrics \
		port-forward service/vmsingle-victoria-metrics-single-server 8428:8428

open-agent:
	kubectl --context=minikube --namespace metrics \
		port-forward deployment/vmagent-victoria-metrics-agent 8429:8429
