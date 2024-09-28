package routes

import (
	"net/http"
	"weblineBackend/internal/handlers"
	"weblineBackend/internal/middleware"
	"weblineBackend/internal/services/i"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupRouter(logger *zap.Logger, handlers *handlers.Handlers, sessionService i.SessionService) *mux.Router {
	r := mux.NewRouter()
	corsMiddleware := middleware.CORS(logger, middleware.CORSOptions{
		AllowedOrigins:   []string{"*"}, // or specify specific origins
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token"},
		AllowCredentials: false, // set to true if you need to allow credentials
	})

	r.Use(corsMiddleware)
	r.Use(middleware.OptionalAuth(logger))
	r.Use(middleware.MetricsMiddleware(logger))

	r.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

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
	userRouter.HandleFunc("", handlers.UserHandler.ListUsers).Methods(http.MethodGet)
	userRouter.HandleFunc("/{id}", handlers.UserHandler.GetUserProfile).Methods(http.MethodGet)
	userRouter.HandleFunc("/register", handlers.UserHandler.RegisterUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/refresh", handlers.UserHandler.RefreshToken).Methods(http.MethodPost)
	userRouter.HandleFunc("/login", handlers.UserHandler.LoginUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/reset-password", handlers.UserHandler.ResetPassword).Methods(http.MethodPost)
	userRouter.HandleFunc("/reset-password/request", handlers.UserHandler.RequestPasswordReset).Methods(http.MethodPost)
	userRouter.HandleFunc("/login/google", handlers.UserHandler.LoginWithGoogle).Methods(http.MethodPost)
	userRouter.HandleFunc("/login/email-verified", handlers.UserHandler.EmailVerified).Methods(http.MethodPost)
	userRouter.HandleFunc("/{id}/profile", handlers.UserHandler.GetUserInfo).Methods(http.MethodGet)

	protected := userRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(logger))
	protected.HandleFunc("/admin-requests", handlers.UserHandler.RequestAdminRole).Methods(http.MethodPost)
	protected.HandleFunc("/approve", handlers.UserHandler.ApproveAdminRole).Methods(http.MethodPost)
	protected.HandleFunc("/profile", handlers.UserHandler.UpdateUserProfile).Methods(http.MethodPut)
}

// registerCategoryRoutes registers category-related routes.
func registerCategoryRoutes(router *mux.Router, handlers *handlers.Handlers) {
	categoryRouter := router.PathPrefix("/api/categories").Subrouter()
	categoryRouter.HandleFunc("/{id}/", handlers.CategoryHandler.GetCategoryByIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("", handlers.CategoryHandler.GetCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", handlers.CategoryHandler.SoftDeleteCategoryHandler).Methods(http.MethodDelete)
	NamedHandleFunc(categoryRouter, "/collections", handlers.CategoryHandler.GetCollectionCategoriesHandler, []string{http.MethodGet}, "GetCollectionCategories")
	categoryRouter.HandleFunc("/name/{name}", handlers.CategoryHandler.GetCategoryByNameHandler).Methods(http.MethodOptions, http.MethodGet)
	categoryRouter.HandleFunc("/parent/{parentId}", handlers.CategoryHandler.GetCategoriesByParentIDHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/products/count", handlers.CategoryHandler.GetCategoriesWithProductsCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/tree", handlers.CategoryHandler.GetCategoryTreeHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/hierarchy", handlers.CategoryHandler.GetV2CategoryHierarchyHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/parent", handlers.CategoryHandler.GetParentCategoriesHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/{id}/", handlers.CategoryHandler.CheckCategoryExistenceHandler).Methods(http.MethodHead)
	categoryRouter.HandleFunc("/subcategories/count", handlers.CategoryHandler.GetCategoriesWithSubcategoryCountHandler).Methods(http.MethodGet)
	categoryRouter.HandleFunc("/upload-image", handlers.CategoryHandler.UploadCategoryImageHandler).Methods(http.MethodPost)
}

// registerAdminCategoryRoutes registers admin category-related routes.
func registerAdminCategoryRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	categoryRouter := router.PathPrefix("/api/v2/categories").Subrouter()
	NamedHandleFunc(categoryRouter, "/{slug}/details", handlers.CategoryHandler.GetCategoryDetailsHandler, []string{http.MethodGet}, "GetCategoryDetails")
	NamedHandleFunc(categoryRouter, "/hierarchy", handlers.CategoryHandler.GetV2CategoryHierarchyHandler, []string{http.MethodGet}, "GetV2CategoryHierarchy")
	NamedHandleFunc(categoryRouter, "/{slug}/seo", handlers.CategoryHandler.GetCategorySEOHandler, []string{http.MethodGet}, "GetCategorySEO")

	protected := categoryRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(logger))
	protected.HandleFunc("", handlers.CategoryHandler.CreateCategoryHandler).Methods(http.MethodPost)
	protected.HandleFunc("/{id}", handlers.CategoryHandler.DeleteCategoryHandler).Methods(http.MethodDelete)
	protected.HandleFunc("/{id}", handlers.CategoryHandler.SoftDeleteCategoryHandler).Methods(http.MethodPut)
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
	productRouter.HandleFunc("/category/{id}", handlers.ProductHandler.GetProductsByCategoryIDHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/filter/{category_id}", handlers.ProductHandler.GetFilteredCategoryProducts).Methods(http.MethodPost)
	productRouter.HandleFunc("/filter/all/options", handlers.ProductHandler.GetAllProductFilterOptionsHandler).Methods(http.MethodGet)
	productRouter.HandleFunc("/filter/options/{name}", handlers.ProductHandler.GetFilterOptionsByCategoryNameHandler).Methods(http.MethodGet)
}

// registerAdminProductRoutes registers admin product-related routes.
func registerAdminProductRoutes(router *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	productRouter := router.PathPrefix("/api/v2/products").Subrouter()
	NamedHandleFunc(productRouter, "", handlers.ProductHandler.GetProductsHandler, []string{http.MethodGet}, "GetProducts")
	NamedHandleFunc(productRouter, "/{slug}/detail", handlers.ProductHandler.GetProductDetailHandler, []string{http.MethodGet}, "GetProductDetail")
	NamedHandleFunc(productRouter, "/meta-fields/{categoryID}", handlers.ProductHandler.GetProductMetaFieldsByCategoryIDHandler, []string{http.MethodGet}, "GetProductMetaFieldsByCategoryID")

	protected := productRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(logger))
	protected.HandleFunc("", handlers.ProductHandler.CreateV2ProductHandler).Methods(http.MethodPost)
	protected.HandleFunc("", handlers.ProductHandler.DeleteProductsHandler).Methods(http.MethodDelete)
	protected.HandleFunc("/{slug}", handlers.ProductHandler.DeleteProductHandler).Methods(http.MethodDelete)
	protected.HandleFunc("/archive", handlers.ProductHandler.ArchiveProductsHandler).Methods(http.MethodPut)
	protected.HandleFunc("/draft", handlers.ProductHandler.DraftProductsHandler).Methods(http.MethodPut)
	protected.HandleFunc("/active", handlers.ProductHandler.ActivateProductsHandler).Methods(http.MethodPut)
	protected.HandleFunc("/{slug}/archive", handlers.ProductHandler.ArchiveProductHandler).Methods(http.MethodPut)
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
	// Cart routes
	cartRouter := router.PathPrefix("/api/cart").Subrouter()
	cartRouter.Use(middleware.OptionalAuth(logger))
	cartRouter.Use(middleware.Session(logger, sessionService))
	cartRouter.Use(middleware.CSRF(logger))

	// Updated cart routes without cartID in the path
	NamedHandleFunc(cartRouter, "/add", handlers.CartHandler.AddToCartHandler, []string{http.MethodPost}, "AddToCart")
	NamedHandleFunc(cartRouter, "/remove", handlers.CartHandler.RemoveFromCartHandler, []string{http.MethodDelete}, "RemoveFromCart")
	NamedHandleFunc(cartRouter, "/items", handlers.CartHandler.GetCartItemsHandler, []string{http.MethodGet}, "GetCartItems")
	NamedHandleFunc(cartRouter, "/items/update-quantity", handlers.CartHandler.UpdateCartItemQuantityHandler, []string{http.MethodPut}, "UpdateCartItemQuantity")
	NamedHandleFunc(cartRouter, "/clear", handlers.CartHandler.ClearCartHandler, []string{http.MethodPost}, "ClearCart")
	NamedHandleFunc(cartRouter, "/total", handlers.CartHandler.CalculateCartTotalHandler, []string{http.MethodGet}, "CalculateCartTotal")
	NamedHandleFunc(cartRouter, "/replace", handlers.CartHandler.ReplaceCartItemsHandler, []string{http.MethodPut}, "ReplaceCartItems")
	NamedHandleFunc(cartRouter, "", handlers.CartHandler.GetShoppingCartBySessionIDHandler, []string{http.MethodGet}, "GetShoppingCart")
}

// registerAdditionalRoutes registers other related routes like variants, images, specifications, options, cart, orders, promotions, etc.
func registerAdditionalRoutes(r *mux.Router, handlers *handlers.Handlers, logger *zap.Logger) {
	// Product Variant routes
	variantRouter := r.PathPrefix("/api/product-variants").Subrouter()
	variantRouter.HandleFunc("", handlers.ProductVariantHandler.CreateProductVariantHandler).Methods(http.MethodPost)
	variantRouter.HandleFunc("/{id}", handlers.ProductVariantHandler.GetProductVariantByIDHandler).Methods(http.MethodGet)
	variantRouter.HandleFunc("/product/{id}", handlers.ProductVariantHandler.ListProductVariantsByProductIDHandler).Methods(http.MethodGet)
	variantRouter.HandleFunc("/{id}", handlers.ProductVariantHandler.UpdateProductVariantHandler).Methods(http.MethodPut)
	variantRouter.HandleFunc("/{id}", handlers.ProductVariantHandler.DeleteProductVariantHandler).Methods(http.MethodDelete)

	// Product Image routes
	imageRouter := r.PathPrefix("/api/product-images").Subrouter()
	imageRouter.HandleFunc("", handlers.ProductImageHandler.CreateProductImageHandler).Methods(http.MethodPost)
	imageRouter.HandleFunc("/{id}", handlers.ProductImageHandler.GetProductImageByIDHandler).Methods(http.MethodGet)
	imageRouter.HandleFunc("/product/{product_id}", handlers.ProductImageHandler.GetProductImagesByProductIDHandler).Methods(http.MethodGet)
	imageRouter.HandleFunc("/{id}", handlers.ProductImageHandler.UpdateProductImageHandler).Methods(http.MethodPut)
	imageRouter.HandleFunc("/{id}", handlers.ProductImageHandler.DeleteProductImageHandler).Methods(http.MethodDelete)

	// Product Specification routes
	specRouter := r.PathPrefix("/api/product-specifications").Subrouter()
	specRouter.HandleFunc("", handlers.ProductSpecificationHandler.CreateProductSpecificationHandler).Methods(http.MethodPost)
	specRouter.HandleFunc("", handlers.ProductSpecificationHandler.ListProductSpecificationsByProductIDHandler).Methods(http.MethodGet)
	specRouter.HandleFunc("/{id}", handlers.ProductSpecificationHandler.DeleteProductSpecificationHandler).Methods(http.MethodDelete)

	// Product Option routes
	optionRouter := r.PathPrefix("/api/product-options").Subrouter()
	optionRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.CreateProductOptionHandler).Methods(http.MethodPost)
	optionRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.ListProductOptionsByProductIDHandler).Methods(http.MethodGet)
	optionRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.DeleteProductOptionHandler).Methods(http.MethodDelete)
	optionRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.UpdateProductOptionHandler).Methods(http.MethodPut)

	// Product Option Value routes
	optionValueRouter := r.PathPrefix("/api/product-option-values").Subrouter()
	optionValueRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.CreateProductOptionValueHandler).Methods(http.MethodPost)
	optionValueRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.ListProductOptionValuesByOptionIDHandler).Methods(http.MethodGet)
	optionValueRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.DeleteProductOptionValueHandler).Methods(http.MethodDelete)
	optionValueRouter.HandleFunc("/{id}", handlers.ProductOptionHandler.UpdateProductOptionValueHandler).Methods(http.MethodPut)

	// Order routes
	orderRouter := r.PathPrefix("/api/orders").Subrouter()
	orderRouter.HandleFunc("", handlers.OrderHandler.CreateOrder).Methods(http.MethodPost)
	orderRouter.HandleFunc("", handlers.OrderHandler.ListOrders).Methods(http.MethodGet)
	orderRouter.HandleFunc("/{id}", handlers.OrderHandler.GetOrder).Methods(http.MethodGet)
	orderRouter.HandleFunc("/pay", handlers.OrderHandler.PayOrder).Methods(http.MethodPost)
	orderRouter.HandleFunc("/pay/status", handlers.OrderHandler.GetPaymentStatus).Methods(http.MethodGet)
	orderRouter.HandleFunc("/pay/mpesa-callback", handlers.OrderHandler.HandleMpesaCallback).Methods(http.MethodPost)
	orderRouter.HandleFunc("/{id}/cancel", handlers.OrderHandler.CancelOrder).Methods(http.MethodPut)
	orderRouter.HandleFunc("/{id}/pay", handlers.OrderHandler.ChangeOrderPaymentMethod).Methods(http.MethodPut)

	// Admin order routes
	adminOrderRouter := r.PathPrefix("/api/v2/orders").Subrouter()
	protectedAdminOrderRouter := adminOrderRouter.PathPrefix("").Subrouter()
	protectedAdminOrderRouter.Use(middleware.Auth(logger))
	protectedAdminOrderRouter.HandleFunc("/total-revenue", handlers.OrderHandler.GetTotalRevenue).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/monthly-sales", handlers.OrderHandler.GetMonthlySales).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/monthly-revenue", handlers.OrderHandler.GetMonthlyRevenue).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/sales-trend", handlers.OrderHandler.GetSalesTrend).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/recent-sales", handlers.OrderHandler.GetRecentSales).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/monthly-total-sales", handlers.OrderHandler.GetTotalSalesCurrentMonth).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/exchange-rate", handlers.OrderHandler.GetExchangeRate).Methods(http.MethodGet)
	protectedAdminOrderRouter.HandleFunc("/exchange-rate", handlers.OrderHandler.UpdateExchangeRate).Methods(http.MethodPut)

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
