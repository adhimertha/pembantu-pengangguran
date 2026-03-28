package routes

import (
	"backend-go/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		interview := api.Group("/interview")
		{
			interview.POST("/start", controllers.StartInterview)
			interview.POST("/respond", controllers.RespondToInterview)
			interview.POST("/feedback", controllers.GenerateFeedback)
		}
	}
}
