# Разработка интернет-приложений

## Репозиторий для лабораторных работ

# Лабораторная работа №3

## Цель работы

Создание REST API веб-сервиса для управления данными археологических артефактов и заявок на расчёт TPQ (Terminus Post Quem) с подключением к базе данных PostgreSQL и тестированием через Insomnia/Postman.

## Задание

Разработка полноценного REST API для веб-приложения, реализующего бизнес-логику управления услугами (артефактами), заявками, связями многие-ко-многим (m-m) и пользователями. API должно соответствовать принципам REST, использовать ORM (GORM) для взаимодействия с базой данных PostgreSQL, интегрироваться с Minio для управления изображениями услуг, и поддерживать фильтрацию данных. Все методы начинаются с `/api`. Авторизация пока не реализуется, пользователь фиксирован через константу (singleton). Тестирование API проводится через коллекцию из 21 запроса в Insomnia/Postman.

## Порядок показа на защите

### 1. Демонстрация работы API через Insomnia/Postman

- **Коллекция запросов**: Показать коллекцию из 21 запроса, охватывающую все методы API.
- **GET /api/requests**: Получить список заявок с фильтрацией по статусу (formed, completed, rejected) и диапазону дат формирования (`formed_at`).
- **GET /api/requests/draft**: Получить текущую черновую заявку (draft) для фиксированного пользователя (creator_id=1).
- **DELETE /api/requests/**: Удалить черновую заявку (логическое удаление, status='deleted').
- **GET /api/services**: Получить список услуг (артефактов) с фильтрацией по имени, эпохе, TPQ, датам (start_date, end_date).
- **POST /api/services**: Добавить новую услугу (без изображения).
- **POST /api/services//image**: Добавить/заменить изображение услуги в Minio (генерация имени на латинице).
- **GET /api/requests/draft**: Повторно получить черновую заявку.
- **POST /api/requests/draft/services**: Добавить другую услугу в черновую заявку.
- **GET /api/requests/draft/services**: Получить список услуг в черновой заявке.
- **GET /api/requests/**: Просмотреть заявку с двумя услугами (включая изображения).
- **PUT /api/requests//items**: Изменить поле в m-m связи (e.g., comment).
- **PUT /api/requests/**: Изменить поле заявки (e.g., excavation).
- **PUT /api/requests//complete**: Попытка завершить заявку создателем (показать ошибку — только модератор).
- **PUT /api/requests//form**: Сформировать заявку (установить formed_at, проверить обязательные поля).
- **PUT /api/requests//complete**: Завершить сформированную заявку модератором (установить completed_at, moderator_id, вычислить TPQ как max(TPQ услуг)).
- **POST /api/users/register**: Зарегистрировать нового пользователя.
- **GET /api/users/me**: Получить данные пользователя (для личного кабинета).
- **PUT /api/users/me**: Обновить данные пользователя.
- **POST /api/users/login**: Аутентификация пользователя.
- **POST /api/users/logout**: Деаутентификация.
- **DELETE /api/requests//items**: Удалить услугу из заявки.
- **SQL SELECT**: Показать изменённые данные через SELECT-запросы в Adminer (e.g., SELECT \* FROM tpq_requests WHERE id = '').

### 2. Анализ кода

**Продемонстрировать в коде:**

- **Модели**: Структура ORM-моделей (`Artifact`, `TPQRequest`, `TPQRequestItem`, `User`) в `internal/models`.
- **Сериализаторы**: Код для преобразования моделей в JSON (e.g., structs с `json:"field"` тегами).
- **Контроллеры**: Реализация REST API в `internal/handlers` (GET, POST, PUT, DELETE для доменов services, requests, items, users).
- **Singleton для пользователя**: Константа `CreatorID = 1` для фиксации пользователя (без авторизации).
- **Фильтрация**: Реализация серверной фильтрации в GET /api/services и GET /api/requests (e.g., WHERE с LIKE для имени, диапазон дат).
- **Minio**: Логика загрузки/удаления изображений в POST /api/services//image и DELETE /api/services/.
- **Бизнес-логика**: Проверки статусов (e.g., draft → formed → completed/rejected), расчёт TPQ в PUT /api/requests//complete.
- **ORM**: Использование GORM для всех операций с БД (Find, First, Create, Save, Preload).
- **Запрет изменения системных полей**: Логика в хэндлерах, игнорирующая id, status, creator_id, formed_at, completed_at из клиентских данных.

### 3. Контрольные вопросы

1. **Веб-сервис**: Определение, архитектура, примеры (REST vs RPC).
2. **REST**: Принципы (stateless, client-server, uniform interface), соответствие методов и URL.
3. **RPC**: Отличия от REST, области применения.
4. **Заголовки и методы HTTP**: GET, POST, PUT, DELETE, headers (Content-Type, Authorization).
5. **Версии HTTP**: HTTP/1.1, HTTP/2, HTTP/3 (QUIC), их особенности.
6. **HTTPS**: Шифрование, TLS, сертификаты.
7. **OSI ISO**: 7 уровней модели, их роль в веб-сервисах.

## Диаграмма классов

Требуется создать в StarUML диаграмму классов, описывающую бэкенд:

## Требования к веб-сервису

- **REST API**: Все методы следуют REST (ресурсы: /api/services, /api/requests, /api/users; методы: GET, POST, PUT, DELETE).
- **Фильтрация**: Серверная фильтрация для GET /api/services (name, epoch, etc.) и GET /api/requests (status, formed_at range).
- **ORM**: Все операции с БД через GORM (Find, First, Create, Save, Preload).
- **Статусы заявок**: Ограничения переходов (draft → formed/deleted создателем; formed → completed/rejected модератором).
- **Фиксированный пользователь**: Константа CreatorID=1 (singleton) для всех операций.
- **Системные поля**: id, status, creator_id, formed_at, completed_at, moderator_id не изменяются клиентом.
- **Minio**: Используется только в POST /api/services//image (загрузка/замена) и DELETE /api/services/ (удаление изображения).
- **Исключения**: Записи со status='deleted' не возвращаются. POST для создания заявки отсутствует (автоматически через /api/requests/draft/services).