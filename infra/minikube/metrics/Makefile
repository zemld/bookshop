VICTORIAMETRICS_CHART_VERSION := 0.49.0

.PHONY: install render open-ui

install:
	helm repo add vm https://victoriametrics.github.io/helm-charts/
	helm repo update vm
	helm upgrade --install vmsingle vm/victoria-metrics-single \
		--version $(VICTORIAMETRICS_CHART_VERSION) \
		--kube-context minikube --namespace default \
		--values guide-vmsingle-values.yaml --wait --timeout 5m

render:
	helm template vmsingle victoria-metrics-single \
		--repo https://victoriametrics.github.io/helm-charts/ \
		--version $(VICTORIAMETRICS_CHART_VERSION) \
		--namespace default --values guide-vmsingle-values.yaml

open-ui:
	kubectl --context=minikube --namespace default \
		port-forward service/vmsingle-victoria-metrics-single-server 8428:8428
