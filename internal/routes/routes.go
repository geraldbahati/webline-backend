package routes

import (
	"encoding/json"
	"net/http"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/services/i"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func SetupRouter(cfg appconfig.Config, logger *zap.Logger, handlers *handlers.Handlers, sessionService i.SessionService) *mux.Router {
	r := mux.NewRouter()

	// Add CORS middleware using rs/cors with configuration-based options.
	r.Use(middleware.CORS(&cfg, logger))

	// Initialize Middleware
	recoveryMiddleware := middleware.RecoveryMiddleware(middleware.RecoveryOptions{
		Logger:           logger,
		EnableStackTrace: true,
	})

	requestIDMiddleware := middleware.RequestIDMiddleware(logger)

	loggingMiddleware := middleware.LoggingMiddleware(logger)

	// rateLimitMiddleware := middleware.RateLimitMiddleware(
	// 	10,            // rps: 10 requests per second
	// 	20,            // burst: 20 requests
	// 	5*time.Minute, // cleanupInterval: 5 minutes
	// 	logger,        // logger: your zap.Logger instance
	// )

	cspMiddleware := middleware.CSP(middleware.CSPOptions{
		DefaultSrc:              "'self'",
		ScriptSrc:               "'self' 'unsafe-inline' https://apis.google.com",
		ObjectSrc:               "'none'",
		StyleSrc:                "'self' 'unsafe-inline'",
		ImgSrc:                  "'self' data:",
		ConnectSrc:              "'self'",
		FontSrc:                 "'self'",
		FrameSrc:                "'none'",
		MediaSrc:                "'self'",
		ReportURI:               "/csp-violation-report",
		UpgradeInsecureRequests: true,
	})

	hstsMiddleware := middleware.HSTS(middleware.HSTSOptions{
		MaxAge:            31536000, // 1 year in seconds
		IncludeSubDomains: true,
		Preload:           true,
	})

	// Apply Middleware to the Router in the correct order
	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware)
	// r.Use(rateLimitMiddleware)
	r.Use(cspMiddleware)
	r.Use(hstsMiddleware)
	r.Use(middleware.OptionalAuth(logger))
	r.Use(middleware.MetricsMiddleware(logger))

	// Health Check Endpoint
	r.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

	// Metrics Endpoint
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	// CSP Violation Report Endpoint
	// r.HandleFunc("/csp-violation-report", handlers.CSPViolationHandler).Methods(http.MethodPost)

	// Guest Token Endpoint
	r.HandleFunc("/api/guest/token", handlers.GuestHandler.GenerateGuestTokenHandler).Methods(http.MethodPost)

	// Serve static files
	r.PathPrefix("/uploads/profile/").Handler(http.StripPrefix("/uploads/profile/", http.FileServer(http.Dir("uploads/profile"))))
	r.PathPrefix("/uploads/product-image/").Handler(http.StripPrefix("/uploads/product-image/", http.FileServer(http.Dir("uploads/product-image"))))

	// Register routes
	registerUserRoutes(r, handlers, logger)
	registerCategoryRoutes(r, handlers, logger)
	registerAdminCategoryRoutes(r, handlers, logger)
	registerProductRoutes(r, handlers)
	registerAdminProductRoutes(r, handlers, logger)
	registerProductAnalyticRoutes(r, handlers)
	registerPromotionRoutes(r, handlers)
	registerAdminPromotionRoutes(r, handlers, logger)
	registerCartPromotionRoutes(r, handlers, logger, sessionService)
	registerSessionRoutes(r, cfg, handlers, logger, sessionService)
	registerAdminAnalyticsRoutes(r, handlers, logger)

	// Register order routes
	registerOrderRoutes(r, handlers, logger, sessionService)

	// Register email verification routes
	registerEmailVerificationRoutes(r, handlers, logger)

	// Search routes
	searchRouter := r.PathPrefix("/api/search").Subrouter()
	NamedHandleFunc(searchRouter, "", handlers.SearchHandler.SearchProducts, []string{http.MethodGet}, "SearchProducts")
	NamedHandleFunc(searchRouter, "/paginated", handlers.SearchHandler.SearchProductsPaginated, []string{http.MethodGet}, "SearchProductsPaginated")
	NamedHandleFunc(searchRouter, "/autocomplete", handlers.SearchHandler.AutocompleteSuggestions, []string{http.MethodGet}, "AutocompleteSuggestions")

	return r
}

