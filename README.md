# bot-zanuda-webserver

## Локальный запуск сервиса

Команды выполняются из корня webserver. Требуются Go (версия из go.mod),
Linux, Docker Compose и Python с PyYAML.

```sh
git submodule update --init --recursive
make up
# Настройте свой config.yaml по testkit/configs/webserver/config-host.yaml.
make run
```

`make up` — алиас `make containers-up`: поднимает PostgreSQL, Redis и RabbitMQ,
подготавливает схему БД. Сам webserver запускается отдельным процессом приведённой
выше командой или из IDE. Пресет config-host.yaml соответствует этому окружению.

`make run` сам собирает и запускает приложение: отдельный `make build` не нужен.
Он читает `./config.yaml` из корня webserver. Для окружения `make up` укажите
PostgreSQL `localhost:5432`, пользователя/пароль `testkit`, БД `webserver`;
настройки RabbitMQ и JWT также должны соответствовать пресету
`testkit/configs/webserver/config-host.yaml`. Существующий конфиг автоматически
не заменяется. Если нужно передать другой файл без изменения config.yaml:

```sh
make build
.build/bin/app -config testkit/configs/webserver/config-host.yaml
```

Для запуска из IDE укажите рабочий каталог webserver, пакет `./cmd/app`, аргумент
`-config testkit/configs/webserver/config-host.yaml` (либо свой YAML) и те же
Version/BuildDate/GitCommit ldflags, что в `make build`: приложение проверяет
build-метаданные. Пользовательские конфиги не перезаписываются.

После работы остановите процесс сервиса (Ctrl+C или Stop в IDE), затем выполните:

```sh
make down
```

`make down` — алиас `make containers-down`, он убирает окружение зависимостей.
Порты локального окружения: 5432/6379/5672, сервис слушает 8080. Другой стенд
с занятыми портами нужно предварительно остановить.

## Автотесты локального сервиса

Когда зависимости и сервис запущены, в другом терминале из webserver:

```sh
make integration-run WEBSERVER_MODE=host TEST_MODE=host PYTEST_CONFIG="$PWD/testkit/configs/pytests/webserver-host.yaml"
# Тот же сервис, pytest в контейнере:
make integration-run WEBSERVER_MODE=host TEST_MODE=container PYTEST_CONFIG="$PWD/testkit/configs/pytests/webserver-host.yaml"
```

`PYTEST_CONFIG` здесь выбирает готовый пресет для окружения testkit. Если
`pytests/config.yaml` уже настроен под ваш webserver, параметр можно опустить:
при `WEBSERVER_MODE=host` этот файл используется по умолчанию из `PYTESTS_DIR`.
Тесты не останавливают процесс сервиса и его зависимости. Для удаления jobs
контейнерного pytest используйте `make integration-down WEBSERVER_MODE=host`.
Перед переходом к контейнерному сервису удалите эти jobs, остановите локальный
процесс и выполните `make down`.

При смене `WEBSERVER_MODE` сначала выполните `make integration-down` для
предыдущего тестового окружения. Смена только `TEST_MODE` между host/container
не требует уборки. `make up/down` управляют зависимостями IDE, а
`make integration-*` — тестовым окружением; это разные владельцы ресурсов.
Не задавайте им один и тот же `RUN_ID` или `COMPOSE_PROJECT_NAME`. Для обычной
локальной работы оставляйте эти переменные незаданными — runner выберет разные
имена автоматически.

## Автотесты webserver (BZ-206)

Общий lifecycle находится в закреплённом testkit. Требуются Linux, Docker Compose,
Python с PyYAML; для host pytest — Python 3.13+ и зависимости pytests из uv.lock.

```sh
git submodule update --init --recursive
make integration-run
# Контейнерный сервис, pytest на хосте:
make integration-run TEST_MODE=host
```

По умолчанию PYTESTS_DIR — соседний оригинальный pytests. TESTKIT_DIR/PYTESTS_DIR
можно переопределить абсолютными путями для совместной разработки в оригинальных
репозиториях. WEBSERVER_IMAGE выбирает сервис (default latest); локальный sha256 ID
не скачивается. TEST_MODE и WEBSERVER_MODE поддерживают все четыре host/container
сочетания. Integration up/seed/test/logs/down доступны отдельно; up не запускает тесты.

CI `.github/workflows/pytests.yaml` работает на PR в main/develop, вручную и как reusable
workflow. Он собирает текущий коммит и проверяет опубликованными pytests (latest либо
autotests_tag). Проверочный образ не публикуется, существующая публикация независима.
Нужны Secrets DOCKER_USERNAME, DOCKER_REPO_NAME, PYTESTS_IMAGE_REPOSITORY и токен
CI_GITHUB_TOKEN с read-доступом к private testkit/docker; fork PR пропускается до
checkout. Пустые настройки дают ошибку. CI использует уникальный RUN_ID, выводит
логи/сводку и пытается выполнить cleanup при любом исходе.

Поставка: сначала сделайте доступными исправленный docker и коммит testkit, затем
совместимый pytests-образ с test-webserver и seed_webserver. Первый CI проверьте с
явным опубликованным тегом; только после этого используйте latest. Недоступные SHA
и локальные TESTKIT_DIR overrides не подтверждают готовность удалённого CI.

Checkout webserver выполняется штатным GITHUB_TOKEN. Отдельный CI_GITHUB_TOKEN
используется только для закреплённого testkit и его вложенного docker; достаточно
Contents: Read на эти два репозитория, доступ к webserver ему не требуется.

Для проверки PR до обновления latest можно задать Actions variable AUTOTESTS_TAG
(например, bz-206). Приоритет: явный input autotests_tag → переменная репозитория
AUTOTESTS_TAG → latest. После поставки совместимого latest переменную можно удалить.
