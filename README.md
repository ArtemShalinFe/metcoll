# Metcoll

- ![Made](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)
- [![codecov](https://codecov.io/gh/ArtemShalinFe/metcoll/branch/main/graph/badge.svg?token=1H84IB1DO1)](https://codecov.io/gh/ArtemShalinFe/metcoll)
- [![Go Report Card](https://goreportcard.com/badge/github.com/ArtemShalinFe/metcoll)](https://goreportcard.com/report/github.com/ArtemShalinFe/metcoll)

Инкрементальный проекта курса «Go-разработчик» трека "Сервис сбора метрик"

## Требования к окружению

- [go](https://go.dev/doc/install)
- [make](https://www.gnu.org/software/make/manual/make.html)
- [godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc)
- [graphviz](https://graphviz.org)
- [PostgreSQL](https://www.postgresql.org)
- [OpenSSL](https://www.openssl.org)

## Как собрать

### Шифрование

Для включения шифрования необходимо сгенерировать файлы приватного(для сервера) и публичного(для агента) ключа. Это можно сделать командой:

```sh
make cryptokeys
```

Файлы будут сгенерированы в корневой директории проекта в каталоге `keys`

### Сборка сервиса metcoll-agent

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
1. Из каталога репозитория выполните команду

```sh
make build-agent
```

1. Собраный файл `agent` будет находится в подкаталоге репозитория `./cmd/agent/agent`

### Сборка сервиса metcoll-server

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
1. Из каталога репозитория выполните команду

```sh
make build-server
```

1. Собраный файл `server` будет находится в подкаталоге репозитория `./cmd/server/server`

### Сборка статического анализатора

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
1. Из каталога репозитория выполните команду

```sh
make build-staticlint
```

1. Собраный файл `staticlint` будет находится в подкаталоге репозитория `./cmd/staticlint/staticlint`
2. Чтобы запустить статический анализатор необходимо передать исполняемому файлу пути к анализируемым пакетам, например:

```sh
cmd/staticlint/staticlint ПутьКПроверяемомуПакету1 ПутьКПроверяемомуПакету2
```

```sh
cmd/staticlint/staticlint ./internal/storage ./internal/stats
```

Или, например, вот такой командой можно проверить все *.go файлы:

```sh
cmd/staticlint/staticlint ./...
```


## Запуск тестов

```sh
go test ./... -v -race
```

## Профилирование

Для профилирования должен быть развернут PostgreSQL с базой "praktikum" и установлена утилита `graphviz`.

1. Перейдите в корневую директорию проекта и соберите бинарные файлы сервера и агента, возспользовавшись командой:

```sh
make build-agent && make build-server
```

1. Запустите сервер командой:

```sh
./cmd/server/server -d "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable" -a localhost:32323 -k nope   
```

1. В отдельном окне терминала запустите агент командой:

```sh
./cmd/agent/agent -a localhost:32323 -l 5 -k nope -p 1 -r 2
```

> Агент начнет собирать метрики и отправлять их на сервер, таким образом будет генерироваться нагрузка.

1. Чтобы собрать профиль приложения по СPU выполните команду, не закрывая окон с сервером и агентом:

```sh
curl -s -v http://localhost:32323/debug/pprof/profile > profiles/cpu.pprof
```

1. Посмотреть собраный профиль можно утилитой pprof:

```sh
go tool pprof -http=":9090" -seconds=60 profiles/cpu.pprof
```

1. Чтобы собрать профиль приложения по выделяемой памяти выполните команду, не закрывая окон с сервером и агентом:

```sh
curl -s -v http://localhost:32323/debug/pprof/heap > profiles/heap.pprof
```

1. Посмотреть собраный профиль можно утилитой pprof:

```sh
go tool pprof -http=":9090" -seconds=60 profiles/heap.pprof
```

## Локальное отображения godoc-документации

1. Установите пакет `godoc`

```sh
go install -v golang.org/x/tools/cmd/godoc@latest
```

1. Перейдите в корневую директорию проекта и выполнить команду

```sh
godoc -http=:8080 -play   
```

1. Для просмотра откройте в браузере `http://localhost:8080`
