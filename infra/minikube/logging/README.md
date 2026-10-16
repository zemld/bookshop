# VictoriaLogs в minikube

Отдельный namespace `logging` содержит VictoriaLogs single и сборщик `vlagent`
в DaemonSet. Удаление namespace `bookshop` не удаляет хранилище логов.
Код, образы и манифесты приложения не меняются.

```text
stdout/stderr контейнеров bookshop
  → файлы логов на узле minikube
  → victoria-logs-collector в logging
  → Service victoria-logs:9428
  → VictoriaLogs → PVC 2 GiB
```

## Конфигурация

| Файл | Назначение |
|---|---|
| `victorialogs-values.yaml` | Один экземпляр VictoriaLogs, PVC, retention и ресурсы |
| `collector-values.yaml` | Адрес назначения, фильтр namespace, разбор полей и ресурсы сборщика |
| `Makefile` | Установка, генерация манифестов и локальный доступ; фиксированные версии charts |

Используются официальные charts `victoria-logs-single` **0.13.10**
и `victoria-logs-collector` **0.3.8**. Оба задают образ приложения **v1.53.0**.
Оператор, CRD, Grafana и Prometheus не нужны.

Сборщик автоматически разбирает JSON, включая `slog` приложений:
`msg` становится `_msg`, `time` — `_time`, `level` и `error` остаются
отдельными полями. Обычный текст тоже принимается; для него используется время
контейнерного runtime. Kubernetes metadata включает `kubernetes.pod_namespace`,
`kubernetes.pod_name`, `kubernetes.container_name` и метки Pod.

Фильтр исключает все namespace, кроме `bookshop`, до чтения файлов.
Он не является границей доступа: chart создаёт ClusterRole с чтением
`nodes`, `namespaces`, `pods` и ClusterRoleBinding, а DaemonSet монтирует
`/var/log` и `/var/lib` узла только для чтения. Последний путь нужен в том
числе для ссылок на файлы логов в minikube.
Буфер и checkpoints сборщика хранятся отдельно в hostPath
`/var/lib/vl-collector` на каждом узле и переживают замену Pod сборщика.

### Расход диска

- PVC VictoriaLogs запрашивает **2 GiB** у StorageClass по умолчанию.
- Записи хранятся до **3 дней**. Порог удаления старых дневных разделов —
  **1 GiB**, с запасом относительно размера PVC.
- Это не жёсткий лимит: VictoriaLogs сохраняет как минимум последние два дня
  данных, а очистка выполняется периодически. При большом потоке порог может
  быть превышен. Локальный hostPath provisioner minikube также может не
  обеспечивать квоту по размеру PVC.
- Буфер сборщика ограничен **513 MiB на узел** для единственного назначения.
  У vlagent v1.53.0 минимальный размер файловой очереди — 512 MiB + 128 байт:
  меньшие значения округляются вверх и не дают ожидаемого ограничения.
  При заполнении очереди старые записи отбрасываются.
- Контейнерные логи runtime, checkpoints и скачанные образы занимают
  дополнительное место. Эта конфигурация не ограничивает весь расход диска
  minikube до 2 GiB.

Это локальное окружение без резервного копирования и HA. PVC переживает
перезапуск Pod VictoriaLogs, но не гарантирует сохранность при удалении
namespace `logging`, тома или самого minikube.

## Установка и обновление

Нужны Helm, kubectl, работающий minikube с Kubernetes >= 1.25,
доступ к репозиторию charts и registry образов, а также StorageClass
по умолчанию. Для сбора логов приложения должен существовать `bookshop`.

Проверить окружение:

```sh
kubectl --context=minikube get nodes
kubectl --context=minikube get storageclass
kubectl --context=minikube get namespace bookshop
```

Из корня репозитория:

```sh
make -C infra/minikube/logging render
make -C infra/minikube/logging install
```

`render` генерирует манифесты без обращения к Kubernetes API.
`install` создаёт namespace `logging`, устанавливает сначала хранилище,
потом сборщик и ждёт их готовности. Повторный запуск обновляет те же releases.
Все команды установки явно используют контекст `minikube`.