// healthCheckHandler handles the health check endpoint.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// registerUserRoutes registers user-related routes.
func registerUserRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	userRouter := router.PathPrefix("/api/users").Subrouter()
	NamedHandleFunc(userRouter, "", handlers.UserHandler.ListUsers, []string{http.MethodGet}, "ListUsers")
	NamedHandleFunc(userRouter, "/{id}", handlers.UserHandler.GetUserProfile, []string{http.MethodGet}, "GetUserProfile")
	NamedHandleFunc(userRouter, "/register", handlers.UserHandler.RegisterUser, []string{http.MethodPost}, "RegisterUser")
	NamedHandleFunc(userRouter, "/refresh", handlers.UserHandler.RefreshToken, []string{http.MethodPost}, "RefreshToken")
	NamedHandleFunc(userRouter, "/login", handlers.UserHandler.LoginUser, []string{http.MethodPost}, "LoginUser")
	NamedHandleFunc(userRouter, "/reset-password", handlers.UserHandler.ResetPassword, []string{http.MethodPost}, "ResetPassword")
	NamedHandleFunc(userRouter, "/reset-password/request", handlers.UserHandler.RequestPasswordReset, []string{http.MethodPost}, "RequestPasswordReset")
	NamedHandleFunc(userRouter, "/login/google", handlers.UserHandler.LoginWithGoogle, []string{http.MethodPost}, "LoginWithGoogle")
	NamedHandleFunc(userRouter, "/login/email-verified", handlers.UserHandler.EmailVerified, []string{http.MethodPost}, "EmailVerified")
	NamedHandleFunc(userRouter, "/{id}/profile", handlers.UserHandler.GetUserInfo, []string{http.MethodGet}, "GetUserInfo")

	protected := userRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(logger))
	NamedHandleFunc(protected, "/admin-requests", handlers.UserHandler.RequestAdminRole, []string{http.MethodPost}, "RequestAdminRole")
	NamedHandleFunc(protected, "/approve", handlers.UserHandler.ApproveAdminRole, []string{http.MethodPost}, "ApproveAdminRole")
	NamedHandleFunc(protected, "/profile", handlers.UserHandler.UpdateUserProfile, []string{http.MethodPut}, "UpdateUserProfile")
}

