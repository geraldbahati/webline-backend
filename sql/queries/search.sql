-- name: V2SearchProducts :many
-- Search products using full-text search. This query supports partial matches
-- by splitting the input string into tokens and appending the :* operator
-- for prefix matching. Matching products are ranked by relevance.
WITH
  search_query AS (
    SELECT to_tsquery('english', string_agg(word || ':*', ' & ')) AS query
    FROM regexp_split_to_table($1, '\s+') AS word
  ),
  search_term AS (
    SELECT lower($1) AS term
  ),
  usd_rate AS (
    SELECT rate_to_kes
    FROM exchange_rates
    WHERE currency_code = 'USD'
      AND now() BETWEEN valid_from AND COALESCE(valid_to, now() + interval '100 years')
    ORDER BY valid_from DESC
    LIMIT 1
  ),
  ft_results AS (
    SELECT
      p.id,
      p.name,
      (p.usd_price * ur.rate_to_kes) AS price_in_kes,
      p.slug,
      ts_rank_cd(p.search_keyword, sq.query) AS rank
    FROM products p
    CROSS JOIN usd_rate ur
    CROSS JOIN search_query sq
    WHERE p.status = 'active'
      AND p.search_keyword @@ sq.query
  ),
  ilike_results AS (
    SELECT DISTINCT
      p.id,
      p.name,
      (p.usd_price * ur.rate_to_kes) AS price_in_kes,
      p.slug,
      0::double precision AS rank
    FROM products p
    LEFT JOIN categories c ON p.category_id = c.id
    CROSS JOIN usd_rate ur
    CROSS JOIN search_term st
    WHERE p.status = 'active'
      AND (
        lower(p.name) ILIKE '%' || st.term || '%'
        OR lower(p.description) ILIKE '%' || st.term || '%'
        OR lower(c.name) ILIKE '%' || st.term || '%'
      )
  )
SELECT id, name, price_in_kes, slug, rank
FROM (
  SELECT * FROM ft_results
  UNION
  SELECT * FROM ilike_results
) AS combined
ORDER BY rank DESC
LIMIT $2;

-- name: AutocompleteSuggestions :many
-- Returns distinct suggestions for auto-completion based on product and category names.
WITH term AS (
  SELECT lower($1) AS term
)
SELECT suggestion
FROM (
  SELECT name AS suggestion
  FROM products
  CROSS JOIN term
  WHERE lower(name) ILIKE '%' || term.term || '%'
  UNION
  SELECT name AS suggestion
  FROM categories
  CROSS JOIN term
  WHERE lower(name) ILIKE '%' || term.term || '%'
) AS sub
ORDER BY suggestion
LIMIT $2;

-- name: V2SearchProductsPaginated :many
WITH
  search_query AS (
    SELECT to_tsquery('english', string_agg(word || ':*', ' & ')) AS query
    FROM regexp_split_to_table($1, '\s+') AS word
  ),
  search_term AS (
    SELECT lower($1) AS term
  ),
  usd_rate AS (
    SELECT rate_to_kes
    FROM exchange_rates
    WHERE currency_code = 'USD'
      AND now() BETWEEN valid_from AND COALESCE(valid_to, now() + interval '100 years')
    ORDER BY valid_from DESC
    LIMIT 1
  ),
  ft_results AS (
    SELECT
      p.id,
      p.name,
      (p.usd_price * ur.rate_to_kes) AS price_in_kes,
      p.slug,
      ts_rank_cd(p.search_keyword, sq.query) AS rank
    FROM products p
    CROSS JOIN usd_rate ur
    CROSS JOIN search_query sq
    WHERE p.status = 'active'
      AND p.search_keyword @@ sq.query
  ),
  ilike_results AS (
    SELECT DISTINCT
      p.id,
      p.name,
      (p.usd_price * ur.rate_to_kes) AS price_in_kes,
      p.slug,
      0 AS rank
    FROM products p
    CROSS JOIN usd_rate ur
    CROSS JOIN search_term st
    WHERE p.status = 'active'
      AND (
        lower(p.name) ILIKE '%' || st.term || '%'
        OR lower(p.description) ILIKE '%' || st.term || '%'
      )
  )
SELECT id, name, price_in_kes, slug, rank
FROM (
  SELECT * FROM ft_results
  UNION
  SELECT * FROM ilike_results
) AS combined
ORDER BY rank DESC
OFFSET $3 LIMIT $2;

-- name: V2SearchProductsCount :one
WITH
  search_query AS (
    SELECT to_tsquery('english', string_agg(word || ':*', ' & ')) AS query
    FROM regexp_split_to_table($1, '\s+') AS word
  ),
  search_term AS (
    SELECT lower($1) AS term
  ),
  usd_rate AS (
    SELECT rate_to_kes
    FROM exchange_rates
    WHERE currency_code = 'USD'
      AND now() BETWEEN valid_from AND COALESCE(valid_to, now() + interval '100 years')
    ORDER BY valid_from DESC
    LIMIT 1
  ),
  ft_results AS (
    SELECT p.id
    FROM products p
    CROSS JOIN usd_rate ur
    CROSS JOIN search_query sq
    WHERE p.status = 'active'
      AND p.search_keyword @@ sq.query
  ),
  ilike_results AS (
    SELECT DISTINCT p.id
    FROM products p
    CROSS JOIN search_term st
    WHERE p.status = 'active'
      AND (
        lower(p.name) ILIKE '%' || st.term || '%'
        OR lower(p.description) ILIKE '%' || st.term || '%'
      )
  )
SELECT COUNT(*) FROM (
  SELECT * FROM ft_results
  UNION
  SELECT * FROM ilike_results
) AS combined;
