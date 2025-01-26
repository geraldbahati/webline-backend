package routes

import (
	"net/http"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/services/i"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func SetupRouter(logger *zap.Logger, handlers *handlers.Handlers, sessionService i.SessionService) *mux.Router {
	r := mux.NewRouter()

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

	corsMiddleware := middleware.CORS(logger, middleware.CORSOptions{
		AllowedOrigins:   []string{"http://localhost:3000", "https://www.weblineshop.co.ke"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token", "Authorization"},
		AllowCredentials: true,
	})

	// Apply Middleware to the Router in the correct order
	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware)
	// r.Use(rateLimitMiddleware)
	r.Use(cspMiddleware)
	r.Use(hstsMiddleware)
	r.Use(corsMiddleware)
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
	registerCategoryRoutes(r, handlers)
	registerAdminCategoryRoutes(r, handlers, logger)
	registerProductRoutes(r, handlers)
	registerAdminProductRoutes(r, handlers, logger)
	registerPromotionRoutes(r, handlers)
	registerAdminPromotionRoutes(r, handlers, logger)
	registerAdditionalRoutes(r, handlers, logger)
	registerCartPromotionRoutes(r, handlers, logger, sessionService)

	return r
}

// healthCheckHandler handles the health check endpoint.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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
func registerCategoryRoutes(router *mux.Router, handlers *handlers.Handlers) {
	categoryRouter := router.PathPrefix("/api/categories").Subrouter()
	NamedHandleFunc(categoryRouter, "/{id}/", handlers.CategoryHandler.GetCategoryByIDHandler, []string{http.MethodGet}, "GetCategoryByID")
	NamedHandleFunc(categoryRouter, "", handlers.CategoryHandler.GetCategoriesHandler, []string{http.MethodGet}, "GetCategories")
	NamedHandleFunc(categoryRouter, "/{id}/", handlers.CategoryHandler.SoftDeleteCategoryHandler, []string{http.MethodDelete}, "SoftDeleteCategory")
	NamedHandleFunc(categoryRouter, "/collections", handlers.CategoryHandler.GetCollectionCategoriesHandler, []string{http.MethodGet}, "GetCollectionCategories")
	NamedHandleFunc(categoryRouter, "/name/{name}", handlers.CategoryHandler.GetCategoryByNameHandler, []string{http.MethodOptions, http.MethodGet}, "GetCategoryByName")
	NamedHandleFunc(categoryRouter, "/parent/{parentId}", handlers.CategoryHandler.GetCategoriesByParentIDHandler, []string{http.MethodGet}, "GetCategoriesByParentID")
	NamedHandleFunc(categoryRouter, "/products/count", handlers.CategoryHandler.GetCategoriesWithProductsCountHandler, []string{http.MethodGet}, "GetCategoriesWithProductsCount")
	NamedHandleFunc(categoryRouter, "/tree", handlers.CategoryHandler.GetCategoryTreeHandler, []string{http.MethodGet}, "GetCategoryTree")
	NamedHandleFunc(categoryRouter, "/hierarchy", handlers.CategoryHandler.GetV2CategoryHierarchyHandler, []string{http.MethodGet}, "GetV2CategoryHierarchy")
	NamedHandleFunc(categoryRouter, "/parent", handlers.CategoryHandler.GetParentCategoriesHandler, []string{http.MethodGet}, "GetParentCategories")
	NamedHandleFunc(categoryRouter, "/{id}/", handlers.CategoryHandler.CheckCategoryExistenceHandler, []string{http.MethodHead}, "CheckCategoryExistence")
	NamedHandleFunc(categoryRouter, "/subcategories/count", handlers.CategoryHandler.GetCategoriesWithSubcategoryCountHandler, []string{http.MethodGet}, "GetCategoriesWithSubcategoryCount")
	NamedHandleFunc(categoryRouter, "/upload-image", handlers.CategoryHandler.UploadCategoryImageHandler, []string{http.MethodPost}, "UploadCategoryImage")
}

// registerAdminCategoryRoutes registers admin category-related routes.
func registerAdminCategoryRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	categoryRouter := router.PathPrefix("/api/v2/categories").Subrouter()
	NamedHandleFunc(categoryRouter, "/{slug}/details", handlers.CategoryHandler.GetCategoryDetailsHandler, []string{http.MethodGet}, "GetCategoryDetails")
	NamedHandleFunc(categoryRouter, "/hierarchy", handlers.CategoryHandler.GetV2CategoryHierarchyHandler, []string{http.MethodGet}, "GetV2CategoryHierarchy")
	NamedHandleFunc(categoryRouter, "/{slug}/seo", handlers.CategoryHandler.GetCategorySEOHandler, []string{http.MethodGet}, "GetCategorySEO")

	protected := categoryRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(logger))
	NamedHandleFunc(protected, "", handlers.CategoryHandler.CreateCategoryHandler, []string{http.MethodPost}, "CreateCategory")
	NamedHandleFunc(protected, "/{id}", handlers.CategoryHandler.DeleteCategoryHandler, []string{http.MethodDelete}, "DeleteCategory")
	NamedHandleFunc(protected, "/{id}", handlers.CategoryHandler.SoftDeleteCategoryHandler, []string{http.MethodPut}, "SoftDeleteCategory")
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

