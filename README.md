# Summary
Реализовано:
- round robin и weighted round robin.
- token bucket.

# Запуск
Конфигурируется через json конфиг. Пример:
```json
{
  "services": {
    "ponger_service": {
      "server_name": "http://localhost:8080",
      "algorithm": "weighted_round_robin",
      "servers": [
        { "url": "http://ponger_1:8080", "weight": 1 },
        { "url": "http://ponger_2:8080", "weight": 2 },
        { "url": "http://ponger_3:8080", "weight": 3 }
      ]
    }
  }
}
```
Поднимет балансировщик с лимитером и 3 заглушками понгерами.
```shell
docker compose up
```
Чек
```shell
curl localhost:8080
```
