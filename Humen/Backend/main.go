package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ----- Models -----
type Company struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:128"`
}

func (Company) TableName() string { return "company" }

type Job struct {
	ID        uint `gorm:"primaryKey"`
	CompanyID uint
	Title     string `gorm:"size:128"`
	SalaryMin int
	SalaryMax int
	Company   Company `gorm:"foreignKey:CompanyID"`
}

func (Job) TableName() string { return "job" }

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;size:64"`
	PasswordHash string
	IsAdmin      bool `gorm:"default:false"`
}

func (User) TableName() string { return "users" }

var (
	DB        *gorm.DB
	RDB       *redis.Client
	ctx       = context.Background()
	jwtSecret = []byte("your-secret-key") // 建議用環境變數設定
)

// ----- DB/Redis Init -----
func initDB() {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("POSTGRES_PASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "jobdb"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, pass, dbname, port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to initialize database, got error %v", err)
	}
	DB = db
}

func initRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
}

// ----- Auth Handlers -----

func RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}
	user := User{Username: req.Username, PasswordHash: string(hash)}
	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user exists"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "註冊成功"})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	var user User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}
	// 回傳 is_admin, user_id 讓前端存
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "is_admin": user.IsAdmin, "user_id": user.ID})
}

// ---- JWT Middleware ----
func AuthMiddleware(requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if len(auth) < 8 || auth[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token"})
			return
		}
		tokenStr := auth[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		if requireAdmin {
			isAdmin, ok := claims["is_admin"].(bool)
			if !ok {
				f, ok2 := claims["is_admin"].(float64)
				if ok2 && f == 1 {
					isAdmin = true
				}
			}
			if !isAdmin {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "管理員限定"})
				return
			}
		}
		c.Set("user_id", claims["user_id"])
		c.Set("is_admin", claims["is_admin"])
		c.Next()
	}
}

// ----- Job Handlers -----

func GetJobs(c *gin.Context) {
	keyword := c.Query("keyword")
	redisKey := "jobs:all"
	if keyword != "" {
		redisKey = "jobs:search:" + keyword
	}

	data, err := RDB.Get(ctx, redisKey).Bytes()
	if err == nil {
		RDB.Expire(ctx, redisKey, 10*time.Minute)
		var jobs []map[string]interface{}
		if err := json.Unmarshal(data, &jobs); err == nil {
			c.JSON(http.StatusOK, jobs)
			return
		}
	}

	var jobs []Job
	q := DB.Preload("Company")
	if keyword != "" {
		// 關鍵字搜尋同時查公司名稱與職缺名稱
		q = q.Joins("JOIN company ON company.id = job.company_id").
			Where("job.title ILIKE ? OR company.name ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		result = append(result, map[string]interface{}{
			"id":         job.ID,
			"company":    job.Company.Name,
			"title":      job.Title,
			"salary_min": job.SalaryMin,
			"salary_max": job.SalaryMax,
		})
	}
	encoded, _ := json.Marshal(result)
	RDB.Set(ctx, redisKey, encoded, 10*time.Minute)
	c.JSON(http.StatusOK, result)
}

// 建立職缺，同步刪掉快取
func CreateJob(c *gin.Context) {
	var req struct {
		CompanyID uint   `json:"company_id"`
		Title     string `json:"title"`
		SalaryMin int    `json:"salary_min"`
		SalaryMax int    `json:"salary_max"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	job := Job{
		CompanyID: req.CompanyID,
		Title:     req.Title,
		SalaryMin: req.SalaryMin,
		SalaryMax: req.SalaryMax,
	}
	if err := DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create job failed"})
		return
	}
	// 刪所有相關快取
	RDB.Del(ctx, "jobs:all")
	RDB.Del(ctx, "jobs:search:"+req.Title)
	c.JSON(http.StatusOK, job)
}

func DeleteJob(c *gin.Context) {
	id := c.Param("id")
	if err := DB.Delete(&Job{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
		return
	}
	RDB.Del(ctx, "jobs:all")
	c.JSON(http.StatusOK, gin.H{"msg": "已刪除"})
}

// 公司統計
func GetCompanyStat(c *gin.Context) {
	type Stat struct {
		Company    string  `json:"company"`
		AvgSalary  float64 `json:"avg_salary"`
		HighSalary int     `json:"high_salary"`
	}
	var stats []Stat
	DB.Raw(`
		SELECT c.name as company,
			ROUND(AVG((j.salary_min + j.salary_max) / 2)) as avg_salary,
			SUM(CASE WHEN j.salary_min > 100000 THEN 1 ELSE 0 END) as high_salary
		FROM company c
		JOIN job j ON c.id = j.company_id
		GROUP BY c.name
		ORDER BY c.name
	`).Scan(&stats)
	c.JSON(http.StatusOK, stats)
}

// ------ User 管理 (admin only) --------

type SimpleUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

func ListUsers(c *gin.Context) {
	var users []SimpleUser
	if err := DB.Model(&User{}).Select("id, username, is_admin").Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	// 防止 admin 刪自己
	userID := fmt.Sprintf("%v", c.MustGet("user_id"))
	if userID == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能刪除自己"})
		return
	}
	if err := DB.Delete(&User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "帳號已刪除"})
}

func main() {
	initDB()
	initRedis()

	r := gin.Default()
	// Auth API
	r.POST("/api/register", RegisterHandler)
	r.POST("/api/login", LoginHandler)

	api := r.Group("/api")
	api.Use(AuthMiddleware(false))
	{
		api.GET("/jobs", GetJobs)
		api.GET("/companies/stat", GetCompanyStat)
	}

	admin := r.Group("/api")
	admin.Use(AuthMiddleware(true))
	{
		admin.POST("/jobs", CreateJob)
		admin.DELETE("/jobs/:id", DeleteJob)
		admin.GET("/users", ListUsers)
		admin.DELETE("/users/:id", DeleteUser)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server started on :%s\n", port)
	r.Run(":" + port)
}
