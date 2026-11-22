ENVOY_GATEWAY_VERSION := v1.10.0
ENVOY_NAMESPACE := envoy-gateway-system
KUBECTL := kubectl --context=minikube
PROXY_SELECTOR := gateway.envoyproxy.io/owning-gateway-namespace=bookshop,gateway.envoyproxy.io/owning-gateway-name=bookshop
LOCAL_PORT ?= 8088

.PHONY: install open

install:
	helm upgrade --install eg oci://docker.io/envoyproxy/gateway-helm \
		--version $(ENVOY_GATEWAY_VERSION) --kube-context minikube \
		--namespace $(ENVOY_NAMESPACE) --create-namespace --wait --timeout 5m
	$(KUBECTL) apply -f ../namespace.yaml -f gatewayclass.yaml -f envoyproxy.yaml -f gateway.yaml -f httproute.yaml
	$(KUBECTL) wait gatewayclass/bookshop-envoy --for=condition=Accepted --timeout=120s
	$(KUBECTL) --namespace bookshop wait gateway/bookshop --for=condition=Programmed --timeout=120s

open:
	@proxy_service="$$( $(KUBECTL) --namespace $(ENVOY_NAMESPACE) get service \
		--selector '$(PROXY_SELECTOR)' -o name )" && test -n "$$proxy_service" && \
		exec $(KUBECTL) --namespace $(ENVOY_NAMESPACE) port-forward "$$proxy_service" $(LOCAL_PORT):80
