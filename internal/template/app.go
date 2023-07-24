package template

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/middleware"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/router"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/auth"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/fileupload"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/user"
	configpkg "github.com/tanveerprottoy/stdlib-go-template/pkg/config"
	"github.com/tanveerprottoy/stdlib-go-template/pkg/data/sqlxpkg"
	"github.com/tanveerprottoy/stdlib-go-template/pkg/file"
	"github.com/tanveerprottoy/stdlib-go-template/pkg/s3pkg"
)

// App struct
type App struct {
	Server           *http.Server
	idleConnsClosed  chan struct{}
	DBClient         *sqlxpkg.Client
	router           *router.Router
	Middlewares      []any
	AuthModule       *auth.Module
	UserModule       *user.Module
	FileUploadModule *fileupload.Module
	Validate         *validator.Validate
	ClientsS3        *s3pkg.Clients
}

func NewApp() *App {
	a := new(App)
	a.initComponents()
	return a
}

func (a *App) initDB() {
	a.DBClient = sqlxpkg.GetInstance()
}

func (a *App) initDir() {
	file.CreateDirIfNotExists("./uploads")
}

func (a *App) initS3() {
	s3Region := configpkg.GetEnvValue("S3_REGION")
	s3Endpoint := configpkg.GetEnvValue("S3_ENDPOINT")
	endpointResolverFunc := s3.EndpointResolverFunc(func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
		if s3Endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           s3Endpoint,
				SigningRegion: s3Region,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	a.ClientsS3 = s3pkg.GetInstance()
	a.ClientsS3.Init(s3.Options{
		Region: configpkg.GetEnvValue("S3_REGION"),
		// Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(config.GetEnvValue("S3_ACCESS_KEY"), config.GetEnvValue("S3_SECRET_KEY"), "")),
		// EndpointResolver: endpointResolverFunc,
		// UsePathStyle: true,
	}, func(o *s3.Options) {
		o.EndpointResolver = endpointResolverFunc
		o.UsePathStyle = true
	})
}

func (a *App) initMiddlewares() {
	authMiddleWare := middleware.NewAuthMiddleware(a.AuthModule.Service)
	a.Middlewares = append(a.Middlewares, authMiddleWare)
}

func (a *App) initModules() {
	a.UserModule = user.NewModule(a.DBClient.DB, a.Validate)
	a.AuthModule = auth.NewModule(a.UserModule.Service)
	a.FileUploadModule = fileupload.NewModule(a.ClientsS3)
}

func (a *App) initModuleRouters() {
	m := a.Middlewares[0].(*middleware.AuthMiddleware)
	router.RegisterUserRoutes(a.router, constant.V1, a.UserModule, m)
	router.RegisterFileUploadRoutes(a.router, constant.V1, a.FileUploadModule)
}

// Init app
func (a *App) initComponents() {
	a.initDB()
	a.initDir()
	a.router = router.NewRouter()
	a.initS3()
	a.initModules()
	a.initMiddlewares()
	a.initModuleRouters()
	a.initServer()
}

func (a *App) initServer() {
	a.Server = &http.Server{
		Addr:    ":" + configpkg.GetEnvValue("APP_PORT"),
		Handler: a.router.Mux,
	}
	// code to support graceful shutdown
	a.idleConnsClosed = make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		// We received an interrupt signal, shut down.
		if err := a.Server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(a.idleConnsClosed)
	}()
}

func (a *App) ShutdownServer(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := a.Server.Shutdown(ctx); err != nil {
		panic(err)
	} else {
		log.Println("application shutdowned")
		// add code
	}
}

// Run app
func (a *App) Run() {
	if err := a.Server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-a.idleConnsClosed
	log.Println("shutdown")
}

// Run app with TLS
func (a *App) RunTLS() {
	if err := a.Server.ListenAndServeTLS("cert.crt", "key.key"); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTPS server ListenAndServe: %v", err)
	}
	<-a.idleConnsClosed
}

// Run app
func (a *App) RunListenAndServe() {
	err := http.ListenAndServe(":"+configpkg.GetEnvValue("APP_PORT"), a.router.Mux)
	if err != nil {
		panic(err)
	}
}

// Run app
func (a *App) RunListenAndServeTLS() {
	err := http.ListenAndServeTLS(":443", "cert.crt", "key.key", a.router.Mux)
	if err != nil {
		panic(err)
	}
}