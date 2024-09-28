package services

import (
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

	// Subnamespaces
	SubNamespaceAdmin    = "admin"
	SubNamespaceCart     = "cart"
	SubNamespaceSEO      = "seo"
	SubNamespaceSession  = "session"
	SubNamespaceDetails  = "details"
	SubNamespaceAll      = "all"
	SubNamespaceSitemap  = "sitemap"
	SubNamespaceRate     = "rate"
	SubNamespaceAnalytic = "analytic"
	SubNamespacePricing  = "pricing"
	SubNamespaceSpecs    = "specs"
	SubNamespaceImages   = "images"
	SubNamespaceItems    = "items"
	SubNamespaceTotal    = "total"
	// Add more subnamespaces as needed
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

// CartItemsKey generates a Redis key for cart items.
func CartItemsKey(cartID string) string {
	return GenerateCacheKey(NamespaceCart, SubNamespaceItems, cartID)
}

// CartTotalKey generates a Redis key for cart total.
func CartTotalKey(cartID string) string {
	return GenerateCacheKey(NamespaceCart, SubNamespaceTotal, cartID)
}

// SessionKey generates a Redis key for session information.
func SessionKey(sessionID string) string {
	return GenerateCacheKey(NamespaceUser, SubNamespaceSession, sessionID)
}
