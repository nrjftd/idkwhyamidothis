package middleware

import (
	"context"
	"fmt"
	"jwt2/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func AuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		userType, ok := claims["user_type"].(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden",
			})
			c.Abort()
			return
		}
		if len(requiredRoles) > 0 && !isRoleAuthorized(userType, requiredRoles) {
			authorized := false
			for _, role := range requiredRoles {
				if userType == role {
					authorized = true
					break
				}
			}
			if !authorized {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Forbidden",
				})
				c.Abort()
				return
			}
		}
		c.Set("user_id", fmt.Sprintf("%v", claims["user_id"]))
		c.Set("user_type", userType)
		c.Next()
	}

}

func isRoleAuthorized(userType string, requiredRoles []string) bool {
	for _, role := range requiredRoles {
		if userType == role {
			return true
		}
	}
	return false
}
func AuthInterceptor(requiredRoles ...string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("missing metadata")
		}
		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, fmt.Errorf("missing authorization header")
		}
		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			return nil, fmt.Errorf("unauthorized: invalid token, %s", err) //
		}
		userType, ok := claims["user_type"].(string)
		if !ok {
			return nil, fmt.Errorf("forbidden: user_type not found in token")
		}
		if len(requiredRoles) > 0 && !isRoleAuthorized(userType, requiredRoles) {
			return nil, fmt.Errorf("forbidden: insufficient permissions")
		}
		ctx = context.WithValue(ctx, "user_id", fmt.Sprintf("%v", claims["user_id"]))
		ctx = context.WithValue(ctx, "user_type", userType)
		return handler(ctx, req)
	}
}
func RoleInterceptor(apiRoles map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userType, ok := ctx.Value("user_type").(string)
		if !ok {
			return nil, fmt.Errorf("forbidden: missing user_type")
		}
		requiredRoles, exists := apiRoles[info.FullMethod]
		if exists && !isRoleAuthorized(userType, requiredRoles) {
			return nil, fmt.Errorf("forbidden: insufficient permission")
		}
		return handler(ctx, req)
	}
}

func AuthStreamInterceptor(requiredRoles ...string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Println("Missing metadata in stream")
			return fmt.Errorf("missing metadata")
		}
		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			log.Println("Missing authorization header in stream")
			return fmt.Errorf("missing authorization header")
		}
		token := strings.TrimPrefix(authHeader[0], "Bearer ")

		claims, err := utils.ValidateJWT(token)
		if err != nil {
			log.Printf("token validation error in stream, am: %v\n", err)
			return fmt.Errorf("unauthorized: invalid token, %s", err)
		}
		userId, userType := claims["user_id"], claims["user_type"]
		if userId == nil {
			log.Println("user_id not found (stream)")
			return fmt.Errorf("not found user id")
		}
		log.Println("AuthInterceptor - Extracted user_id:", claims["user_id"])

		ctx = context.WithValue(ctx, "user_id", fmt.Sprintf("%v", userId))
		ctx = context.WithValue(ctx, "user_type", fmt.Sprintf("%v", userType))
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
