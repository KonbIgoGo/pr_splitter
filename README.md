# avito_test

# Сборка и запуск
- `docker compose up` - поднимет контейнер с postgres и с самим сервисом.

Основные команды Makefile:

- `make generate` – генерирует эндпоинты, моки и собирает проект
- `make test` – запускает юнит- и интеграционные тесты
- `make lint` – запускает линтер (настройки описаны в `.golangci.yaml`)

Для запуска всех шагов разом можно использовать:

- `make all`


# k6 load test

run load test
k6 run ./tank/load-test.js

TOTAL RESULTS

    - `checks_total`.......: 2360    41.859116/s
    - `checks_succeeded`...: 100.00% 2360 out of 2360
    - `checks_failed`......: 0.00%   0 out of 2360

    ✓ team created or already exists (201/400)
    ✓ getReview fallback (200)
    ✓ merge (200/404)
    ✓ setIsActive (200/404)

    HTTP
    http_req_duration..............: avg=4.03ms min=951.68µs med=2.72ms max=28.98ms p(90)=11.53ms p(95)=13.53ms
      { expected_response:true }...: avg=4.73ms min=951.68µs med=3.44ms max=28.98ms p(90)=12.39ms p(95)=13.9ms
    http_req_failed................: 25.00% 590 out of 2360
    http_reqs......................: 2360   41.859116/s

    EXECUTION
    iteration_duration.............: avg=1.01s  min=1s       med=1.01s  max=1.05s   p(90)=1.02s   p(95)=1.02s
    iterations.....................: 590    10.464779/s
    vus............................: 10     min=10          max=10
    vus_max........................: 10     min=10          max=10

# Проблемы
В спецификации openapi ошибка в example в ручке pullRequests/reassign 
old_reviewer_id, когда в схеме old_user_id -> исправлено значение



