# Правила именования монет

## Формат названия монеты в таблице

В колонке **"Монета"** нужно указывать **ТОЛЬКО символ монеты**, без USDT или других пар!

### ✅ Правильно:
```
BTC
ETH
XVG
DOGE
SOL
```

### ❌ Неправильно:
```
BTCUSDT    ❌ (убрать USDT)
BTC/USDT   ❌ (убрать /USDT)
XVG-USDT   ❌ (убрать -USDT)
bitcoin    ❌ (использовать символ, а не полное название)
```

## Как узнать правильный символ

### Вариант 1: На Bybit
1. Найдите монету на Bybit
2. Торговая пара: **XVGUSDT**
3. Символ монеты: **XVG** ← используйте это в таблице

### Вариант 2: На CoinGecko
1. Найдите монету на CoinGecko
2. Например: https://www.coingecko.com/en/coins/verge/usd
3. В URL видно ID: **verge**
4. На странице в шапке: **Verge** (XVG)
5. Символ: **XVG** ← используйте это в таблице

## Примеры популярных монет

| Bybit пара | Символ для таблицы | CoinGecko ID |
|-----------|-------------------|--------------|
| BTCUSDT   | BTC               | bitcoin      |
| ETHUSDT   | ETH               | ethereum     |
| XVGUSDT   | XVG               | verge        |
| DOGEUSDT  | DOGE              | dogecoin     |
| SOLUSDT   | SOL               | solana       |
| SHIBUSDT  | SHIB              | shiba-inu    |
| MATICUSDT | MATIC             | matic-network|
| LINKUSDT  | LINK              | chainlink    |

## Поддерживаемые монеты

### Автоматически поддерживаются:

**Major (топ-20):**
- BTC, ETH, USDT, BNB, SOL, XRP, USDC, ADA, AVAX, DOGE
- DOT, MATIC, LINK, UNI, LTC, ATOM, ETC, XLM, BCH

**DeFi токены:**
- LINK, UNI, AAVE, SUSHI, LDO, IMX

**Layer 2 & Scaling:**
- MATIC, ARB, OP

**Новые проекты:**
- SUI, SEI, TIA, APT, INJ, NEAR, ALGO, VET, ICP, FIL, HBAR

**Meme & Others:**
- DOGE, SHIB, XVG, TRX, TON, OKB, LEO, DAI, WBTC

### Как добавить новую монету?

Если вашей монеты нет в списке:

1. **Найдите на CoinGecko:**
   - Зайдите на https://www.coingecko.com/
   - Найдите вашу монету
   - Посмотрите ID в URL (например: `/coins/verge` → ID: `verge`)
   - Посмотрите символ на странице (например: XVG)

2. **Запишите в таблицу символ:** XVG

3. **Программа автоматически попробует найти:**
   - Сначала проверит внутренний маппинг
   - Если не найдет, попробует использовать символ как ID
   - В большинстве случаев сработает автоматически

4. **Если не работает, создайте Issue на GitHub** с информацией:
   ```
   Символ: XVG
   CoinGecko URL: https://www.coingecko.com/en/coins/verge
   CoinGecko ID: verge
   ```

## Регистр букв

**Не имеет значения!** Программа автоматически приводит к нижнему регистру:
- `BTC` = `btc` = `Btc` ✅
- `XVG` = `xvg` = `Xvg` ✅

## Частые ошибки

### 1. Указана торговая пара вместо символа
```
❌ Написано в таблице: XVGUSDT
✅ Должно быть: XVG
```

### 2. Используется название вместо символа
```
❌ Написано в таблице: Bitcoin
✅ Должно быть: BTC
```

### 3. Используется CoinGecko ID вместо символа
```
❌ Написано в таблице: verge
✅ Должно быть: XVG
```

## Проверка правильности

После запуска программы проверьте логи:

### ✅ Успешно:
```
Record 1 (XVG): Missing Bybit price, fetching from CoinGecko...
  ✅ Updated Bybit price: $0.00523
```

### ❌ Ошибка - монета не найдена:
```
Record 1 (XVGUSDT): failed to get price: price not found for coin: XVGUSDT
```
**Решение:** Уберите USDT, оставьте только XVG

### ⚠️ Rate Limit:
```
Record 1 (XVG): failed to get price: CoinGecko rate limit exceeded
```
**Решение:** Подождите 1-2 минуты и запустите снова. Программа автоматически добавит задержки.

## Тестирование новой монеты

Перед добавлением большого количества данных протестируйте:

1. Добавьте одну строку с новой монетой
2. Запустите `go run ./cmd/cli/... process`
3. Проверьте в логах, что цена получена успешно
4. Если успешно - добавляйте остальные данные

