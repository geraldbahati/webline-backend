--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+1)
-- Dumped by pg_dump version 16.4 (Debian 16.4-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: ltree; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS ltree WITH SCHEMA public;


--
-- Name: EXTENSION ltree; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION ltree IS 'data type for hierarchical tree-like structures';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: format_phone_number(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.format_phone_number() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- For users table
    IF TG_TABLE_NAME = 'users' THEN
        NEW.phone_number = '254' || SUBSTRING(NEW.phone_number FROM 2);
    END IF;

    -- For guest_checkouts table
    IF TG_TABLE_NAME = 'guest_checkouts' THEN
        NEW.phone = '254' || SUBSTRING(NEW.phone FROM 2);
    END IF;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.format_phone_number() OWNER TO geraldbahati;

--
-- Name: generate_ltree_path(uuid); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.generate_ltree_path(category_id uuid) RETURNS public.ltree
    LANGUAGE plpgsql
    AS $$
DECLARE
    path ltree;
BEGIN
    SELECT
        string_agg(regexp_replace(lower(name), '[^a-z0-9]+', '_', 'g'), '.')::ltree
    INTO path
    FROM (
             WITH RECURSIVE parent_categories AS (
                 SELECT id, parent_id, name, 0 AS depth
                 FROM categories
                 WHERE id = category_id
                 UNION ALL
                 SELECT c.id, c.parent_id, c.name, pc.depth + 1
                 FROM categories c
                          INNER JOIN parent_categories pc ON pc.parent_id = c.id
             )
             SELECT name FROM parent_categories ORDER BY depth DESC
         ) AS paths;

    RETURN path;
END;
$$;


ALTER FUNCTION public.generate_ltree_path(category_id uuid) OWNER TO geraldbahati;

--
-- Name: generate_order_number(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.generate_order_number() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.order_number := TO_CHAR(NEXTVAL('order_number_seq'), 'FM000000');
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.generate_order_number() OWNER TO geraldbahati;

--
-- Name: generate_promotion_slug(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.generate_promotion_slug() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    temp_slug VARCHAR(255);
    suffix INT := 1;
BEGIN
    -- Generate the initial slug
    temp_slug := lower(regexp_replace(NEW.title, '[^a-zA-Z0-9]+', '-', 'g'));

    -- Check for uniqueness and append a suffix if necessary
    WHILE EXISTS (SELECT 1 FROM promotions WHERE slug = temp_slug) LOOP
            temp_slug := lower(regexp_replace(NEW.title, '[^a-zA-Z0-9]+', '-', 'g')) || '-' || suffix;
            suffix := suffix + 1;
        END LOOP;

    NEW.slug := temp_slug;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.generate_promotion_slug() OWNER TO geraldbahati;

--
-- Name: generate_slug(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.generate_slug() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    generated_slug text;
BEGIN
    -- Convert the name to a slug
    generated_slug := lower(regexp_replace(NEW.name, '[^a-zA-Z0-9]+', '-', 'g'));

    -- Ensure the slug is unique by appending a number if necessary
    WHILE EXISTS (SELECT 1 FROM products WHERE slug = generated_slug) LOOP
            generated_slug := generated_slug || '-' || (SELECT count(*) + 1 FROM products WHERE slug LIKE generated_slug || '%');
        END LOOP;

    NEW.slug := generated_slug;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.generate_slug() OWNER TO geraldbahati;

--
-- Name: generate_slug_from_hierarchy(uuid); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.generate_slug_from_hierarchy(category_id uuid) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
DECLARE
    path ltree;
    slug VARCHAR;
BEGIN
    -- Generate the ltree path
    path := generate_ltree_path(category_id);

    -- Convert the ltree path to a slug by replacing dots with hyphens
    slug := regexp_replace(path::text, '\.', '-', 'g');

    RETURN slug;
END;
$$;


ALTER FUNCTION public.generate_slug_from_hierarchy(category_id uuid) OWNER TO geraldbahati;

--
-- Name: set_cart_item_price(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.set_cart_item_price() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.price := (SELECT price FROM products WHERE id = NEW.product_id);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_cart_item_price() OWNER TO geraldbahati;

--
-- Name: set_category_position(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.set_category_position() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- If no position is provided, find the next available position within the same parent
    IF NEW.position = -1 THEN
        NEW.position := COALESCE(
                (SELECT MAX(position) + 1 FROM categories WHERE parent_id = NEW.parent_id),
                1
                        );
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_category_position() OWNER TO geraldbahati;

--
-- Name: set_default_position(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.set_default_position() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.position IS NULL THEN
        NEW.position := COALESCE(
                (SELECT MAX(position) + 1 FROM product_images WHERE product_id = NEW.product_id),
                1
                        );
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_default_position() OWNER TO geraldbahati;

--
-- Name: set_valid_to_based_on_category(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.set_valid_to_based_on_category() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    category_name text;
BEGIN
    -- Fetch the category name
    SELECT name INTO category_name FROM public.categories WHERE id = NEW.category_id;

    -- Determine the validity period based on the category
    IF category_name = 'Laptops' OR category_name = 'Desktops' THEN
        NEW.valid_to := NEW.valid_from + INTERVAL '3 months';
    ELSE
        NEW.valid_to := NEW.valid_from + INTERVAL '1 year';
    END IF;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_valid_to_based_on_category() OWNER TO geraldbahati;

--
-- Name: update_cart_totals(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.update_cart_totals() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE shopping_carts
        SET total_items = (SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE shopping_cart_id = OLD.shopping_cart_id),
            total_price = (SELECT COALESCE(SUM(quantity * price), 0) FROM cart_items WHERE shopping_cart_id = OLD.shopping_cart_id),
            updated_at = NOW()
        WHERE id = OLD.shopping_cart_id;
    ELSE
        UPDATE shopping_carts
        SET total_items = (SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE shopping_cart_id = NEW.shopping_cart_id),
            total_price = (SELECT COALESCE(SUM(quantity * price), 0) FROM cart_items WHERE shopping_cart_id = NEW.shopping_cart_id),
            updated_at = NOW()
        WHERE id = NEW.shopping_cart_id;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_cart_totals() OWNER TO geraldbahati;

--
-- Name: update_category_path(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.update_category_path() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        -- Set path based on category name if no parent
        NEW.path := COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'root')::ltree;
    ELSE
        -- Fetch the parent path
        SELECT path INTO NEW.path
        FROM categories
        WHERE id = NEW.parent_id;

        -- Handle case where parent path is NULL
        IF NEW.path IS NULL THEN
            NEW.path := 'root'::ltree;
        END IF;

        -- Safely append current category name to the parent's path
        NEW.path := (NEW.path || COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'child'))::ltree;
    END IF;

    -- Update the level based on the new path
    NEW.level := nlevel(NEW.path);

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_category_path() OWNER TO geraldbahati;

--
-- Name: update_category_slug(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.update_category_slug() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Generate the ltree path based on the parent categories
    IF NEW.parent_id IS NULL THEN
        -- Root category
        NEW.path := COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'root')::ltree;
    ELSE
        -- Get the path of the parent category
        SELECT path
        INTO NEW.path
        FROM categories
        WHERE id = NEW.parent_id;

        -- Handle the case where the parent path is NULL
        IF NEW.path IS NULL THEN
            NEW.path := 'root'::ltree;
        END IF;

        -- Append the current category's name to the parent's path
        NEW.path := NEW.path || COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'child')::ltree;
    END IF;

    -- Convert the ltree path to a slug by replacing dots with hyphens
    NEW.slug := COALESCE(regexp_replace(NEW.path::text, '\.', '-', 'g'), 'unknown-slug');

    -- Handle slug duplicates by appending a counter
    DECLARE
        counter INTEGER := 1;
        base_slug VARCHAR := NEW.slug;
    BEGIN
        WHILE EXISTS (SELECT 1 FROM categories WHERE slug = NEW.slug AND id <> NEW.id) LOOP
                NEW.slug := base_slug || '-' || counter::text;
                counter := counter + 1;
            END LOOP;
    END;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_category_slug() OWNER TO geraldbahati;

--
-- Name: update_search_keyword(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.update_search_keyword() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.search_keyword :=
            setweight(to_tsvector(coalesce(NEW.name, '')), 'A') ||
            setweight(to_tsvector(coalesce(NEW.description, '')), 'B') ||
            setweight(to_tsvector(coalesce(NEW.part_number, '')), 'A') ||
            setweight(to_tsvector(coalesce(NEW.meta_title, '')), 'C') ||
            setweight(to_tsvector(coalesce(NEW.meta_description, '')), 'D') ||
            setweight(to_tsvector(coalesce(NEW.meta_keywords, '')), 'D');
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_search_keyword() OWNER TO geraldbahati;

--
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_timestamp() OWNER TO geraldbahati;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: geraldbahati
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO geraldbahati;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_approval_tokens; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.admin_approval_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    request_id uuid NOT NULL,
    token character varying(512) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.admin_approval_tokens OWNER TO geraldbahati;

--
-- Name: admin_requests; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.admin_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    reason text NOT NULL,
    status character varying(50) DEFAULT 'PENDING'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT admin_requests_reason_check CHECK ((char_length(reason) > 0)),
    CONSTRAINT admin_requests_status_check CHECK (((status)::text = ANY ((ARRAY['PENDING'::character varying, 'APPROVED'::character varying, 'REJECTED'::character varying])::text[])))
);


ALTER TABLE public.admin_requests OWNER TO geraldbahati;

--
-- Name: attribute_types; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.attribute_types (
    id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.attribute_types OWNER TO geraldbahati;

--
-- Name: attribute_types_id_seq; Type: SEQUENCE; Schema: public; Owner: geraldbahati
--

CREATE SEQUENCE public.attribute_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.attribute_types_id_seq OWNER TO geraldbahati;

--
-- Name: attribute_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: geraldbahati
--

ALTER SEQUENCE public.attribute_types_id_seq OWNED BY public.attribute_types.id;


--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.cart_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    shopping_cart_id uuid,
    product_id uuid,
    quantity integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    price numeric(10,2) NOT NULL,
    CONSTRAINT check_quantity CHECK ((quantity > 0))
);


ALTER TABLE public.cart_items OWNER TO geraldbahati;

--
-- Name: categories; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    parent_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    is_active boolean DEFAULT true NOT NULL,
    "position" integer DEFAULT '-1'::integer NOT NULL,
    image_url character varying(255),
    description text,
    meta_title character varying(255),
    meta_description text,
    is_featured boolean DEFAULT false NOT NULL,
    level integer DEFAULT 0 NOT NULL,
    path public.ltree,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, (((((((COALESCE(name, ''::character varying))::text || ' '::text) || COALESCE(description, ''::text)) || ' '::text) || (COALESCE(meta_title, ''::character varying))::text) || ' '::text) || COALESCE(meta_description, ''::text)))) STORED,
    slug character varying(255) NOT NULL,
    last_updated_by uuid
);


ALTER TABLE public.categories OWNER TO geraldbahati;

--
-- Name: category_hierarchy_mv; Type: MATERIALIZED VIEW; Schema: public; Owner: geraldbahati
--

CREATE MATERIALIZED VIEW public.category_hierarchy_mv AS
 WITH RECURSIVE category_hierarchy AS (
         SELECT categories.id,
            categories.name,
            categories.parent_id,
            categories."position"
           FROM public.categories
          WHERE (categories.parent_id IS NULL)
        UNION ALL
         SELECT c.id,
            c.name,
            c.parent_id,
            c."position"
           FROM (public.categories c
             JOIN category_hierarchy ch ON ((c.parent_id = ch.id)))
        )
 SELECT id,
    name,
    parent_id,
    "position"
   FROM category_hierarchy
  ORDER BY "position"
  WITH NO DATA;


ALTER MATERIALIZED VIEW public.category_hierarchy_mv OWNER TO geraldbahati;

--
-- Name: discounts; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.discounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    discount_percentage numeric(5,2) NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.discounts OWNER TO geraldbahati;

--
-- Name: exchange_rates; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.exchange_rates (
    id integer NOT NULL,
    currency_code character varying(3) NOT NULL,
    rate_to_kes numeric(10,4) NOT NULL,
    valid_from timestamp with time zone NOT NULL,
    valid_to timestamp with time zone,
    CONSTRAINT check_validity_range CHECK (((valid_to IS NULL) OR (valid_to > valid_from))),
    CONSTRAINT exchange_rates_rate_to_kes_check CHECK ((rate_to_kes > (0)::numeric))
);


ALTER TABLE public.exchange_rates OWNER TO geraldbahati;

--
-- Name: exchange_rates_id_seq; Type: SEQUENCE; Schema: public; Owner: geraldbahati
--

CREATE SEQUENCE public.exchange_rates_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.exchange_rates_id_seq OWNER TO geraldbahati;

--
-- Name: exchange_rates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: geraldbahati
--

ALTER SEQUENCE public.exchange_rates_id_seq OWNED BY public.exchange_rates.id;


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


ALTER TABLE public.goose_db_version OWNER TO geraldbahati;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: geraldbahati
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.goose_db_version_id_seq OWNER TO geraldbahati;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: geraldbahati
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: guest_checkouts; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.guest_checkouts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    phone character varying(20),
    street_address character varying(255) NOT NULL,
    city character varying(255) NOT NULL,
    state character varying(255) NOT NULL,
    country character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.guest_checkouts OWNER TO geraldbahati;

--
-- Name: order_item_options; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.order_item_options (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_item_id uuid,
    option_type character varying(50) NOT NULL,
    option_value character varying(50) NOT NULL,
    additional_price numeric(10,2) DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.order_item_options OWNER TO geraldbahati;

--
-- Name: order_items; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid,
    product_id uuid,
    quantity integer NOT NULL,
    price numeric(10,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT check_quantity CHECK ((quantity > 0))
);


ALTER TABLE public.order_items OWNER TO geraldbahati;

--
-- Name: order_number_seq; Type: SEQUENCE; Schema: public; Owner: geraldbahati
--

CREATE SEQUENCE public.order_number_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.order_number_seq OWNER TO geraldbahati;

--
-- Name: order_payments; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.order_payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    amount numeric(10,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    payment_method_id integer,
    payment_status_id integer,
    checkout_request_id character varying(255) NOT NULL,
    result_code integer,
    result_desc text,
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.order_payments OWNER TO geraldbahati;

--
-- Name: order_shipments; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.order_shipments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid,
    tracking_id character varying(255) NOT NULL,
    status character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT check_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'shipped'::character varying, 'delivered'::character varying])::text[])))
);


ALTER TABLE public.order_shipments OWNER TO geraldbahati;

--
-- Name: order_status_history; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.order_status_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid,
    status character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT check_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'shipped'::character varying, 'delivered'::character varying, 'cancelled'::character varying])::text[])))
);


ALTER TABLE public.order_status_history OWNER TO geraldbahati;

--
-- Name: orders; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    status character varying(50) NOT NULL,
    total numeric(10,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    payment_status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    guest_checkout_id uuid,
    order_number character varying(20),
    CONSTRAINT check_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'shipped'::character varying, 'delivered'::character varying, 'cancelled'::character varying])::text[]))),
    CONSTRAINT check_user_or_guest CHECK (((user_id IS NOT NULL) OR (guest_checkout_id IS NOT NULL)))
);


ALTER TABLE public.orders OWNER TO geraldbahati;

--
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.password_reset_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    token character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.password_reset_tokens OWNER TO geraldbahati;

--
-- Name: payment_methods; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.payment_methods (
    id integer NOT NULL,
    method character varying(50) NOT NULL
);


ALTER TABLE public.payment_methods OWNER TO geraldbahati;

--
-- Name: payment_methods_id_seq; Type: SEQUENCE; Schema: public; Owner: geraldbahati
--

CREATE SEQUENCE public.payment_methods_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payment_methods_id_seq OWNER TO geraldbahati;

--
-- Name: payment_methods_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: geraldbahati
--

ALTER SEQUENCE public.payment_methods_id_seq OWNED BY public.payment_methods.id;


--
-- Name: payment_statuses; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.payment_statuses (
    id integer NOT NULL,
    status character varying(50) NOT NULL
);


ALTER TABLE public.payment_statuses OWNER TO geraldbahati;

--
-- Name: payment_statuses_id_seq; Type: SEQUENCE; Schema: public; Owner: geraldbahati
--

CREATE SEQUENCE public.payment_statuses_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payment_statuses_id_seq OWNER TO geraldbahati;

--
-- Name: payment_statuses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: geraldbahati
--

ALTER SEQUENCE public.payment_statuses_id_seq OWNED BY public.payment_statuses.id;


--
-- Name: product_attribute_values; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_attribute_values (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    attribute_id uuid,
    value character varying(255) NOT NULL,
    category_id uuid,
    hex_value character varying(7),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.product_attribute_values OWNER TO geraldbahati;

--
-- Name: product_attributes; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_attributes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    attribute_type_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.product_attributes OWNER TO geraldbahati;

--
-- Name: product_images; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    image_url text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    "position" integer
);


ALTER TABLE public.product_images OWNER TO geraldbahati;

--
-- Name: product_interactions; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_interactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    interaction_type character varying(50) NOT NULL,
    user_id uuid,
    interaction_time timestamp with time zone DEFAULT now(),
    CONSTRAINT check_interaction_type CHECK (((interaction_type)::text = ANY ((ARRAY['view'::character varying, 'click'::character varying, 'purchase'::character varying])::text[])))
);


ALTER TABLE public.product_interactions OWNER TO geraldbahati;

--
-- Name: product_option_values; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_option_values (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    option_id uuid,
    value_name character varying(255) NOT NULL,
    additional_price numeric(10,2) DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.product_option_values OWNER TO geraldbahati;

--
-- Name: product_options; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_options (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    option_name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.product_options OWNER TO geraldbahati;

--
-- Name: product_reviews; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    user_id uuid,
    rating integer,
    comment text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT product_reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


ALTER TABLE public.product_reviews OWNER TO geraldbahati;

--
-- Name: product_specifications; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_specifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    spec_name character varying(255) NOT NULL,
    spec_value text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.product_specifications OWNER TO geraldbahati;

--
-- Name: product_to_attribute_values; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_to_attribute_values (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    attribute_value_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.product_to_attribute_values OWNER TO geraldbahati;

--
-- Name: product_variants; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.product_variants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid,
    variant_name character varying(255) NOT NULL,
    variant_value character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    price numeric(10,2) NOT NULL,
    stock integer DEFAULT 0
);


ALTER TABLE public.product_variants OWNER TO geraldbahati;

--
-- Name: products; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    stock integer DEFAULT 0,
    category_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    featured boolean DEFAULT false,
    search_keyword tsvector DEFAULT ''::tsvector,
    part_number character varying(100) NOT NULL,
    meta_title character varying(255) DEFAULT ''::character varying,
    meta_description text DEFAULT ''::text,
    meta_keywords character varying(255) DEFAULT ''::character varying,
    slug character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    usd_price numeric(15,10) DEFAULT 0 NOT NULL,
    price_per_unit numeric(15,10) DEFAULT 0 NOT NULL,
    valid_from timestamp with time zone DEFAULT now() NOT NULL,
    valid_to timestamp with time zone,
    CONSTRAINT status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'archived'::character varying, 'active'::character varying])::text[])))
);


ALTER TABLE public.products OWNER TO geraldbahati;

--
-- Name: promotion_products; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.promotion_products (
    promotion_id uuid NOT NULL,
    product_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.promotion_products OWNER TO geraldbahati;

--
-- Name: promotions; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.promotions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    image_url character varying(255),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    slug character varying(255) DEFAULT ''::character varying NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    CONSTRAINT status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'archived'::character varying, 'draft'::character varying])::text[])))
);


ALTER TABLE public.promotions OWNER TO geraldbahati;

--
-- Name: rate_mv; Type: MATERIALIZED VIEW; Schema: public; Owner: geraldbahati
--

CREATE MATERIALIZED VIEW public.rate_mv AS
 SELECT COALESCE(( SELECT exchange_rates.rate_to_kes
           FROM public.exchange_rates
          WHERE (((exchange_rates.currency_code)::text = 'USD'::text) AND ((exchange_rates.valid_to IS NULL) OR (exchange_rates.valid_to >= now())) AND (exchange_rates.valid_from <= now()))
          ORDER BY exchange_rates.valid_from DESC
         LIMIT 1), (135)::numeric) AS rate_to_kes
  WITH NO DATA;


ALTER MATERIALIZED VIEW public.rate_mv OWNER TO geraldbahati;

--
-- Name: recommendations; Type: MATERIALIZED VIEW; Schema: public; Owner: geraldbahati
--

CREATE MATERIALIZED VIEW public.recommendations AS
 SELECT user_id,
    product_id,
    count(*) AS score
   FROM public.product_interactions
  GROUP BY user_id, product_id
  WITH NO DATA;


ALTER MATERIALIZED VIEW public.recommendations OWNER TO geraldbahati;

--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.refresh_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone
);


ALTER TABLE public.refresh_tokens OWNER TO geraldbahati;

--
-- Name: related_products; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.related_products (
    product_id uuid NOT NULL,
    related_product_id uuid NOT NULL
);


ALTER TABLE public.related_products OWNER TO geraldbahati;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role_name character varying(50) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.roles OWNER TO geraldbahati;

--
-- Name: shipment; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.shipment (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid,
    shipment_status character varying(50) NOT NULL,
    tracking_number character varying(50),
    shipped_date timestamp with time zone,
    delivery_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT check_shipment_status CHECK (((shipment_status)::text = ANY ((ARRAY['pending'::character varying, 'shipped'::character varying, 'delivered'::character varying])::text[])))
);


ALTER TABLE public.shipment OWNER TO geraldbahati;

--
-- Name: shopping_carts; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.shopping_carts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    session_id uuid DEFAULT gen_random_uuid(),
    total_items integer DEFAULT 0 NOT NULL,
    total_price numeric(10,2) DEFAULT 0.0 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT check_total_items CHECK ((total_items >= 0)),
    CONSTRAINT check_total_price CHECK ((total_price >= (0)::numeric))
);


ALTER TABLE public.shopping_carts OWNER TO geraldbahati;

--
-- Name: user_addresses; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.user_addresses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    address jsonb NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.user_addresses OWNER TO geraldbahati;

--
-- Name: user_preferences; Type: MATERIALIZED VIEW; Schema: public; Owner: geraldbahati
--

CREATE MATERIALIZED VIEW public.user_preferences AS
 SELECT user_id,
    product_id,
    count(*) AS interaction_count
   FROM public.product_interactions
  GROUP BY user_id, product_id
  WITH NO DATA;


ALTER MATERIALIZED VIEW public.user_preferences OWNER TO geraldbahati;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.user_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    role_id uuid
);


ALTER TABLE public.user_roles OWNER TO geraldbahati;

--
-- Name: users; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    hashed_password character varying(255),
    first_name character varying(255),
    last_name character varying(255),
    phone_number character varying(20),
    profile_image_url text,
    date_of_birth date,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    last_login timestamp with time zone,
    provider character varying(50),
    provider_id character varying(255),
    email_verified_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO geraldbahati;

--
-- Name: verification_tokens; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.verification_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    token character varying(512) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.verification_tokens OWNER TO geraldbahati;

--
-- Name: wishlist_items; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.wishlist_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wishlist_id uuid,
    product_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.wishlist_items OWNER TO geraldbahati;

--
-- Name: wishlists; Type: TABLE; Schema: public; Owner: geraldbahati
--

CREATE TABLE public.wishlists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    name character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.wishlists OWNER TO geraldbahati;

--
-- Name: attribute_types id; Type: DEFAULT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.attribute_types ALTER COLUMN id SET DEFAULT nextval('public.attribute_types_id_seq'::regclass);


--
-- Name: exchange_rates id; Type: DEFAULT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.exchange_rates ALTER COLUMN id SET DEFAULT nextval('public.exchange_rates_id_seq'::regclass);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: payment_methods id; Type: DEFAULT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.payment_methods ALTER COLUMN id SET DEFAULT nextval('public.payment_methods_id_seq'::regclass);


--
-- Name: payment_statuses id; Type: DEFAULT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.payment_statuses ALTER COLUMN id SET DEFAULT nextval('public.payment_statuses_id_seq'::regclass);


--
-- Data for Name: admin_approval_tokens; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.admin_approval_tokens (id, request_id, token, expires_at, created_at) FROM stdin;
\.


