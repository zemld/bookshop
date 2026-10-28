# Мониторинг

Из этой директории:

```sh
make install
make test
make open-ui
```

`install` разворачивает VictoriaMetrics, Blackbox Exporter и vmagent в контексте `minikube`, namespace `metrics`. Приложение должно быть запущено в namespace `bookshop`. `test` требует Python 3 и Docker и проверяет настоящий exporter через HTTP.

## Пробер

vmagent вызывает `/probe` exporter каждые 30 с. Проверяется `GET http://frontend.bookshop.svc:8081/books`: HTTP 200 за ≤ 3 с, без редиректов, повторов и проверки тела. Таймаут сбора — 5 с.

| Метрика | Значение |
|---|---|
| `probe_success{job="bookshop-availability"}` | 1 — успех, 0 — неуспешная проверка |
| `probe_duration_seconds{job="bookshop-availability"}` | Длительность проверки |
| `up{job="bookshop-availability"}` | Удалось получить результат `/probe` |
| `up{job="blackbox-exporter"}` | Доступны служебные метрики exporter |

`up = 1` при `probe_success = 0` означает, что результат получен, но каталог не прошёл проверку. `up = 0` или отсутствие ожидаемых данных — проблема измерений, не успех приложения.

Проверка в кластере:

```sh
kubectl --context=minikube -n metrics rollout status deployment/blackbox-exporter
make open-agent
```

На `http://localhost:8429/targets` должны быть задачи `bookshop-availability` и `blackbox-exporter`. В VictoriaMetrics через `make open-ui` проверить метрики из таблицы. Полный отказ кластера оставляет пробер и сборщик без возможности записывать результаты.
