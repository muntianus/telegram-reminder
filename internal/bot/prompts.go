package bot

// TelegramDigestPrompts содержит промпты для различных типов дайджестов
// Все промпты адаптированы под Telegram с эмодзи и четкой структурой

const (
	// CryptoDigestPrompt - промпт для криптовалютного дайджеста
	CryptoDigestPrompt = `
🔥 КРИПТО-ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — эксперт по криптовалютам. ОБЯЗАТЕЛЬНО используй веб-поиск для получения НОВОСТЕЙ ИМЕННО СЕГОДНЯШНЕГО ДНЯ.

ВАЖНО: Все новости должны быть от СЕГОДНЯШНЕГО ДНЯ (текущая дата), а не старые новости из 2023 или 2024 года.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

📊 РЫНОЧНЫЕ МЕТРИКИ:
- Текущие цены топ-10 криптовалют с изменением за 24ч
- Общая капитализация рынка и объем торгов
- Доминирование Bitcoin и Ethereum
- Топ-10 движущихся монет (gainers/losers)

🔗 ON-CHAIN АНАЛИЗ:
- Крупные транзакции (whale movements)
- Активность адресов и новых кошельков
- Метрики сети (hash rate, difficulty)
- DeFi протоколы: TVL, доходность, риски

📈 ДЕРИВАТИВЫ И ТОРГОВЛЯ:
- Открытый интерес по фьючерсам
- Ликвидации за 24 часа
- Опционные контракты и волатильность
- Институциональные потоки

🌍 РЕГУЛЯТИВНЫЕ НОВОСТИ:
- Новые законы и постановления
- Действия SEC, CFTC, других регуляторов
- Санкции и ограничения
- Позиции политиков и чиновников

🚀 ЭКОСИСТЕМНЫЕ НОВОСТИ:
- Обновления протоколов и форков
- Новые DeFi протоколы и токены
- NFT проекты и коллекции
- Метавселенные и Web3

💼 ИНСТИТУЦИОНАЛЬНЫЕ НОВОСТИ:
- ETF новости и заявки
- Корпоративные инвестиции
- Банковские услуги для крипто
- Институциональные продукты

🔍 ВЕБ-ПОИСК: Найди НОВОСТИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ по запросам:
- "bitcoin news today [текущая дата]"
- "cryptocurrency news today [текущая дата]" 
- "crypto market news today [текущая дата]"
- "defi news today [текущая дата]"
- "ethereum news today [текущая дата]"
- "crypto regulation news today [текущая дата]"

ПОЛЕЗНЫЕ ССЫЛКИ:
- CoinGecko (https://coingecko.com) - цены и метрики
- CoinMarketCap (https://coinmarketcap.com) - рыночные данные
- Binance (https://binance.com) - торговля
- TradingView (https://tradingview.com) - графики и анализ
- CryptoSlate (https://cryptoslate.com) - новости
- Decrypt (https://decrypt.co) - аналитика
- Cointelegraph (https://cointelegraph.com) - крипто-новости
- The Block (https://theblock.co) - DeFi и институциональные новости

ВАЖНО: 
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #crypto #bitcoin #altcoins #defi #trading #blockchain

Составь детальный криптовалютный дайджест на русском для Telegram с НОВОСТЯМИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ.
`

	// TechDigestPrompt - промпт для технологического дайджеста
	TechDigestPrompt = `
💻 ТЕХ-ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — эксперт по технологиям и стартапам. Используй веб-поиск для получения актуальных новостей о технологиях и стартапах.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

🤖 ИИ И МАШИННОЕ ОБУЧЕНИЕ:
- Новые модели и алгоритмы (названия, характеристики, производительность)
- Крупные инвестиции в AI (суммы, компании, инвесторы)
- Партнерства и слияния в AI секторе
- Регулятивные новости (законы, этические принципы)
- Практические применения (кейсы, результаты)

📱 МОБИЛЬНЫЕ ТЕХНОЛОГИИ:
- Обновления iOS/Android с конкретными фичами
- Новые приложения с метриками (загрузки, доходы)
- AR/VR новости (продукты, инвестиции, технологии)
- 5G/6G развитие (развертывание, скорости, покрытие)
- Мобильные игры (топ-чарты, доходы, тренды)

☁️ ОБЛАЧНЫЕ ТЕХНОЛОГИИ:
- Цены и тарифы облачных провайдеров
- Новые сервисы и API
- Крупные миграции в облако
- Безопасность и инциденты
- Edge computing и CDN

🔗 БЛОКЧЕЙН И WEB3:
- Новые блокчейн-проекты с техническими деталями
- DeFi протоколы (TVL, доходность, риски)
- NFT проекты (объемы, коллекции, художники)
- DAO и децентрализованные организации
- Регулятивные новости в крипто-сфере

🚀 СТАРТАПЫ И ИНВЕСТИЦИИ:
- Размеры раундов финансирования с именами инвесторов
- Оценки стартапов (pre-money, post-money)
- Выходы (IPO, M&A) с суммами сделок
- Имена основателей и их предыдущие проекты
- Технологические стеки и архитектуры

🔬 НАУЧНЫЕ ОТКРЫТИЯ:
- Исследования в области квантовых вычислений
- Биотехнологии и медицина
- Космические технологии (SpaceX, Blue Origin, другие)
- Энергетические инновации (батареи, возобновляемая энергия)

🔍 ВЕБ-ПОИСК: Найди актуальные новости по запросам:
- "AI news today"
- "startup news today"
- "tech company news today"
- "product hunt today"
- "new AI models today"
- "tech IPO news today"

ПОЛЕЗНЫЕ ССЫЛКИ:
- TechCrunch (https://techcrunch.com) - конкретные статьи и новости
- ProductHunt (https://producthunt.com) - конкретные продукты
- GitHub Trending (https://github.com/trending) - популярные репозитории
- Hacker News (https://news.ycombinator.com) - актуальные обсуждения
- VentureBeat (https://venturebeat.com) - новости стартапов и технологий
- The Verge (https://theverge.com) - обзоры и новости
- Ars Technica (https://arstechnica.com) - технические статьи
- Wired (https://wired.com) - технологические тренды
- MIT Technology Review (https://technologyreview.com) - научные открытия

ВАЖНО:
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Добавляй конкретные ссылки на источники
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #tech #ai #startups #innovation #software #blockchain #web3

Составь детальный технологический дайджест на русском для Telegram с ссылками на источники.
`

	// RealEstateDigestPrompt - промпт для дайджеста недвижимости
	RealEstateDigestPrompt = `
🏠 НЕДВИЖИМОСТЬ: ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — эксперт по недвижимости. Используй веб-поиск для анализа СВЕЖИХ НОВОСТЕЙ ЗА ПОСЛЕДНИЕ 24 ЧАСА.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

📈 ДИНАМИКА ЦЕН В ПОДМОСКОВЬЕ:
- Точные цены за квадратный метр по районам (в рублях)
- Процент изменения цен за месяц/квартал
- Топ-5 самых дорогих и дешевых районов с ценами
- Новые транспортные проекты и их влияние на цены
- Прогнозы экспертов с конкретными цифрами

🏞️ НОВЫЕ ЛОТЫ НА ТОРГАХ:
- Конкретные аукционы ГИС-Торги с датами и стартовыми ценами
- Земельные участки: площадь, назначение, район, цена
- Коммерческая недвижимость: тип, площадь, район, цена
- Жилая недвижимость: ЖК, этаж, площадь, цена
- Дедлайны подачи заявок

💰 ИНВЕСТИЦИОННЫЕ ВОЗМОЖНОСТИ:
- ROI по типам недвижимости (квартиры, коммерческая, земля)
- Ипотечные ставки в конкретных банках
- Налоговые изменения с датами вступления в силу
- Субсидии и льготы с условиями получения
- Арендные ставки по районам

📊 АНАЛИТИКА РЫНКА:
- Объем сделок с точными цифрами
- Время экспозиции объектов в днях
- Соотношение спроса и предложения в процентах
- Демографические данные по районам
- Экономические показатели (инфляция, ставки)

🔍 ИНТЕРЕСНЫЕ ОБЪЕКТЫ:
- Конкретные ЖК с датами сдачи и ценами
- Новые проекты с деталями
- Рекордные сделки с суммами
- Интересные лоты на торгах

🔍 ВЕБ-ПОИСК: Найди НОВОСТИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ по запросам:
- "Москва недвижимость новости сегодня"
- "Подмосковье недвижимость новости сегодня"
- "ГИС-Торги новости сегодня"
- "недвижимость новости Москва сегодня"
- "цены на недвижимость новости сегодня"
- "земельные участки новости сегодня"

ПОЛЕЗНЫЕ ССЫЛКИ:
- ЦИАН (https://cian.ru) - конкретные объявления
- Авито Недвижимость (https://avito.ru/rossiya/nedvizhimost) - актуальные цены
- ГИС-Торги (https://torgi.gov.ru) - конкретные лоты
- DomClick (https://domclick.ru) - ипотечные ставки
- РБК Недвижимость (https://realty.rbc.ru) - аналитика
- Н1 (https://n1.ru) - новостройки
- Мир Квартир (https://mirkvartir.ru) - вторичный рынок
- Realty.yandex.ru (https://realty.yandex.ru) - цены и аналитика

ВАЖНО:
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Добавляй конкретные ссылки на источники
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #realestate #property #investment #moscow #land #mortgage

Составь детальный дайджест недвижимости на русском для Telegram с конкретными ссылками на источники.
`

	// BusinessDigestPrompt - промпт для бизнес-дайджеста
	BusinessDigestPrompt = `
💼 БИЗНЕС-ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — бизнес-аналитик. Используй веб-поиск для анализа СВЕЖИХ НОВОСТЕЙ ЗА ПОСЛЕДНИЕ 24 ЧАСА.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

📊 РЫНОЧНЫЕ ТРЕНДЫ:
- Изменения в отраслевых индексах (точные значения и %)
- Корпоративные отчеты (выручка, прибыль, прогнозы)
- Слияния и поглощения (суммы сделок, компании)
- IPO и SPAC (размеры, оценки, инвесторы)
- Реструктуризации и банкротства

💡 ИННОВАЦИОННЫЕ ИДЕИ:
- Новые бизнес-модели с примерами
- Технологические прорывы в бизнесе
- Партнерства между компаниями
- Эксперименты с AI в бизнесе
- Устойчивое развитие и ESG

💰 ИНВЕСТИЦИОННЫЕ ВОЗМОЖНОСТИ:
- Венчурные инвестиции (размеры раундов, оценки)
- Private Equity сделки
- Корпоративные инвестиции
- Государственные программы поддержки
- Налоговые льготы и субсидии

🚀 СТАРТАПЫ И МАСШТАБИРОВАНИЕ:
- Успешные кейсы масштабирования
- Проблемы и вызовы роста
- Международная экспансия
- Партнерства с корпорациями
- Выходы (IPO, M&A)

🌍 МЕЖДУНАРОДНАЯ ТОРГОВЛЯ:
- Новые торговые соглашения
- Тарифы и пошлины
- Логистические изменения
- Цепочки поставок
- Валютные колебания и их влияние

🏢 ОТРАСЛЕВЫЕ НОВОСТИ:
- E-commerce: новые платформы, тренды, метрики
- FinTech: новые продукты, регуляция, инвестиции
- GreenTech: экологические решения, инвестиции, политика
- HealthTech: медицинские инновации, регуляция, партнерства
- EdTech: образовательные технологии, рынок, тренды

📈 ЭКОНОМИЧЕСКИЕ ИНДИКАТОРЫ:
- ВВП, инфляция, безработица
- Потребительские расходы
- Бизнес-климат и уверенность
- Процентные ставки и их влияние
- Валютные курсы

🔍 ВЕБ-ПОИСК: Найди НОВОСТИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ по запросам:
- "business news today [текущая дата]"
- "startup news today [текущая дата]"
- "IPO news today"
- "venture capital news today"
- "company earnings news today"
- "mergers acquisitions news today"

ПОЛЕЗНЫЕ ССЫЛКИ:
- Bloomberg (https://bloomberg.com) - бизнес-новости
- Reuters (https://reuters.com) - финансовые новости
- Forbes (https://forbes.com) - предпринимательство
- Harvard Business Review (https://hbr.org) - стратегия
- Startup Grind (https://startupgrind.com) - стартапы
- Inc. (https://inc.com) - рост бизнеса
- Entrepreneur (https://entrepreneur.com) - предпринимательство
- Wall Street Journal (https://wsj.com) - финансовые рынки
- Financial Times (https://ft.com) - международная экономика

ВАЖНО:
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Добавляй конкретные ссылки на источники
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #business #entrepreneurship #strategy #growth #startups #innovation

Составь детальный бизнес-обзор на русском для Telegram с ссылками на источники.
`

	// InvestmentDigestPrompt - промпт для инвестиционного дайджеста
	InvestmentDigestPrompt = `
💰 ИНВЕСТИЦИОННЫЙ ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — инвестиционный аналитик. Используй веб-поиск для анализа СВЕЖИХ НОВОСТЕЙ ЗА ПОСЛЕДНИЕ 24 ЧАСА.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

📈 АКЦИИ И ФОНДОВЫЙ РЫНОК:
- Индексы (S&P 500, NASDAQ, Dow Jones) с точными значениями и % изменения
- Топ-10 движущихся акций (цены и причины)
- Объемы торгов по секторам
- Корреляции между активами
- Рекомендации аналитиков (покупать/продавать/держать)

🪙 КРИПТОВАЛЮТЫ И DEFI:
- Топ-10 криптовалют по капитализации с ценами
- DeFi протоколы (TVL, доходность, риски)
- NFT рынок (объемы, топ-коллекции)
- Институциональные новости (ETF, корпоративные инвестиции)

🏠 НЕДВИЖИМОСТЬ:
- Индексы недвижимости (REIT) с динамикой
- Цены на жилье в ключевых городах
- Ипотечные ставки и их влияние
- Коммерческая недвижимость (офисы, склады, торговые центры)

💎 АЛЬТЕРНАТИВНЫЕ ИНВЕСТИЦИИ:
- Золото, серебро, платина (цены и тренды)
- Коллекционные предметы (вина, искусство, часы)
- Венчурные инвестиции (размеры раундов, оценки)
- Private Equity сделки

🌍 МЕЖДУНАРОДНЫЕ РЫНКИ:
- Европейские индексы (FTSE, DAX, CAC)
- Азиатские рынки (Nikkei, Hang Seng, Shanghai)
- Валютные пары (EUR/USD, GBP/USD, USD/JPY)
- Экономические индикаторы (ВВП, инфляция, безработица)

📊 ТЕХНИЧЕСКИЙ АНАЛИЗ:
- Уровни поддержки и сопротивления для ключевых активов
- RSI, MACD, Bollinger Bands
- Паттерны свечного анализа
- Прогнозы на основе технических индикаторов

🔍 ВЕБ-ПОИСК: Найди НОВОСТИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ по запросам:
- "stock market news today [текущая дата]"
- "investment news today"
- "market analysis today"
- "financial news today"
- "stock prices news today"
- "market trends today"

ПОЛЕЗНЫЕ ССЫЛКИ:
- Yahoo Finance (https://finance.yahoo.com) - акции и рынки
- CoinGecko (https://coingecko.com) - криптовалюты
- TradingView (https://tradingview.com) - технический анализ
- Bloomberg (https://bloomberg.com) - финансовые новости
- Reuters (https://reuters.com) - рыночные данные
- CNBC (https://cnbc.com) - инвестиционные новости
- Investing.com (https://investing.com) - рыночная аналитика
- MarketWatch (https://marketwatch.com) - рыночные индексы
- Seeking Alpha (https://seekingalpha.com) - инвестиционные идеи

ВАЖНО:
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Добавляй конкретные ссылки на источники
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #investing #stocks #crypto #realestate #finance #wealth #portfolio

Составь детальный инвестиционный обзор на русском для Telegram с ссылками на источники.
`

	// StartupDigestPrompt - промпт для стартап-дайджеста
	StartupDigestPrompt = `
🚀 СТАРТАП-ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — эксперт по стартапам. Используй веб-поиск для анализа СВЕЖИХ НОВОСТЕЙ ЗА ПОСЛЕДНИЕ 24 ЧАСА.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

🔥 ГОРЯЧИЕ СТАРТАПЫ ДНЯ:
- Топ-10 новых стартапов с описанием и метриками
- Имена основателей и их предыдущие проекты
- Технологические стеки и архитектуры
- Целевые рынки и аудитория
- Конкурентный анализ

💡 ИННОВАЦИОННЫЕ ИДЕИ:
- Новые бизнес-модели с примерами
- Технологические прорывы
- Решение реальных проблем
- Уникальные подходы к рынку
- Экспериментальные концепции

💰 ФАНДИНГ И ИНВЕСТИЦИИ:
- Размеры раундов с именами инвесторов
- Оценки стартапов (pre-money, post-money)
- Условия сделок (доли, права, опционы)
- Новые фонды и их стратегии
- Корпоративные инвестиции

📱 НОВЫЕ ПРОДУКТЫ:
- Запуски с метриками (загрузки, доходы, пользователи)
- Обновления существующих продуктов
- Бета-версии и тестирование
- Партнерства и интеграции
- Пользовательский опыт и отзывы

🌍 ГЛОБАЛЬНЫЕ ТРЕНДЫ:
- Региональные экосистемы (США, Европа, Азия)
- Новые акселераторы и инкубаторы
- Государственные программы поддержки
- Международная экспансия стартапов
- Кросс-граничные инвестиции

🏢 КАТЕГОРИИ СТАРТАПОВ:
- SaaS и B2B: новые решения для бизнеса
- Consumer Tech: продукты для массового рынка
- FinTech: финансовые технологии и инновации
- HealthTech: медицинские и биотехнологии
- EdTech: образовательные технологии
- GreenTech: экологические решения

📊 МЕТРИКИ И АНАЛИТИКА:
- Время до выхода (time to exit)
- Средние оценки по отраслям
- Успешность раундов финансирования
- Географическое распределение инвестиций
- Тренды в размерах сделок

🎯 ВЫХОДЫ И ЛИКВИДАЦИИ:
- IPO с деталями (размер, оценка, инвесторы)
- M&A сделки (покупатели, суммы, стратегии)
- SPAC слияния
- Ликвидации и неудачные проекты
- Уроки и выводы

🔍 ВЕБ-ПОИСК: Найди НОВОСТИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ по запросам:
- "startup news today [текущая дата]"
- "startup funding news today"
- "new startups launched today"
- "venture capital news today"
- "startup acquisitions news today"
- "startup IPO news today"

ПОЛЕЗНЫЕ ССЫЛКИ:
- ProductHunt (https://producthunt.com) - новые продукты
- TechCrunch (https://techcrunch.com) - стартап-новости
- AngelList (https://angel.co) - инвестиции
- Crunchbase (https://crunchbase.com) - данные о стартапах
- Startup Grind (https://startupgrind.com) - экосистема
- Y Combinator (https://ycombinator.com) - акселератор
- PitchBook (https://pitchbook.com) - венчурные данные
- CB Insights (https://cbinsights.com) - аналитика
- Dealroom (https://dealroom.co) - сделки

ВАЖНО:
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Добавляй конкретные ссылки на источники
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #startups #funding #entrepreneurship #innovation #tech #venture

Составь детальный стартап-обзор на русском для Telegram с ссылками на источники.
`

	// GlobalDigestPrompt - промпт для глобального дайджеста
	GlobalDigestPrompt = `
🌍 ГЛОБАЛЬНЫЙ ДАЙДЖЕСТ НА СЕГОДНЯ

Ты — международный аналитик. Используй веб-поиск для анализа СВЕЖИХ НОВОСТЕЙ ЗА ПОСЛЕДНИЕ 24 ЧАСА.

МАКСИМУМ ПОЛЕЗНОЙ ИНФОРМАЦИИ:

🏛️ ПОЛИТИКА И ГЕОПОЛИТИКА:
- Ключевые политические решения с именами лидеров
- Международные встречи и саммиты (участники, результаты)
- Дипломатические отношения и конфликты
- Выборы и референдумы (результаты, явка, победители)
- Санкции и торговые ограничения

💰 ЭКОНОМИКА И ФИНАНСЫ:
- ВВП и экономические индикаторы по странам
- Центральные банки (решения по ставкам, заявления)
- Валютные курсы и их влияние
- Торговые соглашения и тарифы
- Инфляция и безработица по регионам

🌱 ЭКОЛОГИЯ И КЛИМАТ:
- Климатические события (температуры, стихийные бедствия)
- Экологические политики и решения
- Возобновляемая энергетика (инвестиции, проекты)
- Углеродные налоги и торговля квотами
- Международные климатические соглашения

🔬 НАУКА И ТЕХНОЛОГИИ:
- Научные открытия с именами исследователей
- Космические миссии и достижения
- Медицинские прорывы и исследования
- Технологические инновации
- Международные научные коллаборации

🎭 КУЛЬТУРА И ОБЩЕСТВО:
- Крупные культурные события и фестивали
- Социальные движения и протесты
- Демографические изменения
- Образовательные инициативы
- Спортивные события международного значения

🌍 РЕГИОНАЛЬНЫЕ НОВОСТИ:
- США и Европа: политика, экономика, технологии
- Азия и Тихоокеанский регион: рост, инновации, конфликты
- Ближний Восток: энергетика, политика, безопасность
- Латинская Америка: экономика, политика, природные ресурсы
- Африка: развитие, инвестиции, вызовы

📊 СТАТИСТИКА И ДАННЫЕ:
- Демографические показатели
- Экономические метрики
- Социальные индикаторы
- Технологические тренды
- Экологические данные

🔍 АНАЛИЗ И ПРОГНОЗЫ:
- Экспертные мнения по ключевым событиям
- Прогнозы экономистов и аналитиков
- Политические сценарии
- Технологические тренды
- Геополитические риски

🔍 ВЕБ-ПОИСК: Найди НОВОСТИ ИМЕННО СЕГОДНЯШНЕГО ДНЯ по запросам:
- "world news today [текущая дата]"
- "global economy news today"
- "international news today"
- "geopolitical news today"
- "world markets news today"
- "global trade news today"

ПОЛЕЗНЫЕ ССЫЛКИ:
- BBC World (https://bbc.com/news/world) - мировые новости
- CNN International (https://cnn.com) - международные события
- Reuters (https://reuters.com) - глобальные новости
- Al Jazeera (https://aljazeera.com) - международная аналитика
- Associated Press (https://ap.org) - мировые события
- The Guardian (https://theguardian.com) - международная политика
- Financial Times (https://ft.com) - глобальная экономика
- The Economist (https://economist.com) - международная аналитика
- Foreign Policy (https://foreignpolicy.com) - внешняя политика

ВАЖНО:
1. Используй веб-поиск для получения актуальных новостей
2. Фокусируйся на актуальных новостях
3. Включай ссылки на источники
4. Используй актуальную информацию

Форматирование:
- НЕ используй жирное форматирование (**)
- Добавляй конкретные ссылки на источники
- Используй эмодзи для структурирования
- Пиши кратко и по делу
- В конце добавь хештеги: #world #geopolitics #economy #news #global #politics

Составь детальный глобальный обзор на русском для Telegram с ссылками на источники.
`
)

