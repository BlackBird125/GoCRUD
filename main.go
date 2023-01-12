package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"

	_ "main.go/docs" // ←追記

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type User struct {
	gorm.Model
	Name     string    `json:"name"`
	Age      int       `json:"age"`
	Birthday time.Time `json:"birthday" time_format:"2006-01-02"`
}

func NewUser() User {
	return User{}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func gormConnect() *gorm.DB {
	DBMS := "mysql"
	USER := os.Getenv("DBUSER")
	PASS := os.Getenv("PASSWORD")
	PROTOCOL := "tcp(" + os.Getenv("DOMAIN") + ":" + os.Getenv("PORT") + ")"
	DBNAME := os.Getenv("DBNAME") + "?parseTime=true&loc=Asia%2FTokyo"
	CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME

	db, err := gorm.Open(DBMS, CONNECT)

	if err != nil {
		panic(err.Error())
	}
	fmt.Println("db connected: ", &db)
	return db
}

func setRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/test", test)

	//CREATE
	r.POST("/user", func(c *gin.Context) {
		data := NewUser()
		now := time.Now()
		data.CreatedAt = now
		data.UpdatedAt = now

		if err := c.BindJSON(&data); err != nil {
			c.String(http.StatusBadRequest, "Request is failed: "+err.Error())
		}
		db.NewRecord(data)
		db.Create(&data)
		if db.NewRecord(data) == false {
			c.JSON(http.StatusOK, data)
		}
	})

	//READ
	//全レコード
	r.GET("/users", func(c *gin.Context) {
		users := []User{}
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})
	//1レコード
	r.GET("/user/:id", func(c *gin.Context) {
		user := NewUser()
		id := c.Param("id")

		db.Where("ID = ?", id).First(&user)
		c.JSON(http.StatusOK, user)
	})

	//UPDATE
	r.PUT("/user/:id", func(c *gin.Context) {
		user := NewUser()
		id := c.Param("id")

		data := NewUser()
		if err := c.BindJSON(&data); err != nil {
			c.String(http.StatusBadRequest, "Request is failed: "+err.Error())
		}

		db.Where("ID = ?", id).First(&user).Updates(&data)
	})

	//DELETE
	r.DELETE("/user/:id", func(c *gin.Context) {
		user := NewUser()
		id := c.Param("id")

		db.Where("ID = ?", id).Delete(&user)
	})

	return r
}

// @description テスト用APIの詳細
// @version 1.0
// @accept application/x-json-stream
// @param none query string false "必須ではありません。"
// @Success 200 {object} gin.H {"code":200,"msg":"ok"}
// @router /test/ [get]
func test(c *gin.Context){
    c.JSON(http.StatusOK, gin.H{ "msg": "ok"})
}

// @title gin-swagger todos 
// @version 1.0
// @license.name kosuke
// @description このswaggerはgin-swaggerの見本apiです
func main() {
	loadEnv()
	db := gormConnect()
	r := setRouter(db)

	defer db.Close()

	db.Set("gorm:table_options", "ENGINE=InnoDB")
	db.AutoMigrate(&User{})
	db.LogMode(true)

	r.Run(":8082")
}
