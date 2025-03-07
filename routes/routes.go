package routes

import (
	"context"
	controller "jwt2/controller/src"
	_ "jwt2/docs"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

func RegisterRoutes(r *gin.Engine, userControl *controller.UserController, refreshToken *controller.RefreshTokenController) {
	r.POST("/register", userControl.Register)
	r.POST("/login", userControl.Login)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/refresh", refreshToken.RefreshToken)

	protected := r.Group("/users")
	protected.Use(middleware.AuthMiddleware())
	{
		// protected.GET("/tasks", taskController.GetAllTask)
		//	protected.POST("/tasks", taskControl.CreateTask)
		//  protected.GET("/tasks/:id", taskControl.GetTaskByID)
		protected.PUT("/profile/update/:id", userControl.UpdateUser)
		//  protected.PUT("/tasks/update/:id", taskControl.UpdateTaskByID)
		//	protected.DELETE("/tasks/delete/:id", taskControl.DeleteTaskByID)
		protected.DELETE("/logout", userControl.Logout)
		// conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		// if err != nil {
		// 	log.Fatalf("failed to connect gRPC: %v", err)
		// }
		// taskClient := v1.NewTaskServiceClient(conn)
		// protected.GET("/tasks", func(c *gin.Context){
		// 	userID :=c.Query("user_id")
		// 	req :=&v1.GetTaskByIdRequest{Id: userID}
		// 	resp, err
		// }
		//=================================================TEST==============================================================//
		protected.GET("/test-grpc", func(c *gin.Context) {
			conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to connect gRPC"})
				return
			}
			defer conn.Close()

			grpcClient := v1.NewTaskServiceClient(conn)
			ctx := context.Background()

			resp, err := grpcClient.GetTaskById(ctx, &v1.GetTaskByIdRequest{Id: 1})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"task": resp})
		})
		//=================================================TEST==============================================================//

	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware("ADMIN"))
	{
		admin.GET("/users/:id", userControl.GetUserByID)
		admin.GET("/users", userControl.GetAllUser)
		admin.DELETE("/user/delete/:id", userControl.DeleteUserById)
	}
}