// WebSearchDoc provides documentation for the web search tool output format.
const WebSearchDoc = `Output and citations
Model responses that use the web search tool will include two parts:

A web_search_call output item with the ID of the search call, along with the action taken in web_search_call.action. The action is one of:
search, which represents a web search. It will usually (but not always) includes the search query and domains which were searched. Search actions incur a tool call cost (see pricing).
open_page, which represents a page being opened. Only emitted by Deep Research models.
find_in_page, which represents searching within a page. Only emitted by Deep Research models.
A message output item containing:
The text result in message.content[0].text
Annotations message.content[0].annotations for the cited URLs
By default, the model's response will include inline citations for URLs found in the web search results. In addition to this, the url_citation annotation object will contain the URL, title and location of the cited source.

When displaying web results or information contained in web results to end users, inline citations must be made clearly visible and clickable in your user interface.

[
    {
        "type": "web_search_call",
        "id": "ws_67c9fa0502748190b7dd390736892e100be649c1a5ff9609",
        "status": "completed"
    },
    {
        "id": "msg_67c9fa077e288190af08fdffda2e34f20be649c1a5ff9609",
        "type": "message",
        "status": "completed",
        "role": "assistant",
        "content": [
            {
                "type": "output_text",
                "text": "On March 6, 2025, several news...",
                "annotations": [
                    {
                        "type": "url_citation",
                        "start_index": 2606,
                        "end_index": 2758,
                        "url": "https://...",
                        "title": "Title..."
                    }
                ]
            }
        ]
    }
]
User location
To refine search results based on geography, you can specify an approximate user location using country, city, region, and/or timezone.

The city and region fields are free text strings, like Minneapolis and Minnesota respectively.
The country field is a two-letter ISO country code, like US.
The timezone field is an IANA timezone like America/Chicago.
Note that user location is not supported for deep research models using web search.

Customizing user location
curl "https://api.openai.com/v1/responses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -d '{
        "model": "o4-mini",
        "tools": [{
            "type": "web_search_preview",
            "user_location": {
                "type": "approximate",
                "country": "GB",
                "city": "London",
                "region": "London"
            }
        }],
        "input": "What are the best restaurants around Granary Square?"
    }'
Search context size
When using this tool, the search_context_size parameter controls how much context is retrieved from the web to help the tool formulate a response. The tokens used by the search tool do not affect the context window of the main model specified in the model parameter in your response creation request. These tokens are also not carried over from one turn to another — they're simply used to formulate the tool response and then discarded.

Choosing a context size impacts:

Cost: Search content tokens are free for some models, but may be billed at a model's text token rates for others. Refer to pricing for details.
Quality: Higher search context sizes generally provide richer context, resulting in more accurate, comprehensive answers.
Latency: Higher context sizes require processing more tokens, which can slow down the tool's response time.
Available values:

high: Most comprehensive context, slower response.
medium (default): Balanced context and latency.
low: Least context, fastest response, but potentially lower answer quality.
Context size configuration is not supported for o3, o3-pro, o4-mini, and deep research models.

Customizing search context size
curl "https://api.openai.com/v1/responses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -d '{
        "model": "gpt-4.1",
        "tools": [{
            "type": "web_search_preview",
            "search_context_size": "low"
        }],
        "input": "What movie won best picture in 2025?"
    }'
Usage notes
API AvailabilityRate limitsNotes
Responses
Chat Completions
Assistants
Same as tiered rate limits for underlying model used with the tool.

Pricing
ZDR and data residency

Limitations
Web search is currently not supported in the 
gpt-4.1-nano
 model.
The 
gpt-4o-search-preview
 and 
gpt-4o-mini-search-preview
 models used in Chat Completions only support a subset of API parameters - view their model data pages for specific information on rate limits and feature support.
When used as a tool in the Responses API, web search has the same tiered rate limits as the models above.
Web search is limited to a context window size of 128000 (even with 
gpt-4.1
 and 
gpt-4.1-mini
 models).
Refer to this guide for data handling, residency, and retention information.`
