# go-dep-analyzer

CLI-утилита для анализа зависимостей Go-репозитория.

## Что делает

Принимает адрес Git-репозитория и возвращает:
- Имя Go-модуля
- Версию Go
- Список прямых зависимостей, доступных для обновления

## Установка

Требования: Go 1.21+

```bash
git clone https://github.com/yourusername/go-dep-analyzer
cd go-dep-analyzer
go build -o go-dep-analyzer .
```

## Использование

```bash
./go-dep-analyzer <repository-url>
```

### Пример

```bash
./go-dep-analyzer https://github.com/spf13/cobra
```

### Пример вывода
Cloning https://github.com/spf13/cobra ...
Module:     github.com/spf13/cobra
Go version: 1.15
Dependencies available for update:
github.com/cpuguy83/go-md2man/v2                             v2.0.6  →  v2.0.7
github.com/spf13/pflag                                       v1.0.9  →  v1.0.10

## Как это работает

1. Клонирует репозиторий во временную папку (shallow clone)
2. Читает и парсит `go.mod`
3. Для каждой прямой зависимости запрашивает последнюю версию через `proxy.golang.org`
4. Сравнивает версии и выводит те, которые можно обновить
5. Временная папка автоматически удаляется после завершения