// registerCategoryRoutes registers category-related routes.
func registerCategoryRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	// Create a subrouter for category endpoints
	categoryRouter := router.PathPrefix("/api/categories").Subrouter()

	// Register read-only public endpoints - fixed order with specific routes first
	NamedHandleFunc(categoryRouter, "", handlers.CategoryHandler.GetCategoriesHandler, []string{http.MethodGet}, "GetCategories")
	NamedHandleFunc(categoryRouter, "/tree", handlers.CategoryHandler.GetCategoryTreeHandler, []string{http.MethodGet}, "GetCategoryTree")
	NamedHandleFunc(categoryRouter, "/popular", handlers.CategoryHandler.GetPopularCategoriesHandler, []string{http.MethodGet}, "GetPopularCategories")
	NamedHandleFunc(categoryRouter, "/products/count", handlers.CategoryHandler.GetCategoriesWithProductsCountHandler, []string{http.MethodGet}, "GetCategoriesWithProductsCount")
	NamedHandleFunc(categoryRouter, "/subcategories/count", handlers.CategoryHandler.GetCategoriesWithSubcategoryCountHandler, []string{http.MethodGet}, "GetCategoriesWithSubcategoryCount")
	NamedHandleFunc(categoryRouter, "/parent", handlers.CategoryHandler.GetParentCategoriesHandler, []string{http.MethodGet}, "GetParentCategories")
	NamedHandleFunc(categoryRouter, "/parents", handlers.CategoryHandler.GetParentCategoriesHandler, []string{http.MethodGet}, "GetParentCategoriesPlural")
	NamedHandleFunc(categoryRouter, "/collections", handlers.CategoryHandler.GetCollectionCategoriesHandler, []string{http.MethodGet}, "GetCollectionCategories")
	NamedHandleFunc(categoryRouter, "/v2/hierarchy", handlers.CategoryHandler.GetV2CategoryHierarchyHandler, []string{http.MethodGet}, "GetV2CategoryHierarchy")

	// Slug-based routes
	NamedHandleFunc(categoryRouter, "/slug/{slug}/details", handlers.CategoryHandler.GetCategoryDetailsHandler, []string{http.MethodGet}, "GetCategoryDetailsHandler")
	NamedHandleFunc(categoryRouter, "/slug/{slug}/seo", handlers.CategoryHandler.GetCategorySEOHandler, []string{http.MethodGet}, "GetCategorySEOHandler")
	NamedHandleFunc(categoryRouter, "/children/{slug}", handlers.CategoryHandler.GetDirectChildrenWithStatsHandler, []string{http.MethodGet}, "GetDirectChildrenWithStats")
	NamedHandleFunc(categoryRouter, "/name/{name}", handlers.CategoryHandler.GetCategoryByNameHandler, []string{http.MethodGet}, "GetCategoryByName")
	NamedHandleFunc(categoryRouter, "/parent/{parentId}", handlers.CategoryHandler.GetCategoriesByParentIDHandler, []string{http.MethodGet}, "GetCategoriesByParentID")

	// Put wildcard ID routes last to avoid catching other routes
	NamedHandleFunc(categoryRouter, "/{id}/check", handlers.CategoryHandler.CheckCategoryExistenceHandler, []string{http.MethodHead}, "CheckCategoryExistence")
	NamedHandleFunc(categoryRouter, "/{id}", handlers.CategoryHandler.GetCategoryByIDHandler, []string{http.MethodGet}, "GetCategoryByID")

	// Admin routes that require authentication
	protectedRouter := router.PathPrefix("/api/admin/categories").Subrouter()

	// Apply authentication middleware to all admin routes
	protectedRouter.Use(middleware.Auth(logger))

	// Configure rate limiting for admin routes
	rateLimitMiddleware := middleware.RateLimitMiddleware(
		5,             // rps: 5 requests per second
		10,            // burst: 10 requests
		5*time.Minute, // cleanupInterval: 5 minutes
		logger,        // logger: your zap.Logger instance
	)
	protectedRouter.Use(rateLimitMiddleware)

	// Admin routes
	NamedHandleFunc(protectedRouter, "", handlers.CategoryHandler.CreateCategoryHandler, []string{http.MethodPost}, "CreateCategory")
	NamedHandleFunc(protectedRouter, "/{id}", handlers.CategoryHandler.SoftDeleteCategoryHandler, []string{http.MethodDelete}, "SoftDeleteCategory")
	NamedHandleFunc(protectedRouter, "/{id}/image", handlers.CategoryHandler.UploadCategoryImageHandler, []string{http.MethodPost}, "UploadCategoryImage")
	NamedHandleFunc(protectedRouter, "/batch-positions", handlers.CategoryHandler.BatchUpdateCategoryPositionsHandler, []string{http.MethodPut}, "BatchUpdateCategoryPositions")
	NamedHandleFunc(protectedRouter, "/{id}/hard", handlers.CategoryHandler.DeleteCategoryHandler, []string{http.MethodDelete}, "HardDeleteCategory")
}

