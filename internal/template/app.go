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
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/config"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/constant"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/data/sqlxext"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/httpext"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/middleware"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/router"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/s3ext"
	"github.com/tanveerprottoy/stdlib-go-template/internal/pkg/validatorext"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/auth"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/content"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/fileupload"
	"github.com/tanveerprottoy/stdlib-go-template/internal/template/module/user"
	modulerouter "github.com/tanveerprottoy/stdlib-go-template/internal/template/router"
	"github.com/tanveerprottoy/stdlib-go-template/pkg/file"
)

// App struct
type App struct {
	Server             *http.Server
	idleConnsClosed    chan struct{}
	DBClient           *sqlxext.Client
	ClientsS3          *s3ext.Clients
	HTTPClientProvider *httpext.ClientProvider
	router             *router.Router
	Middlewares        []any
	AuthModule         *auth.Module
	UserModule         *user.Module
	ContentModule      *content.Module
	FileUploadModule   *fileupload.Module
	Validate           *validator.Validate
}

// NewApp creates App
func NewApp() *App {
	a := new(App)
	a.initComponents()
	a.initServer()
	a.configureGracefulShutdown()
	return a
}

// initDB initializes DB client
func (a *App) initDB() {
	a.DBClient = sqlxext.GetInstance()
}

// // createDir creates uploads directory
func (a *App) createDir() {
	file.CreateDirIfNotExists("./uploads")
}

func (a *App) initRouter() {
	a.router = router.NewRouter()
}

// initS3 initializes s3
func (a *App) initS3() {
	s3Region := config.GetEnvValue("S3_REGION")
	s3Endpoint := config.GetEnvValue("S3_ENDPOINT")
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
	a.ClientsS3 = s3ext.GetInstance()
	a.ClientsS3.Init(s3.Options{
		Region: config.GetEnvValue("S3_REGION"),
		// Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(config.GetEnvValue("S3_ACCESS_KEY"), config.GetEnvValue("S3_SECRET_KEY"), "")),
		// EndpointResolver: endpointResolverFunc,
		// UsePathStyle: true,
	}, func(o *s3.Options) {
		o.EndpointResolver = endpointResolverFunc
		o.UsePathStyle = true
	})
	// tmp create bucket
	/* s3ext.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(config.GetEnvValue("BUCKET_NAME")),
	}, a.ClientsS3.S3Client, context.Background()) */
}

func (a *App) initHTTPClientProvider() {
	a.HTTPClientProvider = httpext.NewClientProvider(90*time.Second, nil, nil)
}

// initValidator initializes validator
func (a *App) initValidator() {
	a.Validate = validator.New()
	validatorext.RegisterTagNameFunc(a.Validate)
	_ = a.Validate.RegisterValidation("notempty", validatorext.NotEmpty)
}

// initModules initializes application modules
func (a *App) initModules() {
	a.UserModule = user.NewModule(a.DBClient.DB, a.Validate)
	a.ContentModule = content.NewModule(a.DBClient.DB, a.Validate)
	a.AuthModule = auth.NewModule(a.UserModule.Service)
	a.FileUploadModule = fileupload.NewModule(a.ClientsS3)
}

// initMiddlewares initializes middlewares
func (a *App) initMiddlewares() {
	am := middleware.NewAuth(a.AuthModule.Service)
	rm := middleware.NewRBAC(a.AuthModule.Service)
	a.Middlewares = append(a.Middlewares, am)
	a.Middlewares = append(a.Middlewares, rm)
}

// initModuleRouters initializes module routers and routes
func (a *App) initModuleRouters() {
	m := a.Middlewares[0].(*middleware.Auth)
	r := a.Middlewares[1].(*middleware.RBAC)
	modulerouter.RegisterUserRoutes(a.router, constant.V1, a.UserModule, m)
	modulerouter.RegisterContentRoutes(a.router, constant.V1, a.ContentModule, m, r)
	modulerouter.RegisterFileUploadRoutes(a.router, constant.V1, a.FileUploadModule)
}

// initServer initializes the server
func (a *App) initServer() {
	a.Server = &http.Server{
		Addr:    ":" + config.GetEnvValue("APP_PORT"),
		Handler: a.router.Mux,
	}
}

// configureGracefulShutdown configures graceful shutdown
func (a *App) configureGracefulShutdown() {
	// code to support graceful shutdown
	a.idleConnsClosed = make(chan struct{})
	go func() {
		// this func listens for SIGINT and initiates
		// shutdown when SIGINT is received
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		// We received an interrupt signal, shut down.
		log.Printf("Received an interrupt signal")
		if err := a.Server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP Server shutdown error: %v", err)
		}
		close(a.idleConnsClosed)
	}()
}

// ShutdownServer shuts down the server
func (a *App) ShutdownServer(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := a.Server.Shutdown(ctx); err != nil {
		panic(err)
	} else {
		log.Println("Server shutdown")
		// add code
	}
}

// initComponents initializes application components
func (a *App) initComponents() {
	a.initDB()
	a.createDir()
	a.initRouter()
	a.initS3()
	a.initValidator()
	a.initModules()
	a.initMiddlewares()
	a.initModuleRouters()
}

// Run runs the server
func (a *App) Run() {
	// if err == http.ErrServerClosed do nothing
	if err := a.Server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP Server ListenAndServe: %v", err)
	}
	<-a.idleConnsClosed
	log.Println("Server shutdown")
}

// RunTLS runs the server with TLS
func (a *App) RunTLS() {
	// if err == http.ErrServerClosed do nothing
	if err := a.Server.ListenAndServeTLS("cert.crt", "key.key"); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTPS server ListenAndServe: %v", err)
	}
	<-a.idleConnsClosed
}

// RunListenAndServe runs the server
func (a *App) RunListenAndServe() {
	err := http.ListenAndServe(":"+config.GetEnvValue("APP_PORT"), a.router.Mux)
	if err != nil {
		panic(err)
	}
}

// RunListenAndServeTLS runs the server with TLS
func (a *App) RunListenAndServeTLS() {
	err := http.ListenAndServeTLS(":443", "cert.crt", "key.key", a.router.Mux)
	if err != nil {
		panic(err)
	}
}
