package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/codersirojiddin/reddit-leads/internal/ai"
	"github.com/codersirojiddin/reddit-leads/internal/config"
	"github.com/codersirojiddin/reddit-leads/internal/database"
	"github.com/codersirojiddin/reddit-leads/internal/discovery"
	"github.com/codersirojiddin/reddit-leads/internal/jobs"
	"github.com/codersirojiddin/reddit-leads/internal/keywords"
	"github.com/codersirojiddin/reddit-leads/internal/opportunities"
	"github.com/codersirojiddin/reddit-leads/internal/projects"
	"github.com/codersirojiddin/reddit-leads/internal/reports"
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

	// The worker also ensures migrations are applied, in case it starts
	// before the API container on first boot.
	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	projectsRepo := projects.NewRepository(pool)
	keywordsRepo := keywords.NewRepository(pool)

	redditProvider := discovery.NewRedditProvider(cfg.RedditUserAgent)
	aiClient := ai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	classifier := ai.NewClassifier(aiClient)

	oppsRepo := opportunities.NewRepository(pool)
	oppsSvc := opportunities.NewService(oppsRepo, redditProvider, classifier, cfg.MaxPostsPerKeyword)

	reportsRepo := reports.NewRepository(pool)

	queue := jobs.NewQueue(reportsRepo)
	reportJob := jobs.NewReportJob(projectsRepo, keywordsRepo, oppsSvc, reportsRepo)
	worker := jobs.NewWorker(queue, reportJob, time.Duration(cfg.WorkerPollIntervalSec)*time.Second)

	log.Printf("reddit-leads worker starting (env=%s)", cfg.AppEnv)
	worker.Run(ctx)
	log.Println("worker stopped")
}