--
-- Data for Name: admin_requests; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.admin_requests (id, user_id, reason, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: attribute_types; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.attribute_types (id, name) FROM stdin;
1	size
2	color
3	RAM
4	storage
5	processor
\.


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.cart_items (id, shopping_cart_id, product_id, quantity, created_at, updated_at, price) FROM stdin;
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.categories (id, name, parent_id, created_at, updated_at, is_active, "position", image_url, description, meta_title, meta_description, is_featured, level, path, slug, last_updated_by) FROM stdin;
adaa23e0-3781-4756-82c9-41ebbe25a846	Flash Drives	ad681ad0-3440-4300-8963-6ce0195e08a3	2024-05-30 08:41:54.855332+00	2024-05-30 08:41:54.855332+00	t	1	\N	\N	\N	\N	f	3	computer_accessories.storage_devices.flash_drives	computer_accessories-storage_devices-flash_drives	\N
a40df5f1-575a-4510-bd3d-0d006b52743f	Printers	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-05-30 07:52:44.086394+00	2024-07-19 10:15:00.286838+00	t	3	category_images/1719996591-printer-3.png	\N	\N	\N	f	2	computing.printers	computing-printers	\N
4dbbf43a-904c-481d-ac3d-cb07947d61fc	HP	a40df5f1-575a-4510-bd3d-0d006b52743f	2024-05-30 07:53:00.29539+00	2024-05-30 07:53:00.29539+00	t	0	\N	\N	\N	\N	f	3	computing.printers.hp	computing-printers-hp	\N
5db5f6db-a6a8-40ab-a3e2-e7b53da06371	Epson	a40df5f1-575a-4510-bd3d-0d006b52743f	2024-05-30 07:53:07.412552+00	2024-05-30 07:53:07.412552+00	t	1	\N	\N	\N	\N	f	3	computing.printers.epson	computing-printers-epson	\N
9a10f89d-bb7f-4bfa-ac88-20445e4d24e5	Projectors	dc9d3918-0853-45c4-955f-0045278776c2	2024-05-30 08:40:16.186464+00	2024-05-30 08:40:16.186464+00	t	1	\N	\N	\N	\N	f	2	computer_accessories.projectors	computer_accessories-projectors	\N
7da2b2af-671c-45d5-b004-9f828fdf8487	Mercury Toners	7efe0308-d1a8-4205-aee7-e14d58fec374	2024-05-30 07:55:45.300119+00	2024-05-30 07:55:45.300119+00	t	3	\N	\N	\N	\N	f	3	computing.printer_supplies.mercury_toners	computing-printer_supplies-mercury_toners	\N
702d8fb1-3236-434c-b051-e0c684a82a02	DELL	f786f97c-20f8-4068-a54e-3ad83b476b9a	2024-05-30 08:39:36.310787+00	2024-05-30 08:39:36.310787+00	t	1	\N	\N	\N	\N	f	3	computer_accessories.monitors.dell	computer_accessories-monitors-dell	\N
b1ec66b3-6d64-461b-8d7f-0572a735df87	DELL	d34aa231-50cf-4106-85fc-3a15cf962714	2024-08-22 15:18:14.168038+00	2024-08-22 15:18:14.168038+00	t	-1	\N	\N	\N	\N	f	3	computing.servers.dell	computing-servers-dell	\N
14a4f3ff-8ea9-49b2-93f9-0020b2a90df1	Analog Cameras	fbe3d373-5ac9-4e1c-826e-ac314a99adc9	2024-05-30 08:45:16.857945+00	2024-05-30 08:45:16.857945+00	t	0	\N	\N	\N	\N	f	3	security_systems.hikvision.analog_cameras	security_systems-hikvision-analog_cameras	\N
8f67c6d1-2fa6-4acd-b074-0bb9e26115c8	IP & PTZ Cameras	fbe3d373-5ac9-4e1c-826e-ac314a99adc9	2024-05-30 08:45:40.401967+00	2024-05-30 08:45:40.401967+00	t	1	\N	\N	\N	\N	f	3	security_systems.hikvision.ip_ptz_cameras	security_systems-hikvision-ip_ptz_cameras	\N
8f2438c5-2da2-405d-ae6c-c5990d367a0e	NVR and DVR	fbe3d373-5ac9-4e1c-826e-ac314a99adc9	2024-05-30 08:46:08.545885+00	2024-05-30 08:46:08.545885+00	t	1	\N	\N	\N	\N	f	3	security_systems.hikvision.nvr_and_dvr	security_systems-hikvision-nvr_and_dvr	\N
d5698351-7503-414d-a011-8c81b19b5cba	Security Systems	\N	2024-05-30 07:45:51.199763+00	2024-05-30 07:45:51.199763+00	t	3	category_images/1721390843-security-systems.png	\N	\N	\N	f	1	security_systems	security_systems	\N
3824f4c8-c838-46fa-bf1b-d94f9c89f990	AVS	2ffb984b-9cce-4f9b-87bb-8868aca8a9ac	2024-05-30 08:20:00.032011+00	2024-05-30 08:20:00.032011+00	t	0	\N	\N	\N	\N	f	3	power_solutions.sollatek.avs	power_solutions-sollatek-avs	\N
5f909c85-d27e-425f-9808-47c6016afb7e	Espon	9a10f89d-bb7f-4bfa-ac88-20445e4d24e5	2024-05-30 08:40:36.579479+00	2024-05-30 08:40:36.579479+00	t	0	\N	\N	\N	\N	f	3	computer_accessories.projectors.espon	computer_accessories-projectors-espon	\N
348f94f9-4d4c-4476-9762-112d5fe6db5c	Lenovo	9a33d027-08b4-4b77-adb9-139c5175edb0	2024-05-30 07:51:02.774795+00	2024-05-30 07:51:02.774795+00	t	1	\N	\N	\N	\N	f	3	computing.desktops.lenovo	computing-desktops-lenovo	\N
8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	HP	b86ff805-4764-49c5-a0e6-e8e80b2377b7	2024-05-30 07:51:45.955314+00	2024-05-30 07:51:45.955314+00	t	0	\N	\N	\N	\N	f	3	computing.laptops.hp	computing-laptops-hp	\N
6c05d9f0-d36b-4980-a396-7641304bf770	AVR	2ffb984b-9cce-4f9b-87bb-8868aca8a9ac	2024-05-30 08:20:07.389711+00	2024-05-30 08:20:07.389711+00	t	1	\N	\N	\N	\N	f	3	power_solutions.sollatek.avr	power_solutions-sollatek-avr	\N
84f836d6-4eba-47c4-9d94-6ee5d05842a8	SVS	2ffb984b-9cce-4f9b-87bb-8868aca8a9ac	2024-05-30 08:20:16.433234+00	2024-05-30 08:20:16.433234+00	t	2	\N	\N	\N	\N	f	3	power_solutions.sollatek.svs	power_solutions-sollatek-svs	\N
bfb5bb0e-2765-46bf-89e7-ca807199e373	Lenovo	f786f97c-20f8-4068-a54e-3ad83b476b9a	2024-05-30 08:39:51.749101+00	2024-05-30 08:39:51.749101+00	t	2	\N	\N	\N	\N	f	3	computer_accessories.monitors.lenovo	computer_accessories-monitors-lenovo	\N
46f781b6-a1d2-4487-8a8d-57652a2b2f6f	Software	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-05-30 07:53:46.583594+00	2024-07-19 10:15:50.209799+00	t	4	category_images/1719996080-software.png	\N	\N	\N	f	2	computing.software	computing-software	\N
f7f0bebb-d508-4543-a49a-6911b8acb14c	Kingsons	e21370fd-0e28-4c30-9d09-044e16d1b326	2024-05-30 08:43:42.583445+00	2024-05-30 08:43:42.583445+00	t	1	\N	\N	\N	\N	f	3	computer_accessories.accessories.kingsons	computer_accessories-accessories-kingsons	\N
88a0b5b3-c284-4a62-8aba-587328b649cd	Point of Sale	a40df5f1-575a-4510-bd3d-0d006b52743f	2024-05-30 07:53:19.832102+00	2024-05-30 07:53:19.832102+00	t	2	\N	\N	\N	\N	f	3	computing.printers.point_of_sale	computing-printers-point_of_sale	\N
9bce2acf-3f55-43fa-82ff-a2f048cce0c4	Mercury	92cea60d-3f54-49af-b1f2-9c285aa53682	2024-05-30 08:02:54.441637+00	2024-05-30 08:02:54.441637+00	t	2	\N	\N	\N	\N	f	3	power_solutions.ups.mercury	power_solutions-ups-mercury	\N
b5cbdc81-ed9c-4c9f-8622-0a0ccd1cc65f	UPS Batteries	556ec16a-6f8b-474f-8abe-a4e8eb8b4f11	2024-05-30 08:03:30.398323+00	2024-05-30 08:03:30.398323+00	t	1	\N	\N	\N	\N	f	2	power_solutions.ups_batteries	power_solutions-ups_batteries	\N
cd4ae299-2fee-4860-a4cf-acf3b6d6a5e2	Brother	7efe0308-d1a8-4205-aee7-e14d58fec374	2024-05-30 07:55:34.509163+00	2024-05-30 07:55:34.509163+00	t	2	\N	\N	\N	\N	f	3	computing.printer_supplies.brother	computing-printer_supplies-brother	\N
556ec16a-6f8b-474f-8abe-a4e8eb8b4f11	Power Solutions	\N	2024-05-30 07:45:15.490854+00	2024-05-30 07:45:15.490854+00	t	1	category_images/1721390187-power-supply.png	\N	\N	\N	f	1	power_solutions	power_solutions	\N
dc9d3918-0853-45c4-955f-0045278776c2	Computer Accessories	\N	2024-05-30 07:45:35.403657+00	2024-05-30 07:45:35.403657+00	t	2	category_images/1721390251-computer-acccessories.png	\N	\N	\N	f	1	computer_accessories	computer_accessories	\N
77b30e88-1c02-4be9-8015-724c13c43cfc	Surveillance Hard Disks	fbe3d373-5ac9-4e1c-826e-ac314a99adc9	2024-05-30 08:46:45.660904+00	2024-05-30 08:46:45.660904+00	t	3	\N	\N	\N	\N	f	3	security_systems.hikvision.surveillance_hard_disks	security_systems-hikvision-surveillance_hard_disks	\N
466e57a8-bbfe-45ed-8cf0-bbaacffd500a	HP	d34aa231-50cf-4106-85fc-3a15cf962714	2024-08-22 15:16:48.976552+00	2024-08-22 15:16:48.976552+00	t	0	\N	\N	\N	\N	f	3	computing.servers.hp	computing-servers-hp	\N
b3e98195-e396-4f85-8417-e8b908709751	Antivirus	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-05-30 07:57:36.167015+00	2024-07-19 10:16:56.209343+00	t	6	category_images/1719996371-Kaspersky.png	\N	\N	\N	f	2	computing.antivirus	computing-antivirus	\N
ccdd2c93-cca6-4198-99d9-fe24beca47ff	Quickheal	b3e98195-e396-4f85-8417-e8b908709751	2024-05-30 07:58:10.139881+00	2024-05-30 07:58:10.139881+00	t	0	\N	\N	\N	\N	f	3	computing.antivirus.quickheal	computing-antivirus-quickheal	\N
e57fb18a-9944-4bfb-b9c8-6c56fefb8f7c	Kaspersky	b3e98195-e396-4f85-8417-e8b908709751	2024-05-30 07:58:25.199149+00	2024-05-30 07:58:25.199149+00	t	1	\N	\N	\N	\N	f	3	computing.antivirus.kaspersky	computing-antivirus-kaspersky	\N
4021f0b5-4bfe-4e35-9b32-d7936d57304b	Seqrite	b3e98195-e396-4f85-8417-e8b908709751	2024-05-30 07:58:35.22392+00	2024-05-30 07:58:35.22392+00	t	2	\N	\N	\N	\N	f	3	computing.antivirus.seqrite	computing-antivirus-seqrite	\N
92cea60d-3f54-49af-b1f2-9c285aa53682	UPS	556ec16a-6f8b-474f-8abe-a4e8eb8b4f11	2024-05-30 08:01:38.50382+00	2024-05-30 08:01:38.50382+00	t	0	\N	\N	\N	\N	f	2	power_solutions.ups	power_solutions-ups	\N
cb1fd6a6-18d2-430d-a44b-3425dd4f94c0	Mecer	92cea60d-3f54-49af-b1f2-9c285aa53682	2024-05-30 08:01:56.488307+00	2024-05-30 08:01:56.488307+00	t	0	\N	\N	\N	\N	f	3	power_solutions.ups.mecer	power_solutions-ups-mecer	\N
1da8f090-7491-4d05-8a4d-4bd63f8c8792	APC	92cea60d-3f54-49af-b1f2-9c285aa53682	2024-05-30 08:02:45.462093+00	2024-05-30 08:02:45.462093+00	t	1	\N	\N	\N	\N	f	3	power_solutions.ups.apc	power_solutions-ups-apc	\N
15472daa-427b-4fbd-ae4a-3df3d68ac72a	Cables	e21370fd-0e28-4c30-9d09-044e16d1b326	2024-05-30 08:43:51.024664+00	2024-05-30 08:43:51.024664+00	t	2	\N	\N	\N	\N	f	3	computer_accessories.accessories.cables	computer_accessories-accessories-cables	\N
97b071ac-8c17-4b17-9935-dede78b80404	Solar Panels	2ffb984b-9cce-4f9b-87bb-8868aca8a9ac	2024-05-30 08:20:51.798163+00	2024-05-30 08:20:51.798163+00	t	4	\N	\N	\N	\N	f	3	power_solutions.sollatek.solar_panels	power_solutions-sollatek-solar_panels	\N
f786f97c-20f8-4068-a54e-3ad83b476b9a	Monitors	dc9d3918-0853-45c4-955f-0045278776c2	2024-05-30 08:39:09.904979+00	2024-05-30 08:39:09.904979+00	t	0	\N	\N	\N	\N	f	2	computer_accessories.monitors	computer_accessories-monitors	\N
263fe9a0-4b39-4335-b8ce-79fa460b5759	HP	f786f97c-20f8-4068-a54e-3ad83b476b9a	2024-05-30 08:39:27.478864+00	2024-05-30 08:39:27.478864+00	t	0	\N	\N	\N	\N	f	3	computer_accessories.monitors.hp	computer_accessories-monitors-hp	\N
32b1910f-ae05-4a83-814a-1c52a5d979e4	Mercury Elite	b5cbdc81-ed9c-4c9f-8622-0a0ccd1cc65f	2024-05-30 08:04:01.573194+00	2024-05-30 08:04:01.573194+00	t	0	\N	\N	\N	\N	f	3	power_solutions.ups_batteries.mercury_elite	power_solutions-ups_batteries-mercury_elite	\N
c02fa807-ed3b-447b-a630-8b6bf1e3b54f	Inverters and Solar	556ec16a-6f8b-474f-8abe-a4e8eb8b4f11	2024-05-30 08:05:09.791105+00	2024-05-30 08:05:09.791105+00	t	2	\N	\N	\N	\N	f	2	power_solutions.inverters_and_solar	power_solutions-inverters_and_solar	\N
2ffb984b-9cce-4f9b-87bb-8868aca8a9ac	Sollatek	556ec16a-6f8b-474f-8abe-a4e8eb8b4f11	2024-05-30 08:06:34.302295+00	2024-05-30 08:06:34.302295+00	t	3	\N	\N	\N	\N	f	2	power_solutions.sollatek	power_solutions-sollatek	\N
74018456-4a2b-4e12-8bdb-6118d7013477	HP Toner	7efe0308-d1a8-4205-aee7-e14d58fec374	2024-05-30 07:56:00.598309+00	2024-05-30 07:56:00.598309+00	t	4	\N	\N	\N	\N	f	3	computing.printer_supplies.hp_toner	computing-printer_supplies-hp_toner	\N
fc14c8a6-f4a2-471b-9197-b8fd61287c15	Lenovo	b86ff805-4764-49c5-a0e6-e8e80b2377b7	2024-05-30 07:51:54.014622+00	2024-05-30 07:51:54.014622+00	t	1	\N	\N	\N	\N	f	3	computing.laptops.lenovo	computing-laptops-lenovo	\N
4542fb36-7119-4ed1-8b60-65c43541ba42	DELL	b86ff805-4764-49c5-a0e6-e8e80b2377b7	2024-05-30 07:52:01.898494+00	2024-05-30 07:52:01.898494+00	t	2	\N	\N	\N	\N	f	3	computing.laptops.dell	computing-laptops-dell	\N
7a24f933-90d6-417f-8b7b-e827f6875fd0	ASUS	b86ff805-4764-49c5-a0e6-e8e80b2377b7	2024-05-30 07:52:10.539434+00	2024-05-30 07:52:10.539434+00	t	3	\N	\N	\N	\N	f	3	computing.laptops.asus	computing-laptops-asus	\N
494456b3-f06c-4ff2-afa5-ef56248257ad	Hard Drives	ad681ad0-3440-4300-8963-6ce0195e08a3	2024-05-30 08:41:44.80047+00	2024-05-30 08:41:44.80047+00	t	0	\N	\N	\N	\N	f	3	computer_accessories.storage_devices.hard_drives	computer_accessories-storage_devices-hard_drives	\N
854542d4-ffa6-4ccd-8008-425065d3b6fa	Keyboards and Mouse	e21370fd-0e28-4c30-9d09-044e16d1b326	2024-05-30 08:44:11.193744+00	2024-05-30 08:44:11.193744+00	t	3	\N	\N	\N	\N	f	3	computer_accessories.accessories.keyboards_and_mouse	computer_accessories-accessories-keyboards_and_mouse	\N
fbe3d373-5ac9-4e1c-826e-ac314a99adc9	Hikvision	d5698351-7503-414d-a011-8c81b19b5cba	2024-05-30 08:44:51.11404+00	2024-05-30 08:44:51.11404+00	t	0	\N	\N	\N	\N	f	2	security_systems.hikvision	security_systems-hikvision	\N
c3c75bc0-4b1d-449a-93ae-9dab0132d84c	Surveillance Hard Disks	d5698351-7503-414d-a011-8c81b19b5cba	2024-05-30 08:46:59.380244+00	2024-05-30 08:46:59.380244+00	t	3	\N	\N	\N	\N	f	2	security_systems.surveillance_hard_disks	security_systems-surveillance_hard_disks	\N
87cd79b5-00bb-482a-a523-7c567221a130	Dahua	d5698351-7503-414d-a011-8c81b19b5cba	2024-05-30 08:47:24.508246+00	2024-05-30 08:47:24.508246+00	t	1	\N	\N	\N	\N	f	2	security_systems.dahua	security_systems-dahua	\N
fad4f0a9-1bad-4de3-a387-6d9df08327b6	Zkteco	d5698351-7503-414d-a011-8c81b19b5cba	2024-05-30 08:47:34.553915+00	2024-05-30 08:47:34.553915+00	t	2	\N	\N	\N	\N	f	2	security_systems.zkteco	security_systems-zkteco	\N
d34aa231-50cf-4106-85fc-3a15cf962714	Servers	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-07-19 10:14:11.96085+00	2024-07-19 10:14:11.96085+00	t	2	category_images/1721390063-server.png	\N	\N	\N	f	2	computing.servers	computing-servers	\N
a43e29ec-763c-4bca-999a-035b0b7404a7	Smart Classroom Solutions	\N	2024-07-19 10:26:44.652318+00	2024-07-19 10:26:44.652318+00	t	5	category_images/1721391739-smart-classroom-2.png	\N	\N	\N	f	1	smart_classroom_solutions	smart_classroom_solutions	\N
f7b8cba3-825f-4ec1-ac40-dd7cb4992eb9	Networking Solutions	\N	2024-07-19 10:26:57.614814+00	2024-07-19 10:26:57.614814+00	t	4	category_images/1721391128-network-solutions.webp	\N	\N	\N	f	1	networking_solutions	networking_solutions	\N
9a33d027-08b4-4b77-adb9-139c5175edb0	Desktops	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-05-30 07:50:35.835482+00	2024-05-30 07:50:35.835482+00	t	0	category_images/1719988942-Dell OptiPlex 3000 Tower-1.png	\N	\N	\N	f	2	computing.desktops	computing-desktops	\N
51223939-57c9-4b21-a222-1dfd6451c251	HP	9a33d027-08b4-4b77-adb9-139c5175edb0	2024-05-30 07:50:54.698269+00	2024-05-30 07:50:54.698269+00	t	0	\N	\N	\N	\N	f	3	computing.desktops.hp	computing-desktops-hp	\N
b4dcc1e4-9599-4df4-8831-64913a204bdb	Memory cards	ad681ad0-3440-4300-8963-6ce0195e08a3	2024-05-30 08:42:27.623751+00	2024-05-30 08:42:27.623751+00	t	2	\N	\N	\N	\N	f	3	computer_accessories.storage_devices.memory_cards	computer_accessories-storage_devices-memory_cards	\N
23474a07-d393-4af9-bdc2-e53067e7cc1b	Sony	9a10f89d-bb7f-4bfa-ac88-20445e4d24e5	2024-05-30 08:40:44.25378+00	2024-05-30 08:40:44.25378+00	t	1	\N	\N	\N	\N	f	3	computer_accessories.projectors.sony	computer_accessories-projectors-sony	\N
ad681ad0-3440-4300-8963-6ce0195e08a3	Storage Devices	dc9d3918-0853-45c4-955f-0045278776c2	2024-05-30 08:41:21.602276+00	2024-05-30 08:41:21.602276+00	t	2	\N	\N	\N	\N	f	2	computer_accessories.storage_devices	computer_accessories-storage_devices	\N
458fcf07-1c3e-4f34-9f0d-fd4203802bd2	Computing	\N	2024-05-30 07:45:03.044702+00	2024-05-30 07:45:03.044702+00	t	0	\N	\N	\N	\N	f	1	computing	computing	\N
37701388-7578-4570-9e0b-1fdd3bac555b	Power Supressors	2ffb984b-9cce-4f9b-87bb-8868aca8a9ac	2024-05-30 08:20:34.582456+00	2024-05-30 08:20:34.582456+00	t	3	\N	\N	\N	\N	f	3	power_solutions.sollatek.power_supressors	power_solutions-sollatek-power_supressors	\N
d4b15a4c-8796-419b-9aba-92e19cd3f2bf	DELL	9a33d027-08b4-4b77-adb9-139c5175edb0	2024-05-30 07:51:10.459105+00	2024-05-30 07:51:10.459105+00	t	2	\N	\N	\N	\N	f	3	computing.desktops.dell	computing-desktops-dell	\N
b86ff805-4764-49c5-a0e6-e8e80b2377b7	Laptops	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-05-30 07:51:30.534331+00	2024-05-30 07:51:30.534331+00	t	1	category_images/1719988986-asus-1.png	\N	\N	\N	f	2	computing.laptops	computing-laptops	\N
fd05e513-6648-4bb0-a397-858c914a7619	Apple	b86ff805-4764-49c5-a0e6-e8e80b2377b7	2024-05-30 07:52:26.071573+00	2024-05-30 07:52:26.071573+00	t	4	\N	\N	\N	\N	f	3	computing.laptops.apple	computing-laptops-apple	\N
c3b8678b-d00b-4511-b95c-bc460de84b1f	Microsoft	46f781b6-a1d2-4487-8a8d-57652a2b2f6f	2024-05-30 07:54:05.801815+00	2024-05-30 07:54:05.801815+00	t	0	\N	\N	\N	\N	f	3	computing.software.microsoft	computing-software-microsoft	\N
7efe0308-d1a8-4205-aee7-e14d58fec374	Printer Supplies	458fcf07-1c3e-4f34-9f0d-fd4203802bd2	2024-05-30 07:54:33.354166+00	2024-07-19 10:16:22.42467+00	t	5	category_images/1719994378-printer-accessory.png	\N	\N	\N	f	2	computing.printer_supplies	computing-printer_supplies	\N
00c07cec-21aa-4e81-9b27-d5b972573820	HP Cartridges	7efe0308-d1a8-4205-aee7-e14d58fec374	2024-05-30 07:55:09.172158+00	2024-05-30 07:55:09.172158+00	t	0	\N	\N	\N	\N	f	3	computing.printer_supplies.hp_cartridges	computing-printer_supplies-hp_cartridges	\N
f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	Espon Cartridges	7efe0308-d1a8-4205-aee7-e14d58fec374	2024-05-30 07:55:20.390583+00	2024-05-30 07:55:20.390583+00	t	1	\N	\N	\N	\N	f	3	computing.printer_supplies.espon_cartridges	computing-printer_supplies-espon_cartridges	\N
e21370fd-0e28-4c30-9d09-044e16d1b326	Accessories	dc9d3918-0853-45c4-955f-0045278776c2	2024-05-30 08:43:08.992514+00	2024-07-19 09:43:42.456599+00	t	3	\N	\N	\N	\N	f	2	computer_accessories.accessories	computer_accessories-accessories	\N
51e6ba8f-43ef-4367-8faa-d70820b47712	Longitech	e21370fd-0e28-4c30-9d09-044e16d1b326	2024-05-30 08:43:33.126295+00	2024-05-30 08:43:33.126295+00	t	0	\N	\N	\N	\N	f	3	computer_accessories.accessories.longitech	computer_accessories-accessories-longitech	\N
21d8bb3a-5521-45c4-be7b-b8c2af5ddc31	zvubi	e21370fd-0e28-4c30-9d09-044e16d1b326	2024-08-23 13:02:36.032361+00	2024-08-23 13:02:36.032361+00	t	1	\N	\N	\N	\N	f	3	computer_accessories.accessories.zvubi	computer_accessories-accessories-zvubi	\N
\.


--
-- Data for Name: discounts; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.discounts (id, product_id, discount_percentage, start_date, end_date, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: exchange_rates; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.exchange_rates (id, currency_code, rate_to_kes, valid_from, valid_to) FROM stdin;
1	USD	135.0000	2024-09-09 14:50:07.205708+00	2024-10-09 14:50:07.205708+00
\.


--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.goose_db_version (id, version_id, is_applied, tstamp) FROM stdin;
1	0	t	2024-08-14 07:00:38.216086
2	1	t	2024-08-14 07:00:38.223185
3	2	t	2024-08-14 07:00:38.257459
4	20240521120328	t	2024-08-14 07:00:38.277374
5	20240521173141	t	2024-08-14 07:00:38.279436
6	20240523234433	t	2024-08-14 07:00:38.283001
7	20240523235901	t	2024-08-14 07:00:38.29724
8	20240524000526	t	2024-08-14 07:00:38.301298
9	20240524003331	t	2024-08-14 07:00:38.303151
10	20240524085250	t	2024-08-14 07:00:38.304952
11	20240529172218	t	2024-08-14 07:00:38.305738
12	20240603092437	t	2024-08-14 07:00:38.306432
13	20240605124424	t	2024-08-14 07:00:38.308477
14	20240610062206	t	2024-08-14 07:00:38.316971
15	20240613105144	t	2024-08-14 07:00:38.320282
16	20240614164711	t	2024-08-14 07:00:38.322597
17	20240614184752	t	2024-08-14 07:00:38.324409
18	20240614184940	t	2024-08-14 07:00:38.325509
19	20240614185347	t	2024-08-14 07:00:38.329768
20	20240614190105	t	2024-08-14 07:00:38.331708
21	20240614191456	t	2024-08-14 07:00:38.333014
22	20240615061135	t	2024-08-14 07:00:38.335277
23	20240615061453	t	2024-08-14 07:00:38.338539
24	20240615061714	t	2024-08-14 07:00:38.341361
25	20240615154221	t	2024-08-14 07:00:38.342941
26	20240615155810	t	2024-08-14 07:00:38.349259
27	20240617023146	t	2024-08-14 07:00:38.35022
28	20240617023457	t	2024-08-14 07:00:38.351172
29	20240618123105	t	2024-08-14 07:00:38.352052
30	20240618123408	t	2024-08-14 07:00:38.359756
31	20240619060730	t	2024-08-14 07:00:38.361161
32	20240619060919	t	2024-08-14 07:00:38.362114
33	20240619061421	t	2024-08-14 07:00:38.363352
34	20240619061459	t	2024-08-14 07:00:38.364682
35	20240619120023	t	2024-08-14 07:00:38.36597
36	20240619120306	t	2024-08-14 07:00:38.368241
37	20240619132449	t	2024-08-14 07:00:38.369675
38	20240620125837	t	2024-08-14 07:00:38.370951
39	20240620183941	t	2024-08-14 07:00:38.375777
40	20240624070646	t	2024-08-14 07:00:38.376668
41	20240624072315	t	2024-08-14 07:00:38.377957
42	20240624072440	t	2024-08-14 07:00:38.381306
43	20240624092208	t	2024-08-14 07:00:38.383852
44	20240624092431	t	2024-08-14 07:00:38.386692
45	20240624092729	t	2024-08-14 07:00:38.390948
46	20240624093154	t	2024-08-14 07:00:38.396518
47	20240625065442	t	2024-08-14 07:00:38.411094
48	20240625071652	t	2024-08-14 07:00:38.417024
49	20240625072448	t	2024-08-14 07:00:38.418892
50	20240703083628	t	2024-08-14 07:00:38.419733
51	20240710091121	t	2024-08-14 07:00:38.422673
52	20240710091521	t	2024-08-14 07:00:38.425125
53	20240710112058	t	2024-08-14 07:00:38.425985
54	20240710164656	t	2024-08-14 07:00:38.434683
55	20240717113413	t	2024-08-14 07:00:38.435469
56	20240722084706	t	2024-08-14 07:00:38.438737
57	20240722085050	t	2024-08-14 07:00:38.441904
58	20240722112703	t	2024-08-14 07:00:38.445505
59	20240726120410	t	2024-08-14 07:00:38.44849
60	20240729102215	t	2024-08-14 07:00:38.453705
61	20240730060112	t	2024-08-14 07:00:38.464175
62	20240730134448	t	2024-08-14 07:00:38.465252
63	20240730144105	t	2024-08-14 07:00:38.474914
64	20240730150433	t	2024-08-14 07:00:38.476235
65	20240730200648	t	2024-08-14 07:00:38.476915
66	20240731114631	t	2024-08-14 07:00:38.4835
67	20240731123048	t	2024-08-14 07:00:38.488536
68	20240802125340	t	2024-08-14 07:00:38.493631
69	20240806093655	t	2024-08-14 07:00:38.494433
70	20240809115550	t	2024-08-14 07:00:38.498809
71	20240809132212	t	2024-08-14 07:00:38.499242
72	20240809142219	t	2024-08-14 07:00:38.504489
73	20240812124042	t	2024-08-14 07:00:38.504983
74	20240812130206	t	2024-08-14 07:00:38.508892
75	20240814112810	t	2024-08-14 08:36:00.456228
79	20240814114706	t	2024-08-14 09:50:07.905122
80	20240816084326	t	2024-08-16 05:44:53.98997
156	20240823114534	t	2024-08-23 13:06:51.9438
82	20240816084543	t	2024-08-16 05:59:34.53
83	20240820155551	t	2024-08-20 12:57:03.293354
84	20240821100405	t	2024-08-21 07:05:01.997614
163	20240823124354	t	2024-08-23 13:23:07.459223
164	20240823132948	t	2024-08-23 13:23:07.486561
121	20240821134808	t	2024-08-23 11:52:13.88034
\.


--
-- Data for Name: guest_checkouts; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.guest_checkouts (id, email, first_name, last_name, phone, street_address, city, state, country, created_at, updated_at) FROM stdin;
4220788c-3910-4ce7-92a3-392cfa044776	geraldbahati@gmail.com	Gerald	Bahati	254790329620	Kahawa Sukari	Nairobi	Nairobi	Kenya	2024-08-22 15:24:58.739632+00	2024-08-22 15:24:58.739632+00
\.


--
-- Data for Name: order_item_options; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.order_item_options (id, order_item_id, option_type, option_value, additional_price, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.order_items (id, order_id, product_id, quantity, price, created_at, updated_at) FROM stdin;
73995c7a-f518-47d2-af76-2c8558fa6d5e	165a6012-6408-4446-b574-824e80d7daec	498018a3-175b-432f-96c5-58af7b509a3a	1	3375.00	2024-08-22 15:24:58.748925+00	2024-08-22 15:24:58.748925+00
\.


--
-- Data for Name: order_payments; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.order_payments (id, order_id, amount, created_at, payment_method_id, payment_status_id, checkout_request_id, result_code, result_desc, updated_at) FROM stdin;
f58c225e-0ca9-4f33-bef9-02fd5ea0e34a	165a6012-6408-4446-b574-824e80d7daec	3375.00	2024-08-22 15:24:58.750256+00	2	2	ws_CO_22082024182517087790329620	\N	\N	2024-08-27 14:14:14.604323+00
\.


--
-- Data for Name: order_shipments; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.order_shipments (id, order_id, tracking_id, status, created_at) FROM stdin;
\.


--
-- Data for Name: order_status_history; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.order_status_history (id, order_id, status, created_at) FROM stdin;
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.orders (id, user_id, status, total, created_at, updated_at, payment_status, guest_checkout_id, order_number) FROM stdin;
165a6012-6408-4446-b574-824e80d7daec	\N	pending	3375.00	2024-08-22 15:24:58.743477+00	2024-08-22 15:24:58.743477+00	pending	4220788c-3910-4ce7-92a3-392cfa044776	000001
\.


--
-- Data for Name: password_reset_tokens; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.password_reset_tokens (id, email, token, expires_at, created_at) FROM stdin;
\.


--
-- Data for Name: payment_methods; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.payment_methods (id, method) FROM stdin;
1	cash
2	credit_card
3	debit_card
4	paypal
5	mpesa
\.


--
-- Data for Name: payment_statuses; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.payment_statuses (id, status) FROM stdin;
1	pending
2	paid
3	failed
\.


--
-- Data for Name: product_attribute_values; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_attribute_values (id, attribute_id, value, category_id, hex_value, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_attributes; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_attributes (id, name, attribute_type_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_images; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_images (id, product_id, image_url, created_at, updated_at, "position") FROM stdin;
2f4517c1-c518-4ffb-9c48-62df532b999f	7425ba85-fff4-4c72-a2ca-583e8462b5a5	product-images/1717395064-Dell OptiPlex 3000 Tower-1.png	2024-06-03 06:11:05.436328+00	2024-06-03 06:11:05.436328+00	1
0ca91a1a-9035-4bb0-9b90-ccb6e43ea076	7425ba85-fff4-4c72-a2ca-583e8462b5a5	product-images/1717395074-Dell OptiPlex 3000 Tower-2.png	2024-06-03 06:11:16.049007+00	2024-06-03 06:11:16.049007+00	2
f02e36aa-c14e-4563-b672-a40f547d48f1	7425ba85-fff4-4c72-a2ca-583e8462b5a5	product-images/1717395082-Dell OptiPlex 3000 Tower-3.png	2024-06-03 06:11:24.362451+00	2024-06-03 06:11:24.362451+00	3
cb489971-09cb-4975-b579-9bc490a5b402	ab773cb9-cf21-4288-adec-f1bd78050c84	product-images/1722609587-816N0EA-HP-laptop.png	2024-08-02 14:39:48.855574+00	2024-08-02 14:39:48.855574+00	1
3d10260e-8f1c-413c-bc65-61d402a53bc0	e3519734-ea27-419f-928d-81d90d8399d9	product-images/1722606747-816N0EA-HP-laptop.png	2024-08-02 13:52:28.231768+00	2024-08-02 13:52:28.231768+00	1
57789a22-99d9-40a7-855f-20ac901a6a4a	e3519734-ea27-419f-928d-81d90d8399d9	product-images/1722606747-816N0EA-HP-laptop-4.png	2024-08-02 13:52:28.22749+00	2024-08-02 13:52:28.22749+00	2
3a50d983-235e-4664-9aa3-5719d85a6f20	e3519734-ea27-419f-928d-81d90d8399d9	product-images/1722606747-816N0EA-HP-laptop-5.png	2024-08-02 13:52:28.230337+00	2024-08-02 13:52:28.230337+00	3
8b0718e5-b5a1-411e-9072-a29b56b67840	e3519734-ea27-419f-928d-81d90d8399d9	product-images/1722606747-816N0EA-HP-laptop-2.png	2024-08-02 13:52:28.233032+00	2024-08-02 13:52:28.233032+00	4
189e4f37-f83a-4c7f-91bf-cd0eba07b4c7	e3519734-ea27-419f-928d-81d90d8399d9	product-images/1722606747-816N0EA-HP-laptop-3.png	2024-08-02 13:52:28.233993+00	2024-08-02 13:52:28.233993+00	5
6e8695a8-3c90-4021-8c7b-8d2414e2640c	6ee18311-c27c-43d5-a07f-9560c5f20105	product-images/1723104820-HP-Pro-Tower-290-G9-Intel-Core-i3-12th-Gen-8GB-RAM-1TB-HDD-18.5-Inch-HD-Monitor-Business-Desktop.jpg	2024-08-08 08:13:41.151637+00	2024-08-08 08:13:41.151637+00	1
1ef82fc8-4f29-4a97-b772-582bb928b955	6258e14f-8fe4-4755-a7e4-931b212a6fbf	product-images/1723105274-HP-Pro-Tower-290-G9-Intel-Core-i3-12th-Gen-8GB-RAM-1TB-HDD-18.5-Inch-HD-Monitor-Business-Desktop.jpg	2024-08-08 08:21:14.986051+00	2024-08-08 08:21:14.986051+00	1
ce9abc75-ff4c-4a67-842e-cce0b552743e	a1401070-cd3a-4caa-9e1d-02ecd64ce290	product-images/1723106223-pc-hp-pro-tower-200-g9-6d419eabed31095.jpg	2024-08-08 08:37:04.231431+00	2024-08-08 08:37:04.231431+00	1
7e55cd85-9799-413c-8031-030a9fbfb1d9	7fbbe4c0-8c78-4f32-832d-056a220612b4	product-images/1723107660-HP-P204V-19.5.jpg	2024-08-08 09:01:00.887767+00	2024-08-08 09:01:00.887767+00	1
4043378e-4f8c-4760-81ad-27df7423096b	7fbbe4c0-8c78-4f32-832d-056a220612b4	product-images/1723107660-HP-P204V-19.5-monitor-in-kenya.jpg	2024-08-08 09:01:00.891312+00	2024-08-08 09:01:00.891312+00	2
c997da00-94df-40ce-92ef-1c4682df2681	d994e04b-1915-4173-b4dd-c78869cb3d0a	product-images/1723109618-HP M22F 21.5-inch 2.JPG	2024-08-08 09:33:38.765289+00	2024-08-08 09:33:38.765289+00	1
8fe5aad3-a7d2-4ab3-bcba-a40bdfc9652e	d994e04b-1915-4173-b4dd-c78869cb3d0a	product-images/1723109618-HP M22F 21.5-inch.JPG	2024-08-08 09:33:38.767143+00	2024-08-08 09:33:38.767143+00	2
e4c9b4cb-f1bf-4244-b32b-02d806ce4549	d994e04b-1915-4173-b4dd-c78869cb3d0a	product-images/1723109819-HP M22F 21.5-inch 4.JPG	2024-08-08 09:37:00.346241+00	2024-08-08 09:37:00.346241+00	3
6077faf1-f464-444e-9c8a-105c247f5608	d994e04b-1915-4173-b4dd-c78869cb3d0a	product-images/1723109819-HP M22F 21.5-inch FHD Monitor,3.JPG	2024-08-08 09:37:00.348124+00	2024-08-08 09:37:00.348124+00	4
5484cc14-bbdd-45ea-b8ac-e6b7bbaa1efc	c2c1f3ec-dc41-4724-89c2-e01a9336e247	product-images/1723116666-HP E23 G4 23-inch Diagonal IPS FHD Monitor 3.JPG	2024-08-08 11:31:07.120717+00	2024-08-08 11:31:07.120717+00	2
3089e9fb-19b7-4772-a955-a9202311387b	c2c1f3ec-dc41-4724-89c2-e01a9336e247	product-images/1723116666-HP E23 G4 23-inch Diagonal IPS FHD Monitor 1.JPG	2024-08-08 11:31:07.122545+00	2024-08-08 11:31:07.122545+00	1
2f74e31b-d3f2-421f-8877-342b181d15e8	c2c1f3ec-dc41-4724-89c2-e01a9336e247	product-images/1723116666-HP E23 G4 23-inch Diagonal IPS FHD Monitor 2.JPG	2024-08-08 11:31:07.125021+00	2024-08-08 11:31:07.125021+00	3
68c63f3f-68da-4eb6-8a78-9ebf37cedeee	c2c1f3ec-dc41-4724-89c2-e01a9336e247	product-images/1723116666-HP E23 G4 23-inch Diagonal IPS FHD Monitor 4.JPG	2024-08-08 11:31:07.126197+00	2024-08-08 11:31:07.126197+00	4
1eee22c2-deb4-4c83-926f-f8e241fb483b	cfdacdb2-a03d-4539-ab93-9f61e6d6b148	product-images/1722953002-Lexar 512 SSD.JPG	2024-08-06 14:03:22.479952+00	2024-08-06 14:03:22.479952+00	1
10ce98af-4c01-4e61-bdcb-16a98abe8934	cfdacdb2-a03d-4539-ab93-9f61e6d6b148	product-images/1722951564-Lexar External SSD.jpg	2024-08-06 13:39:25.099455+00	2024-08-06 13:39:25.099455+00	3
54d3bcda-0880-4c6c-bf74-c0ad77f0c86d	cfdacdb2-a03d-4539-ab93-9f61e6d6b148	product-images/1722952739-Lexar SSD.jpg	2024-08-06 13:58:59.900785+00	2024-08-06 13:58:59.900785+00	2
2c159e05-b7d4-4cd3-a371-a23f9fd63a84	cfdacdb2-a03d-4539-ab93-9f61e6d6b148	product-images/1722952739-Lexar SSD 2.jpg	2024-08-06 13:58:59.902644+00	2024-08-06 13:58:59.902644+00	4
5738420f-3697-4691-b692-84bee119372c	aa5602f2-3382-41fd-8ad2-678d0ac7c903	product-images/1723104110-Lexar 1 TB.JPG	2024-08-08 08:01:51.003965+00	2024-08-08 08:01:51.003965+00	1
d76d95c5-c072-46ce-afe1-45dfe4ba05fb	7ec8298b-688b-4862-92ae-b82c0c22cddc	product-images/1723121717-189T0AA-hp-e24u-g4-23-8-inch-ips-fhd-usb-c-monitor-3.png	2024-08-08 12:55:17.689416+00	2024-08-08 12:55:17.689416+00	2
8f0a07a5-5635-4eb7-9c0f-002e0898eb0f	7ec8298b-688b-4862-92ae-b82c0c22cddc	product-images/1723121717-189T0AA-hp-e24u-g4-23-8-inch-ips-fhd-usb-c-monitor-2.png	2024-08-08 12:55:17.693973+00	2024-08-08 12:55:17.693973+00	3
d628963c-15cc-4c8c-9b43-f1f1d463c24c	7ec8298b-688b-4862-92ae-b82c0c22cddc	product-images/1723121717-189T0AA-hp-e24u-g4-23-8-inch-ips-fhd-usb-c-monitor-4.png	2024-08-08 12:55:17.695962+00	2024-08-08 12:55:17.695962+00	4
ab0a6c32-0227-425e-bbfd-e3b9d598300f	7ec8298b-688b-4862-92ae-b82c0c22cddc	product-images/1723121717-189T0AA-hp-e24u-g4-23-8-inch-ips-fhd-usb-c-monitor.png	2024-08-08 12:55:17.697441+00	2024-08-08 12:55:17.697441+00	1
9afde7c0-b0f2-4949-af8c-c5598a2aa8e9	3b74114e-33ee-4bac-b182-f1256048009f	product-images/1723123532-CE411A-hp-305a-cyan-laserjet-toner-cartridge.png	2024-08-08 13:25:33.006545+00	2024-08-08 13:25:33.006545+00	1
ab6f941d-80f9-4111-94c8-691bd4a41c28	3b74114e-33ee-4bac-b182-f1256048009f	product-images/1723123532-CE411A-hp-305a-cyan-laserjet-toner-cartridge-2.png	2024-08-08 13:25:33.008384+00	2024-08-08 13:25:33.008384+00	2
ab5f8376-0c5a-4c0e-923c-585785d3ad0e	3b74114e-33ee-4bac-b182-f1256048009f	product-images/1723123532-CE411A-hp-305a-cyan-laserjet-toner-cartridge-3.png	2024-08-08 13:25:33.009805+00	2024-08-08 13:25:33.009805+00	3
a57ecf18-158d-45f4-b682-3cec064d3a1b	30bcce26-48ab-4465-85c9-c610b6cd46d2	product-images/1724230355-promotion-2.png	2024-08-21 08:52:38.442447+00	2024-08-21 08:52:38.442447+00	1
c76034b7-0ccc-4ed1-80aa-fc3567a6b973	30bcce26-48ab-4465-85c9-c610b6cd46d2	product-images/1724230355-promotion-1.png	2024-08-21 08:52:38.456799+00	2024-08-21 08:52:38.456799+00	2
\.


--
-- Data for Name: product_interactions; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_interactions (id, product_id, interaction_type, user_id, interaction_time) FROM stdin;
\.


--
-- Data for Name: product_option_values; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_option_values (id, option_id, value_name, additional_price, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_options; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_options (id, product_id, option_name, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_reviews; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_reviews (id, product_id, user_id, rating, comment, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_specifications; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_specifications (id, product_id, spec_name, spec_value, created_at, updated_at) FROM stdin;
f010e4de-c15a-4732-bdfd-fc4609fa49b4	30bcce26-48ab-4465-85c9-c610b6cd46d2	Test1	Test spec description	2024-08-21 08:52:38.458869+00	2024-08-21 08:52:38.458869+00
\.


--
-- Data for Name: product_to_attribute_values; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_to_attribute_values (id, product_id, attribute_value_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_variants; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.product_variants (id, product_id, variant_name, variant_value, created_at, updated_at, price, stock) FROM stdin;
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.products (id, name, description, stock, category_id, created_at, updated_at, created_by, updated_by, featured, search_keyword, part_number, meta_title, meta_description, meta_keywords, slug, status, usd_price, price_per_unit, valid_from, valid_to) FROM stdin;
30bcce26-48ab-4465-85c9-c610b6cd46d2	HP 130A Black LaserJet Toner Cartridge	HP 130A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-21 08:52:35.126271+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'130a':2A,8B,14C,20 'black':3A,9B,15C,21 'cartridg':6A,12B,18C,24 'hp':1A,7B,13C,19 'laserjet':4A,10B,16C,22 'toner':5A,11B,17C,23		HP 130A Black LaserJet Toner Cartridge	HP 130A Black LaserJet Toner Cartridge	\N	hp-130a-black-laserjet-toner-cartridge	active	9.6300000000	9.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
63cd4e8b-b225-4681-85e0-96b396a93e4d	HP 507A Magenta LaserJet Toner Cartridge	HP 507A  Magenta LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'507a':2A,8B 'cartridg':6A,12B 'ce403a':13A 'hp':1A,7B 'laserjet':4A,10B 'magenta':3A,9B 'toner':5A,11B	CE403A				hp-507a-magenta-laserjet-toner-cartridge	active	56.0000000000	56.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
554aeeab-c1e1-47e7-b330-b3e5e95b2727	HP 59A Black LaserJet Toner Cartridge	HP 59A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-21 09:17:38.60495+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'59a':2A,8B,15C,21 'black':3A,9B,16C,22 'cartridg':6A,12B,19C,25 'cf259a':13A 'hp':1A,7B,14C,20 'laserjet':4A,10B,17C,23 'toner':5A,11B,18C,24	CF259A	HP 59A Black LaserJet Toner Cartridge	HP 59A Black LaserJet Toner Cartridge	\N	hp-59a-black-laserjet-toner-cartridge	active	88.8900000000	78.9800000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
d04d95c7-fb6b-4b73-9959-dda943c07367	Transcend 256GB M.2  	Transcend 256GB, M.2 2280,PCIe Gen3x4, M-Key, 3D TLC, DRAM-less	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2280':7B '256gb':2A,5B '3d':13B 'dram':16B 'dram-less':15B 'gen3x4':9B 'key':12B 'less':17B 'm':11B 'm-key':10B 'm.2':3A,6B 'pcie':8B 'tlc':14B 'transcend':1A,4B 'ts256gmte110s':18A	TS256GMTE110S				transcend-256gb-m-2-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
20387f3c-ce2e-42b9-b0ee-de4463eb159a	Epson SIDM Black Ribbon Cartridge for LQ-350/LQ-300	Epson SIDM Black Ribbon Cartridge for LQ-350/LQ-300	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-350':8A,17B '/lq-300':9A,18B 'black':3A,12B 'c13s015633ba':19A 'cartridg':5A,14B 'epson':1A,10B 'lq':7A,16B 'ribbon':4A,13B 'sidm':2A,11B	C13S015633BA				epson-sidm-black-ribbon-cartridge-for-lq-350-lq-300	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
bb34e602-6f1d-4536-a1af-7cfced56c7f9	Epson SIDM Black Ribbon Cartridge for LX-350 / LX-300	Epson SIDM Black Ribbon Cartridge for LX-350 / LX-300	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-300':10A,20B '-350':8A,18B 'black':3A,13B 'c13s015637ba':21A 'cartridg':5A,15B 'epson':1A,11B 'lx':7A,9A,17B,19B 'ribbon':4A,14B 'sidm':2A,12B	C13S015637BA				epson-sidm-black-ribbon-cartridge-for-lx-350-lx-300	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
b9459ecc-4027-4a8c-8e0f-3e2cc7443e67	Epson SIDM Black Ribbon Cartridge for LQ-670/LQ-680	Epson SIDM Black Ribbon Cartridge for LQ-670/LQ-680	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-670':8A,17B '/lq-680':9A,18B 'black':3A,12B 'c13s015262ba':19A 'cartridg':5A,14B 'epson':1A,10B 'lq':7A,16B 'ribbon':4A,13B 'sidm':2A,11B	C13S015262BA				epson-sidm-black-ribbon-cartridge-for-lq-670-lq-680	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
81b0ff69-f7e6-414a-a71b-8d994db61229	Epson 110 XL 120ml EcoTank Pigment black ink bottle	Epson 110 XL 120ml EcoTank Pigment black ink bottle	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'110':2A,11B '120ml':4A,13B 'black':7A,16B 'bottl':9A,18B 'c13t03p14a':19A 'ecotank':5A,14B 'epson':1A,10B 'ink':8A,17B 'pigment':6A,15B 'xl':3A,12B	C13T03P14A				epson-110-xl-120ml-ecotank-pigment-black-ink-bottle	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
6853bc4b-81cc-4f9d-a606-03d77e5a5ef6	Epson T6641 EcoTank Black Ink Bottle 70.0 ml	Epson T6641 EcoTank Black Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'70.0':7A,15B 'black':4A,12B 'bottl':6A,14B 'c13t66414a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 't6641':2A,10B	C13T66414A				epson-t6641-ecotank-black-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
498018a3-175b-432f-96c5-58af7b509a3a	HP DisplayPort To VGA Adapter    	HP DisplayPort To VGA Adapter    	0	15472daa-427b-4fbd-ae4a-3df3d68ac72a	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'adapt':5A,10B 'as615aa':11A 'displayport':2A,7B 'hp':1A,6B 'vga':4A,9B	AS615AA				hp-displayport-to-vga-adapter-	active	25.0000000000	25.0000000000	2024-08-21 07:05:01.997614+00	\N
34675d33-338f-4403-a03c-91dc17735cb7	Logitech Rally Mounting Kit	Logitech Rally Mounting Kit	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-001644':10A '939':9A 'kit':4A,8B 'logitech':1A,5B 'mount':3A,7B 'ralli':2A,6B	939-001644				logitech-rally-mounting-kit	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
774b59d3-04c0-4f93-a25f-7facda2e89ee	Epson T6642 EcoTank Cyan Ink Bottle 70.0 ml	Epson T6642 EcoTank Cyan Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'70.0':7A,15B 'bottl':6A,14B 'c13t66424a':17A 'cyan':4A,12B 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 't6642':2A,10B	C13T66424A				epson-t6642-ecotank-cyan-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7daf32db-1a7f-46a3-86a4-c2b9d70c8c52	Epson T6644 EcoTank Yellow Ink Bottle 70.0 ml	Epson T6644 EcoTank Yellow Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'70.0':7A,15B 'bottl':6A,14B 'c13t66444a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 't6644':2A,10B 'yellow':4A,12B	C13T66444A				epson-t6644-ecotank-yellow-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
f7f0f9da-2c8a-4a7c-a376-4046c97d6c3c	Epson T7741 EcoTank Pigment Black ink bottle 140ml	Epson T7741 EcoTank Pigment Black ink bottle 140ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'140ml':8A,16B 'black':5A,13B 'bottl':7A,15B 'c13t77414a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':6A,14B 'pigment':4A,12B 't7741':2A,10B	C13T77414A				epson-t7741-ecotank-pigment-black-ink-bottle-140ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
ca365fd8-6235-4b28-97ce-259ccc570fac	Epson SIDM Black Ribbon Cartridge for LQ-690 Series	Epson SIDM Black Ribbon Cartridge for LQ-690 Series	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-690':8A,17B 'black':3A,12B 'c13s015610ba':19A 'cartridg':5A,14B 'epson':1A,10B 'lq':7A,16B 'ribbon':4A,13B 'seri':9A,18B 'sidm':2A,11B	C13S015610BA				epson-sidm-black-ribbon-cartridge-for-lq-690-series	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
cc9404f4-07e4-44b2-a341-bd6d0409c5c3	Epson SIDM Black Ribbon Cartridge LQ-2180/2190	Epson SIDM Black Ribbon Cartridge LQ-2180/2190	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-2180':7A,15B '/2190':8A,16B 'black':3A,11B 'c13s015086ba':17A 'cartridg':5A,13B 'epson':1A,9B 'lq':6A,14B 'ribbon':4A,12B 'sidm':2A,10B	C13S015086BA				epson-sidm-black-ribbon-cartridge-lq-2180-2190	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
a10da638-32b4-4b66-8e98-1fe28b5a1ed7	Epson 101 EcoTank Cyan Ink Bottle 70.0 ml	Epson 101 EcoTank Cyan Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'101':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t03v24a':17A 'cyan':4A,12B 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B	C13T03V24A				epson-101-ecotank-cyan-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
b6061350-825e-4688-ab22-8aaa5bc46b5c	Epson 101 EcoTank Magenta Ink Bottle 70.0 ml	Epson 101 EcoTank Magenta Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'101':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t03v34a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'magenta':4A,12B 'ml':8A,16B	C13T03V34A				epson-101-ecotank-magenta-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
f393a011-8519-465e-9ad5-2e2602ece7c0	Epson 101 EcoTank Yellow Ink Bottle 70.0 ml	Epson 101 EcoTank Yellow Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'101':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t03v44a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 'yellow':4A,12B	C13T03V44A				epson-101-ecotank-yellow-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
0e735410-2271-42f3-849d-d0023ac94b74	Epson 103 EcoTank Cyan Ink Bottle 65.0 ml	Epson 103 EcoTank Cyan Ink Bottle 65.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'103':2A,10B '65.0':7A,15B 'bottl':6A,14B 'c13t00s24a':17A 'cyan':4A,12B 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B	C13T00S24A				epson-103-ecotank-cyan-ink-bottle-65-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
cc7ad0b7-dde3-4f89-aced-dbb87b07e158	Epson 103 EcoTank Magenta Ink Bottle 65.0 ml	Epson 103 EcoTank Magenta Ink Bottle 65.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'103':2A,10B '65.0':7A,15B 'bottl':6A,14B 'c13t00s34a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'magenta':4A,12B 'ml':8A,16B	C13T00S34A				epson-103-ecotank-magenta-ink-bottle-65-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
c0f7b3f0-1046-4e28-8e78-f5cfa767957d	Epson 112 EcoTank Pigment Black Ink Bottle 127.0 ml	Epson 112 EcoTank Pigment Black Ink Bottle 127.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'112':2A,11B '127.0':8A,17B 'black':5A,14B 'bottl':7A,16B 'c13t06c14a':19A 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'ml':9A,18B 'pigment':4A,13B	C13T06C14A				epson-112-ecotank-pigment-black-ink-bottle-127-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
20e39790-b6b1-43f1-b31c-a9c8da07d53c	Epson 112 EcoTank Pigment Magenta Ink Bottle 70.0 ml	Epson 112 EcoTank Pigment Magenta Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'112':2A,11B '70.0':8A,17B 'bottl':7A,16B 'c13t06c34a':19A 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'magenta':5A,14B 'ml':9A,18B 'pigment':4A,13B	C13T06C34A				epson-112-ecotank-pigment-magenta-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
1dd35d2f-e0c9-4fbe-a67b-2e6b0f33e78c	Epson 112 EcoTank Pigment Cyan Ink Bottle 70.0 ml	Epson 112 EcoTank Pigment Cyan Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'112':2A,11B '70.0':8A,17B 'bottl':7A,16B 'c13t06c24a':19A 'cyan':5A,14B 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'ml':9A,18B 'pigment':4A,13B	C13T06C24A				epson-112-ecotank-pigment-cyan-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
6f2bb74b-72a4-4d01-8f05-99f5c53773b1	Epson 112 EcoTank Pigment Yellow Ink Bottle 70.0 ml	Epson 112 EcoTank Pigment Yellow Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'112':2A,11B '70.0':8A,17B 'bottl':7A,16B 'c13t06c44a':19A 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'ml':9A,18B 'pigment':4A,13B 'yellow':5A,14B	C13T06C44A				epson-112-ecotank-pigment-yellow-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
6c50e0ac-84c8-4e7f-9ab9-cae6a21abf56	Epson T6731 EcoTank Black Ink Bottle 70.0 ml	Epson T6731 EcoTank Black Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'70.0':7A,15B 'black':4A,12B 'bottl':6A,14B 'c13t67314a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 't6731':2A,10B	C13T67314A				epson-t6731-ecotank-black-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
86a57f2e-81cf-4f02-8c19-90f0cebbda38	Epson T6735 EcoTank Light Cyan Ink Bottle 70.0 ml	Epson T6735 EcoTank Light Cyan Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'70.0':8A,17B 'bottl':7A,16B 'c13t67354a':19A 'cyan':5A,14B 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'light':4A,13B 'ml':9A,18B 't6735':2A,11B	C13T67354A				epson-t6735-ecotank-light-cyan-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
3b572d39-a052-48bd-a070-a18aaa3ff829	Epson T6736 EcoTank Light Magenta Ink Bottle 70.0 ml	Epson T6736 EcoTank Light Magenta Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'70.0':8A,17B 'bottl':7A,16B 'c13t67364a':19A 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'light':4A,13B 'magenta':5A,14B 'ml':9A,18B 't6736':2A,11B	C13T67364A				epson-t6736-ecotank-light-magenta-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
6cf9e6a0-e7e6-4605-a4d5-68d8a2f1ede1	Epson 101 EcoTank Black Ink Bottle 127.0 ml	Epson 101 EcoTank Black Ink Bottle 127.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'101':2A,10B '127.0':7A,15B 'black':4A,12B 'bottl':6A,14B 'c13t03v14a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B	C13T03V14A				epson-101-ecotank-black-ink-bottle-127-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
52a57789-d29d-40c6-9e4e-a92b66aac8fc	Epson 115 EcoTank Photo Black ink bottle 70.0 ml	Epson 115 EcoTank Photo Black ink bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'115':2A,11B '70.0':8A,17B 'black':5A,14B 'bottl':7A,16B 'c13t07d14a':19A 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'ml':9A,18B 'photo':4A,13B	C13T07D14A				epson-115-ecotank-photo-black-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
b330951c-a214-46f4-8298-d654b1af4db5	Epson 115 EcoTank Grey ink bottle 70.0 ml	Epson 115 EcoTank Grey ink bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'115':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t07d54a':17A 'ecotank':3A,11B 'epson':1A,9B 'grey':4A,12B 'ink':5A,13B 'ml':8A,16B	C13T07D54A				epson-115-ecotank-grey-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
1a63b6d0-8585-4e79-b637-c1b962be3dce	Epson T41F5 Singlepack UltraChrome XD2 Black 350ml	Epson T41F5 Singlepack UltraChrome XD2 Black 350ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'350ml':7A,14B 'black':6A,13B 'c13t41f540':15A 'epson':1A,8B 'singlepack':3A,10B 't41f5':2A,9B 'ultrachrom':4A,11B 'xd2':5A,12B	C13T41F540        				epson-t41f5-singlepack-ultrachrome-xd2-black-350ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
f6330d6c-5d64-4b41-851b-5848b9418752	Epson T41F3 Singlepack UltraChrome XD2 Magenta 350ml	Epson T41F3 Singlepack UltraChrome XD2 Magenta 350ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'350ml':7A,14B 'c13t41f340':15A 'epson':1A,8B 'magenta':6A,13B 'singlepack':3A,10B 't41f3':2A,9B 'ultrachrom':4A,11B 'xd2':5A,12B	C13T41F340    				epson-t41f3-singlepack-ultrachrome-xd2-magenta-350ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
028f05a9-bc76-4673-a258-f9661a441cb3	Epson T41F4 Singlepack UltraChrome XD2 Yellow 350ml	Epson T41F4 Singlepack UltraChrome XD2 Yellow 350ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'350ml':7A,14B 'c13t41f440':15A 'epson':1A,8B 'singlepack':3A,10B 't41f4':2A,9B 'ultrachrom':4A,11B 'xd2':5A,12B 'yellow':6A,13B	C13T41F440     				epson-t41f4-singlepack-ultrachrome-xd2-yellow-350ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
cbaeb24e-d293-4221-b6ab-0f316e7af97f	Epson WF-C5890/C5390 Series Maintenance Box 	Epson WF-C5890/C5390 Series Maintenance Box 	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'/c5390':5A,13B 'box':8A,16B 'c12c938211':17A 'c5890':4A,12B 'epson':1A,9B 'mainten':7A,15B 'seri':6A,14B 'wf':3A,11B 'wf-c5890':2A,10B	C12C938211				epson-wf-c5890-c5390-series-maintenance-box-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
6b418d64-33f2-4944-9ef3-319b207679b9	Logitech USB Mouse - M90	Logitech USB Mouse - M90	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-001793':10A '910':9A 'logitech':1A,5B 'm90':4A,8B 'mous':3A,7B 'usb':2A,6B	910-001793				logitech-usb-mouse-m90	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7461b897-2baa-4a62-8bf3-da5e766e1b69	Logitech Wireless Mouse M170	Logitech Wireless Mouse M170	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-004642':10A '910':9A 'logitech':1A,5B 'm170':4A,8B 'mous':3A,7B 'wireless':2A,6B	910-004642				logitech-wireless-mouse-m170	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
0aaff266-a017-433a-a5b9-1db40c90d5be	Logitech Full size Wireless Mouse M190 Charcoal 	Logitech Full size Wireless Mouse M190 Charcoal 	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-005905':16A '910':15A 'charcoal':7A,14B 'full':2A,9B 'logitech':1A,8B 'm190':6A,13B 'mous':5A,12B 'size':3A,10B 'wireless':4A,11B	910-005905				logitech-full-size-wireless-mouse-m190-charcoal-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
ba002991-1381-4bf4-99e4-933c8f208d6d	Logitech Silent Wireless Mouse M220 Charcoal	Logitech Silent Wireless Mouse M220 Charcoal	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-004878':14A '910':13A 'charcoal':6A,12B 'logitech':1A,7B 'm220':5A,11B 'mous':4A,10B 'silent':2A,8B 'wireless':3A,9B	910-004878				logitech-silent-wireless-mouse-m220-charcoal	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
929b02b3-0ee2-4b4b-b2f9-667c3e0476c5	Logitech Wireless Mini Mouse M187 Black	Logitech Wireless Mini Mouse M187 Black	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-002731':14A '910':13A 'black':6A,12B 'logitech':1A,7B 'm187':5A,11B 'mini':3A,9B 'mous':4A,10B 'wireless':2A,8B	910-002731				logitech-wireless-mini-mouse-m187-black	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
94a4586c-b95d-4b6e-86ff-3fbce5c5ddaf	Logitech Pebble M350 Wireless & Bluetooth Mouse - Graphite	Logitech Pebble M350 Wireless & Bluetooth Mouse - Graphite	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-005718':16A '910':15A 'bluetooth':5A,12B 'graphit':7A,14B 'logitech':1A,8B 'm350':3A,10B 'mous':6A,13B 'pebbl':2A,9B 'wireless':4A,11B	910-005718				logitech-pebble-m350-wireless-bluetooth-mouse-graphite	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
faa1d32a-c22a-43b0-aa96-4f853346eb93	Eposn 115 EcoTank Cyan ink bottle 70.0 ml	Eposn 115 EcoTank Cyan ink bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'115':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t07d24a':17A 'cyan':4A,12B 'ecotank':3A,11B 'eposn':1A,9B 'ink':5A,13B 'ml':8A,16B	C13T07D24A				eposn-115-ecotank-cyan-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
6c8da4d1-6967-4c62-806e-eb7452493958	Logitech M350S Pebble 2 Bluetooth Mouse - Tonal Graphite - Dongleless	Logitech M350S Pebble 2 Bluetooth Mouse - Tonal Graphite - Dongleless	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-007015':20A '2':4A,13B '910':19A 'bluetooth':5A,14B 'dongleless':9A,18B 'graphit':8A,17B 'logitech':1A,10B 'm350s':2A,11B 'mous':6A,15B 'pebbl':3A,12B 'tonal':7A,16B	910-007015				logitech-m350s-pebble-2-bluetooth-mouse-tonal-graphite-dongleless	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7df3d335-57c7-4a88-b2d3-d189af29800e	Logitech USB  Keyboard K120	Logitech USB  Keyboard K120	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-002508':10A '920':9A 'k120':4A,8B 'keyboard':3A,7B 'logitech':1A,5B 'usb':2A,6B	920-002508				logitech-usb-keyboard-k120	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
946441d4-c8ac-47ba-99bd-6c84a2997b17	Logitech Wireless Keyboard & Mouse Combo MK220	Logitech Wireless Keyboard & Mouse Combo MK220	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-003160':14A '920':13A 'combo':5A,11B 'keyboard':3A,9B 'logitech':1A,7B 'mk220':6A,12B 'mous':4A,10B 'wireless':2A,8B	920-003160				logitech-wireless-keyboard-mouse-combo-mk220	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7ef26f8d-3251-40a9-86d9-f808992c04fc	Logitech Wireless Keyboard & Mouse Combo MK235	Logitech Wireless Keyboard & Mouse Combo MK235	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-007931':14A '920':13A 'combo':5A,11B 'keyboard':3A,9B 'logitech':1A,7B 'mk235':6A,12B 'mous':4A,10B 'wireless':2A,8B	920-007931				logitech-wireless-keyboard-mouse-combo-mk235	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
64d0e9c2-b56f-4d61-a286-fda0ee68f543	Logitech Wireless Keyboard & Mouse Combo MK270	Logitech Wireless Keyboard & Mouse Combo MK270	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-004509':14A '920':13A 'combo':5A,11B 'keyboard':3A,9B 'logitech':1A,7B 'mk270':6A,12B 'mous':4A,10B 'wireless':2A,8B	920-004509				logitech-wireless-keyboard-mouse-combo-mk270	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
eaebe973-a371-4d1f-8336-09520b6a3a22	Logitech Wireless Keyboard with TouchPad K400 Plus – Black	Logitech Wireless Keyboard with TouchPad K400 Plus – Black	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-007145':18A '920':17A 'black':8A,16B 'k400':6A,14B 'keyboard':3A,11B 'logitech':1A,9B 'plus':7A,15B 'touchpad':5A,13B 'wireless':2A,10B	920-007145				logitech-wireless-keyboard-with-touchpad-k400-plus-black	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
548169fb-0b50-4f9b-9696-ca3edae7b19b	Logitech Stereo Headset H111 - Black (3.5 mm Jack )	Logitech Stereo Headset H111 - Black (3.5 mm Jack )	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-000593':18A '3.5':6A,14B '981':17A 'black':5A,13B 'h111':4A,12B 'headset':3A,11B 'jack':8A,16B 'logitech':1A,9B 'mm':7A,15B 'stereo':2A,10B	981-000593				logitech-stereo-headset-h111-black-3-5-mm-jack-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
c62c76ae-50b3-4592-985d-aa5db39c3d14	Logitech USB Headset H340	Logitech USB Headset H340	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-000475':10A '981':9A 'h340':4A,8B 'headset':3A,7B 'logitech':1A,5B 'usb':2A,6B	981-000475				logitech-usb-headset-h340	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
f96f1b43-0806-4677-b374-5f64b4632960	Logitech C270 HD Webcam	Logitech C270 HD Webcam	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-000584':10A '960':9A 'c270':2A,6B 'hd':3A,7B 'logitech':1A,5B 'webcam':4A,8B	960-000584				logitech-c270-hd-webcam	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7a024e55-fb8e-48b2-a815-0f0c273929b1	Logitech C920 HD Pro Webcam	Logitech C920 HD Pro Webcam	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-001055':12A '960':11A 'c920':2A,7B 'hd':3A,8B 'logitech':1A,6B 'pro':4A,9B 'webcam':5A,10B	960-001055				logitech-c920-hd-pro-webcam	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
4269a523-6252-49bf-85fd-d6a039fc51b8	Logitech C922 Pro Stream Webcam  with Tripod Stand	Logitech C922 Pro Stream Webcam  with Tripod Stand	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-001088':18A '960':17A 'c922':2A,10B 'logitech':1A,9B 'pro':3A,11B 'stand':8A,16B 'stream':4A,12B 'tripod':7A,15B 'webcam':5A,13B	960-001088				logitech-c922-pro-stream-webcam-with-tripod-stand	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
e4f1a90a-d791-4baa-b293-cd7e72cedb61	Logitech C925e Business HD Webcam	Logitech C925e Business HD Webcam	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-001076':12A '960':11A 'busi':3A,8B 'c925e':2A,7B 'hd':4A,9B 'logitech':1A,6B 'webcam':5A,10B	960-001076				logitech-c925e-business-hd-webcam	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
97980ab4-9e83-44eb-af6c-70a12c771e5f	Logitech MeetUp	Logitech MeetUp ( All-in-one conferencecam with an ultra-wide lens for small rooms )	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-001102':20A '960':19A 'all-in-on':5B 'conferencecam':9B 'len':15B 'logitech':1A,3B 'meetup':2A,4B 'one':8B 'room':18B 'small':17B 'ultra':13B 'ultra-wid':12B 'wide':14B	960-001102				logitech-meetup	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
fecc3c33-5462-4281-9940-651f6cea76d3	Epson EcoTank L5590	Epson EcoTank L5590 A4 Wi-Fi All-in-One Ink Tank Printer With ADF	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'a4':7B 'adf':19B 'all-in-on':11B 'c11ck57405':20A 'ecotank':2A,5B 'epson':1A,4B 'fi':10B 'ink':15B 'l5590':3A,6B 'one':14B 'printer':17B 'tank':16B 'wi':9B 'wi-fi':8B	C11CK57405				epson-ecotank-l5590	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
dbbc1df0-a94d-4b26-8f12-b1244e55f164	Logitech Conference Cam  Group - USB	Logitech Conference Cam  Group - USB	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-001057':12A '960':11A 'cam':3A,8B 'confer':2A,7B 'group':4A,9B 'logitech':1A,6B 'usb':5A,10B	960-001057				logitech-conference-cam-group-usb	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
edd3b746-5b06-40cb-99d0-ed67a8f50715	Epson EcoTank L6550 	Epson EcoTank L6550 Wi-Fi Duplex AlO Ink Tank A4 Printer	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'a4':14B 'alo':11B 'c11cj30403da':16A 'duplex':10B 'ecotank':2A,5B 'epson':1A,4B 'fi':9B 'ink':12B 'l6550':3A,6B 'printer':15B 'tank':13B 'wi':8B 'wi-fi':7B	C11CJ30403DA				epson-ecotank-l6550-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
1d2a85e6-daf7-471f-9208-6629682f9a9d	Epson EcoTank L6570 	Epson EcoTank L6570 Wi-Fi Duplex AlO Ink Tank A4 Printer	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'a4':14B 'alo':11B 'c11cj29403da':16A 'duplex':10B 'ecotank':2A,5B 'epson':1A,4B 'fi':9B 'ink':12B 'l6570':3A,6B 'printer':15B 'tank':13B 'wi':8B 'wi-fi':7B	C11CJ29403DA				epson-ecotank-l6570-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
3093b263-0991-412d-bf3c-634b747330b2	Epson EcoTank L14150 	Epson EcoTank L14150 4 in 1 MFP A3 Printer	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1':9B '4':7B 'a3':11B 'c11ch96403da':13A 'ecotank':2A,5B 'epson':1A,4B 'l14150':3A,6B 'mfp':10B 'printer':12B	C11CH96403DA				epson-ecotank-l14150-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7a6d1b71-1009-4f0c-9e8f-7d4fc1aa27d1	Epson EcoTank L15150 	Epson EcoTank L15150 MEAF All-in-one  A3 Printer	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'a3':12B 'all-in-on':8B 'c11ch72403da':14A 'ecotank':2A,5B 'epson':1A,4B 'l15150':3A,6B 'meaf':7B 'one':11B 'printer':13B	C11CH72403DA				epson-ecotank-l15150-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
ef09a5d8-0afd-4e91-8346-b3a5da96274c	Epson EcoTank Monochrome A4 M3170	Epson EcoTank Monochrome A4 M3170, All-in-One Wi-Fi Duplex InkTank Printer Scan, Copy, Fax, with ADF , Hi-Speed USB - compatible with USB 2.0 specification, Ethernet interface (100 Base-TX / 10 Base-T), Wi-Fi Direct,  12 months Carry in, 100.000 pages	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'10':41B '100':37B '100.000':53B '12':49B '2.0':33B 'a4':4A,9B 'adf':25B 'all-in-on':11B 'base':39B,43B 'base-t':42B 'base-tx':38B 'c11cg92404by':55A 'carri':51B 'compat':30B 'copi':22B 'direct':48B 'duplex':18B 'ecotank':2A,7B 'epson':1A,6B 'ethernet':35B 'fax':23B 'fi':17B,47B 'hi':27B 'hi-spe':26B 'inktank':19B 'interfac':36B 'm3170':5A,10B 'monochrom':3A,8B 'month':50B 'one':14B 'page':54B 'printer':20B 'scan':21B 'specif':34B 'speed':28B 'tx':40B 'usb':29B,32B 'wi':16B,46B 'wi-fi':15B,45B	C11CG92404BY				epson-ecotank-monochrome-a4-m3170	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
d5f3f4cb-0982-4d87-a24c-82ca4ad73989	EPSON Perfection V39II Photo and Document Scanner 	EPSON Perfection V39II Photo and Document Scanner 	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'b11b268401':15A 'document':6A,13B 'epson':1A,8B 'perfect':2A,9B 'photo':4A,11B 'scanner':7A,14B 'v39ii':3A,10B	B11B268401				epson-perfection-v39ii-photo-and-document-scanner-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
7d2a4265-654a-4a9e-b4ef-1200cb9671d3	Epson TM-T20X	Epson TM-T20X (052): Ethernet + USB + Serial, PS, Blk, UK – 200mm/s, RS-232, USB 2.0 Type B, Drawer kick-out. Reliability: 60,000,000 MCBF (Lines), 360,000 MTBF (Hours). Print Head Life: 100 km – 100,000,000 pulses   	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-232':18B '000':29B,30B,34B,43B,44B '052':9B '100':40B,42B '2.0':20B '200mm/s':16B '360':33B '60':28B 'b':22B 'blk':14B 'c31ch26052a0':46A 'drawer':23B 'epson':1A,5B 'ethernet':10B 'head':38B 'hour':36B 'kick':25B 'kick-out':24B 'km':41B 'life':39B 'line':32B 'mcbf':31B 'mtbf':35B 'print':37B 'ps':13B 'puls':45B 'reliabl':27B 'rs':17B 'serial':12B 't20x':4A,8B 'tm':3A,7B 'tm-t20x':2A,6B 'type':21B 'uk':15B 'usb':11B,19B	C31CH26052A0				epson-tm-t20x	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
a1401070-cd3a-4caa-9e1d-02ecd64ce290	HP Pro 290 G9 Intel Core I7	HP Pro Tower 290 G9 Intel Core I7-12700, 8GB RAM, 1TB HDD, DVDRW, DOS ,  P27 G5 27 Inch MONITOR, 1 Year Warranty 	0	51223939-57c9-4b21-a222-1dfd6451c251	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-12700':16B '1':28B '1tb':19B '27':25B '290':3A,11B '6d419ea':31A '8gb':17B 'core':6A,14B 'dos':22B 'dvdrw':21B 'g5':24B 'g9':4A,12B 'hdd':20B 'hp':1A,8B 'i7':7A,15B 'inch':26B 'intel':5A,13B 'monitor':27B 'p27':23B 'pro':2A,9B 'ram':18B 'tower':10B 'warranti':30B 'year':29B	6D419EA				hp-pro-290-g9-intel-core-i7	active	453.0000000000	453.0000000000	2024-08-21 07:05:01.997614+00	\N
1009b950-e3c4-49d7-9417-c8507deb5238	Epson CO-W01 Projector	Epson CO-W01 Projector 3LCD Technology, WXGA, 1280 x 800, 16:10, 3000 Lumen - 2000 Lumen (economy), 16,000 : 1, HDMI 1.4, USB 2.0-A, USB 2.0 Type B (Service Only), 2.4 kg, 5W, Main unit, Power cable, Quick Start Guide, Remote control incl. batteries, User manual (CD), 	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'000':25B '1':26B '1.4':28B '10':18B '1280':14B '16':17B,24B '2.0':30B,33B '2.4':38B '2000':21B '3000':19B '3lcd':11B '5w':40B '800':16B 'b':35B 'batteri':51B 'cabl':44B 'cd':54B 'co':3A,8B 'co-w01':2A,7B 'control':49B 'economi':23B 'epson':1A,6B 'guid':47B 'hdmi':27B 'incl':50B 'kg':39B 'lumen':20B,22B 'main':41B 'manual':53B 'power':43B 'projector':5A,10B 'quick':45B 'remot':48B 'servic':36B 'start':46B 'technolog':12B 'type':34B 'unit':42B 'usb':29B,32B 'user':52B 'v11ha86040':55A 'w01':4A,9B 'wxga':13B 'x':15B	V11HA86040				epson-co-w01-projector	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
aeb486e9-53c9-47bb-b08c-a8c856ff790d	Epson EcoTank L6490 	Epson EcoTank L6490 A4 Printer, Print, Copy, Scan and Fax, Duplex Printing – ADF, Wi-Fi, Wi-Fi Direct, Ethernet, USB Interface with LCD Touchscreen	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'a4':7B 'adf':16B 'c11cj88404da':30A 'copi':10B 'direct':23B 'duplex':14B 'ecotank':2A,5B 'epson':1A,4B 'ethernet':24B 'fax':13B 'fi':19B,22B 'interfac':26B 'l6490':3A,6B 'lcd':28B 'print':9B,15B 'printer':8B 'scan':11B 'touchscreen':29B 'usb':25B 'wi':18B,21B 'wi-fi':17B,20B	C11CJ88404DA				epson-ecotank-l6490-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
6dc8af45-3918-49d3-95f3-e26c39b02038	32GB Lexar	32GB Lexar® JumpDrive® M22 USB 2.0 Light Gold Flash Drive	0	adaa23e0-3781-4756-82c9-41ebbe25a846	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2.0':8B '32gb':1A,3B 'bnjng':15A 'drive':12B 'flash':11B 'gold':10B 'jumpdriv':5B 'lexar':2A,4B 'light':9B 'ljdm022032g':14A 'ljdm022032g-bnjng':13A 'm22':6B 'usb':7B	LJDM022032G-BNJNG				32gb-lexar	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
07f1e4fa-da53-44a8-84ce-bb095949f89a	64GB Lexar	64GB Lexar® JumpDrive® M22 USB 2.0 Light Gold Flash Drive	0	adaa23e0-3781-4756-82c9-41ebbe25a846	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2.0':8B '64gb':1A,3B 'bnjng':15A 'drive':12B 'flash':11B 'gold':10B 'jumpdriv':5B 'lexar':2A,4B 'light':9B 'ljdm022064g':14A 'ljdm022064g-bnjng':13A 'm22':6B 'usb':7B	LJDM022064G-BNJNG				64gb-lexar	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
a0eaf32c-0708-454e-80ac-5388997047d7	HP PROBOOK 450 G10	HP PROBOOK 450 G10 LAPTOP (CI5-1335U/8GB/512GB/15.6 FHD/DOS) - PIKE SILVER ALUMINUM 	0	8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'/8gb/512gb/15.6':13B '1335u':12B '450':3A,7B '816n8ea':18A 'aluminum':17B 'ci5':11B 'ci5-1335u':10B 'fhd/dos':14B 'g10':4A,8B 'hp':1A,5B 'laptop':9B 'pike':15B 'probook':2A,6B 'silver':16B	816N8EA				hp-probook-450-g10	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	\N
a1cba806-8536-4acd-bf89-f13fe0ae4c56	HP 126A LaserJet Imaging Drum	HP 126A LaserJet Imaging Drum	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'126a':2A,7B 'ce314a':11A 'drum':5A,10B 'hp':1A,6B 'imag':4A,9B 'laserjet':3A,8B	CE314A				hp-126a-laserjet-imaging-drum	active	36.0000000000	36.0000000000	2024-08-21 07:05:01.997614+00	\N
ec5edbdf-282a-44c9-a069-555888ed434f	Epson ERC 38B Black Ribbon Cartridge 	Epson ERC 38B Black Ribbon Cartridge 	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'38b':3A,9B 'black':4A,10B 'c13s015374':13A 'cartridg':6A,12B 'epson':1A,7B 'erc':2A,8B 'ribbon':5A,11B	C13S015374				epson-erc-38b-black-ribbon-cartridge-	active	45.0000000000	45.0000000000	2024-08-21 07:05:01.997614+00	\N
324d3a51-1cb1-4c1a-868d-c5530c21436d	Logitech Expansion Mic For Meetup 	Logitech Expansion Mic For Meetup 	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-000405':12A '989':11A 'expans':2A,7B 'logitech':1A,6B 'meetup':5A,10B 'mic':3A,8B	989-000405				logitech-expansion-mic-for-meetup-	active	76.0000000000	76.0000000000	2024-08-21 07:05:01.997614+00	\N
c57f19c4-bd78-4662-9839-dc419f36f83b	HP 87A Black LaserJet Toner Cartridge	HP 87A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'87a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf287a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF287A				hp-87a-black-laserjet-toner-cartridge	active	78.0000000000	78.0000000000	2024-08-21 07:05:01.997614+00	\N
4896b1e2-0f37-4e97-9847-fcf4d1ef4e1d	HP 14A Black LaserJet Toner Cartridge	HP 14A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'14a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf214a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF214A				hp-14a-black-laserjet-toner-cartridge	active	733.0000000000	733.0000000000	2024-08-21 07:05:01.997614+00	\N
e3519734-ea27-419f-928d-81d90d8399d9	HP ProBook 440 G10 Laptop	The HP ProBook 440 G10 is a versatile business laptop designed for professionals and students who require robust performance in a sleek and portable package. This laptop is equipped with the latest hardware and features, making it ideal for multitasking, productivity, and entertainment.	0	8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'440':3A,9B '816n0ea':49A 'busi':14B 'design':16B 'entertain':48B 'equip':34B 'featur':40B 'g10':4A,10B 'hardwar':38B 'hp':1A,7B 'ideal':43B 'laptop':5A,15B,32B 'latest':37B 'make':41B 'multitask':45B 'packag':30B 'perform':24B 'portabl':29B 'probook':2A,8B 'product':46B 'profession':18B 'requir':22B 'robust':23B 'sleek':27B 'student':20B 'versatil':13B	816N0EA				hp-probook-440-g10-laptop	active	45.0000000000	45.0000000000	2024-08-21 07:05:01.997614+00	\N
ab2af2c0-da9c-4b7e-9534-1807e59a5d04	Lexar 1TB NS100 2.5” SATA Internal Hard Disk 	Lexar 1TB NS100 2.5” SATA (6Gb/s) Solid-State Drive, up to 550MB/s Read	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1tb':2A,10B '1tbrb':25A '2.5':4A,12B '550mb/s':21B '6gb/s':14B 'disk':8A 'drive':18B 'hard':7A 'intern':6A 'lexar':1A,9B 'lns100':24A 'lns100-1tbrb':23A 'ns100':3A,11B 'read':22B 'sata':5A,13B 'solid':16B 'solid-st':15B 'state':17B	LNS100-1TBRB				lexar-1tb-ns100-2-5-sata-internal-hard-disk-	active	34.0000000000	34.0000000000	2024-08-21 07:05:01.997614+00	\N
9f8caa44-07f5-44c7-928d-64be8595f8b1	16GB Lexar	16GB Lexar® JumpDrive® M22 USB 2.0 Light Gold Flash Drive	0	adaa23e0-3781-4756-82c9-41ebbe25a846	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'16gb':1A,3B '2.0':8B 'bnjng':15A 'drive':12B 'flash':11B 'gold':10B 'jumpdriv':5B 'lexar':2A,4B 'light':9B 'ljdm022016g':14A 'ljdm022016g-bnjng':13A 'm22':6B 'usb':7B	LJDM022016G-BNJNG				16gb-lexar	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
7425ba85-fff4-4c72-a2ca-583e8462b5a5	Dell OptiPlex 3000 Tower	Intel Core i5 12500, 4GB DDR4 3200, 256GB PCIe NVMe M.2 SSD, Ubuntu, DVD±RW, Wired Keyboard and Mouse, Black, 1 Year Warranty, Front Ports: One Universal Audio Jack, Two USB 2.0 Ports, Two USB 3.2 Gen 1 Ports, Rear Ports: Two USB 3.2 Gen 1 Ports, Two USB 2.0 Ports with Smart Power On, One RJ45 Ethernet Port, One DisplayPort 1.4 Port, One HDMI 1.4b Port, Kensington Security Slot - N015O3000MTAC_EM	0	d4b15a4c-8796-419b-9aba-92e19cd3f2bf	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1':25B,42B,50B '1.4':66B,70B '12500':8B '2.0':36B,54B '256gb':12B '3.2':40B,48B '3000':3A '3200':11B '4gb':9B '7374slv':80A 'audio':32B 'b':71B 'black':24B 'core':6B 'ddr4':10B 'dell':1A 'displayport':65B 'dvd':18B 'em':77B 'ethernet':62B 'front':28B 'gen':41B,49B 'hdmi':69B 'i5':7B 'i7430':79A 'i7430-7374slv-pus':78A 'intel':5B 'jack':33B 'kensington':73B 'keyboard':21B 'm.2':15B 'mous':23B 'n015o3000mtac':76B 'nvme':14B 'one':30B,60B,64B,68B 'optiplex':2A 'pcie':13B 'port':29B,37B,43B,45B,51B,55B,63B,67B,72B 'power':58B 'pus':81A 'rear':44B 'rj45':61B 'rw':19B 'secur':74B 'slot':75B 'smart':57B 'ssd':16B 'tower':4A 'two':34B,38B,46B,52B 'ubuntu':17B 'univers':31B 'usb':35B,39B,47B,53B 'warranti':27B 'wire':20B 'year':26B	I7430-7374SLV-PUS				dell-optiplex-3000-tower	active	10.0000000000	10.0000000000	2024-08-21 07:05:01.997614+00	\N
603efb93-8bf5-451f-866d-fa362f8f7c3f	HP 90A Black LaserJet Toner Cartridge	HP 90A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'90a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce390a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE390A				hp-90a-black-laserjet-toner-cartridge	active	346.0000000000	346.0000000000	2024-08-21 07:05:01.997614+00	\N
cb408c48-5a1c-4aad-9898-7681fe59a6aa	Lexar 2TB External Portable SSD Hard Disk 	Lexar 2TB External Portable SSD Hard Disk , up to 550MB/s Read and 400MB/s Write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2tb':2A,9B '400mb/s':20B '550mb/s':17B 'disk':7A,14B 'extern':3A,10B 'hard':6A,13B 'lexar':1A,8B 'lsl200x002t':23A 'lsl200x002t-rnnng':22A 'portabl':4A,11B 'read':18B 'rnnng':24A 'ssd':5A,12B 'write':21B	LSL200X002T-RNNNG				lexar-2tb-external-portable-ssd-hard-disk-	active	43.0000000000	43.0000000000	2024-08-21 07:05:01.997614+00	\N
46525d71-2234-4b0b-8157-2f1087cef391	Lexar 512GB M.2 NVMe	Lexar 512GB High Speed PCIe Gen3 with 4 Lanes M.2 NVMe, up to 3300 MB/s read and 2400 MB/s write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2400':22B '3300':18B '4':12B '512gb':2A,6B 'gen3':10B 'high':7B 'lane':13B 'lexar':1A,5B 'lnm620x512g':26A 'lnm620x512g-rnnng':25A 'm.2':3A,14B 'mb/s':19B,23B 'nvme':4A,15B 'pcie':9B 'read':20B 'rnnng':27A 'speed':8B 'write':24B	LNM620X512G-RNNNG				lexar-512gb-m-2-nvme	active	3.0000000000	3.0000000000	2024-08-21 07:05:01.997614+00	\N
bfc5b9fc-d805-4dee-978c-f97ae3cf9a8e	HP 507A Black LaserJet Toner Cartridge	HP 507A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'507a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce400a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE400A				hp-507a-black-laserjet-toner-cartridge	active	68.0000000000	68.0000000000	2024-08-21 07:05:01.997614+00	\N
871939cc-75aa-491f-a3ac-90ac71ed7b40	Lexar 2TB NS100 2.5” SATA Internal Hard Disk 	Lexar 2TB NS100 2.5” SATA (6Gb/s) Solid-State Drive, up to 550MB/s Read	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2.5':4A,12B '2tb':2A,10B '2tbrb':25A '550mb/s':21B '6gb/s':14B 'disk':8A 'drive':18B 'hard':7A 'intern':6A 'lexar':1A,9B 'lns100':24A 'lns100-2tbrb':23A 'ns100':3A,11B 'read':22B 'sata':5A,13B 'solid':16B 'solid-st':15B 'state':17B	LNS100-2TBRB				lexar-2tb-ns100-2-5-sata-internal-hard-disk-	active	23.0000000000	23.0000000000	2024-08-21 07:05:01.997614+00	\N
a25788ef-c538-402c-a78e-9878ef4c09e0	Lexar 256GB NS100 2.5” SATA Internal Hard Disk	Lexar 256GB NS100 2.5” SATA (6Gb/s) Solid-State Drive, up to 520MB/s Read	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2.5':4A,12B '256gb':2A,10B '256rb':25A '520mb/s':21B '6gb/s':14B 'disk':8A 'drive':18B 'hard':7A 'intern':6A 'lexar':1A,9B 'lns100':24A 'lns100-256rb':23A 'ns100':3A,11B 'read':22B 'sata':5A,13B 'solid':16B 'solid-st':15B 'state':17B	LNS100-256RB				lexar-256gb-ns100-2-5-sata-internal-hard-disk	active	54.0000000000	54.0000000000	2024-08-21 07:05:01.997614+00	\N
84df0f4e-3b7d-4409-a88a-782e00792b0f	HP 05A Black Laserjet Toner Cartridge	HP 05A Black Laserjet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'05a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce505a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE505A				hp-05a-black-laserjet-toner-cartridge	active	74.0000000000	74.0000000000	2024-08-21 07:05:01.997614+00	\N
3b74114e-33ee-4bac-b182-f1256048009f	HP 305A Cyan LaserJet Toner Cartridge	HP 305A Cyan LaserJet Toner Cartridge	0	00c07cec-21aa-4e81-9b27-d5b972573820	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'305a':2A,8B 'cartridg':6A,12B 'ce411a':13A 'cyan':3A,9B 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE411A				hp-305a-cyan-laserjet-toner-cartridge	active	78.0000000000	78.0000000000	2024-08-21 07:05:01.997614+00	\N
54cc4539-8511-49a9-8301-c1aab84f9933	HP 305A Magenta LaserJet Toner Cartridge	HP 305A Magenta LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'305a':2A,8B 'cartridg':6A,12B 'ce413a':13A 'hp':1A,7B 'laserjet':4A,10B 'magenta':3A,9B 'toner':5A,11B	CE413A				hp-305a-magenta-laserjet-toner-cartridge	active	5.0000000000	5.0000000000	2024-08-21 07:05:01.997614+00	\N
4df5efa4-566d-47a8-bb95-99fdf40aa350	HP 131A Magenta LaserJet Toner Cartridge	HP 131A Magenta LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'131a':2A,8B 'cartridg':6A,12B 'cf213a':13A 'hp':1A,7B 'laserjet':4A,10B 'magenta':3A,9B 'toner':5A,11B	CF213A				hp-131a-magenta-laserjet-toner-cartridge	active	37.0000000000	37.0000000000	2024-08-21 07:05:01.997614+00	\N
cfb01089-6770-480a-a255-57f293eb1bde	HP M27fw 27-inch IPS FHD	HP M27fw 27-inch IPS FHD (1920 x 1080) LED Backlit Monitor 1 VGA, 2 HDMI 1.4	0	263fe9a0-4b39-4335-b8ce-79fa460b5759	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1':19B '1.4':23B '1080':15B '1920':13B '2':21B '27':3A,9B '2h1a4as':24A 'backlit':17B 'fhd':6A,12B 'hdmi':22B 'hp':1A,7B 'inch':4A,10B 'ip':5A,11B 'led':16B 'm27fw':2A,8B 'monitor':18B 'vga':20B 'x':14B	2H1A4AS				hp-m27fw-27-inch-ips-fhd	active	68.0000000000	68.0000000000	2024-08-21 07:05:01.997614+00	\N
930bdd12-9ba0-475c-aec9-756207770184	HP 305A Yellow LaserJet Toner Cartridge	HP 305A Yellow LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'305a':2A,8B 'cartridg':6A,12B 'ce412a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B 'yellow':3A,9B	CE412A				hp-305a-yellow-laserjet-toner-cartridge	active	34.0000000000	34.0000000000	2024-08-21 07:05:01.997614+00	\N
c865006d-68c4-47bc-b9e5-2ef584d93f31	HP 26A Black LaserJet Toner Cartridge	HP 26A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'26a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf226a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF226A				hp-26a-black-laserjet-toner-cartridge	active	797.0000000000	797.0000000000	2024-08-21 07:05:01.997614+00	\N
82f3b358-410f-401c-8e82-5fcd51c4ca48	Lexar 512GB NS100 2.5” SATA  Internal Hard Disk	Lexar 512GB NS100 2.5” SATA (6Gb/s) Solid-State Drive, up to 550MB/s Read	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2.5':4A,12B '512gb':2A,10B '512rb':25A '550mb/s':21B '6gb/s':14B 'disk':8A 'drive':18B 'hard':7A 'intern':6A 'lexar':1A,9B 'lns100':24A 'lns100-512rb':23A 'ns100':3A,11B 'read':22B 'sata':5A,13B 'solid':16B 'solid-st':15B 'state':17B	LNS100-512RB				lexar-512gb-ns100-2-5-sata-internal-hard-disk	active	324.0000000000	324.0000000000	2024-08-21 07:05:01.997614+00	\N
d994e04b-1915-4173-b4dd-c78869cb3d0a	HP M22F 21.5-inch FHD Monitor	HP M22F 21.5-inch FHD Monitor, 1 HDMI 1.4 (with HDCP support); 1 VGA	0	263fe9a0-4b39-4335-b8ce-79fa460b5759	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1':13B,19B '1.4':15B '21.5':3A,9B '2e2':21A 'fhd':5A,11B 'hdcp':17B 'hdmi':14B 'hp':1A,7B 'inch':4A,10B 'm22f':2A,8B 'monitor':6A,12B 'support':18B 'vga':20B 'y3aa':22A	2E2Y3AA				hp-m22f-21-5-inch-fhd-monitor	active	434.0000000000	434.0000000000	2024-08-21 07:05:01.997614+00	\N
257a68ac-ccdf-4588-95b7-b99e7e8fd46f	HP 83A Black LaserJet Toner Cartridge	HP 83A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'83a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf283a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF283A				hp-83a-black-laserjet-toner-cartridge	active	7.0000000000	7.0000000000	2024-08-21 07:05:01.997614+00	\N
cfdacdb2-a03d-4539-ab93-9f61e6d6b148	Lexar 512GB External Portable SSD Hard Disk	Lexar 512GB External Portable SSD, up to 550MB/s Read and 400MB/s Write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'400mb/s':18B '512gb':2A,9B '550mb/s':15B 'disk':7A 'extern':3A,10B 'hard':6A 'lexar':1A,8B 'lsl200x512g':21A 'lsl200x512g-rnnng':20A 'portabl':4A,11B 'read':16B 'rnnng':22A 'ssd':5A,12B 'write':19B	LSL200X512G-RNNNG				lexar-512gb-external-portable-ssd-hard-disk	active	86.0000000000	86.0000000000	2024-08-21 07:05:01.997614+00	\N
09a32f41-86be-4e2d-b903-718ee31f8b33	HP 130A Yellow LaserJet Toner Cartridge	HP 130A Yellow LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'130a':2A,8B 'cartridg':6A,12B 'cf352a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B 'yellow':3A,9B	CF352A				hp-130a-yellow-laserjet-toner-cartridge	active	98.0000000000	98.0000000000	2024-08-21 07:05:01.997614+00	\N
91c4c7b6-4173-4a32-832a-4ac3bba9f0af	HP 131A Yellow LaserJet Toner Cartridge	HP 131A Yellow LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'131a':2A,8B 'cartridg':6A,12B 'cf212a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B 'yellow':3A,9B	CF212A				hp-131a-yellow-laserjet-toner-cartridge	active	37.0000000000	37.0000000000	2024-08-21 07:05:01.997614+00	\N
e9df4302-1ee0-4b7c-bc37-5b059dc692e9	HP 85A 2-pack Black Original LaserJet Toner Cartridges	HP 85A 2-pack Black Original LaserJet Toner Cartridges	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'2':3A,12B '85a':2A,11B 'black':5A,14B 'cartridg':9A,18B 'ce285af':19A 'hp':1A,10B 'laserjet':7A,16B 'origin':6A,15B 'pack':4A,13B 'toner':8A,17B	CE285AF				hp-85a-2-pack-black-original-laserjet-toner-cartridges	active	12.0000000000	12.0000000000	2024-08-21 07:05:01.997614+00	\N
5f759187-cab4-4da8-a26b-fb4af2bd13ee	HP 17A Black LaserJet Toner Cartridge	HP 17A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'17a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf217a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF217A				hp-17a-black-laserjet-toner-cartridge	active	775.0000000000	775.0000000000	2024-08-21 07:05:01.997614+00	\N
c8f2758a-e50e-468e-ad08-3c897c3ef22a	HP 507A Yellow LaserJet Toner Cartridge	HP 507A Yellow LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'507a':2A,8B 'cartridg':6A,12B 'ce402a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B 'yellow':3A,9B	CE402A				hp-507a-yellow-laserjet-toner-cartridge	active	24.0000000000	24.0000000000	2024-08-21 07:05:01.997614+00	\N
7fbbe4c0-8c78-4f32-832d-056a220612b4	HP 19.5-inch HD Monitor	HP P204v 19.5-inch HD Monitor, 1 HDMI 1.4 (with HDCP support); 1 VGA	0	263fe9a0-4b39-4335-b8ce-79fa460b5759	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1':12B,18B '1.4':14B '19.5':2A,8B '5rd66aa':20A 'hd':4A,10B 'hdcp':16B 'hdmi':13B 'hp':1A,6B 'inch':3A,9B 'monitor':5A,11B 'p204v':7B 'support':17B 'vga':19B	5RD66AA				hp-19-5-inch-hd-monitor	active	12.0000000000	12.0000000000	2024-08-21 07:05:01.997614+00	\N
f0413c8e-ce2c-4cc3-a713-a0ccd0fee111	Lexar 1TB M.2 NVMe	Lexar 1TB High Speed PCIe Gen3 with 4 Lanes M.2 NVMe, up to 3300 MB/s read and 3000 MB/s write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1tb':2A,6B '3000':22B '3300':18B '4':12B 'gen3':10B 'high':7B 'lane':13B 'lexar':1A,5B 'lnm620x001t':26A 'lnm620x001t-rnnng':25A 'm.2':3A,14B 'mb/s':19B,23B 'nvme':4A,15B 'pcie':9B 'read':20B 'rnnng':27A 'speed':8B 'write':24B	LNM620X001T-RNNNG				lexar-1tb-m-2-nvme	active	2345.0000000000	2345.0000000000	2024-08-21 07:05:01.997614+00	\N
3ae55826-3a0b-4814-95a6-2cbfdb6b3ecd	HP 507A Cyan LaserJet Toner Cartridge	HP 507A Cyan LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'507a':2A,8B 'cartridg':6A,12B 'ce401a':13A 'cyan':3A,9B 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE401A				hp-507a-cyan-laserjet-toner-cartridge	active	86.0000000000	86.0000000000	2024-08-21 07:05:01.997614+00	\N
4069773e-108e-4647-86d4-204d72092f11	HP V27i 27-inch IPS FHD	HP V27i 27-inch IPS FHD (1920 x 1080) LED Backlit Monitor 1 VGA, 1 HDMI 1.4	0	263fe9a0-4b39-4335-b8ce-79fa460b5759	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'1':19B,21B '1.4':23B '1080':15B '1920':13B '27':3A,9B '9sv94as':24A 'backlit':17B 'fhd':6A,12B 'hdmi':22B 'hp':1A,7B 'inch':4A,10B 'ip':5A,11B 'led':16B 'monitor':18B 'v27i':2A,8B 'vga':20B 'x':14B	9SV94AS				hp-v27i-27-inch-ips-fhd	active	78.0000000000	78.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
ab773cb9-cf21-4288-adec-f1bd78050c84	HP EliteBook 840 G10 Laptop	The HP EliteBook 840 G10 is a premium business laptop designed for professionals who require a powerful and versatile machine. It offers a blend of performance, mobility, and security features, making it ideal for corporate environments and on-the-go productivity.	0	8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'81a31ea':48A '840':3A,9B 'blend':29B 'busi':14B 'corpor':40B 'design':16B 'elitebook':2A,8B 'environ':41B 'featur':35B 'g10':4A,10B 'go':46B 'hp':1A,7B 'ideal':38B 'laptop':5A,15B 'machin':25B 'make':36B 'mobil':32B 'offer':27B 'on-the-go':43B 'perform':31B 'power':22B 'premium':13B 'product':47B 'profession':18B 'requir':20B 'secur':34B 'versatil':24B	81A31EA				hp-elitebook-840-g10-laptop	active	57.0000000000	57.0000000000	2024-08-21 07:05:01.997614+00	\N
39a15032-25d6-401a-89db-1c4c4bdc2f62	HP 80A Black LaserJet Toner Cartridge	HP 80A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'80a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf280a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF280A				hp-80a-black-laserjet-toner-cartridge	active	78.0000000000	78.0000000000	2024-08-21 07:05:01.997614+00	\N
153fb270-5362-4d08-9716-75bef45d03c8	HP 128A Black LaserJet Toner Cartridge	HP 128A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'128a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce320a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE320A				hp-128a-black-laserjet-toner-cartridge	active	364.0000000000	364.0000000000	2024-08-21 07:05:01.997614+00	\N
1a0c5ae6-94d1-4bb8-abaa-5fed96fe823e	Lexar 256GB  M.2 NVME	Lexar 256GB High Speed PCIe Gen3 with 4 Lanes M.2 NVMe, up to 3300 MB/s read and 1300 MB/s write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'1300':22B '256gb':2A,6B '3300':18B '4':12B 'gen3':10B 'high':7B 'lane':13B 'lexar':1A,5B 'lnm620x256g':26A 'lnm620x256g-rnnng':25A 'm.2':3A,14B 'mb/s':19B,23B 'nvme':4A,15B 'pcie':9B 'read':20B 'rnnng':27A 'speed':8B 'write':24B	LNM620X256G-RNNNG				lexar-256gb-m-2-nvme	active	423.0000000000	423.0000000000	2024-08-21 07:05:01.997614+00	\N
6c156ae1-e546-43ef-8866-dd729257cf96	Epson ERC 32B Black Ribbon Cartridge 	Epson ERC 32B Black Ribbon Cartridge 	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'32b':3A,9B 'black':4A,10B 'c43s015371':13A 'cartridg':6A,12B 'epson':1A,7B 'erc':2A,8B 'ribbon':5A,11B	C43S015371				epson-erc-32b-black-ribbon-cartridge-	active	5445.0000000000	5445.0000000000	2024-08-21 07:05:01.997614+00	\N
6258e14f-8fe4-4755-a7e4-931b212a6fbf	HP 290 G9 Intel Core i7	HP 290 G9 Tower Intel Core i7-12700 / 8GB / 1TB HDD / DOS / DVD-WR ODD  / 125 BLKkbd / 125mouse / P22v G5 Monitor, 1 Year Warraty	0	51223939-57c9-4b21-a222-1dfd6451c251	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'-12700':14B '1':29B '125':23B '125mouse':25B '1tb':16B '290':2A,8B '5w838es':32A '8gb':15B 'blkkbd':24B 'core':5A,12B 'dos':18B 'dvd':20B 'dvd-wr':19B 'g5':27B 'g9':3A,9B 'hdd':17B 'hp':1A,7B 'i7':6A,13B 'intel':4A,11B 'monitor':28B 'odd':22B 'p22v':26B 'tower':10B 'warrati':31B 'wr':21B 'year':30B	5W838ES				hp-290-g9-intel-core-i7	active	876.0000000000	876.0000000000	2024-08-21 07:05:01.997614+00	\N
224f2c5b-ae41-4d33-aa6d-ba487e22f441	HP 78A Black LaserJet Toner Cartridge	HP 78A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	\N	f	'78a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce278a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE278A				hp-78a-black-laserjet-toner-cartridge	active	34.0000000000	34.0000000000	2024-08-21 07:05:01.997614+00	\N
7ec8298b-688b-4862-92ae-b82c0c22cddc	HP E24u G4 23.8 inch IPS FHD USB-C Monitor	HP E24u G4 23.8 inch IPS FHD USB-C Monitor, 1 USB Type-C™, 1 DisplayPort™, 1 HDMI 1.4, 4 USB-A 3.2 Gen 1, Tilt and Height Adjustable, Pivot, Swivel Stand	0	263fe9a0-4b39-4335-b8ce-79fa460b5759	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'1':23B,28B,30B,39B '1.4':32B '189t0aa':47A '23.8':4A,15B '3.2':37B '4':33B 'adjust':43B 'c':10A,21B,27B 'displayport':29B 'e24u':2A,13B 'fhd':7A,18B 'g4':3A,14B 'gen':38B 'hdmi':31B 'height':42B 'hp':1A,12B 'inch':5A,16B 'ip':6A,17B 'monitor':11A,22B 'pivot':44B 'stand':46B 'swivel':45B 'tilt':40B 'type':26B 'type-c':25B 'usb':9A,20B,24B,35B 'usb-a':34B 'usb-c':8A,19B	189T0AA				hp-e24u-g4-23-8-inch-ips-fhd-usb-c-monitor	active	4.0000000000	4.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
4a5f304f-c54b-4b46-89c3-904d6499b1ca	Epson T41F2 Singlepack UltraChrome XD2 Cyan 350ml	Epson T41F2 Singlepack UltraChrome XD2 Cyan 350ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-21 09:16:42.68424+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'350ml':7A,14B,22C,29 'c13t41f240':15A 'cyan':6A,13B,21C,28 'epson':1A,8B,16C,23 'singlepack':3A,10B,18C,25 't41f2':2A,9B,17C,24 'ultrachrom':4A,11B,19C,26 'xd2':5A,12B,20C,27	C13T41F240	Epson T41F2 Singlepack UltraChrome XD2 Cyan 350ml	Epson T41F2 Singlepack UltraChrome XD2 Cyan 350ml	\N	epson-t41f2-singlepack-ultrachrome-xd2-cyan-350ml	active	10.3700000000	8.8900000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
381f8011-5643-42ab-bd48-dc359c52380a	Epson 103 EcoTank Yellow Ink Bottle 65.0 ml	Epson 103 EcoTank Yellow Ink Bottle 65.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'103':2A,10B '65.0':7A,15B 'bottl':6A,14B 'c13t00s44a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 'yellow':4A,12B	C13T00S44A				epson-103-ecotank-yellow-ink-bottle-65-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
ed445874-1562-46b7-ad76-1912b3e016f5	Epson T6732 EcoTank Cyan Ink Bottle 70.0 ml	Epson T6732 EcoTank Cyan Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'70.0':7A,15B 'bottl':6A,14B 'c13t67324a':17A 'cyan':4A,12B 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 't6732':2A,10B	C13T67324A				epson-t6732-ecotank-cyan-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
0c6fecb6-2d32-4173-83f1-cfe30d763597	HP 130A Cyan LaserJet Toner Cartridge	HP 130A Cyan LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'130a':2A,8B 'cartridg':6A,12B 'cf351a':13A 'cyan':3A,9B 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF351A				hp-130a-cyan-laserjet-toner-cartridge	active	7.0000000000	7.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
29bae2f6-b7f2-45f5-a4cf-cc606747c55a	Epson T6733 EcoTank Magenta Ink Bottle 70.0 ml	Epson T6733 EcoTank Magenta Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'70.0':7A,15B 'bottl':6A,14B 'c13t67334a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'magenta':4A,12B 'ml':8A,16B 't6733':2A,10B	C13T67334A				epson-t6733-ecotank-magenta-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
698a5668-5813-4f6b-86eb-24bcecbc2a83	Epson T6734 EcoTank Yellow Ink Bottle 70.0 ml	Epson T6734 EcoTank Yellow Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'70.0':7A,15B 'bottl':6A,14B 'c13t67344a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 't6734':2A,10B 'yellow':4A,12B	C13T67344A				epson-t6734-ecotank-yellow-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
d81e2797-ac06-4d89-a22e-bb2b352cdcec	Epson 115 EcoTank Pigment Black ink bottle 70.0 ml	Epson 115 EcoTank Pigment Black ink bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'115':2A,11B '70.0':8A,17B 'black':5A,14B 'bottl':7A,16B 'c13t07c14a':19A 'ecotank':3A,12B 'epson':1A,10B 'ink':6A,15B 'ml':9A,18B 'pigment':4A,13B	C13T07C14A				epson-115-ecotank-pigment-black-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
51e19126-a3b0-4493-9466-7f3c2953a647	Epson 115 EcoTank Yellow ink bottle 70.0 ml	Epson 115 EcoTank Yellow ink bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'115':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t07d44a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B 'yellow':4A,12B	C13T07D44A				epson-115-ecotank-yellow-ink-bottle-70-0-ml	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
6c392753-410c-4eab-aae5-0cc34b1b8a47	HP 126A Magenta LaserJet Toner Cartridge	HP 126A Magenta LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'126a':2A,8B 'cartridg':6A,12B 'ce313a':13A 'hp':1A,7B 'laserjet':4A,10B 'magenta':3A,9B 'toner':5A,11B	CE313A				hp-126a-magenta-laserjet-toner-cartridge	active	54.0000000000	54.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
943cb219-7173-4d04-8a63-8bd217817f2d	Epson T04D1 Maintenance Box	Epson T04D1 Maintenance Box For Printer, L6270 L6290 L6490 L14150 M1100 M1120 M1140 M1170 M1180 M2120 M2140 M2170 M3140 M3170 M3180	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'box':4A,8B 'c13t04d100':26A 'epson':1A,5B 'l14150':14B 'l6270':11B 'l6290':12B 'l6490':13B 'm1100':15B 'm1120':16B 'm1140':17B 'm1170':18B 'm1180':19B 'm2120':20B 'm2140':21B 'm2170':22B 'm3140':23B 'm3170':24B 'm3180':25B 'mainten':3A,7B 'printer':10B 't04d1':2A,6B	C13T04D100				epson-t04d1-maintenance-box	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
d2b5a122-c84b-41f5-9a9c-a30b1c91fc32	Epson C9345 Maintenance Box	Epson C9345 Maintenance Box For Printer, L1160 L6550 6570 L6580 M15140 M15180 L15180 L8180 L8160 	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'6570':13B 'box':4A,8B 'c12c934591':20A 'c9345':2A,6B 'epson':1A,5B 'l1160':11B 'l15180':17B 'l6550':12B 'l6580':14B 'l8160':19B 'l8180':18B 'm15140':15B 'm15180':16B 'mainten':3A,7B 'printer':10B	C12C934591				epson-c9345-maintenance-box	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
fb3032f9-5e61-4c49-8c44-9a94a01ab1bc	Epson T6716 Maintenance Box	Epson T6716 Maintenance Box For WF-C5790/M5799/C529R/C579R/M5299/M5799/C5710/C5790/ET-8700	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'/m5799/c529r/c579r/m5299/m5799/c5710/c5790/et-8700':13B 'box':4A,8B 'c13t671600':14A 'c5790':12B 'epson':1A,5B 'mainten':3A,7B 't6716':2A,6B 'wf':11B 'wf-c5790':10B	C13T671600				epson-t6716-maintenance-box	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
0edfe0c8-6308-42b2-b42e-2708d2f6786b	Logitech Silent Wireless Mouse M221 Charcoal	Logitech Silent Wireless Mouse M221 Charcoal	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-006510':14A '910':13A 'charcoal':6A,12B 'logitech':1A,7B 'm221':5A,11B 'mous':4A,10B 'silent':2A,8B 'wireless':3A,9B	910-006510				logitech-silent-wireless-mouse-m221-charcoal	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
32354d91-2372-4bda-8684-9f88c05667b9	Logitech Rally Plus Ultra	Logitech Rally Plus Ultra-HD Conference Cam - BLACK - USB 	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-001242':16A '960':15A 'black':13B 'cam':12B 'confer':11B 'hd':10B 'logitech':1A,5B 'plus':3A,7B 'ralli':2A,6B 'ultra':4A,9B 'ultra-hd':8B 'usb':14B	960-001242				logitech-rally-plus-ultra	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
ceb5bcd1-3cb8-4910-95a1-ade3a81289cb	Logitech Expansion Mic For Group 	Logitech Expansion Mic For Group 	0	51e6ba8f-43ef-4367-8faa-d70820b47712	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-000171':12A '989':11A 'expans':2A,7B 'group':5A,10B 'logitech':1A,6B 'mic':3A,8B	989-000171				logitech-expansion-mic-for-group-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
a2e407a0-639a-4dae-8f26-a32577eb2811	Epson EcoTank L4260 	Epson EcoTank L4260 A4 Wi-Fi Duplex All-in-One Ink Tank Printer	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'a4':7B 'all-in-on':12B 'c11cj63415':19A 'duplex':11B 'ecotank':2A,5B 'epson':1A,4B 'fi':10B 'ink':16B 'l4260':3A,6B 'one':15B 'printer':18B 'tank':17B 'wi':9B 'wi-fi':8B	C11CJ63415				epson-ecotank-l4260-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
3efe2e74-69d9-4d81-af6a-ea10543148d4	Epson EcoTank L6270 	Epson EcoTank L6270 Wi-Fi A4 Duplex All-in-One Ink Printer	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'a4':10B 'all-in-on':12B 'c11cj61407':18A 'duplex':11B 'ecotank':2A,5B 'epson':1A,4B 'fi':9B 'ink':16B 'l6270':3A,6B 'one':15B 'printer':17B 'wi':8B 'wi-fi':7B	C11CJ61407				epson-ecotank-l6270-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
bf3000bf-1b22-426c-8987-44a378335b79	Epson LQ-350	Epson LQ-350 Dot matrix printer, 24 pin	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-350':3A,6B '24':10B 'c11cc25002':12A 'dot':7B 'epson':1A,4B 'lq':2A,5B 'matrix':8B 'pin':11B 'printer':9B	C11CC25002				epson-lq-350	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
90a25cf3-60cc-4611-ac47-2a3e73c5bbd7	HP 131A Black LaserJet Toner Cartridge	HP 131A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'131a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf210a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF210A				hp-131a-black-laserjet-toner-cartridge	active	75.0000000000	75.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
93a64578-1807-4889-bfc7-233ad723dcfb	Epson TM-T20X Printer 	Epson TM-T20X (051): USB + Serial, PS, Blk, UK – 200mm/s, RS-232, USB 2.0 Type B, Drawer kick-out. Reliability: 60,000,000 MCBF (Lines), 360,000 MTBF (Hours). Print Head Life: 100 km – 100,000,000 pulses   	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-232':18B '000':29B,30B,34B,43B,44B '051':10B '100':40B,42B '2.0':20B '200mm/s':16B '360':33B '60':28B 'b':22B 'blk':14B 'c31ch26051a0':46A 'drawer':23B 'epson':1A,6B 'head':38B 'hour':36B 'kick':25B 'kick-out':24B 'km':41B 'life':39B 'line':32B 'mcbf':31B 'mtbf':35B 'print':37B 'printer':5A 'ps':13B 'puls':45B 'reliabl':27B 'rs':17B 'serial':12B 't20x':4A,9B 'tm':3A,8B 'tm-t20x':2A,7B 'type':21B 'uk':15B 'usb':11B,19B	C31CH26051A0				epson-tm-t20x-printer-	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
08a0767f-8a3a-4f3e-86ae-72b22f4b1650	Epson T6643 EcoTank Magenta Ink Bottle 70.0 ml	Epson T6643 EcoTank Magenta Ink Bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'70.0':7A,15B 'bottl':6A,14B 'c13t66434a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'magenta':4A,12B 'ml':8A,16B 't6643':2A,10B	C13T66434A				epson-t6643-ecotank-magenta-ink-bottle-70-0-ml	archived	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
6bd556f1-4ebc-41a9-b955-187e10c5f33d	Epson CO-FH02 Smart Full HD	Epson CO-FH02 Smart Full HD 3LCD Technology Projector 3000 lumens, Project a big screen experience in the home or office. This affordable solution is easy-to-use, Full HD and also has Android TV3	0	5db5f6db-a6a8-40ab-a3e2-e7b53da06371	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'3000':18B '3lcd':15B 'afford':31B 'also':41B 'android':43B 'big':22B 'co':3A,10B 'co-fh02':2A,9B 'easi':35B 'easy-to-us':34B 'epson':1A,8B 'experi':24B 'fh02':4A,11B 'full':6A,13B,38B 'hd':7A,14B,39B 'home':27B 'lumen':19B 'offic':29B 'project':20B 'projector':17B 'screen':23B 'smart':5A,12B 'solut':32B 'technolog':16B 'tv3':44B 'use':37B 'v11ha85040':45A	V11HA85040				epson-co-fh02-smart-full-hd	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
52447da5-eba7-498e-973c-216f729c2e84	HP Spectre X360 14	HP SPECTRE X360 14-EF0036NA LAPTOP (CI5-1235U/8GB/512GB/13.5 TOUCH/WIN11H) - NATURAL SILVER	0	8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'/8gb/512gb/13.5':14B '1235u':13B '14':4A,8B '6m020ea':18A 'ci5':12B 'ci5-1235u':11B 'ef0036na':9B 'hp':1A,5B 'laptop':10B 'natur':16B 'silver':17B 'spectr':2A,6B 'touch/win11h':15B 'x360':3A,7B	6M020EA				hp-spectre-x360-14	active	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
dc65edbe-d41b-4036-b88b-0dc5e7a1aa62	Lexar 128GB  NS100 2.5” SATA Internal Hard Disk	Lexar 128GB  NS100 2.5” SATA (6Gb/s) Solid-State Drive, up to 520MB/s Read	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'128gb':2A,10B '128rb':25A '2.5':4A,12B '520mb/s':21B '6gb/s':14B 'disk':8A 'drive':18B 'hard':7A 'intern':6A 'lexar':1A,9B 'lns100':24A 'lns100-128rb':23A 'ns100':3A,11B 'read':22B 'sata':5A,13B 'solid':16B 'solid-st':15B 'state':17B	LNS100-128RB				lexar-128gb-ns100-2-5-sata-internal-hard-disk	active	45.0000000000	45.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
0f462280-7973-4e2b-a272-b9f28877f218	HP 19A LaserJet Imaging Drum 	HP 19A LaserJet Imaging Drum 	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'19a':2A,7B 'cf219a':11A 'drum':5A,10B 'hp':1A,6B 'imag':4A,9B 'laserjet':3A,8B	CF219A				hp-19a-laserjet-imaging-drum-	active	758.0000000000	758.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
a1f45997-fe8c-4a16-9a2e-0ef8ffe9a78f	HP 305A Black LaserJet Toner Cartridge	HP 305A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'305a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce410a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE410A				hp-305a-black-laserjet-toner-cartridge	active	47.0000000000	47.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
70383742-794b-4895-b3ec-76858cf60421	HP 30A Black LaserJet Toner Cartridge	HP 30A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'30a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'cf230a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF230A				hp-30a-black-laserjet-toner-cartridge	active	78.0000000000	78.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
a6b22e47-8473-4a8c-8888-1f8cc8d3fac3	HP 126A Yellow LaserJet Toner Cartridge	HP 126A Yellow LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'126a':2A,8B 'cartridg':6A,12B 'ce312a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B 'yellow':3A,9B	CE312A				hp-126a-yellow-laserjet-toner-cartridge	active	334.0000000000	334.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
aa5602f2-3382-41fd-8ad2-678d0ac7c903	Lexar 1TB External Portable SSD Hard Disk 	Lexar 1TB External Portable SSD Hard Disk , up to 550MB/s Read and 400MB/s Write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'1tb':2A,9B '400mb/s':20B '550mb/s':17B 'disk':7A,14B 'extern':3A,10B 'hard':6A,13B 'lexar':1A,8B 'lsl200x001t':23A 'lsl200x001t-rnnng':22A 'portabl':4A,11B 'read':18B 'rnnng':24A 'ssd':5A,12B 'write':21B	LSL200X001T-RNNNG				lexar-1tb-external-portable-ssd-hard-disk-	active	45.0000000000	45.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
81a54856-595c-4a25-b337-ad756154d36a	Lexar 2TB  M.2 NVMe	Lexar 2TB High Speed PCIe Gen3 with 4 Lanes M.2 NVMe, up to 3300 MB/s read and 3000 MB/s write	0	494456b3-f06c-4ff2-afa5-ef56248257ad	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'2tb':2A,6B '3000':22B '3300':18B '4':12B 'gen3':10B 'high':7B 'lane':13B 'lexar':1A,5B 'lnm620x002t':26A 'lnm620x002t-rnnng':25A 'm.2':3A,14B 'mb/s':19B,23B 'nvme':4A,15B 'pcie':9B 'read':20B 'rnnng':27A 'speed':8B 'write':24B	LNM620X002T-RNNNG				lexar-2tb-m-2-nvme	active	43.0000000000	43.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
bd83ef1f-e865-4c75-aa05-297c54432e78	HP 130A Magenta LaserJet Toner Cartridge	HP 130A Magenta LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'130a':2A,8B 'cartridg':6A,12B 'cf353a':13A 'hp':1A,7B 'laserjet':4A,10B 'magenta':3A,9B 'toner':5A,11B	CF353A				hp-130a-magenta-laserjet-toner-cartridge	active	92.0000000000	92.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
7f08de77-01f0-4bc6-a479-fd480c91febc	HP 85A Black LaserJet Toner Cartridge	HP 85A Black LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'85a':2A,8B 'black':3A,9B 'cartridg':6A,12B 'ce285a':13A 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CE285A				hp-85a-black-laserjet-toner-cartridge	active	21.0000000000	21.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
1b11603b-ca02-45ce-a43e-5a1a8583173e	HP 131A Cyan LaserJet Toner Cartridge	HP 131A Cyan LaserJet Toner Cartridge	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'131a':2A,8B 'cartridg':6A,12B 'cf211a':13A 'cyan':3A,9B 'hp':1A,7B 'laserjet':4A,10B 'toner':5A,11B	CF211A				hp-131a-cyan-laserjet-toner-cartridge	active	3.0000000000	3.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
6ee18311-c27c-43d5-a07f-9560c5f20105	HP 290 G9 Intel Core i5	HP 290 G9 Tower Intel Core i5-12500 / 8GB / 512GB SSD / DOS / DVD-WR ODD  / 125 BLKkbd / 125mouse / P22v G5 Monitor, 1 Year Warranty	0	51223939-57c9-4b21-a222-1dfd6451c251	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-12500':14B '1':29B '125':23B '125mouse':25B '290':2A,8B '512gb':16B '5w892es':32A '8gb':15B 'blkkbd':24B 'core':5A,12B 'dos':18B 'dvd':20B 'dvd-wr':19B 'g5':27B 'g9':3A,9B 'hp':1A,7B 'i5':6A,13B 'intel':4A,11B 'monitor':28B 'odd':22B 'p22v':26B 'ssd':17B 'tower':10B 'warranti':31B 'wr':21B 'year':30B	5W892ES				hp-290-g9-intel-core-i5	active	35.0000000000	35.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
c2c1f3ec-dc41-4724-89c2-e01a9336e247	HP E23 G4 23-inch Diagonal IPS FHD Monitor	HP E23 G4 23-inch Diagonal IPS FHD Monitor with 1 VGA; 1 USB Type-B; 1 HDMI 1.4; 1 DisplayPort™ 1.2; 4 USB-A 3.2 Gen 1 ports	0	263fe9a0-4b39-4335-b8ce-79fa460b5759	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'1':20B,22B,27B,30B,39B '1.2':32B '1.4':29B '23':4A,13B '3.2':37B '4':33B '9vf96aa':41A 'b':26B 'diagon':6A,15B 'displayport':31B 'e23':2A,11B 'fhd':8A,17B 'g4':3A,12B 'gen':38B 'hdmi':28B 'hp':1A,10B 'inch':5A,14B 'ip':7A,16B 'monitor':9A,18B 'port':40B 'type':25B 'type-b':24B 'usb':23B,35B 'usb-a':34B 'vga':21B	9VF96AA				hp-e23-g4-23-inch-diagonal-ips-fhd-monitor	active	46.0000000000	46.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
9c77bf3f-d362-499e-8df3-6d23fa3d1a12	Test 1	Test Description	1	8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	2024-08-21 10:38:28.634091+00	2024-08-21 11:25:47.523399+00	2a6a713a-a7b6-401b-8978-23ee77cb370f	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-1':6A '1':2A,8C 'descript':4B,10 'test':1A,3B,5A,7C,9	Test-1	Test 1	Test Description	\N	test-1	active	1320.0000000000	1200.0000000000	2024-08-21 10:38:28.634091+00	2025-08-21 10:38:28.634091+00
a59c96d9-45ec-4f40-99f4-79fea9d74405	HP 32A LaserJet Imaging Drum	HP 32A LaserJet Imaging Drum	0	74018456-4a2b-4e12-8bdb-6118d7013477	2024-08-15 10:15:09.142827+00	2024-09-02 08:01:01.108081+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'32a':2A,7B,13C,18 'cf232a':11A 'drum':5A,10B,16C,21 'hp':1A,6B,12C,17 'imag':4A,9B,15C,20 'laserjet':3A,8B,14C,19	CF232A	HP 32A LaserJet Imaging Drum 	HP 32A LaserJet Imaging Drum 	\N	hp-32a-laserjet-imaging-drum-	active	75.5555555556	74.8740740741	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
884a5e2f-af66-4bbf-84dd-7abfcdcee422	Test 0	abcd	1	8f4f02a7-d485-4fe4-8abc-e2289bfa2dba	2024-09-07 15:45:23.753018+00	2024-09-07 15:45:23.753018+00	2a6a713a-a7b6-401b-8978-23ee77cb370f	\N	f	'-0':5A '0':2A,7C 'abcd':3B,8 'test':1A,4A,6C	test-0	Test 0	abcd	\N	test-0	active	0.7692307692	0.0769230769	2024-09-07 15:45:23.753018+00	2025-09-07 15:45:23.753018+00
a183524c-0aff-45eb-a9f5-6b3228f697f4	Epson SIDM Black Ribbon Cartridge for DFX-9000	Epson SIDM Black Ribbon Cartridge for DFX-9000	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'-9000':8A,16B 'black':3A,11B 'c13s015384ba':17A 'cartridg':5A,13B 'dfx':7A,15B 'epson':1A,9B 'ribbon':4A,12B 'sidm':2A,10B	C13S015384BA				epson-sidm-black-ribbon-cartridge-for-dfx-9000	archived	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
34871d1c-ae6e-4b88-a1a1-77a48207b7a7	Epson 103 EcoTank Black Ink Bottle 65.0 ml	Epson 103 EcoTank Black Ink Bottle 65.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'103':2A,10B '65.0':7A,15B 'black':4A,12B 'bottl':6A,14B 'c13t00s14a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'ml':8A,16B	C13T00S14A				epson-103-ecotank-black-ink-bottle-65-0-ml	archived	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
ea6453da-dd43-4cad-af87-c662de928ad7	Epson 115 EcoTank Magenta ink bottle 70.0 ml	Epson 115 EcoTank Magenta ink bottle 70.0 ml	0	f5a894b5-ddae-4c30-b0b3-61ba0563c6f2	2024-08-15 10:15:09.142827+00	2024-08-15 10:15:09.142827+00	\N	2a6a713a-a7b6-401b-8978-23ee77cb370f	f	'115':2A,10B '70.0':7A,15B 'bottl':6A,14B 'c13t07d34a':17A 'ecotank':3A,11B 'epson':1A,9B 'ink':5A,13B 'magenta':4A,12B 'ml':8A,16B	C13T07D34A				epson-115-ecotank-magenta-ink-bottle-70-0-ml	archived	0.0000000000	0.0000000000	2024-08-21 07:05:01.997614+00	2025-08-21 07:05:01.997614+00
\.


--
-- Data for Name: promotion_products; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.promotion_products (promotion_id, product_id, created_at, updated_at) FROM stdin;
000c20c8-0bae-4d28-a3e5-92a147123e61	6bd556f1-4ebc-41a9-b955-187e10c5f33d	2024-08-20 15:34:13.926261+00	2024-08-20 15:34:13.926261+00
000c20c8-0bae-4d28-a3e5-92a147123e61	9f8caa44-07f5-44c7-928d-64be8595f8b1	2024-08-20 15:34:13.926261+00	2024-08-20 15:34:13.926261+00
000c20c8-0bae-4d28-a3e5-92a147123e61	70383742-794b-4895-b3ec-76858cf60421	2024-08-20 15:34:13.926261+00	2024-08-20 15:34:13.926261+00
000c20c8-0bae-4d28-a3e5-92a147123e61	81a54856-595c-4a25-b337-ad756154d36a	2024-08-20 15:34:13.926261+00	2024-08-20 15:34:13.926261+00
\.


--
-- Data for Name: promotions; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.promotions (id, title, description, image_url, created_at, updated_at, start_date, end_date, slug, status) FROM stdin;
000c20c8-0bae-4d28-a3e5-92a147123e61	Black Friday	Description	promotions/1724168051-promotion-1.png	2024-08-20 15:34:13.876208+00	2024-08-20 15:34:13.876208+00	2024-08-19 21:00:00+00	2024-08-26 21:00:00+00	black-friday	active
\.


--
-- Data for Name: refresh_tokens; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.refresh_tokens (id, user_id, token, created_at, expires_at, revoked_at) FROM stdin;
35deffa9-1808-4d9f-9c9a-83db6d712a6b	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzMTIxLCJpYXQiOjE3MjM3MTcxMjF9.uvq4U4AJskWKI-s59ZNJSWAfnHUPm8KPadDCvHJLCw8	2024-08-15 10:18:41.850214+00	2024-11-13 10:18:41.850102+00	\N
9701fddb-bd17-49f2-9537-c263db7d2f4f	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzMjIzLCJpYXQiOjE3MjM3MTcyMjN9.cLBB6cuZ2HFimapE1QbTIDjCaRXCkoorMOuKTVxGrmI	2024-08-15 10:20:23.668371+00	2024-11-13 10:20:23.668271+00	\N
9fbdb66c-ef00-4a11-b7b0-9c4eec1dc9b1	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzMjI5LCJpYXQiOjE3MjM3MTcyMjl9.q15DxyPR0rCBv81SsfpYjoO5aXU24YNQshsNLWa5TZg	2024-08-15 10:20:29.727621+00	2024-11-13 10:20:29.727531+00	\N
7de82a11-3651-4ac0-9cc1-32a3c15f1dbf	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzNDcxLCJpYXQiOjE3MjM3MTc0NzF9.Y569iynyWzCQ3CEd-AWyGYtygnzRK6_AYFzVTMPAXgo	2024-08-15 10:24:31.396443+00	2024-11-13 10:24:31.396271+00	\N
a8939e5a-5e04-427d-a2d4-12a6d442fc99	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzNjAwLCJpYXQiOjE3MjM3MTc2MDB9.Bc55iehqD5f_r_32NooT8ABUJVKLtsEtP9OWF_JlrCQ	2024-08-15 10:26:40.632853+00	2024-11-13 10:26:40.632707+00	\N
46591eee-6072-4753-a4d1-c6239591bfb1	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzNjUxLCJpYXQiOjE3MjM3MTc2NTF9.9Yk0WrmeioaGuyWjbjnFgb6ozBEpTEeA9VlGEaEdzrQ	2024-08-15 10:27:31.269458+00	2024-11-13 10:27:31.269256+00	\N
e3ebbc35-9b52-4d0b-89ce-726a6a98db9b	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzNzE5LCJpYXQiOjE3MjM3MTc3MTl9.BlnjGf9AKyUamMpuowyJhZ11d-R_ph3YW9iW2AasozY	2024-08-15 10:28:39.595314+00	2024-11-13 10:28:39.595146+00	\N
b9b15c97-c3eb-4650-a2fd-d9ab02bb5872	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDkzNzU1LCJpYXQiOjE3MjM3MTc3NTV9.juoxncbpkXRtYAblPBEa2EMYafP9tlW09uSzqFDah-Y	2024-08-15 10:29:15.008668+00	2024-11-13 10:29:15.008519+00	\N
136e739e-7207-4f94-98f8-6284572e6969	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0MDA0LCJpYXQiOjE3MjM3MTgwMDR9.gQKgM5CBwD8RsnNEGaRw1nZb_NOzB7St9-xOH4XnFn0	2024-08-15 10:33:24.161948+00	2024-11-13 10:33:24.161754+00	\N
c75628ea-62d1-418e-b853-932d954c01a3	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0MDMyLCJpYXQiOjE3MjM3MTgwMzJ9.3830GeJfbNOSwaIdxxPVUkUlteHbxCAlYpjovvtg8TY	2024-08-15 10:33:52.97151+00	2024-11-13 10:33:52.971348+00	\N
5fb1d6ab-da9b-413c-b1b9-4fdd57633830	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0MDU1LCJpYXQiOjE3MjM3MTgwNTV9.jtEBjMrnJcVTb4pyp2WWt3AgxgdlwrKbQXafKu2AhM8	2024-08-15 10:34:15.103261+00	2024-11-13 10:34:15.103093+00	\N
42102af2-a772-4756-a756-0049d4117224	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0MTQxLCJpYXQiOjE3MjM3MTgxNDF9.9IMjOzUGwShfqLs1AgMNMn07ghjhhv4umTAooz4yoGo	2024-08-15 10:35:41.001187+00	2024-11-13 10:35:41.001042+00	\N
1ac5fb9d-cfe6-452f-ac96-3477375bb18e	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0MzQwLCJpYXQiOjE3MjM3MTgzNDB9.fIoLp3zw8s04MpgYu_ns7z0vIBrl6De-uEoqHr3HCi8	2024-08-15 10:39:00.348265+00	2024-11-13 10:39:00.348092+00	\N
345317f5-75ad-432e-a935-946b5bc9519c	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0Mzc2LCJpYXQiOjE3MjM3MTgzNzZ9.Hc7-4o4JGkZJt9eOnhU-Hj54O2-7qNfTxRW9xYVjzfE	2024-08-15 10:39:36.498527+00	2024-11-13 10:39:36.498364+00	\N
b25e82b4-1640-47cb-9282-f7b98817fa80	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0NDAxLCJpYXQiOjE3MjM3MTg0MDF9.PrdJmYVv8U-5SMlbCLbYNnqu0PDBEJZzOwlidO-DoEw	2024-08-15 10:40:01.610108+00	2024-11-13 10:40:01.609956+00	\N
dfc6f76e-d988-471b-98ec-b9dd26b28e43	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0NDM1LCJpYXQiOjE3MjM3MTg0MzV9.SHYdPs6w_BnBo94DUCQnYKlZquG209Oy2VCNNh1EDWQ	2024-08-15 10:40:35.066586+00	2024-11-13 10:40:35.066421+00	\N
de4a7dc3-4c30-4211-9f52-307efcccfeab	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0NDc1LCJpYXQiOjE3MjM3MTg0NzV9.4ciwWhHTgRHj90uvGpwMK4Z9ogtTWFtsTkdzYm0W2AQ	2024-08-15 10:41:15.212157+00	2024-11-13 10:41:15.211998+00	\N
c3fccd58-c437-4798-a04c-41d8da2a4a66	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0NTU5LCJpYXQiOjE3MjM3MTg1NTl9.bYVpX8amN5L27e5J_HX8tYRUqEQ2zl_-q-7oaBSDKbc	2024-08-15 10:42:39.414291+00	2024-11-13 10:42:39.414165+00	\N
e0629888-1ef1-4fdd-9a91-8d1991516ca8	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0NTY1LCJpYXQiOjE3MjM3MTg1NjV9.WOjhZJdqKlY1e_FCC_8OxT58asFfcdaxD333IndqLuQ	2024-08-15 10:42:45.55497+00	2024-11-13 10:42:45.554714+00	\N
869f2b4d-e5ad-4e32-9a65-76c4765f2b73	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxNDk0ODIzLCJpYXQiOjE3MjM3MTg4MjN9.i2IFnhWvBdvig-hKaWpJ59rXiVgocZaDhiIRwis3ceg	2024-08-15 10:47:03.350597+00	2024-11-13 10:47:03.350469+00	\N
f983150e-2628-46e6-90d4-68a88a86cc89	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMxOTA3OTc3LCJpYXQiOjE3MjQxMzE5Nzd9.MeJpFDgptTurhx7Y70Hm9vw1XDLyEDmG3Op10Aq6TeU	2024-08-20 05:32:57.341344+00	2024-11-18 05:32:57.341256+00	\N
92ffa628-58e0-40ce-86f2-ad76979f6519	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMyMDA0MTc5LCJpYXQiOjE3MjQyMjgxNzl9.AAwsALRnFxxvTcXoX5LnogD6Z6mNnl1TGg-37M9DLQg	2024-08-21 08:16:19.617426+00	2024-11-19 08:16:19.617245+00	\N
a116087d-de76-45a2-a748-4a929c928b1e	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMyNTE1NzI4LCJpYXQiOjE3MjQ3Mzk3Mjh9.s8lqmrsP_sm87qHWGQm-DjR3wjL7PQrpukxjiU6dIeU	2024-08-27 06:22:08.833924+00	2024-11-25 06:22:08.833749+00	\N
c86bcd53-3512-41ab-8ac5-8ecd016689dd	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMyNjIyMTQyLCJpYXQiOjE3MjQ4NDYxNDJ9.fRkp7y966wW4W6Z7HkD0HhwV6KXfX30I8GtWF-NUEzU	2024-08-28 11:55:42.645204+00	2024-11-26 11:55:42.645058+00	\N
031eb48d-2ce9-4007-9f89-a48aca3657e8	2a6a713a-a7b6-401b-8978-23ee77cb370f	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyYTZhNzEzYS1hN2I2LTQwMWItODk3OC0yM2VlNzdjYjM3MGYiLCJlbWFpbCI6ImJhaGF0aWdlcmFsZDBAZ21haWwuY29tIiwic3ViIjoiMmE2YTcxM2EtYTdiNi00MDFiLTg5NzgtMjNlZTc3Y2IzNzBmIiwiZXhwIjoxNzMzNjQ0MzUxLCJpYXQiOjE3MjU4NjgzNTF9.1TIIdG5RzZdkvda3XDdWwHi6atMSWw5SYLl_AA5_Lyg	2024-09-09 07:52:31.593859+00	2024-12-08 07:52:31.593514+00	\N
\.


--
-- Data for Name: related_products; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.related_products (product_id, related_product_id) FROM stdin;
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.roles (id, role_name, description, created_at, updated_at) FROM stdin;
39781a68-8dc6-47d8-bc92-58028c7f0750	admin	\N	2024-07-31 14:49:32.40906+00	2024-07-31 14:49:32.40906+00
38fbb788-c654-41a9-b241-b2c625f932c1	customer	\N	2024-07-31 14:49:39.974387+00	2024-07-31 14:49:39.974387+00
1462d044-575c-4e31-bf7e-ba54db12fc51	user	\N	2024-07-31 14:49:46.071193+00	2024-07-31 14:49:46.071193+00
\.


--
-- Data for Name: shipment; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.shipment (id, order_id, shipment_status, tracking_number, shipped_date, delivery_date, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: shopping_carts; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.shopping_carts (id, user_id, session_id, total_items, total_price, created_at, updated_at) FROM stdin;
72985fc5-7a52-461f-a5c4-80a8e9d3a2c4	\N	d3cef952-552a-4255-ad57-6a5953a54915	0	0.00	2024-08-14 21:29:00.301142+00	2024-08-14 21:29:00.301142+00
00810ae4-4019-4550-8cf3-b1f8021e5f46	\N	74954e4a-71cd-433e-9be1-c77775750b87	0	0.00	2024-08-14 21:33:21.653992+00	2024-08-14 21:33:21.653992+00
1e6d642c-fb2e-4ca4-bea6-cba2f4f01b17	\N	387d0ca1-e5c3-48ca-a53f-d1745acb6dab	0	0.00	2024-08-27 06:14:55.149865+00	2024-08-27 06:14:55.149865+00
9a2c13e9-433e-4a47-9156-5847f4a17f7f	\N	cac13bcf-dafd-43fc-a869-96967721e2f7	0	0.00	2024-08-27 14:44:08.988524+00	2024-08-27 14:44:08.988524+00
\.


--
-- Data for Name: user_addresses; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.user_addresses (id, user_id, address, is_default, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.user_roles (id, user_id, created_at, updated_at, role_id) FROM stdin;
243d7edc-3f4d-4112-84eb-010890288d4a	2a6a713a-a7b6-401b-8978-23ee77cb370f	2024-07-31 15:37:48.197011+00	2024-07-31 15:37:48.197011+00	38fbb788-c654-41a9-b241-b2c625f932c1
526d7641-7e48-4030-96d7-e324c7de474b	2a6a713a-a7b6-401b-8978-23ee77cb370f	2024-07-31 15:41:37.357733+00	2024-07-31 15:41:37.357733+00	39781a68-8dc6-47d8-bc92-58028c7f0750
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.users (id, email, hashed_password, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login, provider, provider_id, email_verified_at) FROM stdin;
f4168cbb-b7c0-4982-a331-c48313618f89	bahatigerald96@gmail.com	$2a$10$Y/L7Uatyj6ul81qXuBo6s.sDPPCryPp2eBvjtC3nxxxCioPSmQbb6	\N	\N	\N	\N	\N	t	2024-08-15 10:43:00.563193+00	2024-08-15 10:43:00.563193+00	\N	\N	\N	\N
2a6a713a-a7b6-401b-8978-23ee77cb370f	bahatigerald0@gmail.com	$2a$10$4T4v5dil1/Me27vDTURvvOT3ZyeOGb6Gdx5/8mmDNaoYSDBtoyMfC	\N	\N	\N	\N	\N	t	2024-07-31 15:37:48.193684+00	2024-08-15 10:20:16.057696+00	2024-09-09 07:52:31.61373+00	\N	\N	2024-07-31 15:38:08.346049+00
\.


--
-- Data for Name: verification_tokens; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.verification_tokens (id, email, token, expires_at, created_at) FROM stdin;
\.


--
-- Data for Name: wishlist_items; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.wishlist_items (id, wishlist_id, product_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: wishlists; Type: TABLE DATA; Schema: public; Owner: geraldbahati
--

COPY public.wishlists (id, user_id, name, created_at, updated_at) FROM stdin;
\.


--
-- Name: attribute_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: geraldbahati
--

SELECT pg_catalog.setval('public.attribute_types_id_seq', 5, true);


--
-- Name: exchange_rates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: geraldbahati
--

SELECT pg_catalog.setval('public.exchange_rates_id_seq', 1, true);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: geraldbahati
--

SELECT pg_catalog.setval('public.goose_db_version_id_seq', 164, true);


--
-- Name: order_number_seq; Type: SEQUENCE SET; Schema: public; Owner: geraldbahati
--

SELECT pg_catalog.setval('public.order_number_seq', 1, true);


--
-- Name: payment_methods_id_seq; Type: SEQUENCE SET; Schema: public; Owner: geraldbahati
--

SELECT pg_catalog.setval('public.payment_methods_id_seq', 5, true);


--
-- Name: payment_statuses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: geraldbahati
--

SELECT pg_catalog.setval('public.payment_statuses_id_seq', 3, true);


--
-- Name: admin_approval_tokens admin_approval_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.admin_approval_tokens
    ADD CONSTRAINT admin_approval_tokens_pkey PRIMARY KEY (id);


--
-- Name: admin_approval_tokens admin_approval_tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.admin_approval_tokens
    ADD CONSTRAINT admin_approval_tokens_token_key UNIQUE (token);


--
-- Name: admin_requests admin_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.admin_requests
    ADD CONSTRAINT admin_requests_pkey PRIMARY KEY (id);


--
-- Name: attribute_types attribute_types_name_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.attribute_types
    ADD CONSTRAINT attribute_types_name_key UNIQUE (name);


--
-- Name: attribute_types attribute_types_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.attribute_types
    ADD CONSTRAINT attribute_types_pkey PRIMARY KEY (id);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: categories categories_slug_unique; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_slug_unique UNIQUE (slug);


--
-- Name: discounts discounts_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_pkey PRIMARY KEY (id);


--
-- Name: exchange_rates exchange_rates_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT exchange_rates_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: guest_checkouts guest_checkouts_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.guest_checkouts
    ADD CONSTRAINT guest_checkouts_pkey PRIMARY KEY (id);


--
-- Name: order_item_options order_item_options_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_item_options
    ADD CONSTRAINT order_item_options_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: order_payments order_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_payments
    ADD CONSTRAINT order_payments_pkey PRIMARY KEY (id);


--
-- Name: order_shipments order_shipments_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_shipments
    ADD CONSTRAINT order_shipments_pkey PRIMARY KEY (id);


--
-- Name: order_status_history order_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_pkey PRIMARY KEY (id);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: products part_number_unique; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT part_number_unique UNIQUE (part_number);


--
-- Name: password_reset_tokens password_reset_tokens_email_token_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_email_token_key UNIQUE (email, token);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_key UNIQUE (token);


--
-- Name: payment_methods payment_methods_method_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_method_key UNIQUE (method);


--
-- Name: payment_methods payment_methods_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_pkey PRIMARY KEY (id);


--
-- Name: payment_statuses payment_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.payment_statuses
    ADD CONSTRAINT payment_statuses_pkey PRIMARY KEY (id);


--
-- Name: payment_statuses payment_statuses_status_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.payment_statuses
    ADD CONSTRAINT payment_statuses_status_key UNIQUE (status);


--
-- Name: product_attribute_values product_attribute_values_attribute_id_value_category_id_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attribute_values
    ADD CONSTRAINT product_attribute_values_attribute_id_value_category_id_key UNIQUE (attribute_id, value, category_id);


--
-- Name: product_attribute_values product_attribute_values_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attribute_values
    ADD CONSTRAINT product_attribute_values_pkey PRIMARY KEY (id);


--
-- Name: product_attributes product_attributes_name_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attributes
    ADD CONSTRAINT product_attributes_name_key UNIQUE (name);


--
-- Name: product_attributes product_attributes_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attributes
    ADD CONSTRAINT product_attributes_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: product_interactions product_interactions_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_interactions
    ADD CONSTRAINT product_interactions_pkey PRIMARY KEY (id);


--
-- Name: product_option_values product_option_values_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_option_values
    ADD CONSTRAINT product_option_values_pkey PRIMARY KEY (id);


--
-- Name: product_options product_options_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_options
    ADD CONSTRAINT product_options_pkey PRIMARY KEY (id);


--
-- Name: product_reviews product_reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT product_reviews_pkey PRIMARY KEY (id);


--
-- Name: product_specifications product_specifications_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_specifications
    ADD CONSTRAINT product_specifications_pkey PRIMARY KEY (id);


--
-- Name: product_to_attribute_values product_to_attribute_values_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_to_attribute_values
    ADD CONSTRAINT product_to_attribute_values_pkey PRIMARY KEY (id);


--
-- Name: product_to_attribute_values product_to_attribute_values_product_id_attribute_value_id_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_to_attribute_values
    ADD CONSTRAINT product_to_attribute_values_product_id_attribute_value_id_key UNIQUE (product_id, attribute_value_id);


--
-- Name: product_variants product_variants_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT product_variants_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promotion_products promotion_products_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.promotion_products
    ADD CONSTRAINT promotion_products_pkey PRIMARY KEY (promotion_id, product_id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_slug_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_slug_key UNIQUE (slug);


--
-- Name: promotions promotions_slug_unique; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_slug_unique UNIQUE (slug);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_user_id_token_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_token_key UNIQUE (user_id, token);


--
-- Name: related_products related_products_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.related_products
    ADD CONSTRAINT related_products_pkey PRIMARY KEY (product_id, related_product_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: shipment shipment_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.shipment
    ADD CONSTRAINT shipment_pkey PRIMARY KEY (id);


--
-- Name: shopping_carts shopping_carts_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.shopping_carts
    ADD CONSTRAINT shopping_carts_pkey PRIMARY KEY (id);


--
-- Name: shopping_carts shopping_carts_user_id_session_id_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.shopping_carts
    ADD CONSTRAINT shopping_carts_user_id_session_id_key UNIQUE (user_id, session_id);


--
-- Name: exchange_rates unique_currency_valid_range; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT unique_currency_valid_range UNIQUE (currency_code, valid_from, valid_to);


--
-- Name: users unique_email; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_email UNIQUE (email);


--
-- Name: verification_tokens unique_email_token; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT unique_email_token UNIQUE (email, token);


--
-- Name: products unique_product_name_category; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT unique_product_name_category UNIQUE (name, category_id);


--
-- Name: products unique_product_slug; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT unique_product_slug UNIQUE (slug);


--
-- Name: product_specifications unique_product_specification; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_specifications
    ADD CONSTRAINT unique_product_specification UNIQUE (product_id, spec_name);


--
-- Name: product_variants unique_product_variant; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT unique_product_variant UNIQUE (product_id, variant_name, variant_value);


--
-- Name: roles unique_role_name; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT unique_role_name UNIQUE (role_name);


--
-- Name: cart_items unique_shopping_cart_product; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT unique_shopping_cart_product UNIQUE (shopping_cart_id, product_id);


--
-- Name: user_addresses user_addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_pkey PRIMARY KEY (id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: user_roles user_roles_user_id_role_id_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_role_id_key UNIQUE (user_id, role_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: verification_tokens verification_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_pkey PRIMARY KEY (id);


--
-- Name: verification_tokens verification_tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_token_key UNIQUE (token);


--
-- Name: wishlist_items wishlist_items_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.wishlist_items
    ADD CONSTRAINT wishlist_items_pkey PRIMARY KEY (id);


--
-- Name: wishlists wishlists_pkey; Type: CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.wishlists
    ADD CONSTRAINT wishlists_pkey PRIMARY KEY (id);


--
-- Name: cart_items_shopping_cart_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX cart_items_shopping_cart_id_idx ON public.cart_items USING btree (shopping_cart_id);


--
-- Name: categories_name_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX categories_name_idx ON public.categories USING btree (name);


--
-- Name: idx_admin_approval_tokens_expires_at; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_admin_approval_tokens_expires_at ON public.admin_approval_tokens USING btree (expires_at);


--
-- Name: idx_admin_approval_tokens_request_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_admin_approval_tokens_request_id ON public.admin_approval_tokens USING btree (request_id);


--
-- Name: idx_admin_requests_created_at; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_admin_requests_created_at ON public.admin_requests USING btree (created_at);


--
-- Name: idx_admin_requests_status; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_admin_requests_status ON public.admin_requests USING btree (status);


--
-- Name: idx_admin_requests_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_admin_requests_user_id ON public.admin_requests USING btree (user_id);


--
-- Name: idx_cart_item_cart_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_cart_item_cart_id ON public.cart_items USING btree (shopping_cart_id);


--
-- Name: idx_cart_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_cart_user_id ON public.shopping_carts USING btree (user_id);


--
-- Name: idx_categories_parent_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_categories_parent_id ON public.categories USING btree (parent_id);


--
-- Name: idx_categories_parent_id_name; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_categories_parent_id_name ON public.categories USING btree (parent_id, name);


--
-- Name: idx_categories_path; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_categories_path ON public.categories USING gist (path);


--
-- Name: idx_categories_search_vector; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_categories_search_vector ON public.categories USING gin (search_vector);


--
-- Name: idx_discounts_end_date; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_discounts_end_date ON public.discounts USING btree (end_date);


--
-- Name: idx_discounts_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_discounts_product_id ON public.discounts USING btree (product_id);


--
-- Name: idx_discounts_start_date; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_discounts_start_date ON public.discounts USING btree (start_date);


--
-- Name: idx_exchange_rates_validity; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_exchange_rates_validity ON public.exchange_rates USING btree (currency_code, valid_from DESC, valid_to DESC);


--
-- Name: idx_order_item_options_option_type; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_item_options_option_type ON public.order_item_options USING btree (option_type);


--
-- Name: idx_order_item_options_option_value; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_item_options_option_value ON public.order_item_options USING btree (option_value);


--
-- Name: idx_order_item_options_order_item_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_item_options_order_item_id ON public.order_item_options USING btree (order_item_id);


--
-- Name: idx_order_item_order_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_item_order_id ON public.order_items USING btree (order_id);


--
-- Name: idx_order_payments_checkout_request_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_payments_checkout_request_id ON public.order_payments USING btree (checkout_request_id);


--
-- Name: idx_order_payments_order_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_payments_order_id ON public.order_payments USING btree (order_id);


--
-- Name: idx_order_payments_payment_method_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_payments_payment_method_id ON public.order_payments USING btree (payment_method_id);


--
-- Name: idx_order_payments_payment_status_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_payments_payment_status_id ON public.order_payments USING btree (payment_status_id);


--
-- Name: idx_order_shipments_order_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_shipments_order_id ON public.order_shipments USING btree (order_id);


--
-- Name: idx_order_status; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_status ON public.orders USING btree (status);


--
-- Name: idx_order_status_history_order_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_status_history_order_id ON public.order_status_history USING btree (order_id);


--
-- Name: idx_order_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_order_user_id ON public.orders USING btree (user_id);


--
-- Name: idx_password_reset_tokens_email; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_password_reset_tokens_email ON public.password_reset_tokens USING btree (email);


--
-- Name: idx_password_reset_tokens_expires_at; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_password_reset_tokens_expires_at ON public.password_reset_tokens USING btree (expires_at);


--
-- Name: idx_product_attribute_values_attribute_id_category_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_attribute_values_attribute_id_category_id ON public.product_attribute_values USING btree (attribute_id, category_id);


--
-- Name: idx_product_attributes_attribute_type_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_attributes_attribute_type_id ON public.product_attributes USING btree (attribute_type_id);


--
-- Name: idx_product_interactions_interaction_time; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_interactions_interaction_time ON public.product_interactions USING btree (interaction_time);


--
-- Name: idx_product_interactions_interaction_type; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_interactions_interaction_type ON public.product_interactions USING btree (interaction_type);


--
-- Name: idx_product_interactions_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_interactions_product_id ON public.product_interactions USING btree (product_id);


--
-- Name: idx_product_interactions_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_interactions_user_id ON public.product_interactions USING btree (user_id);


--
-- Name: idx_product_option_values_option_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_option_values_option_id ON public.product_option_values USING btree (option_id);


--
-- Name: idx_product_option_values_value_name; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_option_values_value_name ON public.product_option_values USING btree (value_name);


--
-- Name: idx_product_options_option_name; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_options_option_name ON public.product_options USING btree (option_name);


--
-- Name: idx_product_options_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_options_product_id ON public.product_options USING btree (product_id);


--
-- Name: idx_product_to_attribute_values_product_id_attribute_value_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_to_attribute_values_product_id_attribute_value_id ON public.product_to_attribute_values USING btree (product_id, attribute_value_id);


--
-- Name: idx_product_variants_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_variants_product_id ON public.product_variants USING btree (product_id);


--
-- Name: idx_product_variants_variant_name; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_product_variants_variant_name ON public.product_variants USING btree (variant_name);


--
-- Name: idx_products_part_number; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_part_number ON public.products USING btree (part_number);


--
-- Name: idx_products_price_per_unit; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_price_per_unit ON public.products USING btree (price_per_unit);


--
-- Name: idx_products_price_status; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_price_status ON public.products USING btree (usd_price, status);


--
-- Name: idx_products_search_keyword; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_search_keyword ON public.products USING gin (search_keyword);


--
-- Name: idx_products_status; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_status ON public.products USING btree (status);


--
-- Name: idx_products_valid_from; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_valid_from ON public.products USING btree (valid_from);


--
-- Name: idx_products_valid_to; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_products_valid_to ON public.products USING btree (valid_to);


--
-- Name: idx_promotion_products_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_promotion_products_product_id ON public.promotion_products USING btree (product_id);


--
-- Name: idx_promotion_products_promotion_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_promotion_products_promotion_id ON public.promotion_products USING btree (promotion_id);


--
-- Name: idx_promotions_end_date; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_promotions_end_date ON public.promotions USING btree (end_date);


--
-- Name: idx_promotions_slug; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_promotions_slug ON public.promotions USING btree (slug);


--
-- Name: idx_promotions_start_date; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_promotions_start_date ON public.promotions USING btree (start_date);


--
-- Name: idx_promotions_status; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_promotions_status ON public.promotions USING btree (status);


--
-- Name: idx_recommendations_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_recommendations_product_id ON public.recommendations USING btree (product_id);


--
-- Name: idx_recommendations_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_recommendations_user_id ON public.recommendations USING btree (user_id);


--
-- Name: idx_related_products_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_related_products_product_id ON public.related_products USING btree (product_id);


--
-- Name: idx_related_products_related_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_related_products_related_product_id ON public.related_products USING btree (related_product_id);


--
-- Name: idx_user_preferences_product_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_user_preferences_product_id ON public.user_preferences USING btree (product_id);


--
-- Name: idx_user_preferences_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_user_preferences_user_id ON public.user_preferences USING btree (user_id);


--
-- Name: idx_verification_tokens_email; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_verification_tokens_email ON public.verification_tokens USING btree (email);


--
-- Name: idx_verification_tokens_expires_at; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_verification_tokens_expires_at ON public.verification_tokens USING btree (expires_at);


--
-- Name: idx_verification_tokens_token; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_verification_tokens_token ON public.verification_tokens USING btree (token);


--
-- Name: idx_wishlist_item_wishlist_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_wishlist_item_wishlist_id ON public.wishlist_items USING btree (wishlist_id);


--
-- Name: idx_wishlist_user_id; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX idx_wishlist_user_id ON public.wishlists USING btree (user_id);


--
-- Name: order_items_order_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX order_items_order_id_idx ON public.order_items USING btree (order_id);


--
-- Name: order_payments_order_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX order_payments_order_id_idx ON public.order_payments USING btree (order_id);


--
-- Name: order_shipments_order_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX order_shipments_order_id_idx ON public.order_shipments USING btree (order_id);


--
-- Name: order_status_history_order_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX order_status_history_order_id_idx ON public.order_status_history USING btree (order_id);


--
-- Name: orders_status_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX orders_status_idx ON public.orders USING btree (status);


--
-- Name: orders_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX orders_user_id_idx ON public.orders USING btree (user_id);


--
-- Name: orders_user_status_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX orders_user_status_idx ON public.orders USING btree (user_id, status);


--
-- Name: product_images_product_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX product_images_product_id_idx ON public.product_images USING btree (product_id);


--
-- Name: product_reviews_product_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX product_reviews_product_id_idx ON public.product_reviews USING btree (product_id);


--
-- Name: product_reviews_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX product_reviews_user_id_idx ON public.product_reviews USING btree (user_id);


--
-- Name: product_specifications_product_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX product_specifications_product_id_idx ON public.product_specifications USING btree (product_id);


--
-- Name: product_variants_product_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX product_variants_product_id_idx ON public.product_variants USING btree (product_id);


--
-- Name: products_category_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX products_category_id_idx ON public.products USING btree (category_id);


--
-- Name: products_category_name_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX products_category_name_idx ON public.products USING btree (category_id, name);


--
-- Name: products_name_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX products_name_idx ON public.products USING btree (name);


--
-- Name: refresh_tokens_expires_at_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX refresh_tokens_expires_at_idx ON public.refresh_tokens USING btree (expires_at);


--
-- Name: refresh_tokens_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX refresh_tokens_user_id_idx ON public.refresh_tokens USING btree (user_id);


--
-- Name: roles_created_at_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX roles_created_at_idx ON public.roles USING btree (created_at);


--
-- Name: shipment_order_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX shipment_order_id_idx ON public.shipment USING btree (order_id);


--
-- Name: shopping_carts_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX shopping_carts_user_id_idx ON public.shopping_carts USING btree (user_id);


--
-- Name: unique_pending_request; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE UNIQUE INDEX unique_pending_request ON public.admin_requests USING btree (user_id) WHERE ((status)::text = 'PENDING'::text);


--
-- Name: user_addresses_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX user_addresses_user_id_idx ON public.user_addresses USING btree (user_id);


--
-- Name: user_roles_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX user_roles_user_id_idx ON public.user_roles USING btree (user_id);


--
-- Name: users_created_at_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX users_created_at_idx ON public.users USING btree (created_at);


--
-- Name: users_email_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE UNIQUE INDEX users_email_idx ON public.users USING btree (email);


--
-- Name: wishlist_items_wishlist_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX wishlist_items_wishlist_id_idx ON public.wishlist_items USING btree (wishlist_id);


--
-- Name: wishlists_user_id_idx; Type: INDEX; Schema: public; Owner: geraldbahati
--

CREATE INDEX wishlists_user_id_idx ON public.wishlists USING btree (user_id);


--
-- Name: cart_items cart_items_after_insert_update_delete; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER cart_items_after_insert_update_delete AFTER INSERT OR DELETE OR UPDATE ON public.cart_items FOR EACH ROW EXECUTE FUNCTION public.update_cart_totals();


--
-- Name: cart_items cart_items_before_insert; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER cart_items_before_insert BEFORE INSERT ON public.cart_items FOR EACH ROW EXECUTE FUNCTION public.set_cart_item_price();


--
-- Name: guest_checkouts format_phone_number_guest_checkouts; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER format_phone_number_guest_checkouts BEFORE INSERT OR UPDATE ON public.guest_checkouts FOR EACH ROW EXECUTE FUNCTION public.format_phone_number();


--
-- Name: users format_phone_number_users; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER format_phone_number_users BEFORE INSERT OR UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.format_phone_number();


--
-- Name: promotions generate_slug_trigger; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER generate_slug_trigger BEFORE INSERT ON public.promotions FOR EACH ROW EXECUTE FUNCTION public.generate_promotion_slug();


--
-- Name: orders set_order_number; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER set_order_number BEFORE INSERT ON public.orders FOR EACH ROW EXECUTE FUNCTION public.generate_order_number();


--
-- Name: product_images set_position_before_insert; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER set_position_before_insert BEFORE INSERT ON public.product_images FOR EACH ROW EXECUTE FUNCTION public.set_default_position();


--
-- Name: products set_slug_before_insert_or_update; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER set_slug_before_insert_or_update BEFORE INSERT OR UPDATE ON public.products FOR EACH ROW WHEN (((new.slug IS NULL) OR ((new.slug)::text = ''::text))) EXECUTE FUNCTION public.generate_slug();


--
-- Name: categories trg_set_category_position; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trg_set_category_position BEFORE INSERT ON public.categories FOR EACH ROW EXECUTE FUNCTION public.set_category_position();


--
-- Name: categories trg_update_category_path; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trg_update_category_path BEFORE INSERT OR UPDATE ON public.categories FOR EACH ROW EXECUTE FUNCTION public.update_category_path();


--
-- Name: categories trg_update_category_slug; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trg_update_category_slug BEFORE INSERT OR UPDATE ON public.categories FOR EACH ROW EXECUTE FUNCTION public.update_category_slug();


--
-- Name: products trigger_set_valid_to; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_set_valid_to BEFORE INSERT OR UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.set_valid_to_based_on_category();


--
-- Name: product_attribute_values trigger_update_product_attribute_values_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_update_product_attribute_values_updated_at BEFORE UPDATE ON public.product_attribute_values FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: product_attributes trigger_update_product_attributes_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_update_product_attributes_updated_at BEFORE UPDATE ON public.product_attributes FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: product_to_attribute_values trigger_update_product_to_attribute_values_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_update_product_to_attribute_values_updated_at BEFORE UPDATE ON public.product_to_attribute_values FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: promotion_products trigger_update_promotion_products_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_update_promotion_products_updated_at BEFORE UPDATE ON public.promotion_products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: promotions trigger_update_promotions_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_update_promotions_updated_at BEFORE UPDATE ON public.promotions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: products trigger_update_search_keyword; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER trigger_update_search_keyword BEFORE INSERT OR UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_search_keyword();


--
-- Name: cart_items update_cart_item_timestamp; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_cart_item_timestamp BEFORE UPDATE ON public.cart_items FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: order_items update_order_item_timestamp; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_order_item_timestamp BEFORE UPDATE ON public.order_items FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: order_payments update_order_payments_timestamp; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_order_payments_timestamp BEFORE UPDATE ON public.order_payments FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: order_shipments update_order_shipments_timestamp; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_order_shipments_timestamp BEFORE UPDATE ON public.order_shipments FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: orders update_order_timestamp; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_order_timestamp BEFORE UPDATE ON public.orders FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: roles update_roles_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON public.roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: user_roles update_user_roles_updated_at; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE ON public.user_roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: wishlist_items update_wishlist_item_timestamp; Type: TRIGGER; Schema: public; Owner: geraldbahati
--

CREATE TRIGGER update_wishlist_item_timestamp BEFORE UPDATE ON public.wishlist_items FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: admin_approval_tokens admin_approval_tokens_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.admin_approval_tokens
    ADD CONSTRAINT admin_approval_tokens_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.admin_requests(id) ON DELETE CASCADE;


--
-- Name: admin_requests admin_requests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.admin_requests
    ADD CONSTRAINT admin_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_shopping_cart_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_shopping_cart_id_fkey FOREIGN KEY (shopping_cart_id) REFERENCES public.shopping_carts(id) ON DELETE CASCADE;


--
-- Name: categories categories_last_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_last_updated_by_fkey FOREIGN KEY (last_updated_by) REFERENCES public.users(id);


--
-- Name: categories categories_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.categories(id) ON DELETE SET NULL;


--
-- Name: discounts discounts_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: order_payments fk_order; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_payments
    ADD CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_payments fk_payment_method; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_payments
    ADD CONSTRAINT fk_payment_method FOREIGN KEY (payment_method_id) REFERENCES public.payment_methods(id);


--
-- Name: order_payments fk_payment_status; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_payments
    ADD CONSTRAINT fk_payment_status FOREIGN KEY (payment_status_id) REFERENCES public.payment_statuses(id);


--
-- Name: verification_tokens fk_user_email; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT fk_user_email FOREIGN KEY (email) REFERENCES public.users(email) ON DELETE CASCADE;


--
-- Name: order_item_options order_item_options_order_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_item_options
    ADD CONSTRAINT order_item_options_order_item_id_fkey FOREIGN KEY (order_item_id) REFERENCES public.order_items(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE SET NULL;


--
-- Name: order_payments order_payments_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_payments
    ADD CONSTRAINT order_payments_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_shipments order_shipments_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_shipments
    ADD CONSTRAINT order_shipments_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_status_history order_status_history_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: orders orders_guest_checkout_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_guest_checkout_id_fkey FOREIGN KEY (guest_checkout_id) REFERENCES public.guest_checkouts(id) ON DELETE CASCADE;


--
-- Name: orders orders_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: password_reset_tokens password_reset_tokens_email_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_email_fkey FOREIGN KEY (email) REFERENCES public.users(email) ON DELETE CASCADE;


--
-- Name: product_attribute_values product_attribute_values_attribute_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attribute_values
    ADD CONSTRAINT product_attribute_values_attribute_id_fkey FOREIGN KEY (attribute_id) REFERENCES public.product_attributes(id) ON DELETE CASCADE;


--
-- Name: product_attribute_values product_attribute_values_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attribute_values
    ADD CONSTRAINT product_attribute_values_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE;


--
-- Name: product_attributes product_attributes_attribute_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_attributes
    ADD CONSTRAINT product_attributes_attribute_type_id_fkey FOREIGN KEY (attribute_type_id) REFERENCES public.attribute_types(id) ON DELETE CASCADE;


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_interactions product_interactions_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_interactions
    ADD CONSTRAINT product_interactions_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_interactions product_interactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_interactions
    ADD CONSTRAINT product_interactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: product_option_values product_option_values_option_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_option_values
    ADD CONSTRAINT product_option_values_option_id_fkey FOREIGN KEY (option_id) REFERENCES public.product_options(id) ON DELETE CASCADE;


--
-- Name: product_options product_options_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_options
    ADD CONSTRAINT product_options_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_reviews product_reviews_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT product_reviews_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_reviews product_reviews_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT product_reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: product_specifications product_specifications_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_specifications
    ADD CONSTRAINT product_specifications_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_to_attribute_values product_to_attribute_values_attribute_value_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_to_attribute_values
    ADD CONSTRAINT product_to_attribute_values_attribute_value_id_fkey FOREIGN KEY (attribute_value_id) REFERENCES public.product_attribute_values(id) ON DELETE CASCADE;


--
-- Name: product_to_attribute_values product_to_attribute_values_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_to_attribute_values
    ADD CONSTRAINT product_to_attribute_values_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_variants product_variants_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT product_variants_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: products products_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE SET NULL;


--
-- Name: promotion_products promotion_products_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.promotion_products
    ADD CONSTRAINT promotion_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: promotion_products promotion_products_promotion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.promotion_products
    ADD CONSTRAINT promotion_products_promotion_id_fkey FOREIGN KEY (promotion_id) REFERENCES public.promotions(id) ON DELETE CASCADE;


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: related_products related_products_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.related_products
    ADD CONSTRAINT related_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: related_products related_products_related_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.related_products
    ADD CONSTRAINT related_products_related_product_id_fkey FOREIGN KEY (related_product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: shipment shipment_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.shipment
    ADD CONSTRAINT shipment_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: shopping_carts shopping_carts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.shopping_carts
    ADD CONSTRAINT shopping_carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_addresses user_addresses_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: wishlist_items wishlist_items_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.wishlist_items
    ADD CONSTRAINT wishlist_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: wishlist_items wishlist_items_wishlist_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.wishlist_items
    ADD CONSTRAINT wishlist_items_wishlist_id_fkey FOREIGN KEY (wishlist_id) REFERENCES public.wishlists(id) ON DELETE CASCADE;


--
-- Name: wishlists wishlists_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: geraldbahati
--

ALTER TABLE ONLY public.wishlists
    ADD CONSTRAINT wishlists_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: category_hierarchy_mv; Type: MATERIALIZED VIEW DATA; Schema: public; Owner: geraldbahati
--

REFRESH MATERIALIZED VIEW public.category_hierarchy_mv;


--
-- Name: rate_mv; Type: MATERIALIZED VIEW DATA; Schema: public; Owner: geraldbahati
--

REFRESH MATERIALIZED VIEW public.rate_mv;


--
-- Name: recommendations; Type: MATERIALIZED VIEW DATA; Schema: public; Owner: geraldbahati
--

REFRESH MATERIALIZED VIEW public.recommendations;


--
-- Name: user_preferences; Type: MATERIALIZED VIEW DATA; Schema: public; Owner: geraldbahati
--

REFRESH MATERIALIZED VIEW public.user_preferences;


--
-- PostgreSQL database dump complete
--

