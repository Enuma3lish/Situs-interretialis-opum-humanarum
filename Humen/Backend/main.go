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
	"github.com/redis/go-redis/v9"
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

var (
	DB  *gorm.DB
	RDB *redis.Client
	ctx = context.Background()
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

// ----- Handlers -----

func GetJobs(c *gin.Context) {
	redisKey := "jobs:all"

	// 1. Try to get from cache
	data, err := RDB.Get(ctx, redisKey).Bytes()
	if err == nil {
		// Found in cache, refresh TTL
		RDB.Expire(ctx, redisKey, 10*time.Minute)
		var jobs []map[string]interface{}
		if err := json.Unmarshal(data, &jobs); err == nil {
			c.JSON(http.StatusOK, jobs)
			return
		}
		// Unmarshal error, fallback to DB
	}

	// 2. Get from DB
	var jobs []Job
	if err := DB.Preload("Company").Find(&jobs).Error; err != nil {
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

	// 3. Cache the result with 10 min TTL
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
	// 刪快取
	RDB.Del(ctx, "jobs:all")
	c.JSON(http.StatusOK, job)
}

func main() {
	initDB()
	initRedis()

	r := gin.Default()
	r.GET("/jobs", GetJobs)
	r.POST("/jobs", CreateJob) // for testing

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server started on :%s\n", port)
	r.Run(":" + port)
}
