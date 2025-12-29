# Настройка Google Sheets API

Есть два способа аутентификации: через **Service Account** (рекомендуется) или через **API Key**.

## Способ 1: Service Account (Рекомендуется)

Service Account позволяет работать с приватными документами и обеспечивает более надежную аутентификацию.

### Создание Service Account:

1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект или выберите существующий
3. Включите Google Sheets API:
   - Перейдите в "APIs & Services" > "Library"
   - Найдите "Google Sheets API"
   - Нажмите "Enable"
4. Создайте Service Account:
   - Перейдите в "APIs & Services" > "Credentials"
   - Нажмите "Create Credentials" > "Service Account"
   - Заполните название и описание
   - Нажмите "Create and Continue"
   - Роль можно оставить пустой или выбрать "Viewer"
   - Нажмите "Done"
5. Создайте ключ для Service Account:
   - Нажмите на созданный Service Account
   - Перейдите на вкладку "Keys"
   - Нажмите "Add Key" > "Create new key"
   - Выберите тип "JSON"
   - Скачайте файл (например, `service-account-file.json`)

### Настройка доступа к документу:

1. Откройте ваш Google Sheets документ
2. Нажмите "Share" (Поделиться)
3. Скопируйте email Service Account из скачанного JSON файла (поле `client_email`)
4. ⚠️ **ВАЖНО:** Добавьте этот email с правами **"Editor" (Редактор)** - это необходимо для автоматического обновления данных в таблице
   - Если дать только "Viewer", программа сможет только читать, но не записывать данные

### Настройка .env файла:

```env
# Google Sheets API (Service Account)
GOOGLE_SERVICE_ACCOUNT_FILE=service-account-file.json
GOOGLE_SHEET_ID=1zDO5I9ZWnT9AbD--RT9NZX3aQgem6d1FEleq0ISsElk
GOOGLE_SHEET_RANGE=
```

**Примечание:** `GOOGLE_SHEET_RANGE` может быть пустым - в этом случае будет прочитан первый лист полностью.

Примеры значений для `GOOGLE_SHEET_RANGE`:
- Пусто или не указано - читает первый лист полностью
- `Лист1` - читает весь лист с названием "Лист1"
- `Sheet1!A1:C10` - читает диапазон A1:C10 на листе Sheet1
- `Sheet1!A:C` - читает колонки A, B, C на листе Sheet1

Поместите файл `service-account-file.json` в корень проекта.

---

## Способ 2: API Key (Простой, только для публичных документов)

### Создание API Key:

1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект или выберите существующий
3. Включите Google Sheets API (см. выше)
4. Создайте API ключ:
   - Перейдите в "APIs & Services" > "Credentials"
   - Нажмите "Create Credentials" > "API key"
   - Скопируйте созданный API ключ
5. Настройте ограничения (рекомендуется):
   - Ограничьте использование только для Google Sheets API

### Настройка .env файла:

```env
# Google Sheets API (API Key)
GOOGLE_API_KEY=your_api_key_here
GOOGLE_SERVICE_ACCOUNT_FILE=
GOOGLE_SHEET_ID=1zDO5I9ZWnT9AbD--RT9NZX3aQgem6d1FEleq0ISsElk
GOOGLE_SHEET_RANGE=
```

**Важно:** Документ должен иметь права "Anyone with the link can view"

---

## Установка зависимостей

После настройки выполните:

```bash
go mod tidy
```

## Запуск

```bash
go run ./cmd/cli/... process
```

## Безопасность

1. Добавьте `.env` и `service-account-file.json` в `.gitignore`
2. Никогда не коммитьте credentials в репозиторий
3. Для production используйте переменные окружения или secrets manager

