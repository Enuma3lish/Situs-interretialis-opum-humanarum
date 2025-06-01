# Situs-interretialis-opum-humanarum
Demo of human resource like Taiwan's 104

# Detailed System Architecture with Real-time Notification
+-----------------------------+
|        使用者瀏覽器           |
|   (Vue.js SPA, WebSocket)   |
+-------------^---------------+
              | (HTTP/WS/HTTPS)
              v
+-----------------------------+
|           NGINX             |
| (Reverse Proxy, Static SPA, |
|   WebSocket Gateway, SSL)   |
+------+-------------+--------+
       |             |
       |             |
(1) 靜態檔請求   (2) API/WS 代理
       |             |
       v             v
+------|-------------|---------+
|      |      Golang API       |
|      |   (Gin, REST+WS)      |
|      +----------+------------+
|                 |           
|   (3) 查詢或寫入 PostgreSQL/Redis
|         +-------v--------+
|         |   PostgreSQL   |
|         +---------------+
|         |     Redis     |
|         +--^-----^------+
|            |     |       
| (4) Redis  |     | (5) 發布 Pub/Sub
|     查詢快取     |     Channel
|   & 熱門關鍵字  |
|            +-----+
|   <------------------------------->|
|     Redis Pub/Sub (job_update,     |
|        company_update, ...)        |
|                                    |
+------v-----------------------------+
       |
(6) Backend 監聽 Redis channel
       |
(7) 當有資料異動，WebSocket
    推送通知至所有前端
       |
+------v-----------------------------+
|   使用者瀏覽器 Vue.js 前端           |
| (onmessage handler, 主動刷新資料)  |
+------------------------------------+

# Docker Architecture

+-------------------+
|   docker-compose  |
+--------+----------+
         |
   +-----+------------------------------+
   |     |          |          |        |
+--v-+ +--v--+  +---v---+  +---v---+ +--v---+
|Vue | |Golang|  |Postgre|  |Redis | |Nginx |
|App | |API   |  |SQL DB |  |Cache | |proxy |
+----+ +------+  +-------+  +------+ +------+

