package main

import (
	controller "jwt2/controller/src"
	"jwt2/database"
	grpcServer "jwt2/gRPCService"
	"jwt2/models"
	"jwt2/pubsub"
	repo "jwt2/repo/src"
	"jwt2/routes"
	service "jwt2/service/src"
	"log"

	"github.com/gin-gonic/gin"
)

// @title Swagger ?? API
// @version 1.0
// @description This is a idk server
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	database.InitDB()
	if database.DB == nil {
		log.Fatal("Database connection is nil")
	}
	defer database.DB.Close()
	database.InitRedis()

	defer database.RedisClient.Close()
	models.CreateTable(database.DB)

	//pubsub
	go pubsub.SubscribeToTaskEvent(database.RedisClient)
	//

	bunDBAdapter := &repo.BunDBWrapper{DB: database.DB}
	taskRepo := repo.NewTaskRepo(bunDBAdapter)
	//RedisClient := redis.Client(*database.RedisClient)
	redisClient := database.RedisClient
	redisRepo := repo.NewRedisRepo(redisClient)
	go func() { //Goroutine chạy song song
		grpcServer.StartGRPCServer(taskRepo, redisRepo)
	}()
	refreshTokenRepo := repo.NewRefreshToken(database.DB)
	if refreshTokenRepo == nil {
		log.Fatal("Failed to initialize refresh token repo")
	} else {
		log.Println("RefreshTokenRepo initialized successfully")
	}

	refreshService := service.NewRefreshService(refreshTokenRepo, database.RedisClient)
	refreshTokenController := controller.NewRefreshTokenController(refreshService)
	userRepo := repo.NewUserRepo(database.DB)
	userService := service.NewService(userRepo, refreshService)
	userController := controller.NewUserController(userService, refreshService)

	//taskRepo := repo.NewTaskRepo(bunDBAdapter)
	//taskService := service.NewTaskService(taskRepo)
	//taskController := controller.NewTaskController(taskService)

	r := gin.Default()
	routes.RegisterRoutes(r, userController, refreshTokenController)

	r.Run(":8080")
}
