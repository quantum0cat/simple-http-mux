Simple HTTP Multiplexer
- приложение представляет собой http-сервер с одним хендлером
- хендлер на вход получает POST-запрос со списком url в json-формате  
  -:сервер запрашивает данные по всем этим url и возвращает результат клиенту в json-формате
- если в процессе обработки хотя бы одного из url получена ошибка, обработка всего списка прекращается и клиенту возвращается текстовая ошибка
- для реализации задачи следует использовать Go 1.13 или выше
- использовать можно только компоненты стандартной библиотеки Go
- сервер не принимает запрос если количество url в в нем больше 20
- сервер не обслуживает больше чем 100 одновременных входящих http-запросов
- таймаут на обработку одного входящего запроса - 10 секунд
- для каждого входящего запроса должно быть не больше 4 одновременных исходящих
- таймаут на запрос одного url - секунда
- обработка запроса может быть отменена клиентом в любой момент, это должно повлечь за собой остановку всех операций связанных с этим запросом
- сервис должен поддерживать 'graceful shutdown': при получении сигнала от OS перестать принимать входящие запросы, завершить текущие запросы и остановиться
- результат должен быть выложен на github и запускаться docker-compose.  