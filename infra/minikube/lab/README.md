# Выкатки Bookshop в Minikube

Для опытов нужны запущенные Bookshop, Envoy Gateway, vmagent, VictoriaMetrics и Grafana. Команды ниже выполняются из корня репозитория. Два учебных Deployment используют существующий Service backend и общую базу. Версии отличаются идентификатором сборки, поэтому сравнение маршрутов можно начать на одинаковом коде.

## Подготовить версии

```sh
docker build -f frontend/deploy/Dockerfile --build-arg BOOKSHOP_VERSION=v1 -t bookshop-frontend:v1 frontend
docker build -f frontend/deploy/Dockerfile --build-arg BOOKSHOP_VERSION=v2 -t bookshop-frontend:v2 frontend
docker save bookshop-frontend:v1 bookshop-frontend:v2 | minikube -p minikube image load -

kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/workloads
kubectl --context=minikube -n bookshop rollout status deployment/frontend-blue --timeout=120s
kubectl --context=minikube -n bookshop rollout status deployment/frontend-green --timeout=120s
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/blue.yaml
kubectl --context=minikube -n bookshop get pods -L app.kubernetes.io/version
```

Варианты в каталоге `routes` описывают один объект `HTTPRoute/bookshop`. Применяйте отдельный файл. После смены маршрута выполните `kubectl --context=minikube -n bookshop describe httproute bookshop` и проверьте `Accepted=True` и `ResolvedRefs=True` для текущего поколения. Затем проверьте, что через Envoy отвечает ожидаемая версия.

Откройте шлюз в отдельном терминале:

```sh
proxy_service=$(kubectl --context=minikube -n envoy-gateway-system get service \
  -l gateway.envoyproxy.io/owning-gateway-namespace=bookshop,gateway.envoyproxy.io/owning-gateway-name=bookshop \
  -o name)
kubectl --context=minikube -n envoy-gateway-system port-forward "$proxy_service" 8088:80
```

Проверьте, что через шлюз приходит HTTP 200 и `X-Bookshop-Version: v1`:

```sh
curl -fsS -D - -o /dev/null http://127.0.0.1:8088/books
```

## Наблюдать трафик

Импортируйте в Grafana файл `docs/monitoring/bookshop-rollouts-dashboard.json` и выберите существующий источник VictoriaMetrics. Дашборд показывает запросы в секунду и долю принятых запросов по версии. Дашборд учитывает все `GET /books` на Blue и Green, включая открытия каталога в браузере. Kubernetes проверяет готовность через `/readyz`. Blackbox проверяет основной frontend, который находится за пределами дашборда лаборатории.

В отдельном терминале оставьте этот цикл работать во время переключений. Каждая строка показывает HTTP-статус и версию, которая ответила клиенту. На каждой ступени держите нагрузку минимум минуту, чтобы окно расчёта в Grafana успело обновиться. Для остановки нажмите Ctrl+C.

```sh
while true; do
  curl -sS --max-time 3 -o /dev/null -w '%{http_code} %header{x-bookshop-version}\n' \
    -H "X-Bookshop-Cohort: ${COHORT:-default}" \
    http://127.0.0.1:8088/books
  sleep 0.2
done
```

## Rolling

Маршрут `blue.yaml` отправляет весь трафик в один Service. Обновите образ и метку Pod одним патчем, чтобы Kubernetes создал одну новую ревизию Deployment:

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/blue.yaml
kubectl --context=minikube -n bookshop patch deployment frontend-blue --type=strategic \
  -p '{"spec":{"template":{"metadata":{"labels":{"app.kubernetes.io/version":"v2"}},"spec":{"containers":[{"name":"frontend","image":"bookshop-frontend:v2"}]}}}}'
kubectl --context=minikube -n bookshop rollout status deployment/frontend-blue --timeout=120s
```

Пока сменяются реплики, ожидаются ответы обеих версий. После завершения ожидается `v2`. Перед остальными опытами верните Blue на `v1`:

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/workloads/frontend-blue-deployment.yaml
kubectl --context=minikube -n bookshop rollout status deployment/frontend-blue --timeout=120s
```

## Blue/Green и Canary

Blue/Green переводит весь поток на Green, затем возвращает его на Blue. Сверьте версию ответа после каждого переключения.

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/green.yaml
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/blue.yaml
```

Для Canary выполняйте команды по одной. Дождитесь, пока каждая ступень станет видна на графике, прежде чем менять веса. Ожидаемые доли Blue/Green — примерно 90/10, 50/50 и затем 0/100.

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/canary-10.yaml
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/canary-50.yaml
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/green.yaml
```

## A/B

Обычный запрос должен получить `v1`, запрос с заголовком группы beta — `v2`. Чтобы увидеть обе версии на графике, оставьте обычный цикл включённым и запустите такой же во втором терминале, предварительно выполнив `COHORT=beta`.

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/ab.yaml
curl -fsS -D - -o /dev/null http://127.0.0.1:8088/books
curl -fsS -D - -o /dev/null \
  -H 'X-Bookshop-Cohort: beta' http://127.0.0.1:8088/books
```

## Shadow

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/lab/routes/shadow.yaml
```

Оставьте обычный цикл запросов включённым. В терминале отвечает `v1`. В Grafana растут счётчики обеих версий, потому что Green получает копию каждого `GET /books`. Другие методы и пути идут только в Blue. Доли на этом дашборде считаются от принятых запросов вместе с копиями.

## Вернуть исходный маршрут

Сначала верните обычный Service frontend в маршрут и проверьте HTTP 200 через Envoy. Затем можно удалить учебные Deployment и Service:

```sh
kubectl --context=minikube -n bookshop apply -f infra/minikube/gateway/httproute.yaml
curl -fsS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8088/books
kubectl --context=minikube -n bookshop delete -f infra/minikube/lab/workloads
```
