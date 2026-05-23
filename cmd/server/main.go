package main

import (
	"log"

	"ops-platform/config"
	"ops-platform/internal/database"
	"ops-platform/internal/handler"
	"ops-platform/internal/repository"
	"ops-platform/internal/router"
	"ops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	gin.SetMode(cfg.Server.Mode)

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	completionRepo := repository.NewCompletionRepository(db)

	// Services
	userService := service.NewUserService(userRepo, cfg)
	teamService := service.NewTeamService(teamRepo, userRepo)
	ticketService := service.NewTicketService(ticketRepo)
	projectService := service.NewProjectService(projectRepo)
	completionService := service.NewCompletionService(completionRepo)
	ticketService.SetCompletionService(completionService)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	teamHandler := handler.NewTeamHandler(teamService)
	ticketHandler := handler.NewTicketHandler(ticketService, userService)
	projectHandler := handler.NewProjectHandler(projectService)
	completionHandler := handler.NewCompletionHandler(completionService)

	// Router
	r := router.Setup(userHandler, teamHandler, ticketHandler, projectHandler, completionHandler)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
