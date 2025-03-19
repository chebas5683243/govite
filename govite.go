package govite

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/chebas5683243/govite/render"
	"github.com/chebas5683243/govite/session"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

type GoVite struct {
	Appname  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	Session  *scs.SessionManager
	JetViews *jet.Set
	config   config
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
}

func (gv *GoVite) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := gv.Init(pathConfig)
	if err != nil {
		return err
	}

	err = gv.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := gv.startLoggers()
	gv.InfoLog = infoLog
	gv.ErrorLog = errorLog
	gv.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	gv.Version = version
	gv.RootPath = rootPath
	gv.Routes = gv.routes().(*chi.Mux)

	gv.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
	}

	// create session

	sess := session.Session{
		CookieLifetime: gv.config.cookie.lifetime,
		CookiePersist:  gv.config.cookie.persist,
		CookieName:     gv.config.cookie.name,
		CookieDomain:   gv.config.cookie.domain,
		SessionType:    gv.config.sessionType,
	}

	gv.Session = sess.InitSession()

	gv.JetViews = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	gv.createRenderer()

	return nil
}

func (gv *GoVite) Init(p initPaths) error {
	root := p.rootPath

	for _, path := range p.folderNames {
		// create folder if it doesnt exist
		err := gv.CreateDirIfNotExists(root + "/" + path)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListenAndServe starts the web server
func (gv *GoVite) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", gv.config.port),
		ErrorLog:     gv.ErrorLog,
		Handler:      gv.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	gv.InfoLog.Printf("Listening on port %s", gv.config.port)
	err := srv.ListenAndServe()
	gv.ErrorLog.Fatal(err)
}

func (gv *GoVite) checkDotEnv(path string) error {
	err := gv.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}

	return nil
}

func (gv *GoVite) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (gv *GoVite) createRenderer() {
	gv.Render = &render.Render{
		Renderer: gv.config.renderer,
		RootPath: gv.RootPath,
		Port:     gv.config.port,
		JetViews: gv.JetViews,
	}
}
