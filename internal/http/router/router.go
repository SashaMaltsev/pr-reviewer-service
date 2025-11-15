package router

import (
	"net/http"

	handler "github.com/SashaMalcev/pr-reviewer-service/internal/http/handlers"
	custom_middleware "github.com/SashaMalcev/pr-reviewer-service/internal/http/middleware"
	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
	"github.com/go-chi/chi/v5"
)

func New(teamService *service.TeamService, userService *service.UserService, 
	prService *service.PRService, statsService *service.StatsService) http.Handler {
    
	r := chi.NewRouter()

    r.Use(custom_middleware.Recovery)
    r.Use(custom_middleware.Logger)

	// handlers
    teamHandler 	:= handler.NewTeamHandler(teamService)
    userHandler 	:= handler.NewUserHandler(userService)
    prHandler 		:= handler.NewPRHandler(prService)
    statsHandler 	:= handler.NewStatsHandler(statsService)
    healthHandler 	:= handler.NewHealthHandler()

    // routes
    r.Route("/team", func(r chi.Router) {
        r.Post("/add", teamHandler.CreateTeam)
        r.Get("/get", teamHandler.GetTeam)
    })

    r.Route("/users", func(r chi.Router) {
        r.Post("/setIsActive", userHandler.SetIsActive)
        r.Get("/getReview", userHandler.GetReviews)
    })

    r.Route("/pullRequest", func(r chi.Router) {
        r.Post("/create", prHandler.CreatePR)
        r.Post("/merge", prHandler.MergePR)
        r.Post("/reassign", prHandler.ReassignReviewer)
    })

    r.Get("/stats/assignments", statsHandler.GetAssignmentStats)
    r.Get("/health", healthHandler.Check)

    return r
}