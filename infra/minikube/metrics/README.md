# Сбор метрик

`vmagent` обнаруживает targets и опрашивает их каждые 15 секунд, затем отправляет
метрики в VictoriaMetrics Single через `/api/v1/write`. Single отвечает только
за приём, хранение и запросы; встроенный scrape отключён.

Оба Helm-релиза устанавливаются в namespace `metrics` контекста `minikube`:

```sh
make -C infra/minikube/metrics render
make -C infra/minikube/metrics install
```

Если Single ранее установлен в `default`, не запускайте установку поверх него
без отдельного плана переноса: Helm-релизы в разных namespaces независимы,
и эта команда не переносит существующие данные.

Агент работает в одном экземпляре. Стратегия `Recreate` исключает одновременный
сбор старым и новым pod'ами при обновлении deployment, ценой короткого перерыва.
Очередь отправки находится в `emptyDir`, без PVC. Она переживает перезапуск
контейнера внутри pod'а, но теряется при замене pod'а. Сохранность очереди при
сбоях и HA-сбор не входят в гарантии этой конфигурации.

## Проверка

В отдельных терминалах:

```sh
make -C infra/minikube/metrics open-agent
make -C infra/minikube/metrics open-ui
```

Состояние targets агента:

```sh
curl --fail --silent --show-error http://localhost:8429/api/v1/targets
```

В одноузловом Minikube ожидаются восемь активных targets с `health: up`
и пустым `lastError`: backend, frontend, postgres-exporter, API server,
kubelet, cAdvisor, VictoriaMetrics и vmagent.

Проверка доставки свежих измерений в Single:

```sh
curl --fail --silent --show-error --get http://localhost:8428/api/v1/query \
  --data-urlencode 'query=up'
curl --fail --silent --show-error --get http://localhost:8428/api/v1/query \
  --data-urlencode 'query=time() - timestamp(up)'
curl --fail --silent --show-error http://localhost:8428/api/v1/targets
```

У восьми текущих targets `up` должен быть равен `1`, а возраст измерений —
меньше 60 секунд. Список активных targets самого Single должен быть пустым.
Сразу после перехода в запросах ещё могут быть видны старые ряды, в частности
`instance="localhost:8428"`; они больше не обновляются.

Для проверки приложений выполните `GET /books` через их HTTP-порты, затем
проверьте увеличение `ogen_server_request_count_total` backend и
`ogen_client_request_count_total` frontend в Single:

```sh
curl --fail --silent --show-error --get http://localhost:8428/api/v1/query \
  --data-urlencode 'query=sum by (__name__) ({job="bookshop-pods",http_route="/books",http_request_method="GET",__name__=~"ogen_(server|client)_request_count_total"})'
```

Дополнительно проверьте `pg_up == 1` и наличие инфраструктурных метрик
`apiserver_request_total`, `kubelet_running_pods`,
`container_cpu_usage_seconds_total`. Одного `up == 1` недостаточно для проверки
содержимого метрик приложения.
