package services

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Define cache key namespaces and subnamespaces as constants for consistency.
const (
	// Namespaces
	NamespaceProduct      = "product"
	NamespaceUser         = "user"
	NamespaceOrder        = "order"
	NamespaceExchangeRate = "exchange"
	NamespaceCart         = "cart"
	NamespaceSearch       = "search"

	// Subnamespaces
	SubNamespaceAdmin       = "admin"
	SubNamespaceCart        = "cart"
	SubNamespaceSEO         = "seo"
	SubNamespaceSession     = "session"
	SubNamespaceDetails     = "details"
	SubNamespaceAll         = "all"
	SubNamespaceSitemap     = "sitemap"
	SubNamespaceRate        = "rate"
	SubNamespaceAnalytic    = "analytic"
	SubNamespacePricing     = "pricing"
	SubNamespaceSpecs       = "specs"
	SubNamespaceImages      = "images"
	SubNamespaceItems       = "items"
	SubNamespaceTotal       = "total"
	SubNamespaceItem        = "item"
	SubNamespaceNewArrivals = "new-arrivals"
	SubNamespaceBestSellers = "best-sellers"
	SubNamespaceFeatured    = "featured"
	SubNamespaceOnSale      = "on-sale"
	SubNamespaceDailyDeals  = "daily-deals"

	// Session-related cache keys
	NamespaceSession   = "session"
	SubNamespaceUser   = "user"
	SubNamespaceGuest  = "guest"
	SubNamespaceExpiry = "expiry"
)

// GenerateCacheKey constructs a standardized Redis cache key.
// It takes multiple parts and joins them with a colon (:).
func GenerateCacheKey(parts ...string) string {
	sanitizedParts := make([]string, len(parts))
	for i, part := range parts {
		sanitizedParts[i] = url.PathEscape(part)
	}
	return strings.Join(sanitizedParts, ":")
}

// ProductSEOKey generates a Redis key for product SEO information.
func ProductSEOKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceSEO, slug)
}

// UserSessionKey generates a Redis key for user session information.
func UserSessionKey(userID string) string {
	return GenerateCacheKey(NamespaceUser, SubNamespaceSession, userID)
}

// OrderDetailsKey generates a Redis key for order details.
func OrderDetailsKey(orderID string) string {
	return GenerateCacheKey(NamespaceOrder, SubNamespaceDetails, orderID)
}

// ProductDetailKey generates a Redis key for product detail.
func ProductDetailKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceDetails, slug)
}

// AdminProductDetailKey generates a Redis key for admin product detail.
func AdminProductDetailKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceAdmin, SubNamespaceDetails, slug)
}

// ProductAllKey generates a Redis key for all products.
func ProductAllKey() string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceAll)
}

// ProductSitemapKey generates a Redis key for product sitemap.
func ProductSitemapKey() string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceSitemap)
}

// ExchangeRateKey generates a Redis key for exchange rates by currency.
func ExchangeRateKey(currency string) string {
	return GenerateCacheKey(NamespaceExchangeRate, SubNamespaceRate, currency)
}

// ProductAnalyticKey generates a Redis key for product analytic.
func ProductAnalyticKey(category string, limit int32) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceAnalytic, category, strconv.Itoa(int(limit)))
}

// ProductPricingKey generates a Redis key for product pricing.
func ProductPricingKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespacePricing, slug)
}

// ProductSpecsKey generates a Redis key for product specifications.
func ProductSpecsKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceSpecs, slug)
}

// ProductImagesKey generates a Redis key for product images.
func ProductImagesKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceImages, slug)
}

// ProductCartKey generates a Redis key for product cart.
func ProductCartKey(slug string) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceCart, slug)
}

// SessionKey generates a Redis key for session information.
func SessionKey(sessionID string) string {
	return GenerateCacheKey(NamespaceSession, sessionID)
}

// CartItemsKey generates a Redis key for cart items.
func CartItemsKey(cartID string) string {
	return GenerateCacheKey(NamespaceCart, cartID, SubNamespaceItems)
}

// CartItemKey generates a Redis key for a specific cart item.
func CartItemKey(cartID string, productID string) string {
	return GenerateCacheKey(NamespaceCart, cartID, SubNamespaceItem, productID)
}

// CartTotalKey generates a Redis key for cart total.
func CartTotalKey(cartID string) string {
	return GenerateCacheKey(NamespaceCart, cartID, SubNamespaceTotal)
}

// ProductBestSellersKey generates a Redis key for best seller products.
func ProductBestSellersKey(limit int32) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceBestSellers, strconv.Itoa(int(limit)))
}

// ProductFeaturedKey generates a Redis key for featured products.
func ProductFeaturedKey(limit int32) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceFeatured, strconv.Itoa(int(limit)))
}

// ProductOnSaleKey generates a Redis key for on sale products.
func ProductOnSaleKey(limit int32) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceOnSale, strconv.Itoa(int(limit)))
}

// ProductNewArrivalsKey generates a Redis key for new arrivals products.
func ProductNewArrivalsKey(limit int32) string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceNewArrivals, strconv.Itoa(int(limit)))
}

// ProductDailyDealsKey generates a Redis key for daily deals products.
func ProductDailyDealsKey() string {
	return GenerateCacheKey(NamespaceProduct, SubNamespaceDailyDeals)
}

// ProductCachePattern generates a Redis key pattern for product cache.
func ProductCachePattern() string {
	return fmt.Sprintf("%s:*", GenerateCacheKey(NamespaceProduct))
}

// UserSessionsKey generates a Redis key for user sessions
func UserSessionsKey(userID string) string {
	return GenerateCacheKey(NamespaceSession, SubNamespaceUser, userID)
}

// SessionExpiryKey generates a Redis key for session expiry tracking
func SessionExpiryKey(sessionID string) string {
	return GenerateCacheKey(NamespaceSession, SubNamespaceExpiry, sessionID)
}

// SearchProductsKey generates a Redis key for storing search results.
// It includes the search term, page number, and limit.
func SearchProductsKey(searchTerm string, page int, limit int) string {
	return GenerateCacheKey(NamespaceSearch, "products", searchTerm, "page", strconv.Itoa(page), "limit", strconv.Itoa(limit))
}

// AutocompleteSuggestionsKey generates a Redis key for storing autocomplete search suggestions.
func AutocompleteSuggestionsKey(searchTerm string, limit int) string {
	return GenerateCacheKey(NamespaceSearch, "autocomplete", searchTerm, strconv.Itoa(limit))
}
