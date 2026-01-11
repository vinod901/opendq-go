package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/vinod901/opendq-go/api/http"
	"github.com/vinod901/opendq-go/internal/alerting"
	"github.com/vinod901/opendq-go/internal/auth"
	"github.com/vinod901/opendq-go/internal/authorization"
	"github.com/vinod901/opendq-go/internal/check"
	"github.com/vinod901/opendq-go/internal/datasource"
	"github.com/vinod901/opendq-go/internal/lineage"
	"github.com/vinod901/opendq-go/internal/middleware"
	"github.com/vinod901/opendq-go/internal/policy"
	"github.com/vinod901/opendq-go/internal/scheduler"
	"github.com/vinod901/opendq-go/internal/tenant"
	"github.com/vinod901/opendq-go/internal/view"
	"github.com/vinod901/opendq-go/internal/workflow"
	"github.com/vinod901/opendq-go/pkg/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	log.Printf("Starting OpenDQ Control Plane on %s:%d", cfg.Server.Host, cfg.Server.Port)

	// Initialize components
	components, err := initializeComponents(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// Create HTTP handler for core platform features
	handler := api.NewHandler(
		components.tenantManager,
		components.policyManager,
		components.workflowEngine,
	)

	// Create HTTP handler for data quality features
	dqHandler := api.NewDataQualityHandler(
		components.datasourceManager,
		components.checkManager,
		components.schedulerManager,
		components.alertManager,
		components.viewManager,
	)

	// Set up router
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	dqHandler.RegisterRoutes(mux)

	// Build middleware chain
	var httpHandler http.Handler = mux

	// Add CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"})
	httpHandler = corsMiddleware.Handle(httpHandler)

	// Add authentication middleware (if OIDC is configured)
	if components.authManager != nil {
		authMiddleware := middleware.NewAuthMiddleware(components.authManager)
		httpHandler = authMiddleware.Handle(httpHandler)
	}

	// Add tenant middleware (if multi-tenant is enabled)
	if cfg.MultiTenant.Enabled {
		tenantMiddleware := middleware.NewTenantMiddleware(components.tenantManager)
		httpHandler = tenantMiddleware.Handle(httpHandler)
	}

	// Add authorization middleware (if OpenFGA is configured)
	if components.authzManager != nil {
		authzMiddleware := middleware.NewAuthzMiddleware(components.authzManager)
		httpHandler = authzMiddleware.Handle(httpHandler)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server listening on %s", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Printf("Received signal: %v", sig)
	}

	// Graceful shutdown
	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

type components struct {
	authManager       *auth.Manager
	authzManager      *authorization.Manager
	tenantManager     *tenant.Manager
	policyManager     *policy.Manager
	workflowEngine    *workflow.Engine
	lineageClient     *lineage.Client
	datasourceManager *datasource.Manager
	checkManager      *check.Manager
	schedulerManager  *scheduler.Manager
	alertManager      *alerting.Manager
	viewManager       *view.Manager
}

func initializeComponents(ctx context.Context, cfg *config.Config) (*components, error) {
	comp := &components{}

	// Initialize authentication manager (if OIDC is configured)
	if cfg.OIDC.Issuer != "" {
		authManager, err := auth.NewManager(ctx, auth.Config{
			Issuer:       cfg.OIDC.Issuer,
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			RedirectURL:  cfg.OIDC.RedirectURL,
		})
		if err != nil {
			log.Printf("Warning: failed to initialize auth manager: %v", err)
		} else {
			comp.authManager = authManager
			log.Println("Authentication manager initialized")
		}
	}

	// Initialize authorization manager (if OpenFGA is configured)
	if cfg.OpenFGA.APIHost != "" {
		authzManager, err := authorization.NewManager(authorization.Config{
			APIHost:   cfg.OpenFGA.APIHost,
			StoreID:   cfg.OpenFGA.StoreID,
			AuthModel: cfg.OpenFGA.AuthModel,
		})
		if err != nil {
			log.Printf("Warning: failed to initialize authz manager: %v", err)
		} else {
			comp.authzManager = authzManager
			log.Println("Authorization manager initialized")
		}
	}

	// Initialize tenant manager
	comp.tenantManager = tenant.NewManager()
	log.Println("Tenant manager initialized")

	// Initialize policy manager
	comp.policyManager = policy.NewManager()
	log.Println("Policy manager initialized")

	// Initialize workflow engine
	comp.workflowEngine = workflow.NewEngine()
	if err := comp.workflowEngine.RegisterStandardWorkflows(); err != nil {
		return nil, fmt.Errorf("failed to register workflows: %w", err)
	}
	log.Println("Workflow engine initialized")

	// Initialize OpenLineage client (if enabled)
	if cfg.OpenLineage.Enabled {
		comp.lineageClient = lineage.NewClient(lineage.Config{
			Endpoint:  cfg.OpenLineage.Endpoint,
			Namespace: cfg.OpenLineage.Namespace,
		})
		log.Println("OpenLineage client initialized")
	}

	// Initialize data quality components
	
	// Initialize datasource manager
	comp.datasourceManager = datasource.NewManager()
	log.Println("Datasource manager initialized")

	// Initialize alert manager
	comp.alertManager = alerting.NewManager()
	log.Println("Alert manager initialized")

	// Initialize check manager
	comp.checkManager = check.NewManager(comp.datasourceManager)
	log.Println("Check manager initialized")

	// Initialize scheduler manager
	comp.schedulerManager = scheduler.NewManager(comp.checkManager, comp.alertManager)
	log.Println("Scheduler manager initialized")

	// Initialize view manager
	comp.viewManager = view.NewManager(comp.datasourceManager)
	log.Println("View manager initialized")

	return comp, nil
}
