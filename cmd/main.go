package main

import "weblineBackend/pkg/logger"

func main() {
	// initialize the logger
	logger.Init()
	logger.Info("Starting the application...")

	//cfg := config.LoadConfig()
	//
	//// initialize database connection
	//conn, err := config.NewDatabaseConnection(cfg.DbUrl)
	//if err != nil {
	//	log.Printf("Error connecting to database: %v", err)
	//}
	//
	//db := database.New(conn)
	//
	//// setup routes
	//r := mux.NewRouter()
	//r.Use(middleware.CORS)
	//
	//// start server
	//log.Printf("Server listening on port %s", cfg.Port)
	//log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
