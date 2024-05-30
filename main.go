package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"


	
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/joho/godotenv"
	d "disbursement/init"
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {

	var ctx = context.Background()

	err := godotenv.Load()
	if err != nil {
		logrus.Fatalf("Error loading .env file: %s\n", err)
	}
  
	isUsingDatadog := os.Getenv("IS_USING_DATADOG")
  
	if (isUsingDatadog!="FALSE") {
		logrus.Info("is using datadog")
	}

	d.CreateAndOpen("user")
	
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	r := gin.Default()

	r.ForwardedByClientIP = true
	r.Use(CORSMiddleware())

	// Connect to the PostgreSQL database
	connStr := "user=myuser dbname=mydb password=mypassword sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logrus.Fatalf("listen db: %s\n", err)
		panic(err)
	}
	defer db.Close()

	// r.SetTrustedProxies([]string{"127.0.0.1", "0.0.0.0/8080"})
	api := r.Group("/v1")
    {
		api.GET("/users", func(c *gin.Context) {
            c.String(http.StatusOK, "API endpoint")
        })
	}
	r.GET("/getData", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello world!")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	
	r.GET("/healthcheck", HealthCheck)


	// handling metrics by prometheus
	recordMetrics()
	r.GET("/metrics", prometheusHandler())


	// Listen and Server in 0.0.0.0:8080
	logrus.WithField("addr", ":8080").Info("starting server")
	if err := r.Run(":8080"); err != nil {
		logrus.Fatalf("listen: %s\n", err)
	}

	err = rdb.Set(ctx, "server-run:", "value", 0).Err()
	if err != nil {
		logrus.Fatalf("listen redis err: %s\n", err)
		panic(err)
	}

	go gracefulShutdown()
	forever := make(chan int)
	<-forever

}

func gracefulShutdown() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		<-s
		fmt.Println("Sutting down gracefully.")
		// clean up here
		os.Exit(0)
	}()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func HealthCheck(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "API is up and working fine",
	})
}