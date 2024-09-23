package services

import (
	"net/url"
	"strings"
)

// Define cache key namespaces and subnamespaces as constants for consistency.
const (
	// Namespaces
	NamespaceProduct      = "product"
	NamespaceUser         = "user"
	NamespaceOrder        = "order"
	NamespaceExchangeRate = "exchange"

	// Subnamespaces
	SubNamespaceAdmin = "admin"

	SubNamespaceSEO     = "seo"
	SubNamespaceSession = "session"
	SubNamespaceDetails = "details"
	SubNamespaceAll     = "all"
	SubNamespaceSitemap = "sitemap"
	SubNamespaceRate    = "rate"
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
