# Summary
Реализовано:
- round robin и weighted round robin.
- token bucket.
- health check

# Запуск
Конфигурируется через json конфиг. Пример:
```json
{
  "balancer": {
    "type": "round_robin",
    "servers": [
      { "url": "http://localhost:8081", "weight": 3 },
      { "url": "http://localhost:8082", "weight": 2 },
      { "url": "http://localhost:8083", "weight": 1 }
    ]
  },
  "limiter": {
    "type": "token_buket",
    "users": [
      { "token": "vasya", "limit":  1.5 },
      { "token": "petya", "limit":  10 }
    ]
  },
  "active_health_check": {
    "handler": "/ping",
    "method": "GET",
    "interval": 10,
    "timeout": 100
  }
}
```
Поднимет балансировщик с лимитером и 3 заглушками понгерами.
# Запуск в docker compose
```shell
docker compose up
```
Чек
```shell
curl -H "ft-header: vasya" localhost:8080
```
