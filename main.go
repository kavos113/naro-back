package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/srinathgs/mysqlstore"
	"github.com/traPtitech/naro-template-backend/handler"
)

func main() {
	// .envファイルから環境変数を読み込み
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	// データーベースの設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}
	conf := mysql.Config{
		User:      os.Getenv("DB_USERNAME"),
		Passwd:    os.Getenv("DB_PASSWORD"),
		Net:       "tcp",
		Addr:      os.Getenv("DB_HOSTNAME") + ":" + os.Getenv("DB_PORT"),
		DBName:    os.Getenv("DB_DATABASE"),
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
	}

	// データベースに接続
	db, err := sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// create users table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (Username VARCHAR(255) PRIMARY KEY, HashedPass VARCHAR(255))")
	if err != nil {
		log.Fatal(err)
	}

	store, err := mysqlstore.NewMySQLStoreFromConnection(db.DB, "sessions", "/", 60*60*24*14, []byte("secret-token"))
	if err != nil {
		log.Fatal(err)
	}

	h := handler.NewHandler(db)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	e.POST("/signup", h.SignUpHandler)
	e.POST("/login", h.LoginHandler)
	e.GET("/ping", func (c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	withAuth := e.Group("")
	withAuth.Use(handler.UserAuthMiddleware)
	withAuth.GET("/me", handler.GetMeHandler)
	withAuth.GET("/cities/:cityName", h.GetCityInfoHandler)
	withAuth.POST("/cities", h.PostCityHandler)
	withAuth.GET("/allcities", h.GetAllCityHandler)

	err = e.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
