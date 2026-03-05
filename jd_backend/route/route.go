package route

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"jd/controller"
	"jd/service"
	"jd/service/chess"

	"jd/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	origin := os.Getenv("FRONTEND_URL")
	if origin == "" {
		origin = "http://localhost:5173"
	}

	// 设置跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{origin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
	r.Use(middleware.AuthMiddleware())

	r.Static("/media", "./uploads")

	user := controller.NewUserController(service.NewUserService())
	media := controller.NewMediaController(service.NewMediaService())
	post := controller.NewPostController(service.NewPostService())

	// 象棋 WebSocket 路由（全局共享一个房间管理器）
	chessCtrl := controller.NewChessController(chess.NewManager())
	r.GET("/ws/chess", chessCtrl.HandleWS)

	// 设置路由组
	api := r.Group("/api")
	api.POST("/info", user.GetUserInfo)

	mediaRoute := api.Group("/media")
	mediaRoute.POST("/upload", media.UploadImage)

	postRoute := api.Group("/post")
	postRoute.POST("/create", post.CreatePost)

	publicRoute := api.Group("/public")

	publicRoute.POST("/register", user.Register)
	publicRoute.POST("/login", user.Login)
	publicRoute.POST("/reset", user.ResetPassword)
	publicRoute.POST("/send-code", user.SendVCode)

	publicRoute.POST("/posts", post.GetPostsList)
	publicRoute.GET("/posts", post.GetMaxId)
	publicRoute.POST("/posts-by-id", post.GetPostById)

	return r
}