После редактирования values-файлов повторить `install`. Размер существующего
PVC нельзя просто уменьшить; расширение зависит от StorageClass.
Удаление release приложения не затрагивает releases логирования.

Проверить ресурсы:

```sh
helm list --kube-context minikube --namespace logging
kubectl --context=minikube --namespace logging get pods,svc,pvc
kubectl --context=minikube --namespace logging rollout status statefulset/victoria-logs --timeout=120s
kubectl --context=minikube --namespace logging rollout status daemonset/victoria-logs-collector --timeout=120s
```

## Доступ и проверка сбора

```sh
make -C infra/minikube/logging open-ui
```

Оставить терминал открытым и перейти на
<http://localhost:9428/select/vmui/>.
Service имеет тип ClusterIP; Ingress и внешняя аутентификация не настроены.
Port-forward по умолчанию слушает только localhost.

Примеры LogsQL:

```text
{kubernetes.pod_namespace="bookshop"}
{kubernetes.pod_namespace="bookshop",kubernetes.container_name="backend"}
{kubernetes.pod_namespace="bookshop"} level:="ERROR"
```

Для проверки без изменений существующих сервисов можно создать временный Pod
в `bookshop`. Имя должно быть свободно:

```sh
kubectl --context=minikube --namespace bookshop run logging-smoke \
  --image=busybox:1.37.0 --restart=Never \
  --command -- sh -c 'echo "{\"level\":\"INFO\",\"msg\":\"victorialogs-smoke-json\",\"probe\":\"logging-smoke\"}"; echo "victorialogs-smoke-text"; sleep 300'
kubectl --context=minikube --namespace bookshop wait \
  --for=condition=Ready pod/logging-smoke --timeout=120s
kubectl --context=minikube --namespace bookshop logs logging-smoke
```

В другом терминале, пока работает port-forward:

```sh
curl --fail-with-body --get http://localhost:9428/select/logsql/query \
  --data-urlencode 'query={kubernetes.pod_namespace="bookshop",kubernetes.pod_name="logging-smoke"} _time:5m'
```

Дождаться появления обеих записей. У JSON-записи должны быть
`_msg=victorialogs-smoke-json`, `level=INFO`, `probe=logging-smoke`;
у текстовой — сообщение `victorialogs-smoke-text` в `_msg`
(runtime может сохранить завершающий перевод строки).
Обе должны содержать namespace, Pod и container.
Обычный успешный HTTP-запрос к приложению может не писать лог:
сборщик не добавляет access logging.

Чтобы проверить исключение других namespace, создать такой же временный Pod
в `logging`, заменив `--namespace bookshop` на `--namespace logging`.
Его записи не должны появляться в запросе:

```text
{kubernetes.pod_namespace="logging",kubernetes.pod_name="logging-smoke"} _time:5m
```

Проверить сохранность записей после замены Pod VictoriaLogs:

```sh
kubectl --context=minikube --namespace logging rollout restart statefulset/victoria-logs
kubectl --context=minikube --namespace logging rollout status statefulset/victoria-logs --timeout=120s
```

Перезапустить port-forward и повторить запрос, увеличив `_time:5m`, если нужно.
Удалить только созданные для проверки Pod:

```sh
kubectl --context=minikube --namespace bookshop delete pod logging-smoke
kubectl --context=minikube --namespace logging delete pod logging-smoke --ignore-not-found
```

## Остановка

```sh
helm uninstall victoria-logs-collector victoria-logs \
  --kube-context minikube --namespace logging
```

PVC StatefulSet сохраняется после удаления release. Namespace и hostPath
сборщика также остаются. Не удалять их, если данные ещё нужны.
Удаление PVC и очистка hostPath — отдельные необратимые операции и не входят
в команды установки или остановки.

## Документация

- [VictoriaLogs single chart](https://docs.victoriametrics.com/helm/victoria-logs-single/)
- [Collector chart](https://docs.victoriametrics.com/helm/victoria-logs-collector/)
- [Сбор и фильтрация Kubernetes-логов](https://docs.victoriametrics.com/victorialogs/vlagent/#collecting-kubernetes-logs)
- [Ограничения retention по диску](https://docs.victoriametrics.com/victorialogs/#retention-by-disk-space-usage)
