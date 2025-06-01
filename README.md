# Situs-interretialis-opum-humanarum
Demo of human resource like Taiwan's 104

# Detailed System Architecture with Real-time Notification
- 使用者瀏覽器（Vue.js SPA, 支援 WebSocket）
- NGINX 負責反向代理、靜態檔案與 WebSocket Gateway
- Golang API 服務處理所有 REST API 及 WebSocket，對接 PostgreSQL 與 Redis
- Redis 除了查詢快取，也用 Pub/Sub 通知所有 API 實例資料異動
- 當資料異動時（如職缺更新），API server 透過 WebSocket 主動推播通知前端即時刷新
flowchart TD
    A[User Browser<br/>(Vue.js SPA,<br/>WebSocket)] -->|HTTP/WS| B(NGINX<br/>(Proxy / Static SPA / WS Gateway))
    B -->|Static file| A
    B -->|API/WS Proxy| C(Golang API<br/>(Gin, REST+WS))
    C -->|Query/Write| D[PostgreSQL]
    C -->|Cache / Hot keywords| E[Redis]
    E <-->|Pub/Sub (job_update,<br/>company_update, ...)| C
    C -->|WebSocket Push| A
# Docker Architecture
graph LR
    docker-compose
    docker-compose --> vue[Vue App]
    docker-compose --> api[Golang API]
    docker-compose --> db[PostgreSQL]
    docker-compose --> redis[Redis]
    docker-compose --> nginx[NGINX]

    vue -- static files --> nginx
    api -- REST/WS --> nginx
    api -- query --> db
    api -- cache/pubsub --> redis
    nginx -- proxy --> api
