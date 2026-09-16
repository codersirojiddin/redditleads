package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/codersirojiddin/reddit-leads/internal/ai"
	"github.com/codersirojiddin/reddit-leads/internal/auth"
	"github.com/codersirojiddin/reddit-leads/internal/config"
	"github.com/codersirojiddin/reddit-leads/internal/database"
	"github.com/codersirojiddin/reddit-leads/internal/discovery"
	apphttp "github.com/codersirojiddin/reddit-leads/internal/http"
	"github.com/codersirojiddin/reddit-leads/internal/jobs"
	"github.com/codersirojiddin/reddit-leads/internal/keywords"
	"github.com/codersirojiddin/reddit-leads/internal/opportunities"
	"github.com/codersirojiddin/reddit-leads/internal/projects"
	"github.com/codersirojiddin/reddit-leads/internal/reports"
	"github.com/codersirojiddin/reddit-leads/internal/users"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	// --- wire dependencies -------------------------------------------------
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTExpiryHrs)
	authHandler := auth.NewHandler(authSvc)

	usersRepo := users.NewRepository(pool)
	usersSvc := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersSvc)

	projectsRepo := projects.NewRepository(pool)
	projectsSvc := projects.NewService(projectsRepo)
	projectsHandler := projects.NewHandler(projectsSvc)

	keywordsRepo := keywords.NewRepository(pool)
	keywordsSvc := keywords.NewService(keywordsRepo)
	keywordsHandler := keywords.NewHandler(keywordsSvc, projectsSvc)

	redditProvider := discovery.NewProvider(cfg.RedditClientID, cfg.RedditClientSecret, cfg.RedditUserAgent)
	aiClient := ai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	classifier := ai.NewClassifier(aiClient)

	oppsRepo := opportunities.NewRepository(pool)
	oppsSvc := opportunities.NewService(oppsRepo, redditProvider, classifier, cfg.MaxPostsPerKeyword)

	reportsRepo := reports.NewRepository(pool)
	reportsSvc := reports.NewService(reportsRepo, oppsRepo)
	reportsHandler := reports.NewHandler(reportsSvc, projectsSvc)

	// Run the worker loop in-process by default. This lets the whole app
	// (API + background report processing) run as a single deployable
	// service — handy for free-tier hosts (Render, Fly, Railway, etc.) that
	// only give you one free web service. Set RUN_WORKER_INMEMORY=false and
	// run `go run ./apps/worker` as a separate process instead if you'd
	// rather scale them independently.
	if cfg.RunWorkerInProcess {
		queue := jobs.NewQueue(reportsRepo)
		reportJob := jobs.NewReportJob(projectsRepo, keywordsRepo, oppsSvc, reportsRepo)
		worker := jobs.NewWorker(queue, reportJob, time.Duration(cfg.WorkerPollIntervalSec)*time.Second)
		go worker.Run(ctx)
	}

	router := apphttp.NewRouter(apphttp.RouterDeps{
		AllowedOrigin: cfg.FrontendURL,

		AuthMiddleware: auth.Middleware(authSvc),

		RegisterHandler: authHandler.Register,
		LoginHandler:    authHandler.Login,

		MeHandler: usersHandler.Me,

		CreateProjectHandler: projectsHandler.Create,
		ListProjectsHandler:  projectsHandler.List,
		GetProjectHandler:    projectsHandler.Get,
		DeleteProjectHandler: projectsHandler.Delete,

		CreateKeywordHandler: keywordsHandler.Create,
		ListKeywordsHandler:  keywordsHandler.List,
		DeleteKeywordHandler: keywordsHandler.Delete,

		CreateReportHandler: reportsHandler.Create,
		ListReportsHandler:  reportsHandler.List,
		GetReportHandler:    reportsHandler.Get,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("reddit-leads API listening on :%s (env=%s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down api server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
