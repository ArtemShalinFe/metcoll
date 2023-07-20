# Metcoll

- ![Made](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)
- [![codecov](https://codecov.io/gh/ArtemShalinFe/metcoll/branch/main/graph/badge.svg?token=1H84IB1DO1)](https://codecov.io/gh/ArtemShalinFe/gophermart) 
- [![Go Report Card](https://goreportcard.com/badge/github.com/ArtemShalinFe/metcoll)](https://goreportcard.com/report/github.com/ArtemShalinFe/metcoll) 
- [![codebeat badge](https://codebeat.co/badges/43f20fd4-6625-41d5-b9be-56d75b5bfee6)](https://codebeat.co/projects/github-com-artemshalinfe-metcoll-main)
- 
Инкрементальный проекта курса «Go-разработчик» трека "Сервис сбора метрик"

## Требования к окружению

- [go](https://go.dev/doc/install)
- [make](https://www.gnu.org/software/make/manual/make.html)

## Как собрать

### Сборка сервиса metcoll-agent

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. Из каталога репозитория выполните команду

```sh
make build-agent
```

3. Собраный файл `agent` будет находится в подкаталоге репозитория `./cmd/agent/agent`

### Сборка сервиса metcoll-server

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. Из каталога репозитория выполните команду

```sh
make build-server
```

3. Собраный файл `server` будет находится в подкаталоге репозитория `./cmd/server/server`

## Запуск тестов

```sh
go test ./... -v -race
```
