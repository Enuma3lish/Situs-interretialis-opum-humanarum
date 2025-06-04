# Humen 職缺管理系統

本專案是一套全端工程範例，具備職缺管理與公司統計、帳號系統（支援 JWT 登入/註冊）、Redis 快取加速，並採用 Docker 一鍵啟動所有服務。

---

## 目錄結構

```

.
├── Backend/
│   ├── main.go         # Go Gin RESTful API
│   ├── Dockerfile      # Backend 專用 Dockerfile
│   └── init.sql        # 資料庫初始化
├── Frontend/Homepage/
│   ├── src/            # Vue3 + TypeScript 前端原始碼
│   ├── Dockerfile      # Frontend 專用 Dockerfile
│   └── ...
├── docker-compose.yml  # 一鍵啟動所有服務
└── README.md

````

---

## 一鍵啟動

### 1. 準備環境

- **Docker / Docker Desktop**  
  安裝好即可（不用本地裝 Go 或 Node）

### 2. 啟動全部服務

請**務必在專案根目錄**（有 `docker-compose.yml` 那層）執行：

```bash
docker-compose up --build
````

第一次執行會自動 build 並初始化所有環境與資料。

### 3. 開始使用

* 前端服務: [http://localhost](http://localhost)
* 預設管理員帳密：
  帳號：admin
  密碼：root

---

## 啟動成功後，你可以...

* 透過前端 UI

  * **註冊/登入**（JWT 認證，權限管理）
  * **查詢所有職缺、搜尋關鍵字**（支援公司名/職缺名模糊搜尋）
  * **瀏覽公司薪資統計**
  * **管理員權限：新增/刪除職缺、刪除帳號**
* 所有 API 請求自動連到後端（Gin + PostgreSQL + Redis）

---

## 常見問題 Q\&A

### Q1. 啟動時出現 port 被佔用？

> 例如：`Ports are not available: ... 0.0.0.0:5432: bind: address already in use`

* **解法**：
  本機有其他程式或舊的 Docker container 佔用 5432、6379、8080、80 請先關閉或停用。
* 可以用 `lsof -i :5432` 查出來，或直接重開 Docker Desktop。

---

### Q2. 第一次登入用 admin/root 卻失敗？

* 請確認 `Backend/init.sql` 已正確建立預設管理員帳密
* Docker Compose 會自動初始化資料庫，除非你已經有舊 volume 沒被移除（Postgres 預設不會覆蓋舊資料）
* 若要**重設資料庫**，可以移除資料卷後重啟（注意資料會清空）：

```bash
docker-compose down -v
docker-compose up --build
```

---

### Q3. 修改程式碼後要如何重建？

* 只要加上 `--build` 參數即可：

```bash
docker-compose up --build
```

---

### Q4. 前端頁面進不去 or 無法顯示？

* 請確定後端（8080）和前端（80）服務都啟動成功
* 也可以檢查後端 API log 是否有錯誤

---

### Q5. 如何只重啟某個服務？

```bash
docker-compose restart backend
docker-compose restart frontend
```

---

### Q6. 開發測試想進入資料庫或 Redis？

```bash
docker exec -it <資料庫服務名，如 db> psql -U postgres -d jobdb
docker exec -it redis redis-cli
```

---

### Q7. 遇到容器啟動但無法連線？

* 檢查 `.env`、環境變數、資料庫/Redis 連線資訊
* 檢查 backend 的 log，常見於 host 名稱沒設對、沒等 db/redis ready

---

### Q8. 如何清除全部容器及資料？

```bash
docker-compose down -v
```

---

## 系統技術棧

* **Frontend:** Vue 3 + TypeScript + Vite
* **Backend:** Go (Gin) + JWT 認證 + PostgreSQL + Redis
* **資料庫初始化**：`Backend/init.sql`
* **部署工具**：Docker, docker-compose