// registerAdditionalRoutes registers other related routes like variants, images, specifications, options, cart, orders, promotions, etc.
func registerAdditionalRoutes(r *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	// Product Variant routes
	variantRouter := r.PathPrefix("/api/product-variants").Subrouter()
	NamedHandleFunc(variantRouter, "", handlers.ProductVariantHandler.CreateProductVariantHandler, []string{http.MethodPost}, "CreateProductVariant")
	NamedHandleFunc(variantRouter, "/{id}", handlers.ProductVariantHandler.GetProductVariantByIDHandler, []string{http.MethodGet}, "GetProductVariantByID")
	NamedHandleFunc(variantRouter, "/product/{id}", handlers.ProductVariantHandler.ListProductVariantsByProductIDHandler, []string{http.MethodGet}, "ListProductVariantsByProductID")
	NamedHandleFunc(variantRouter, "/{id}", handlers.ProductVariantHandler.UpdateProductVariantHandler, []string{http.MethodPut}, "UpdateProductVariant")
	NamedHandleFunc(variantRouter, "/{id}", handlers.ProductVariantHandler.DeleteProductVariantHandler, []string{http.MethodDelete}, "DeleteProductVariant")

	// Product Image routes
	imageRouter := r.PathPrefix("/api/product-images").Subrouter()
	NamedHandleFunc(imageRouter, "", handlers.ProductImageHandler.CreateProductImageHandler, []string{http.MethodPost}, "CreateProductImage")
	NamedHandleFunc(imageRouter, "/{id}", handlers.ProductImageHandler.GetProductImageByIDHandler, []string{http.MethodGet}, "GetProductImageByID")
	NamedHandleFunc(imageRouter, "/product/{product_id}", handlers.ProductImageHandler.GetProductImagesByProductIDHandler, []string{http.MethodGet}, "GetProductImagesByProductID")
	NamedHandleFunc(imageRouter, "/{id}", handlers.ProductImageHandler.UpdateProductImageHandler, []string{http.MethodPut}, "UpdateProductImage")
	NamedHandleFunc(imageRouter, "/{id}", handlers.ProductImageHandler.DeleteProductImageHandler, []string{http.MethodDelete}, "DeleteProductImage")

	// Product Specification routes
	specRouter := r.PathPrefix("/api/product-specifications").Subrouter()
	NamedHandleFunc(specRouter, "", handlers.ProductSpecificationHandler.CreateProductSpecificationHandler, []string{http.MethodPost}, "CreateProductSpecification")
	NamedHandleFunc(specRouter, "", handlers.ProductSpecificationHandler.ListProductSpecificationsByProductIDHandler, []string{http.MethodGet}, "ListProductSpecificationsByProductID")
	NamedHandleFunc(specRouter, "/{id}", handlers.ProductSpecificationHandler.DeleteProductSpecificationHandler, []string{http.MethodDelete}, "DeleteProductSpecification")

	// Product Option routes
	optionRouter := r.PathPrefix("/api/product-options").Subrouter()
	NamedHandleFunc(optionRouter, "/{id}", handlers.ProductOptionHandler.CreateProductOptionHandler, []string{http.MethodPost}, "CreateProductOption")
	NamedHandleFunc(optionRouter, "/{id}", handlers.ProductOptionHandler.ListProductOptionsByProductIDHandler, []string{http.MethodGet}, "ListProductOptionsByProductID")
	NamedHandleFunc(optionRouter, "/{id}", handlers.ProductOptionHandler.DeleteProductOptionHandler, []string{http.MethodDelete}, "DeleteProductOption")
	NamedHandleFunc(optionRouter, "/{id}", handlers.ProductOptionHandler.UpdateProductOptionHandler, []string{http.MethodPut}, "UpdateProductOption")

	// Product Option Value routes
	optionValueRouter := r.PathPrefix("/api/product-option-values").Subrouter()
	NamedHandleFunc(optionValueRouter, "/{id}", handlers.ProductOptionHandler.CreateProductOptionValueHandler, []string{http.MethodPost}, "CreateProductOptionValue")
	NamedHandleFunc(optionValueRouter, "/{id}", handlers.ProductOptionHandler.ListProductOptionValuesByOptionIDHandler, []string{http.MethodGet}, "ListProductOptionValuesByOptionID")
	NamedHandleFunc(optionValueRouter, "/{id}", handlers.ProductOptionHandler.DeleteProductOptionValueHandler, []string{http.MethodDelete}, "DeleteProductOptionValue")
	NamedHandleFunc(optionValueRouter, "/{id}", handlers.ProductOptionHandler.UpdateProductOptionValueHandler, []string{http.MethodPut}, "UpdateProductOptionValue")

	// Order routes
	orderRouter := r.PathPrefix("/api/orders").Subrouter()
	NamedHandleFunc(orderRouter, "", handlers.OrderHandler.CreateOrder, []string{http.MethodPost}, "CreateOrder")
	NamedHandleFunc(orderRouter, "", handlers.OrderHandler.ListOrders, []string{http.MethodGet}, "ListOrders")
	NamedHandleFunc(orderRouter, "/{id}", handlers.OrderHandler.GetOrder, []string{http.MethodGet}, "GetOrder")
	NamedHandleFunc(orderRouter, "/pay", handlers.OrderHandler.PayOrder, []string{http.MethodPost}, "PayOrder")
	NamedHandleFunc(orderRouter, "/pay/status", handlers.OrderHandler.GetPaymentStatus, []string{http.MethodGet}, "GetPaymentStatus")
	NamedHandleFunc(orderRouter, "/pay/mpesa-callback", handlers.OrderHandler.HandleMpesaCallback, []string{http.MethodPost}, "HandleMpesaCallback")
	NamedHandleFunc(orderRouter, "/{id}/cancel", handlers.OrderHandler.CancelOrder, []string{http.MethodPut}, "CancelOrder")
	NamedHandleFunc(orderRouter, "/{id}/pay", handlers.OrderHandler.ChangeOrderPaymentMethod, []string{http.MethodPut}, "ChangeOrderPaymentMethod")

	// Admin order routes
	adminOrderRouter := r.PathPrefix("/api/v2/orders").Subrouter()
	protectedAdminOrderRouter := adminOrderRouter.PathPrefix("").Subrouter()
	protectedAdminOrderRouter.Use(middleware.Auth(logger))
	NamedHandleFunc(protectedAdminOrderRouter, "/total-revenue", handlers.OrderHandler.GetTotalRevenue, []string{http.MethodGet}, "GetTotalRevenue")
	NamedHandleFunc(protectedAdminOrderRouter, "/monthly-sales", handlers.OrderHandler.GetMonthlySales, []string{http.MethodGet}, "GetMonthlySales")
	NamedHandleFunc(protectedAdminOrderRouter, "/monthly-revenue", handlers.OrderHandler.GetMonthlyRevenue, []string{http.MethodGet}, "GetMonthlyRevenue")
	NamedHandleFunc(protectedAdminOrderRouter, "/sales-trend", handlers.OrderHandler.GetSalesTrend, []string{http.MethodGet}, "GetSalesTrend")
	NamedHandleFunc(protectedAdminOrderRouter, "/recent-sales", handlers.OrderHandler.GetRecentSales, []string{http.MethodGet}, "GetRecentSales")
	NamedHandleFunc(protectedAdminOrderRouter, "/monthly-total-sales", handlers.OrderHandler.GetTotalSalesCurrentMonth, []string{http.MethodGet}, "GetTotalSalesCurrentMonth")
	NamedHandleFunc(protectedAdminOrderRouter, "/exchange-rate", handlers.OrderHandler.GetExchangeRate, []string{http.MethodGet}, "GetExchangeRate")
	NamedHandleFunc(protectedAdminOrderRouter, "/exchange-rate", handlers.OrderHandler.UpdateExchangeRate, []string{http.MethodPut}, "UpdateExchangeRate")

	// Product Analytic routes
	productAnalyticRouter := r.PathPrefix("/api/product-analytics").Subrouter()
	NamedHandleFunc(productAnalyticRouter, "/best-sellers", handlers.ProductAnalyticHandler.GetBestSellerProducts, []string{http.MethodGet}, "GetBestSellerProducts")
	NamedHandleFunc(productAnalyticRouter, "/featured", handlers.ProductAnalyticHandler.GetFeaturedProducts, []string{http.MethodGet}, "GetFeaturedProducts")
	NamedHandleFunc(productAnalyticRouter, "/new-arrivals", handlers.ProductAnalyticHandler.GetNewArrivalProducts, []string{http.MethodGet}, "GetNewArrivalProducts")
	NamedHandleFunc(productAnalyticRouter, "/daily-deals", handlers.ProductAnalyticHandler.GetDailyDealsProducts, []string{http.MethodGet}, "GetDailyDealsProducts")

	// Discount routes
	discountRouter := r.PathPrefix("/api/discounts").Subrouter()
	protectedDiscountRouter := discountRouter.PathPrefix("").Subrouter()
	protectedDiscountRouter.Use(middleware.Auth(logger))
	protectedDiscountRouter.HandleFunc("", handlers.DiscountHandler.CreateDiscountHandler).Methods(http.MethodPost)

	// Role routes
	roleRouter := r.PathPrefix("/api/roles").Subrouter()
	roleRouter.HandleFunc("", handlers.RoleHandler.CreateRole).Methods(http.MethodPost)
}
