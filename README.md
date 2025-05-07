# Summary
Реализован round robin. Готов только обычный, но проведен рефакторинг под weighted round robin.
Реализован token bucket.

# Запуск
Конфигурируется через переменные окружения. Пример:
Для `petya` задан рейтлимит на 1.5 токена в секунду, для `vasya` 60.
```shell
export BACKEND_URLS="http://localhost:8081,http://localhost:8082,http://localhost:8083"
export CLIENT_LIMITS="vasya=60,petya=1.5" 
```
Поднимет балансировщик с лимитером и 3 заглушками понгерами.
```shell
docker compose up
```
Чек
```shell
curl -H "ft-token: petya" localhost:8080
```
