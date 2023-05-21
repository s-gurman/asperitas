package main

import (
	"flag"
	"net/http"

	"asperitas/internal/handlers"
	"asperitas/internal/middleware"
	"asperitas/internal/post"
	"asperitas/internal/session"
	"asperitas/internal/user"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	mySQLAddr = flag.String("mySQLAddr", `root:admin@tcp(localhost:3306)/vk-go?charset=utf8&interpolateParams=true`, "mysql addr")
	mongoAddr = flag.String("mongoAddr", `mongodb://localhost:27017`, "mongo addr")
)

func main() {
	flag.Parse()

	sessionManager, err := session.NewManagerMySQL(*mySQLAddr)
	panicOnErr(err)

	userRepo, err := user.NewRepoMySQL(*mySQLAddr)
	panicOnErr(err)

	postRepo, err := post.NewRepoMongo(*mongoAddr)
	panicOnErr(err)

	zapLogger, err := zap.NewProduction()
	panicOnErr(err)

	defer zapLogger.Sync() // nolint:errcheck
	logger := zapLogger.Sugar()

	usersHandler := &handlers.UserHandler{
		Sess:   sessionManager,
		Repo:   userRepo,
		Logger: logger,
	}

	postsHandler := &handlers.PostHandler{
		Sess:   sessionManager,
		Repo:   postRepo,
		Logger: logger,
	}

	r := router(usersHandler, postsHandler)
	mux := middleware.AccessLog(logger, r)
	mux = middleware.Panic(mux)

	addr := ":8080"
	logger.Infow("starting server",
		"type", "START",
		"addr", addr,
	)

	err = http.ListenAndServe(addr, mux)
	panicOnErr(err)
}

func router(usersHandler *handlers.UserHandler, postsHandler *handlers.PostHandler) *mux.Router {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	))

	r.HandleFunc("/api/register", usersHandler.Register).Methods("POST")
	r.HandleFunc("/api/login", usersHandler.Login).Methods("POST")

	r.HandleFunc("/api/posts/", postsHandler.ListPosts).Methods("GET")
	r.HandleFunc("/api/posts", postsHandler.CreatePost).Methods("POST")
	r.HandleFunc("/api/posts/{categoryName}", postsHandler.ListPostsByCategory).Methods("GET")
	r.HandleFunc("/api/post/{postID}", postsHandler.ShowPost).Methods("GET")
	r.HandleFunc("/api/post/{postID}", postsHandler.DeletePost).Methods("DELETE")
	r.HandleFunc("/api/post/{postID}", postsHandler.CreateComment).Methods("POST")
	r.HandleFunc("/api/post/{postID}/{commentID}", postsHandler.DeleteComment).Methods("DELETE")
	r.HandleFunc("/api/post/{postID}/upvote", postsHandler.UpvotePost).Methods("GET")
	r.HandleFunc("/api/post/{postID}/downvote", postsHandler.DownvotePost).Methods("GET")
	r.HandleFunc("/api/post/{postID}/unvote", postsHandler.UnvotePost).Methods("GET")
	r.HandleFunc("/api/user/{userName}", postsHandler.ListPostsByUser).Methods("GET")

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/html/index.html")
	}).Methods("GET")

	return r
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
