# RSAGEN

Утилита для генерации публичного и приватного ключей ассиметричного шифрования.

## Требования к окружению

- [go](https://go.dev/doc/install)
- [make](https://www.gnu.org/software/make/manual/make.html)
- [godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc)

## Как собрать

```sh
make build-rsagen
```

Собраный файл `rsagen` будет находится в подкаталоге `cmd` основного каталога `rsagen`

## Как запустить тесты

```sh
make tests-rsagen
```

## Подсказка по программе

```sh
rsagen --help
```