// registerAdminCategoryRoutes registers admin category-related routes.
func registerAdminCategoryRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	// Create a subrouter for v2 API category endpoints
	categoryRouter := router.PathPrefix("/api/v2/categories").Subrouter()

	// Public endpoints for v2
	NamedHandleFunc(categoryRouter, "/hierarchy", handlers.CategoryHandler.GetCategoryHierarchyStatsHandler, []string{http.MethodGet}, "GetCategoryHierarchyStats")

	// Admin routes for v2 API
	adminRouter := router.PathPrefix("/api/v2/admin/categories").Subrouter()

	// Apply authentication middleware to all admin routes
	adminRouter.Use(middleware.Auth(logger))

	// Configure rate limiting for admin routes
	rateLimitMiddleware := middleware.RateLimitMiddleware(
		5,             // rps: 5 requests per second
		10,            // burst: 10 requests
		5*time.Minute, // cleanupInterval: 5 minutes
		logger,        // logger: your zap.Logger instance
	)
	adminRouter.Use(rateLimitMiddleware)

	// V2 Admin endpoints
	NamedHandleFunc(adminRouter, "/batch", handlers.CategoryHandler.BatchUpdateCategoryPositionsHandler, []string{http.MethodPut}, "BatchUpdateCategoryPositionsV2")
}

// registerProductRoutes registers product-related routes.
func registerProductRoutes(router *mux.Router, handlers *handlers.Handlers) {
	productRouter := router.PathPrefix("/api/products").Subrouter()
	productRouter.HandleFunc("/{slug}", handlers.ProductHandler.GetProductBySlugHandler).Methods(http.MethodGet)
	NamedHandleFunc(productRouter, "/{slug}/cart", handlers.ProductHandler.GetProductCartHandler, []string{http.MethodGet}, "GetProductCart")
	NamedHandleFunc(productRouter, "/{slug}/images", handlers.ProductHandler.GetProductImagesBySlugHandler, []string{http.MethodGet}, "GetProductImagesBySlug")
	NamedHandleFunc(productRouter, "/{slug}/pricing", handlers.ProductHandler.GetProductPricingBySlugHandler, []string{http.MethodGet}, "GetProductPricingBySlug")
	NamedHandleFunc(productRouter, "/{slug}/specs", handlers.ProductHandler.GetProductSpecsBySlugHandler, []string{http.MethodGet}, "GetProductSpecsBySlug")
	NamedHandleFunc(productRouter, "", handlers.ProductHandler.GetAllProductsHandler, []string{http.MethodPost}, "GetAllProducts")
	NamedHandleFunc(productRouter, "/{slug}/seo", handlers.ProductHandler.GetProductSEOHandler, []string{http.MethodGet}, "GetProductSEO")
	NamedHandleFunc(productRouter, "/all/sitemap", handlers.ProductHandler.GetAllProductSitemapHandler, []string{http.MethodGet}, "GetAllProductSitemap")
	NamedHandleFunc(productRouter, "/actions/search", handlers.ProductHandler.SearchProductsHandler, []string{http.MethodGet}, "SearchProducts")
	NamedHandleFunc(productRouter, "/category/{id}", handlers.ProductHandler.GetProductsByCategoryIDHandler, []string{http.MethodGet}, "GetProductsByCategoryID")
	NamedHandleFunc(productRouter, "/filter/{category_id}", handlers.ProductHandler.GetFilteredCategoryProducts, []string{http.MethodPost}, "GetFilteredCategoryProducts")
	NamedHandleFunc(productRouter, "/filter/all/options", handlers.ProductHandler.GetAllProductFilterOptionsHandler, []string{http.MethodGet}, "GetAllProductFilterOptions")
	NamedHandleFunc(productRouter, "/filter/options/{name}", handlers.ProductHandler.GetFilterOptionsByCategoryNameHandler, []string{http.MethodGet}, "GetFilterOptionsByCategoryName")
}

