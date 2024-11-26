# My Application: GetRates

## Описание
**GetRates** — это приложение для получения и обработки данных курса USDT с биржи **Garantex**. Приложение выводит информацию о **ask** и **bid** ценах, а также метку времени получения курса.


## Вы можете передать параметры при запуске как в слудующем примере
go run cmd/main.go -host=localhost -port=5432 -user=postgres -password=secret -dbname=mydb


### Параметры командной строки:
- `-host` — хост базы данных
- `-port` — порт базы данных 
- `-user` — пользователь базы данных.
- `-password` — пароль пользователя базы данных.
- `-dbname` — имя базы данных.

## Запуск тестов
make test

## Для проверки кода на соответствие стандартам используйте:
make lint

##  Для сборки Docker-образа используйте команду:
make docker-build

## Приложение можно запустить с использованием Docker. Для этого выполните следующие шаги:
```bash
git clone https://github.com/ilyavasilenko227/GetRates
cd GetRates
docker compose up -d
docker compose run --rm rates /main