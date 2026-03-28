package routes

import (
	"backend-go/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, interviewCtrl *controllers.InterviewController, cvCtrl *controllers.CVController, audioCtrl *controllers.AudioController) {
	api := r.Group("/api")
	{
		interview := api.Group("/interview")
		{
			interview.POST("/start", interviewCtrl.StartInterview)
			interview.POST("/respond", interviewCtrl.RespondToInterview)
			interview.POST("/respond-audio", interviewCtrl.RespondToInterviewAudio)
			interview.POST("/feedback", interviewCtrl.GenerateFeedback)
		}
		cv := api.Group("/cv")
		{
			cv.POST("/upload", cvCtrl.Upload)
		}
		audio := api.Group("/audio")
		{
			audio.POST("/transcribe", audioCtrl.Transcribe)
		}
	}
}