// registerAdminProductRoutes registers admin product-related routes.
func registerAdminProductRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	productRouter := router.PathPrefix("/api/v2/products").Subrouter()
	NamedHandleFunc(productRouter, "", handlers.ProductHandler.GetProductsHandler, []string{http.MethodGet}, "GetProducts")
	NamedHandleFunc(productRouter, "/{slug}/detail", handlers.ProductHandler.GetProductDetailHandler, []string{http.MethodGet}, "GetProductDetail")
	NamedHandleFunc(productRouter, "/meta-fields/{categoryID}", handlers.ProductHandler.GetProductMetaFieldsByCategoryIDHandler, []string{http.MethodGet}, "GetProductMetaFieldsByCategoryID")

	protected := productRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(logger))
	NamedHandleFunc(protected, "", handlers.ProductHandler.CreateV2ProductHandler, []string{http.MethodPost}, "CreateV2Product")
	NamedHandleFunc(protected, "", handlers.ProductHandler.DeleteProductsHandler, []string{http.MethodDelete}, "DeleteProducts")
	NamedHandleFunc(protected, "/{slug}", handlers.ProductHandler.DeleteProductHandler, []string{http.MethodDelete}, "DeleteProduct")
	NamedHandleFunc(protected, "/archive", handlers.ProductHandler.ArchiveProductsHandler, []string{http.MethodPut}, "ArchiveProducts")
	NamedHandleFunc(protected, "/draft", handlers.ProductHandler.DraftProductsHandler, []string{http.MethodPut}, "DraftProducts")
	NamedHandleFunc(protected, "/active", handlers.ProductHandler.ActivateProductsHandler, []string{http.MethodPut}, "ActivateProducts")
	NamedHandleFunc(protected, "/{slug}/archive", handlers.ProductHandler.ArchiveProductHandler, []string{http.MethodPut}, "ArchiveProduct")
}

// registerProductAnalyticRoutes registers product analytics related routes.
func registerProductAnalyticRoutes(router *mux.Router, handlers *handlers.Handlers) {
	productAnalyticRouter := router.PathPrefix("/api/product-analytics").Subrouter()
	NamedHandleFunc(
		productAnalyticRouter,
		"/best-sellers",
		handlers.ProductAnalyticHandler.GetBestSellerProducts,
		[]string{http.MethodGet},
		"GetBestSellerProducts",
	)
	NamedHandleFunc(
		productAnalyticRouter,
		"/featured",
		handlers.ProductAnalyticHandler.GetFeaturedProducts,
		[]string{http.MethodGet},
		"GetFeaturedProducts",
	)
	NamedHandleFunc(
		productAnalyticRouter,
		"/new-arrivals",
		handlers.ProductAnalyticHandler.GetNewArrivalProducts,
		[]string{http.MethodGet},
		"GetNewArrivalProducts",
	)
	NamedHandleFunc(
		productAnalyticRouter,
		"/daily-deals",
		handlers.ProductAnalyticHandler.GetDailyDealsProducts,
		[]string{http.MethodGet},
		"GetDailyDealsProducts",
	)
}

// registerPromotionRoutes registers promotion-related routes.
func registerPromotionRoutes(router *mux.Router, handlers *handlers.Handlers) {
	// Promotion routes
	promotionRouter := router.PathPrefix("/api/promotions").Subrouter()
	NamedHandleFunc(promotionRouter, "", handlers.PromotionHandler.GetPromotions, []string{http.MethodGet}, "GetPromotions")
}

// registerAdminPromotionRoutes registers admin promotion-related routes.
func registerAdminPromotionRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	adminPromotionRouter := router.PathPrefix("/api/v2/promotions").Subrouter()
	NamedHandleFunc(adminPromotionRouter, "", handlers.PromotionHandler.GetV2Promotions, []string{http.MethodGet}, "GetV2Promotions")

	protectedAdminPromotionRouter := adminPromotionRouter.PathPrefix("").Subrouter()
	protectedAdminPromotionRouter.Use(middleware.Auth(logger))
	NamedHandleFunc(protectedAdminPromotionRouter, "", handlers.PromotionHandler.CreateOrEditV2Promotion, []string{http.MethodPost}, "CreateOrEditV2Promotion")
	NamedHandleFunc(protectedAdminPromotionRouter, "/{slug}", handlers.PromotionHandler.GetPromotionDetails, []string{http.MethodGet}, "GetPromotionDetails")
	NamedHandleFunc(protectedAdminPromotionRouter, "/{slug}/delete", handlers.PromotionHandler.DeletePromotion, []string{http.MethodDelete}, "DeletePromotion")
	NamedHandleFunc(protectedAdminPromotionRouter, "/archive", handlers.PromotionHandler.ArchivePromotions, []string{http.MethodPut}, "ArchivePromotions")
	NamedHandleFunc(protectedAdminPromotionRouter, "/draft", handlers.PromotionHandler.DraftPromotions, []string{http.MethodPut}, "DraftPromotions")
	NamedHandleFunc(protectedAdminPromotionRouter, "/active", handlers.PromotionHandler.ActivatePromotions, []string{http.MethodPut}, "ActivatePromotions")
	NamedHandleFunc(protectedAdminPromotionRouter, "/delete", handlers.PromotionHandler.DeletePromotions, []string{http.MethodDelete}, "DeletePromotions")
}

// registerCartPromotionRoutes registers cart promotion-related routes.
func registerCartPromotionRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger, sessionService i.SessionService) {
	// Cart routes with session handling
	cartRouter := router.PathPrefix("/api/cart").Subrouter()
	cartRouter.Use(
		middleware.Session(logger, sessionService),
		middleware.OptionalAuth(logger), // Allow both authenticated and guest users
	)

	// Cart routes that work with both guest and authenticated users
	NamedHandleFunc(cartRouter, "/add", handlers.CartHandler.AddToCartHandler, []string{http.MethodPost}, "AddToCart")
	NamedHandleFunc(cartRouter, "/remove", handlers.CartHandler.RemoveFromCartHandler, []string{http.MethodDelete}, "RemoveFromCart")
	NamedHandleFunc(cartRouter, "/items", handlers.CartHandler.GetCartItemsHandler, []string{http.MethodGet}, "GetCartItems")
	NamedHandleFunc(cartRouter, "/items/update-quantity", handlers.CartHandler.UpdateCartItemQuantityHandler, []string{http.MethodPut}, "UpdateCartItemQuantity")
	NamedHandleFunc(cartRouter, "/clear", handlers.CartHandler.ClearCartHandler, []string{http.MethodPost}, "ClearCart")
	NamedHandleFunc(cartRouter, "/total", handlers.CartHandler.CalculateCartTotalHandler, []string{http.MethodGet}, "CalculateCartTotal")
	NamedHandleFunc(cartRouter, "/replace", handlers.CartHandler.ReplaceCartItemsHandler, []string{http.MethodPut}, "ReplaceCartItems")
}

// registerSessionRoutes registers session-related routes.
func registerSessionRoutes(r *mux.Router, cfg appconfig.Config, handlers *handlers.Handlers, logger *zap.Logger, sessionService i.SessionService) {
	// Update the route prefix to match frontend expectations
	sessionRouter := r.PathPrefix("/api/session").Subrouter()

	sessionRouter.Use(middleware.OptionalAuth(logger), middleware.CORS(&cfg, logger))

	sessionBasedRouter := sessionRouter.PathPrefix("").Subrouter()

	sessionBasedRouter.Use(
		middleware.Session(logger, sessionService),
		middleware.OptionalAuth(logger), // Allow both authenticated and guest users
		middleware.CORS(&cfg, logger),
	)

	// Routes now match frontend URLs
	NamedHandleFunc(sessionRouter, "/guest", handlers.SessionHandler.CreateGuestSession, []string{http.MethodPost}, "CreateGuestSession")
	NamedHandleFunc(sessionBasedRouter, "/merge", handlers.SessionHandler.MergeSession, []string{http.MethodPost}, "MergeSession")
}

// registerOrderRoutes registers order-related routes.
func registerOrderRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger, sessionService i.SessionService) {
	orderRouter := router.PathPrefix("/api/orders").Subrouter()

	protectedOrderRouter := orderRouter.PathPrefix("").Subrouter()
	protectedOrderRouter.Use(middleware.OptionalAuth(logger), middleware.Session(logger, sessionService))
	// Checkout endpoint (create new order)
	NamedHandleFunc(protectedOrderRouter, "/checkout", handlers.OrderHandler.CreateOrder, []string{http.MethodPost}, "CreateOrder")

	// List orders for the current user
	NamedHandleFunc(orderRouter, "", handlers.OrderHandler.ListOrders, []string{http.MethodGet}, "ListOrders")

	// Get order details
	orderRouter.HandleFunc("/{id}", handlers.OrderHandler.GetOrder).Methods(http.MethodGet)

	// Payment endpoint to pay for order (query param "method" required, e.g., ?method=mpesa)
	NamedHandleFunc(orderRouter, "/pay", handlers.OrderHandler.PayOrder, []string{http.MethodPost}, "PayOrder")

	// Cancel order endpoint
	orderRouter.HandleFunc("/{id}/cancel", handlers.OrderHandler.CancelOrder).Methods(http.MethodPost)

	// Change payment method endpoint
	orderRouter.HandleFunc("/{id}/change-payment", handlers.OrderHandler.ChangeOrderPaymentMethod).Methods(http.MethodPost)

	// M-Pesa Callback endpoint
	orderRouter.HandleFunc("/mpesa/callback", handlers.OrderHandler.HandleMpesaCallback).Methods(http.MethodPost)

	// Payment status endpoint
	orderRouter.HandleFunc("/payment/status", handlers.OrderHandler.GetPaymentStatus).Methods(http.MethodGet)
}

func registerAdminAnalyticsRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	adminAnalyticsRouter := router.PathPrefix("/api/v2/orders").Subrouter()
	protectedAdminAnalyticsRouter := adminAnalyticsRouter.PathPrefix("").Subrouter()
	protectedAdminAnalyticsRouter.Use(middleware.Auth(logger))
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/total-revenue", handlers.OrderHandler.GetTotalRevenue, []string{http.MethodGet}, "GetTotalRevenue")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/monthly-sales", handlers.OrderHandler.GetMonthlySales, []string{http.MethodGet}, "GetMonthlySales")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/monthly-revenue", handlers.OrderHandler.GetMonthlyRevenue, []string{http.MethodGet}, "GetMonthlyRevenue")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/sales-trend", handlers.OrderHandler.GetSalesTrend, []string{http.MethodGet}, "GetSalesTrend")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/recent-sales", handlers.OrderHandler.GetRecentSales, []string{http.MethodGet}, "GetRecentSales")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/total-sales-current-month", handlers.OrderHandler.GetTotalSalesCurrentMonth, []string{http.MethodGet}, "GetTotalSalesCurrentMonth")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/exchange-rate", handlers.OrderHandler.GetExchangeRate, []string{http.MethodGet}, "GetExchangeRate")
	NamedHandleFunc(protectedAdminAnalyticsRouter, "/exchange-rate", handlers.OrderHandler.UpdateExchangeRate, []string{http.MethodPut}, "UpdateExchangeRate")
}

func registerEmailVerificationRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	emailVerificationRouter := router.PathPrefix("/auth").Subrouter()
	NamedHandleFunc(emailVerificationRouter, "/verify-email", handlers.UserHandler.VerifyEmail, []string{http.MethodGet}, "VerifyEmail")
}
