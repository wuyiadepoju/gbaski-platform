--
-- PostgreSQL database dump
--

-- Dumped from database version 14.17 (Ubuntu 14.17-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 17.4 (Ubuntu 17.4-1.pgdg24.04+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: backup_deleted_record(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.backup_deleted_record() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Insert the old record into the deleted table
    INSERT INTO deleted (table_name, deleted_data, deleted_by)
    VALUES (
        TG_TABLE_NAME,
        row_to_json(OLD)::jsonb,
        OLD.created_by  -- Use the created_by field from the original record
    );
    RETURN OLD;
END;
$$;


ALTER FUNCTION public.backup_deleted_record() OWNER TO postgres;

--
-- Name: calculate_content_relevance(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.calculate_content_relevance(event_name text, event_description text) RETURNS numeric
    LANGUAGE plpgsql
    AS $$
DECLARE
    name_tokens TEXT[];
    desc_tokens TEXT[];
    matching_tokens INT;
    name_length INT;
    desc_length INT;
    similarity_score DECIMAL;
    keyword_density DECIMAL;
    final_weight DECIMAL;
BEGIN
    -- Normalize and tokenize name and description
    event_name := normalize_query(event_name);
    event_description := normalize_query(event_description);
    
    -- Split into tokens (words)
    name_tokens := regexp_split_to_array(event_name, '\s+');
    desc_tokens := regexp_split_to_array(event_description, '\s+');
    
    -- Calculate basic metrics
    name_length := array_length(name_tokens, 1);
    desc_length := array_length(desc_tokens, 1);
    
    -- Calculate text similarity using built-in similarity function
    similarity_score := similarity(event_name, event_description);
    
    -- Calculate keyword density (how many times name words appear in description)
    SELECT COUNT(*)
    INTO matching_tokens
    FROM unnest(name_tokens) nt
    WHERE EXISTS (
        SELECT 1 
        FROM unnest(desc_tokens) dt 
        WHERE dt = nt
    );
    
    -- Calculate keyword density as a ratio
    keyword_density := CASE 
        WHEN desc_length > 0 THEN 
            matching_tokens::DECIMAL / desc_length::DECIMAL
        ELSE 0 
    END;
    
    -- Calculate final weight based on multiple factors:
    -- 1. Text similarity (30% weight)
    -- 2. Keyword density penalty if too high (40% weight)
    -- 3. Length ratio penalty if description is too short (30% weight)
    final_weight := 
        (similarity_score * 0.3) +
        (CASE 
            WHEN keyword_density > 0.3 THEN 0.2  -- Penalty for high keyword density
            WHEN keyword_density > 0.2 THEN 0.3
            WHEN keyword_density > 0.1 THEN 0.4
            ELSE 0.5  -- Reward for natural keyword distribution
        END) +
        (CASE 
            WHEN desc_length < name_length THEN 0.1  -- Penalty for very short descriptions
            WHEN desc_length < name_length * 2 THEN 0.2
            WHEN desc_length < name_length * 3 THEN 0.3
            ELSE 0.4  -- Reward for substantial descriptions
        END);

    RETURN ROUND(final_weight::DECIMAL, 2);
END;
$$;


ALTER FUNCTION public.calculate_content_relevance(event_name text, event_description text) OWNER TO postgres;

--
-- Name: get_event_features(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_event_features(var_event_id uuid) RETURNS TABLE(name character varying, description text, access_type character varying, form_id uuid, id uuid, tag character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    form_record RECORD;
    result_query text := '';
    first_query boolean := true;
    table_name text;
BEGIN
    FOR form_record IN (
        SELECT f.*, e.name as event_name
        FROM forms f
        JOIN events e ON f.event_id = e.id
        WHERE e.id = VAR_EVENT_ID
    ) LOOP
        -- Add UNION ALL except for the first query
        IF NOT first_query THEN
            result_query := result_query || ' UNION ALL ';
        END IF;

        -- Handle irregular plurals
        CASE form_record.registration_type
            WHEN 'entry' THEN table_name := 'entries';
            WHEN 'ticket' THEN table_name := 'tickets';
            ELSE table_name := form_record.registration_type || 's';
        END CASE;

        -- Build query using the correct plural table name
        result_query := result_query || format(
            'SELECT tbl.name::varchar(20) as name, tbl.description::text as description, %L::varchar(20) as access_type, %L::uuid as form_id, tbl.id::uuid as id, %L::varchar(20) as tag 
             FROM %I tbl 
             WHERE tbl.form_id = %L',
            form_record.access_type,
            form_record.id,
            form_record.registration_type,
            table_name,
            form_record.id
        );

        first_query := false;
    END LOOP;

    -- If we have any results, execute the dynamic query
    IF result_query <> '' THEN
        RETURN QUERY EXECUTE result_query;
    END IF;
END;
$$;


ALTER FUNCTION public.get_event_features(var_event_id uuid) OWNER TO postgres;

--
-- Name: get_next_payout(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_next_payout(form_uuid uuid) RETURNS TABLE(form_id uuid, next_payout_date date, pending_amount numeric)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT
        le.form_id,
        MIN(le.payout_date),
        SUM(le.balance)
    FROM ledger_entries le
    WHERE
        le.form_id = form_uuid
        AND le.entry_type = 'host'
        AND le.payout_status = 'pending'
    GROUP BY le.form_id;
END;
$$;


ALTER FUNCTION public.get_next_payout(form_uuid uuid) OWNER TO postgres;

--
-- Name: normalize_query(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.normalize_query(query text) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN LOWER(TRIM(REGEXP_REPLACE(query, '\s+', ' ', 'g')));
END;
$$;


ALTER FUNCTION public.normalize_query(query text) OWNER TO postgres;

--
-- Name: refresh_event_rankings(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.refresh_event_rankings() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY event_rankings;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.refresh_event_rankings() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: banks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.banks (
    id integer NOT NULL,
    bank_name character varying(255) NOT NULL,
    nip_code character varying(50) NOT NULL,
    cbn_code character varying(50),
    country character varying(5) DEFAULT 'NG'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.banks OWNER TO postgres;

--
-- Name: banks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.banks_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.banks_id_seq OWNER TO postgres;

--
-- Name: banks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.banks_id_seq OWNED BY public.banks.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.categories (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    public boolean DEFAULT false NOT NULL
);


ALTER TABLE public.categories OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.categories_id_seq OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: deleted; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.deleted (
    id integer NOT NULL,
    table_name character varying(255),
    deleted_data jsonb,
    deleted_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_by uuid
);


ALTER TABLE public.deleted OWNER TO postgres;

--
-- Name: deleted_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.deleted_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.deleted_id_seq OWNER TO postgres;

--
-- Name: deleted_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.deleted_id_seq OWNED BY public.deleted.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.events (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    description text NOT NULL,
    slug character varying(255) NOT NULL,
    category_id integer NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    mode jsonb NOT NULL,
    form_id uuid,
    image_url character varying(255),
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT events_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'published'::character varying, 'ended'::character varying])::text[])))
);


ALTER TABLE public.events OWNER TO postgres;

--
-- Name: forms; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.forms (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    event_id uuid,
    name character varying(255) NOT NULL,
    registration_type character varying(20) NOT NULL,
    schema jsonb,
    configs jsonb,
    fee jsonb,
    access_type character varying(20) NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT forms_access_type_check CHECK (((access_type)::text = ANY ((ARRAY['public'::character varying, 'exclusive'::character varying])::text[]))),
    CONSTRAINT forms_dates_check CHECK ((end_date > start_date)),
    CONSTRAINT forms_registration_type_check CHECK (((registration_type)::text = ANY ((ARRAY['ticket'::character varying, 'entry'::character varying, 'contest'::character varying, 'vote'::character varying, 'game'::character varying, 'task'::character varying, 'trivia'::character varying, 'tussle'::character varying, 'stealsplit'::character varying])::text[])))
);


ALTER TABLE public.forms OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255),
    phone character varying(15),
    name character varying(255) NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255),
    picture character varying(255),
    account_type character varying(20) DEFAULT 'participant'::character varying NOT NULL,
    account_status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_account_status_check CHECK (((account_status)::text = ANY ((ARRAY['active'::character varying, 'unverified'::character varying, 'suspended'::character varying, 'deleted'::character varying])::text[]))),
    CONSTRAINT users_account_type_check CHECK (((account_type)::text = ANY ((ARRAY['participant'::character varying, 'organizer'::character varying])::text[])))
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: event_rankings; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--

CREATE MATERIALIZED VIEW public.event_rankings AS
 WITH event_weights AS (
         SELECT e.id,
            e.name,
            e.description,
            e.slug,
            e.category_id,
            (e.mode ->> 'location'::text) AS location,
            e.image_url,
            e.start_date,
            e.end_date,
            u.name AS brand_name,
            e.created_at,
            e.created_by,
            c.name AS category_name,
            c.description AS category_description,
            (e.mode ->> 'description'::text) AS mode_description,
                CASE
                    WHEN ((e.mode ->> 'type'::text) = 'physical'::text) THEN 'physical, venue-event, onsite'::text
                    ELSE 'virtual, online-event, remote'::text
                END AS mode_tag,
                CASE
                    WHEN ((f.fee ->> 'enabled'::text) = 'true'::text) THEN 'paid, ticket, registration, paid-event'::text
                    ELSE 'free, registration, free-event'::text
                END AS fee_tag,
                CASE
                    WHEN ((f.fee ->> 'enabled'::text) = 'true'::text) THEN ((f.fee ->> 'value'::text))::numeric
                    ELSE (0)::numeric
                END AS price,
            ((((((((1.0)::double precision + (random() * (0.3)::double precision)) + (public.calculate_content_relevance((e.name)::text, e.description))::double precision) + (
                CASE
                    WHEN ((f.fee ->> 'enabled'::text) = 'true'::text) THEN 0.3
                    ELSE (0)::numeric
                END)::double precision) + (
                CASE
                    WHEN (e.start_date <= (now() + '24:00:00'::interval)) THEN 0.4
                    WHEN (e.start_date <= (now() + '3 days'::interval)) THEN 0.3
                    WHEN (e.start_date <= (now() + '7 days'::interval)) THEN 0.2
                    WHEN (e.start_date <= (now() + '14 days'::interval)) THEN 0.1
                    ELSE (0)::numeric
                END)::double precision) + (
                CASE
                    WHEN (e.created_at >= (now() - '24:00:00'::interval)) THEN 0.3
                    WHEN (e.created_at >= (now() - '3 days'::interval)) THEN 0.2
                    WHEN (e.created_at >= (now() - '7 days'::interval)) THEN 0.1
                    ELSE (0)::numeric
                END)::double precision) + (
                CASE
                    WHEN (e.image_url IS NOT NULL) THEN 0.1
                    ELSE (0)::numeric
                END)::double precision) + (
                CASE
                    WHEN (length(e.description) > 100) THEN 0.1
                    ELSE (0)::numeric
                END)::double precision) AS weight,
            row_number() OVER (PARTITION BY e.created_by ORDER BY
                CASE
                    WHEN ((f.fee ->> 'enabled'::text) = 'true'::text) THEN 0
                    ELSE 1
                END, e.start_date, e.created_at DESC) AS organizer_event_rank
           FROM (((public.events e
             LEFT JOIN public.forms f ON ((e.id = f.event_id)))
             LEFT JOIN public.categories c ON ((e.category_id = c.id)))
             JOIN public.users u ON ((e.created_by = u.id)))
          WHERE (((e.status)::text = 'published'::text) AND (e.end_date > now()))
        ), distributed_weights AS (
         SELECT event_weights.id,
            event_weights.name,
            event_weights.description,
            event_weights.slug,
            event_weights.category_id,
            event_weights.location,
            event_weights.image_url,
            event_weights.start_date,
            event_weights.end_date,
            event_weights.brand_name,
            event_weights.created_at,
            event_weights.created_by,
            event_weights.category_name,
            event_weights.category_description,
            event_weights.mode_description,
            event_weights.mode_tag,
            event_weights.fee_tag,
            event_weights.price,
            event_weights.weight,
            event_weights.organizer_event_rank,
            ceil(((event_weights.organizer_event_rank)::double precision / (2)::double precision)) AS distribution_group
           FROM event_weights
        )
 SELECT distributed_weights.id,
    distributed_weights.name,
    distributed_weights.description,
    distributed_weights.slug,
    distributed_weights.category_id,
    distributed_weights.location,
    COALESCE(NULLIF((distributed_weights.image_url)::text, ''::text), 'https://image.gbaski.app/gbaski/event-image.webp'::text) AS image_url,
    distributed_weights.start_date,
    distributed_weights.end_date,
    date(distributed_weights.start_date) AS date,
    distributed_weights.brand_name,
    distributed_weights.created_at,
    distributed_weights.created_by,
    distributed_weights.category_name,
    distributed_weights.category_description,
    distributed_weights.mode_description,
    distributed_weights.mode_tag,
    distributed_weights.fee_tag,
    distributed_weights.price,
    distributed_weights.weight,
    distributed_weights.organizer_event_rank,
    distributed_weights.distribution_group,
    row_number() OVER (ORDER BY distributed_weights.distribution_group, distributed_weights.weight DESC, distributed_weights.start_date, distributed_weights.created_at DESC) AS rank,
    ((((((((setweight(to_tsvector('english'::regconfig, (COALESCE(distributed_weights.name, ''::character varying))::text), 'A'::"char") || setweight(to_tsvector('english'::regconfig, COALESCE(distributed_weights.description, ''::text)), 'B'::"char")) || setweight(to_tsvector('english'::regconfig, (COALESCE(distributed_weights.category_name, ''::character varying))::text), 'C'::"char")) || setweight(to_tsvector('english'::regconfig, COALESCE(distributed_weights.category_description, ''::text)), 'C'::"char")) || setweight(to_tsvector('english'::regconfig, COALESCE(distributed_weights.location, ''::text)), 'B'::"char")) || setweight(to_tsvector('english'::regconfig, COALESCE(distributed_weights.mode_description, ''::text)), 'B'::"char")) || setweight(to_tsvector('english'::regconfig, COALESCE(distributed_weights.mode_tag, ''::text)), 'B'::"char")) || setweight(to_tsvector('english'::regconfig, COALESCE(distributed_weights.fee_tag, ''::text)), 'B'::"char")) || setweight(to_tsvector('english'::regconfig, (COALESCE(distributed_weights.brand_name, ''::character varying))::text), 'B'::"char")) AS search_vector
   FROM distributed_weights
  WITH NO DATA;


ALTER MATERIALIZED VIEW public.event_rankings OWNER TO postgres;

--
-- Name: ledger_entries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ledger_entries (
    id integer NOT NULL,
    form_id uuid,
    tx_ref character varying(255),
    entry_desc text,
    entry_type character varying(20) NOT NULL,
    credit numeric(15,2) DEFAULT 0.00 NOT NULL,
    debit numeric(15,2) DEFAULT 0.00 NOT NULL,
    balance numeric(15,2) GENERATED ALWAYS AS ((credit - debit)) STORED,
    payout_status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    payout_date date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ledger_entries_entry_type_check CHECK (((entry_type)::text = ANY ((ARRAY['host'::character varying, 'platform'::character varying])::text[]))),
    CONSTRAINT ledger_entries_payout_status_check CHECK (((payout_status)::text = ANY ((ARRAY['pending'::character varying, 'processed'::character varying, 'reversed'::character varying])::text[])))
);


ALTER TABLE public.ledger_entries OWNER TO postgres;

--
-- Name: ledger_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ledger_entries_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ledger_entries_id_seq OWNER TO postgres;

--
-- Name: ledger_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ledger_entries_id_seq OWNED BY public.ledger_entries.id;


--
-- Name: payout_accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payout_accounts (
    id integer NOT NULL,
    bank_id integer NOT NULL,
    account_number character varying(20) NOT NULL,
    beneficiary_ref character varying(255),
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.payout_accounts OWNER TO postgres;

--
-- Name: payout_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payout_accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payout_accounts_id_seq OWNER TO postgres;

--
-- Name: payout_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payout_accounts_id_seq OWNED BY public.payout_accounts.id;


--
-- Name: recommended_events; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.recommended_events AS
 SELECT event_rankings.id,
    event_rankings.name,
    event_rankings.slug,
    event_rankings.category_id,
    event_rankings.location,
    event_rankings.image_url,
    event_rankings.start_date,
    event_rankings.brand_name,
    event_rankings.price,
    event_rankings.weight AS relevance_score,
    event_rankings.rank AS popularity_rank
   FROM public.event_rankings
  WHERE (event_rankings.start_date > now())
  ORDER BY event_rankings.weight DESC, event_rankings.start_date;


ALTER VIEW public.recommended_events OWNER TO postgres;

--
-- Name: refunds; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.refunds (
    id integer NOT NULL,
    registration_id integer NOT NULL,
    amount numeric(15,2) NOT NULL,
    reason text,
    status character varying(20) DEFAULT 'requested'::character varying NOT NULL,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT refunds_status_check CHECK (((status)::text = ANY ((ARRAY['requested'::character varying, 'processed'::character varying, 'failed'::character varying])::text[])))
);


ALTER TABLE public.refunds OWNER TO postgres;

--
-- Name: refunds_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.refunds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.refunds_id_seq OWNER TO postgres;

--
-- Name: refunds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.refunds_id_seq OWNED BY public.refunds.id;


--
-- Name: registrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.registrations (
    id integer NOT NULL,
    form_id uuid NOT NULL,
    buyer_id uuid NOT NULL,
    user_id uuid,
    transaction_id bigint,
    reg_items jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    reg_code character varying(255) NOT NULL,
    reg_status character varying(255) DEFAULT 'registered'::character varying NOT NULL,
    reg_desc text,
    CONSTRAINT registrations_status_check CHECK (((reg_status)::text = ANY ((ARRAY['registered'::character varying, 'checkedin'::character varying, 'cancelled'::character varying])::text[])))
);


ALTER TABLE public.registrations OWNER TO postgres;

--
-- Name: registrations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.registrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.registrations_id_seq OWNER TO postgres;

--
-- Name: registrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.registrations_id_seq OWNED BY public.registrations.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO postgres;

--
-- Name: transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transactions (
    id integer NOT NULL,
    ref character varying(255) NOT NULL,
    tx_id character varying(255),
    tx_pro character varying(100),
    tx_cur character varying(3) NOT NULL,
    qty integer DEFAULT 1 NOT NULL,
    amount numeric(15,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    tx_date timestamp with time zone
);


ALTER TABLE public.transactions OWNER TO postgres;

--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.transactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.transactions_id_seq OWNER TO postgres;

--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- Name: banks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.banks ALTER COLUMN id SET DEFAULT nextval('public.banks_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: deleted id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deleted ALTER COLUMN id SET DEFAULT nextval('public.deleted_id_seq'::regclass);


--
-- Name: ledger_entries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ledger_entries ALTER COLUMN id SET DEFAULT nextval('public.ledger_entries_id_seq'::regclass);


--
-- Name: payout_accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payout_accounts ALTER COLUMN id SET DEFAULT nextval('public.payout_accounts_id_seq'::regclass);


--
-- Name: refunds id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refunds ALTER COLUMN id SET DEFAULT nextval('public.refunds_id_seq'::regclass);


--
-- Name: registrations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registrations ALTER COLUMN id SET DEFAULT nextval('public.registrations_id_seq'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- Data for Name: banks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.banks (id, bank_name, nip_code, cbn_code, country, created_at, updated_at) FROM stdin;
1	Sterling Bank	000001	232	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
2	Keystone Bank	000002	082	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
3	First City Monument Bank	000003	214	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
4	United Bank for Africa	000004	033	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
5	Diamond Bank	000005	063	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
6	JAIZ Bank	000006	301	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
7	Fidelity Bank	000007	070	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
8	Polaris Bank	000008	076	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
9	Citibank Nigeria	000009	023	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
10	Ecobank Bank	000010	050	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
11	Unity Bank	000011	215	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
12	Stanbic IBTC Bank	000012	221	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
13	Guaranty Trust Bank	000013	058	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
14	Access Bank	000014	044	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
15	Zenith Bank	000015	057	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
16	First Bank of Nigeria	000016	011	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
17	Wema Bank	000017	035	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
18	Union Bank	000018	032	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
19	Standard Chartered Bank	000021	068	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
20	Suntrust Bank	000022	100	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
21	Providus Bank	000023	101	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
22	Rand Merchant Bank	000024	502	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
23	Titan Trust Bank	000025	102	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
24	Taj Bank	000026	302	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
25	Globus Bank	000027	00103	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
26	Lotus Bank	000029	303	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
27	Premium Trust Bank	000031	105	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
28	Signature Bank	000034	106	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
29	Optimus Bank	000036	107	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
30	SAGE GREY FINANCE LIMITED	050003	40165	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
31	Branch International Financial Services	050006	FC40163	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
32	FSDH Merchant Bank	400001	501	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
33	Coronation Merchant Bank	060001	559	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
34	Nova Merchant Bank	060003	561	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
35	Greenwich Merchant Bank	060004	562	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
36	Abbey Mortgage Bank	070010	404	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
37	NPF MicroFinance Bank	070001	50629	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
38	Gateway Mortgage Bank	070009	812	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
39	Refuge Mortgage Bank	070011	90067	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
40	Lagos Building Investment Company	070012	90052	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
41	Platinum Mortgage Bank	070013	268	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
42	Infinity Trust Mortgage Bank	070016	070016	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
43	ASO Savings & Loans	090001	401	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
44	9mobile 9Payment Service Bank	120001	120001	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
45	HopePSB	120002	120002	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
46	MTN Momo PSB	120003	120003	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
47	Airtel Smartcash PSB	120004	120004	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
48	Moniepoint MFB	50515	50515	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
49	Kuda Bank	50211	90267	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
50	VFD Microfinance Bank	566	566	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
51	Carbon	565	565	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
52	Rubies MFB	125	125	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
53	Sparkle Microfinance Bank	51310	90325	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
54	Parallex Bank	104	104	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
55	PalmPay	100033	999991	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
56	Opay Digital Services	100004	999992	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
57	Paga	100002	100002	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
58	Fairmoney Microfinance Bank	51318	51318	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
59	ALAT by WEMA	035A	035150103	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
60	CEMCS Microfinance Bank	50823	90154	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
61	GoMoney	100022	100022	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
62	Imperial Homes Mortgage Bank	100024	415	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
63	AG Mortgage Bank	100028	90077	NG	2025-05-11 10:36:56.585296+01	2025-05-11 10:36:56.585296+01
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.categories (id, name, description, public) FROM stdin;
1	Concerts and Shows	Music concerts, comedy shows, talent showcases, and live performances that bring people together for unforgettable entertainment.	t
2	Competitions and Contests	Beauty pageants, talent competitions, voting contests, and gaming tournaments where participants compete and audiences engage.	t
3	Promotional Events	Brand activations, marketing campaigns, and influencer-led events to showcase products and services.	t
4	Corporate Events	Conferences, seminars, workshops, and product launches designed for businesses, professionals, and networking opportunities.	t
5	Sports and Fitness	Football matches, marathons, fitness challenges, and e-sports tournaments that keep participants active and excited.	t
6	Parties and Social Gatherings	Birthday parties, wedding receptions, reunions, and themed parties to celebrate life's special moments.	t
7	Educational Events	Academic fairs, quiz competitions, career expos, and lectures that inspire learning and growth.	t
8	Community Events	Fundraisers, cultural festivals, church programs, and charity events that unite people for a cause.	t
9	Interactive Gaming Events	Trivia games, Tussle games, Steal or Split challenges, and task-based engagements to entertain and involve audiences.	t
10	Other	Custom category not listed above	t
\.


--
-- Data for Name: deleted; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.deleted (id, table_name, deleted_data, deleted_at, deleted_by) FROM stdin;
\.


--
-- Data for Name: events; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.events (id, name, description, slug, category_id, status, mode, form_id, image_url, start_date, end_date, created_at, updated_at, created_by) FROM stdin;
20550061-7e21-49c8-ab8f-df90c85d426b	aliquam	Rerum similique occaecati magnam dolore quo.	aliquam	6	published	{"type": "physical", "location": "Port Harcourt, Nigeria", "description": "Sapiente ut reprehenderit nesciunt provident deserunt."}	ec8cca80-1b0d-46f1-a7eb-ccbf3c752117	https://images.unsplash.com/photo-1496337589254-7e19d01cec44	2025-07-02 15:02:51.060138+01	2025-07-06 15:02:51.060138+01	2025-03-31 15:02:51.219632+01	2025-03-31 15:02:51.219632+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
02c9fa2d-4759-46eb-b0f4-9b0f98059cd5	repellendus	Ut nisi voluptas autem doloremque modi.	repellendus	3	ended	{"type": "virtual", "location": "https://meet.maarhvh.com", "description": "Enim eaque suscipit dicta dolores et."}	fc8c7112-db74-4d45-beb3-f4c2b43c4fe0	https://image.gbaski.app/brand/repellendus-event.webp	2025-07-01 15:03:06.932601+01	2025-07-03 15:03:06.932601+01	2025-03-31 15:03:07.099824+01	2025-03-31 15:03:07.099824+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
65a639a9-a96b-4ef2-88fe-5208d998422c	cupiditate	Quia nesciunt aperiam a iusto voluptatem.	cupiditate	3	published	{"type": "physical", "location": "Ibadan, Nigeria", "description": "Ipsam et voluptatibus deserunt ut neque."}	97bc69b5-39f2-42d5-b932-6e0f6ba5cfb6	https://images.unsplash.com/photo-1528605248644-14dd04022da1	2025-09-08 15:03:00.788119+01	2025-09-11 15:03:00.788119+01	2025-03-31 15:03:00.949583+01	2025-03-31 15:03:00.949583+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
4e0c92c4-75a1-41d4-8cb5-29c17a95497a	facilis	Ut rem eveniet sit tempora doloremque.	facilis	5	published	{"type": "physical", "location": "Kano, Nigeria", "description": "Laborum consectetur debitis sunt ratione odit."}	069eb4f5-8968-4ef0-b665-826fb575589a	https://images.unsplash.com/photo-1517838277536-f5f99be501cd	2025-08-27 15:03:03.860159+01	2025-08-29 15:03:03.860159+01	2025-03-31 15:03:04.020191+01	2025-03-31 15:03:04.020191+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
abec585d-ebc2-4e53-8a28-b9963177902a	fugiat	Velit nobis ut cupiditate itaque nisi.	fugiat	5	published	{"type": "physical", "location": "Abuja, Nigeria", "description": "Consequatur voluptates ut temporibus deserunt aspernatur."}	04879c0c-fc17-49eb-9d3a-cdc2146168f9	https://image.gbaski.app/brand/fugiat-event.webp	2025-07-13 15:02:54.132195+01	2025-07-18 15:02:54.132195+01	2025-03-31 15:02:54.310534+01	2025-03-31 15:02:54.310534+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
5dfc66f0-2fd8-449a-8c6f-4961d0830c43	aut	Est rem pariatur hic quia laborum.	aut	4	published	{"type": "virtual", "location": "https://meet.mrdqdsp.ru", "description": "Rerum optio voluptatem quos officia et."}	67dc091e-8841-4fcc-88f9-9219ba3ac260	https://images.unsplash.com/photo-1515187029135-18ee286d815b	2025-06-06 15:03:09.8731+01	2025-06-09 15:03:09.8731+01	2025-03-31 15:03:10.05074+01	2025-03-31 15:03:10.05074+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
91bd8cfc-7773-4930-a144-fb4156ff1f58	laboriosam	Officia omnis eos pariatur culpa distinctio.	laboriosam	8	published	{"type": "physical", "location": "Kano, Nigeria", "description": "Delectus quos pariatur et harum error."}	df39e848-a49a-468a-9035-6e9dc2cd1ffd	https://images.unsplash.com/photo-1491438590914-bc09fcaaf77a	2025-07-01 15:03:12.564354+01	2025-07-02 15:03:12.564354+01	2025-03-31 15:03:12.739854+01	2025-03-31 15:03:12.739854+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
4084cfc3-4444-4776-b88c-77d11a7004ae	architecto	Corporis est saepe pariatur animi magni.	architecto	9	published	{"type": "physical", "location": "Ibadan, Nigeria", "description": "Quia quis aut quibusdam soluta sed."}	700a6762-a2e0-478d-afcb-c93e13ce171e	https://image.gbaski.app/brand/architecto-event.webp	2025-07-23 15:02:47.989791+01	2025-07-25 15:02:47.989791+01	2025-03-31 15:02:48.162049+01	2025-03-31 15:02:48.162049+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
c062e01d-bb1f-43f7-a33d-28f0d7cc53f3	consequuntur	Sed eum architecto et ratione iusto.	consequuntur	1	published	{"type": "physical", "location": "Kano, Nigeria", "description": "Eveniet est magni magnam quod quas."}	d2fea592-b287-4c53-9dc8-53ff236c92f1	https://images.unsplash.com/photo-1429962714451-bb934ecdc4ec	2025-04-01 15:03:19.017481+01	2025-04-02 15:03:19.017481+01	2025-03-31 15:03:19.190655+01	2025-03-31 15:03:19.190655+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
b4716531-b598-479b-8a9f-c487e9a87b60	sunt	Voluptate quos consectetur voluptatem esse veritatis.	sunt	7	published	{"type": "virtual", "location": "https://meet.ugqwamw.biz", "description": "Vitae quidem eos qui recusandae maxime."}	027835c0-ffb0-411a-bdf8-145d74600f8a	https://images.unsplash.com/photo-1509062522246-3755977927d7	2025-05-07 15:03:22.394863+01	2025-05-08 15:03:22.394863+01	2025-03-31 15:03:22.559456+01	2025-03-31 15:03:22.559456+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
277b524b-a0a3-443d-8ca0-cbee0d0a22f6	sit	Iure suscipit beatae sunt aliquam voluptatem.	sit	1	published	{"type": "virtual", "location": "https://meet.ucfeigl.top", "description": "Minima ea consequatur soluta sapiente qui."}	38090087-6e96-470f-89e4-d9cbf23bfe22	https://images.unsplash.com/photo-1501281668745-f7f57925c3b4	2025-04-24 15:03:25.569748+01	2025-04-29 15:03:25.569748+01	2025-03-31 15:03:25.745457+01	2025-03-31 15:03:25.745457+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
927454ae-4c76-4ad6-9f4d-ba42136ef82e	soluta	Aliquam dolor ut sunt ducimus eligendi.	soluta	3	ended	{"type": "virtual", "location": "https://meet.hiyscrw.biz", "description": "Ut rerum blanditiis quisquam quo nam."}	799039fd-0092-4c95-b8a5-aff7297f667b	https://image.gbaski.app/brand/soluta-event.webp	2025-09-25 15:03:16.045939+01	2025-09-27 15:03:16.045939+01	2025-03-31 15:03:16.209399+01	2025-03-31 15:03:16.209399+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
7fe7d47f-bf5f-4094-999a-37dfb530ed0f	OLEKU RAVE	<p><span style="color: rgb(31, 36, 47);">Oleku&nbsp;Rave&nbsp;serves&nbsp;as&nbsp;the&nbsp;official&nbsp;After&nbsp;Party&nbsp;for&nbsp;all&nbsp;the&nbsp;Owambe&nbsp;Parties&nbsp;during&nbsp;the&nbsp;month.&nbsp;</span></p><p><span style="color: rgb(31, 36, 47);">It&nbsp;gives&nbsp;a&nbsp;perfect&nbsp;opportunity&nbsp;for&nbsp;people&nbsp;to&nbsp;re-rock&nbsp;the&nbsp;Aso&nbsp;Ebi&nbsp;that&nbsp;they&nbsp;usually&nbsp;wear&nbsp;just&nbsp;once.&nbsp;&nbsp;</span></p><p></p><p><span style="color: rgb(31, 36, 47);">For&nbsp;those&nbsp;who&nbsp;just&nbsp;want&nbsp;to&nbsp;wear&nbsp;their&nbsp;agbada&nbsp;for&nbsp;no&nbsp;reason,&nbsp;Oleku&nbsp;Rave&nbsp;creates&nbsp;the&nbsp;right&nbsp;atmosphere&nbsp;for&nbsp;an&nbsp;incredible&nbsp;Owambe&nbsp;Experience&nbsp;with&nbsp;the&nbsp;right&nbsp;music,&nbsp;mouth&nbsp;watering&nbsp;dishes,&nbsp;fashion&nbsp;and&nbsp;great&nbsp;people.</span></p><p></p><p><span style="color: rgb(31, 36, 47);">You&nbsp;don’t&nbsp;want&nbsp;miss&nbsp;this</span></p><p><span style="color: rgb(31, 36, 47);">Music:&nbsp;DJ&nbsp;&amp;&nbsp;Live&nbsp;Band</span></p><p></p><p><span style="color: rgb(31, 36, 47);">See&nbsp;y’all&nbsp;on&nbsp;Saturday&nbsp;the&nbsp;26th&nbsp;of&nbsp;April.&nbsp;It’s&nbsp;about&nbsp;to&nbsp;be&nbsp;lit.&nbsp;🔥&nbsp;</span></p><p></p><p><span style="color: rgb(31, 36, 47);">Dress&nbsp;Code:&nbsp;Owambe&nbsp;Fits&nbsp;</span></p><p></p><p><span style="color: rgb(31, 36, 47);">Venue:&nbsp;Sinatra&nbsp;Place,&nbsp;16B&nbsp;Ladipo&nbsp;Oluwole&nbsp;St,&nbsp;Off&nbsp;Adeniyi&nbsp;Jones,&nbsp;Ikeja.</span></p><p></p><p><span style="color: rgb(31, 36, 47);">Time:&nbsp;6pm&nbsp;-&nbsp;midnight&nbsp;</span></p><p></p><p><span style="color: rgb(31, 36, 47);">RSVP</span></p><p><span style="color: rgb(31, 36, 47);">09030009826</span></p><p><span style="color: rgb(31, 36, 47);">08066187620</span></p><p></p>	oleku-rave	5	published	{"type": "physical", "location": "Sinatra Place, 16B Ladipo Oluwole St, Off Adeniyi Jones, Ikeja.", "description": "Opposite Chelsea House"}	55cc6890-4178-4b1a-a93c-14d852c6d8ed	https://image.gbaski.app/brand/OLEKU-RAVE-event.webp	2025-04-26 18:00:00+01	2025-04-27 00:00:00+01	2025-04-07 17:34:07.164404+01	2025-04-07 17:34:07.164404+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
b575416b-c51f-4bd5-994c-45a0ff9c21f0	qui	<p><strong class="ql-size-huge" style="color: rgb(31, 36, 47);">A&nbsp;Four-Day&nbsp;Camping&nbsp;Experience&nbsp;for&nbsp;Teens</strong><span class="ql-size-huge" style="color: rgb(31, 36, 47);">&nbsp;⛺✨</span></p><p></p><h3><em style="color: rgb(31, 36, 47);" class="ql-size-large">Camp&nbsp;David</em><span style="color: rgb(31, 36, 47);" class="ql-size-large">&nbsp;is&nbsp;an&nbsp;extraordinary,&nbsp;non</span><span class="ql-size-large">&nbsp;</span><span style="color: rgb(31, 36, 47);" class="ql-size-large">-denominational&nbsp;gathering&nbsp;designed&nbsp;to&nbsp;bring&nbsp;together&nbsp;teenagers&nbsp;from&nbsp;across&nbsp;the&nbsp;globe&nbsp;for&nbsp;a&nbsp;transformative&nbsp;experience&nbsp;like&nbsp;no&nbsp;other.&nbsp;Set&nbsp;in&nbsp;a&nbsp;serene&nbsp;and&nbsp;welcoming&nbsp;environment,&nbsp;Camp&nbsp;David&nbsp;is&nbsp;more&nbsp;than&nbsp;just&nbsp;a&nbsp;retreat—it&#39;s&nbsp;a&nbsp;unique&nbsp;opportunity&nbsp;for&nbsp;young&nbsp;individuals&nbsp;to&nbsp;come&nbsp;together,&nbsp;share&nbsp;their&nbsp;cultures,&nbsp;make&nbsp;lifelong&nbsp;friendships,&nbsp;and&nbsp;embark&nbsp;on&nbsp;a&nbsp;journey&nbsp;of&nbsp;personal&nbsp;growth.</span></h3><p></p><h3><span style="color: rgb(31, 36, 47);" class="ql-size-large">Offering&nbsp;an&nbsp;unparalleled&nbsp;experience&nbsp;of&nbsp;a&nbsp;lifetime.&nbsp;Just&nbsp;like&nbsp;in&nbsp;previous&nbsp;years,&nbsp;the&nbsp;2025&nbsp;edition&nbsp;promises&nbsp;impactful&nbsp;Bible&nbsp;teachings,&nbsp;heartfelt&nbsp;prayer&nbsp;sessions,&nbsp;exciting&nbsp;games,&nbsp;and&nbsp;a&nbsp;myriad&nbsp;of&nbsp;engaging&nbsp;activities.</span></h3><h2></h2><h2><strong style="color: rgb(31, 36, 47);">Camp&nbsp;David&nbsp;runs&nbsp;from&nbsp;Wednesday,&nbsp;July&nbsp;30,&nbsp;to&nbsp;Sunday,&nbsp;August&nbsp;3,&nbsp;2025,&nbsp;and&nbsp;is&nbsp;exclusively&nbsp;for&nbsp;teens&nbsp;aged&nbsp;13&nbsp;to&nbsp;19.</strong></h2>	qui	4	published	{"type": "physical", "location": "Ibadan, Nigeria", "description": "Est rem asperiores qui voluptatibus ut."}	bf296a2c-de4b-4140-9e48-565ac33c88b5	https://image.gbaski.app/brand/qui-event.webp	2025-06-18 15:02:57.818496+01	2025-06-18 17:02:57.818496+01	2025-03-31 15:02:57.978906+01	2025-03-31 15:02:57.978906+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
0b7b42fa-65b2-4643-8223-e186b6405987	eligendi	<p><strong style="color: rgb(31, 36, 47);">A&nbsp;Four-Day&nbsp;Camping&nbsp;Experience&nbsp;for&nbsp;Teens</strong><span style="color: rgb(31, 36, 47);">&nbsp;⛺✨</span></p><p></p><p><em style="color: rgb(31, 36, 47);">Camp&nbsp;David</em><span style="color: rgb(31, 36, 47);">&nbsp;is&nbsp;an&nbsp;extraordinary,&nbsp;non-denominational&nbsp;gathering&nbsp;designed&nbsp;to&nbsp;bring&nbsp;together&nbsp;teenagers&nbsp;from&nbsp;across&nbsp;the&nbsp;globe&nbsp;for&nbsp;a&nbsp;transformative&nbsp;experience&nbsp;like&nbsp;no&nbsp;other.&nbsp;Set&nbsp;in&nbsp;a&nbsp;serene&nbsp;and&nbsp;welcoming&nbsp;environment,&nbsp;Camp&nbsp;David&nbsp;is&nbsp;more&nbsp;than&nbsp;just&nbsp;a&nbsp;retreat—it&#39;s&nbsp;a&nbsp;unique&nbsp;opportunity&nbsp;for&nbsp;young&nbsp;individuals&nbsp;to&nbsp;come&nbsp;together,&nbsp;share&nbsp;their&nbsp;cultures,&nbsp;make&nbsp;lifelong&nbsp;friendships,&nbsp;and&nbsp;embark&nbsp;on&nbsp;a&nbsp;journey&nbsp;of&nbsp;personal&nbsp;growth.</span></p><p></p><p><span style="color: rgb(31, 36, 47);">Offering&nbsp;an&nbsp;unparalleled&nbsp;experience&nbsp;of&nbsp;a&nbsp;lifetime.&nbsp;Just&nbsp;like&nbsp;in&nbsp;previous&nbsp;years,&nbsp;the&nbsp;2025&nbsp;edition&nbsp;promises&nbsp;impactful&nbsp;Bible&nbsp;teachings,&nbsp;heartfelt&nbsp;prayer&nbsp;sessions,&nbsp;exciting&nbsp;games,&nbsp;and&nbsp;a&nbsp;myriad&nbsp;of&nbsp;engaging&nbsp;activities.</span></p><p></p><p><strong style="color: rgb(31, 36, 47);">Camp&nbsp;David&nbsp;runs&nbsp;from&nbsp;Wednesday,&nbsp;July&nbsp;30,&nbsp;to&nbsp;Sunday,&nbsp;August&nbsp;3,&nbsp;2025,&nbsp;and&nbsp;is&nbsp;exclusively&nbsp;for&nbsp;teens&nbsp;aged&nbsp;13&nbsp;to&nbsp;19.</strong></p>	eligendi	9	draft	{"type": "physical", "location": "Abuja, Nigeria", "description": "Aliquid sequi nisi est alias possimus."}	0ac6bd9c-547c-4749-a22b-1b7366befd92	https://image.gbaski.app/brand/eligendi-event.webp	2025-04-11 15:03:28.641809+01	2025-04-15 15:03:28.641809+01	2025-03-31 15:03:28.820354+01	2025-03-31 15:03:28.820354+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
\.


--
-- Data for Name: forms; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.forms (id, event_id, name, registration_type, schema, configs, fee, access_type, start_date, end_date, created_at, updated_at, created_by) FROM stdin;
700a6762-a2e0-478d-afcb-c93e13ce171e	4084cfc3-4444-4776-b88c-77d11a7004ae	architecto	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 25141, "enabled": true}	exclusive	2025-07-23 15:02:47.989791+01	2025-07-25 15:02:47.989791+01	2025-03-31 15:02:48.162049+01	2025-03-31 15:02:48.162049+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
ec8cca80-1b0d-46f1-a7eb-ccbf3c752117	20550061-7e21-49c8-ab8f-df90c85d426b	aliquam	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 66180, "enabled": true}	public	2025-07-02 15:02:51.060138+01	2025-07-06 15:02:51.060138+01	2025-03-31 15:02:51.219632+01	2025-03-31 15:02:51.219632+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
97bc69b5-39f2-42d5-b932-6e0f6ba5cfb6	65a639a9-a96b-4ef2-88fe-5208d998422c	cupiditate	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 13877, "enabled": true}	public	2025-09-08 15:03:00.788119+01	2025-09-11 15:03:00.788119+01	2025-03-31 15:03:00.949583+01	2025-03-31 15:03:00.949583+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
fc8c7112-db74-4d45-beb3-f4c2b43c4fe0	02c9fa2d-4759-46eb-b0f4-9b0f98059cd5	repellendus	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 97151, "enabled": true}	public	2025-07-01 15:03:06.932601+01	2025-07-03 15:03:06.932601+01	2025-03-31 15:03:07.099824+01	2025-03-31 15:03:07.099824+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
67dc091e-8841-4fcc-88f9-9219ba3ac260	5dfc66f0-2fd8-449a-8c6f-4961d0830c43	aut	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 96517, "enabled": true}	public	2025-06-06 15:03:09.8731+01	2025-06-09 15:03:09.8731+01	2025-03-31 15:03:10.05074+01	2025-03-31 15:03:10.05074+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
df39e848-a49a-468a-9035-6e9dc2cd1ffd	91bd8cfc-7773-4930-a144-fb4156ff1f58	laboriosam	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 79693, "enabled": true}	public	2025-07-01 15:03:12.564354+01	2025-07-02 15:03:12.564354+01	2025-03-31 15:03:12.739854+01	2025-03-31 15:03:12.739854+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
799039fd-0092-4c95-b8a5-aff7297f667b	927454ae-4c76-4ad6-9f4d-ba42136ef82e	soluta	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 32614, "enabled": true}	public	2025-09-25 15:03:16.045939+01	2025-09-27 15:03:16.045939+01	2025-03-31 15:03:16.209399+01	2025-03-31 15:03:16.209399+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
d2fea592-b287-4c53-9dc8-53ff236c92f1	c062e01d-bb1f-43f7-a33d-28f0d7cc53f3	consequuntur	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 70668, "enabled": true}	exclusive	2025-04-01 15:03:19.017481+01	2025-04-02 15:03:19.017481+01	2025-03-31 15:03:19.190655+01	2025-03-31 15:03:19.190655+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
027835c0-ffb0-411a-bdf8-145d74600f8a	b4716531-b598-479b-8a9f-c487e9a87b60	sunt	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 32192, "enabled": true}	exclusive	2025-05-07 15:03:22.394863+01	2025-05-08 15:03:22.394863+01	2025-03-31 15:03:22.559456+01	2025-03-31 15:03:22.559456+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
04879c0c-fc17-49eb-9d3a-cdc2146168f9	abec585d-ebc2-4e53-8a28-b9963177902a	fugiat	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 5368, "enabled": true}	public	2025-07-13 15:02:54.132195+01	2025-07-18 15:02:54.132195+01	2025-03-31 15:02:54.310534+01	2025-03-31 15:02:54.310534+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
069eb4f5-8968-4ef0-b665-826fb575589a	4e0c92c4-75a1-41d4-8cb5-29c17a95497a	facilis	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 86829, "enabled": true}	public	2025-08-27 15:03:03.860159+01	2025-08-29 15:03:03.860159+01	2025-03-31 15:03:04.020191+01	2025-03-31 15:03:04.020191+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
0ac6bd9c-547c-4749-a22b-1b7366befd92	0b7b42fa-65b2-4643-8223-e186b6405987	eligendi	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 26038, "enabled": true}	public	2025-04-11 15:03:28.641809+01	2025-04-15 15:03:28.641809+01	2025-03-31 15:03:28.820354+01	2025-03-31 15:03:28.820354+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
38090087-6e96-470f-89e4-d9cbf23bfe22	277b524b-a0a3-443d-8ca0-cbee0d0a22f6	sit	entry	[]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 10855, "enabled": true}	public	2025-04-24 15:03:25.569748+01	2025-04-29 15:03:25.569748+01	2025-03-31 15:03:25.745457+01	2025-03-31 15:03:25.745457+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
55cc6890-4178-4b1a-a93c-14d852c6d8ed	7fe7d47f-bf5f-4094-999a-37dfb530ed0f	Madeson Gutierrez	ticket	[{"fee": null, "type": {"name": "Regular", "description": "Standard ticket at regular price"}, "order": null, "price": 0, "isPaid": true, "capacity": null, "discount": null, "remaining": null}]	{"email": {"enabled": false}, "redirect": {"link": {"label": ""}, "enabled": false}}	{"value": 0, "enabled": true}	public	2025-04-08 00:00:00+01	2025-04-10 00:00:00+01	2025-04-07 17:34:07.164404+01	2025-04-07 17:34:07.164404+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
bf296a2c-de4b-4140-9e48-565ac33c88b5	b575416b-c51f-4bd5-994c-45a0ff9c21f0	qui	ticket	[{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "fee": 0, "sold": null, "type": {"name": "Standard", "description": "Basic event access"}, "limit": 5, "price": 1000, "isPaid": true, "capacity": 74, "discount": 90, "description": "Basic event access"}, {"id": "2ce195b5-a8d2-4c25-9055-75748ad37720", "fee": 0, "sold": null, "type": {"name": "Early Bird", "description": "Discounted tickets available for a limited time"}, "limit": 1, "price": 0, "isPaid": false, "capacity": 10, "discount": 30, "description": "Discounted tickets available for a limited time"}, {"id": "f08f5bac-aac8-400e-9a81-8561af0ac587", "fee": 0, "sold": 5, "type": {"name": "Regular", "description": "Standard ticket at regular price"}, "limit": 5, "price": 9500, "isPaid": true, "capacity": null, "discount": 98, "description": "Standard ticket at regular price"}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "fee": 0, "sold": null, "type": {"name": "Student", "description": "Discounted rates for students with valid ID"}, "limit": 5, "price": 150, "isPaid": true, "capacity": null, "discount": null, "description": "Discounted rates for students with valid ID"}]	{"email": {"content": null, "enabled": false}, "redirect": {"type": "website", "enabled": false}, "isPrimaryForm": false}	{"value": 1000, "enabled": true}	public	2025-06-18 00:02:57.818+01	2025-06-21 00:02:57.818+01	2025-03-31 15:02:57.978906+01	2025-05-04 08:59:46.096708+01	bb99244d-c9fb-40c3-a7b6-96a5aee0d35e
\.


--
-- Data for Name: ledger_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ledger_entries (id, form_id, tx_ref, entry_desc, entry_type, credit, debit, payout_status, payout_date, created_at) FROM stdin;
\.


--
-- Data for Name: payout_accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payout_accounts (id, bank_id, account_number, beneficiary_ref, user_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: refunds; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.refunds (id, registration_id, amount, reason, status, processed_at, created_at) FROM stdin;
\.


--
-- Data for Name: registrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.registrations (id, form_id, buyer_id, user_id, transaction_id, reg_items, created_at, reg_code, reg_status, reg_desc) FROM stdin;
6	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	8	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}], "attendees": []}	2025-05-06 13:56:21.425922+01	DK7550	registered	Ticket payment
7	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	9	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 300, "quantity": 3}, {"id": "f08f5bac-aac8-400e-9a81-8561af0ac587", "name": "Regular", "amount": 380, "quantity": 2}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 150, "quantity": 1}], "attendees": [{"email": "mariam.adepoju@gmail.com", "phone": "07052326821", "lastName": "Adepoju", "ticketId": "1e4d20b1-547e-427d-8256-8c70c981721c", "firstName": "Mariam ", "ticketName": "Standard"}, {"email": "tara.adepoju@gmail.com", "phone": "", "lastName": "Adepoju", "ticketId": "1e4d20b1-547e-427d-8256-8c70c981721c", "firstName": "Omotara", "ticketName": "Standard"}, {"email": "bayo.adepoju@gmail.com", "phone": "", "lastName": "Adepoju", "ticketId": "1e4d20b1-547e-427d-8256-8c70c981721c", "firstName": "Bayo", "ticketName": "Standard"}, {"email": "sola.adepoju@gmail.com", "phone": "07052326821", "lastName": "Adepoju", "ticketId": "f08f5bac-aac8-400e-9a81-8561af0ac587", "firstName": "Sola", "ticketName": "Regular"}, {"email": "wuyi.adepoju@gmail.com", "phone": "08034309999", "lastName": "Adepoju", "ticketId": "f08f5bac-aac8-400e-9a81-8561af0ac587", "firstName": "Wuyi", "ticketName": "Regular"}, {"email": "roju.adepoju@gmail.com", "phone": "", "lastName": "", "ticketId": "33b3f097-90cf-4df2-a816-16b346601ab5", "firstName": "Roju", "ticketName": "Student"}]}	2025-05-06 15:34:19.610055+01	ER3827	registered	Ticket payment
8	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	10	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 400, "quantity": 4}, {"id": "f08f5bac-aac8-400e-9a81-8561af0ac587", "name": "Regular", "amount": 380, "quantity": 2}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 750, "quantity": 5}], "attendees": []}	2025-05-06 16:15:49.962869+01	ES4687	registered	Ticket payment
9	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	11	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 400, "quantity": 4}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 600, "quantity": 4}], "attendees": []}	2025-05-07 05:53:09.241694+01	SO4524	registered	Ticket payment
10	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	12	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 400, "quantity": 4}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 600, "quantity": 4}], "attendees": []}	2025-05-07 05:54:54.592444+01	WM5853	registered	Ticket payment
11	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	13	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 400, "quantity": 4}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 600, "quantity": 4}], "attendees": []}	2025-05-07 05:56:43.952662+01	YK2036	registered	Ticket payment
1	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	1	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 200, "quantity": 2}, {"id": "2ce195b5-a8d2-4c25-9055-75748ad37720", "name": "Early Bird", "amount": 0, "quantity": 1}, {"id": "f08f5bac-aac8-400e-9a81-8561af0ac587", "name": "Regular", "amount": 190, "quantity": 1}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 150, "quantity": 1}], "attendees": []}	2025-04-28 16:29:51.575328+01	DN2001	registered	4 tickets(standard, VIP)
12	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	14	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}], "attendees": []}	2025-05-07 06:19:55.704251+01	IX6066	registered	Ticket payment
13	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	15	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}], "attendees": []}	2025-05-07 06:21:09.983371+01	SG1501	registered	Ticket payment
14	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	16	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}], "attendees": []}	2025-05-07 06:23:23.688417+01	NX3774	registered	Ticket payment
3	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	\N	{}	2025-04-28 16:49:23.995479+01	XQ7329	registered	1 Regular 
4	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	4	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 200, "quantity": 2}, {"id": "2ce195b5-a8d2-4c25-9055-75748ad37720", "name": "Early Bird", "amount": 0, "quantity": 1}, {"id": "f08f5bac-aac8-400e-9a81-8561af0ac587", "name": "Regular", "amount": 190, "quantity": 1}, {"id": "33b3f097-90cf-4df2-a816-16b346601ab5", "name": "Student", "amount": 150, "quantity": 1}], "attendees": []}	2025-05-04 09:34:23.120621+01	TE2943	registered	Ticket payment
2	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	2	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}, {"id": "f08f5bac-aac8-400e-9a81-8561af0ac587", "name": "Regular", "amount": 380, "quantity": 2}], "attendees": [{"email": "johnizeighbe47@gmail.com", "phone": "07052326821", "lastName": "Erhabor", "ticketId": "1e4d20b1-547e-427d-8256-8c70c981721c", "firstName": "Mariam", "ticketName": "Standard"}, {"email": "wuyi.adepoju@gmail.com", "phone": "08034309999", "lastName": "Adepoju", "ticketId": "f08f5bac-aac8-400e-9a81-8561af0ac587", "firstName": "Adewuyi", "ticketName": "Regular"}, {"email": "tara.adepoju@gbaski.app", "phone": "", "lastName": "", "ticketId": "f08f5bac-aac8-400e-9a81-8561af0ac587", "firstName": "Omotara", "ticketName": "Regular"}]}	2025-04-28 16:49:23.995479+01	AG4082	registered	2 tickets(standard, VIP)
15	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	18	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}], "attendees": []}	2025-05-07 08:00:57.955319+01	XU6341	registered	Ticket payment
16	bf296a2c-de4b-4140-9e48-565ac33c88b5	98ee2b18-b76c-402c-870d-7dcb7d02bf21	\N	20	{"orders": [{"id": "1e4d20b1-547e-427d-8256-8c70c981721c", "name": "Standard", "amount": 100, "quantity": 1}], "attendees": []}	2025-05-07 11:02:48.421308+01	ND4817	registered	Ticket payment
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.schema_migrations (version, dirty) FROM stdin;
9	f
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.transactions (id, ref, tx_id, tx_pro, tx_cur, qty, amount, created_at, tx_date) FROM stdin;
1	fc_8ghjo	1792236847	flutterwave	NGN	1	100.00	2025-04-28 16:29:51.575328+01	2025-04-28 16:29:51.575328+01
4	pcbrandbf296dghki	4932009708	paystack	NGN	5	540.00	2025-05-04 09:34:23.120621+01	\N
2	pcbrand_bf296_8sf8j	4914103308	paystack	NGN	3	480.00	2025-04-28 16:49:23.995479+01	2025-04-28 16:49:23.995479+01
8	pcbrandbf296iih09	4939241606	paystack	NGN	1	100.00	2025-05-06 13:56:21.425922+01	\N
9	scbrandbf296m0l3o	scbrandbf296m0l3o_1_18_1	squad	NGN	6	830.00	2025-05-06 15:34:19.610055+01	\N
10	scbrandbf296nhtwt	scbrandbf296nhtwt_1_18_1	squad	NGN	11	1530.00	2025-05-06 16:15:49.962869+01	\N
11	scbrandbf296goy19	scbrandbf296goy19_1_18_1	squad	NGN	8	1000.00	2025-05-07 05:53:09.241694+01	\N
12	pcbrandbf296greao	4940884703	paystack	NGN	8	1000.00	2025-05-07 05:54:54.592444+01	\N
13	fcbrandbf296gt0v9	1797598993	flutterwave	NGN	8	1000.00	2025-05-07 05:56:43.952662+01	\N
15	pcbrandbf296hp8yz	4940943631	paystack	NGN	1	100.00	2025-05-07 06:21:09.983371+01	\N
16	fcbrandbf296hr1t1	1797607547	flutterwave	NGN	1	100.00	2025-05-07 06:23:23.688417+01	\N
14	scbrandbf296hno7q1	scbrandbf296hno7q_1_18_1	squad	NGN	1	100.00	2025-05-07 06:19:55.704251+01	\N
18	scbrandbf296hno7q2	scbrandbf296hno7q_1_18_1		NGN	1	100.00	2025-05-07 08:00:57.955319+01	\N
20	scbrandbf296hno7q	scbrandbf296hno7q_1_18_1		NGN	1	100.00	2025-05-07 11:02:48.421308+01	\N
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, email, password, phone, name, first_name, last_name, picture, account_type, account_status, created_at, updated_at) FROM stdin;
bb99244d-c9fb-40c3-a7b6-96a5aee0d35e	wuyi.adepoju@gmail.com	$2a$10$ZukIiypUS.oOi835Ir4ghO0HoDH6MG5N7hMRk37gcA4ZMQITR2.06	+234943-681-02	brand	brand	\N	\N	organizer	active	2025-03-31 10:49:40.021566+01	2025-03-31 10:49:40.021566+01
98ee2b18-b76c-402c-870d-7dcb7d02bf21	aliu.adepoju@gmail.com		\N	aliu	aliu	adepoju	\N	participant	active	2025-04-16 10:06:27.263749+01	2025-04-16 10:06:27.263749+01
\.


--
-- Name: banks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.banks_id_seq', 63, true);


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.categories_id_seq', 10, true);


--
-- Name: deleted_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.deleted_id_seq', 1, false);


--
-- Name: ledger_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ledger_entries_id_seq', 1, false);


--
-- Name: payout_accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payout_accounts_id_seq', 1, false);


--
-- Name: refunds_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.refunds_id_seq', 1, false);


--
-- Name: registrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.registrations_id_seq', 16, true);


--
-- Name: transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.transactions_id_seq', 20, true);


--
-- Name: banks banks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.banks
    ADD CONSTRAINT banks_pkey PRIMARY KEY (id);


--
-- Name: categories categories_name_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_name_unique UNIQUE (name);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: deleted deleted_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deleted
    ADD CONSTRAINT deleted_pkey PRIMARY KEY (id);


--
-- Name: events events_name_created_by_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_name_created_by_unique UNIQUE (name, created_by);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: forms forms_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forms
    ADD CONSTRAINT forms_pkey PRIMARY KEY (id);


--
-- Name: ledger_entries ledger_entries_form_id_tx_ref_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ledger_entries
    ADD CONSTRAINT ledger_entries_form_id_tx_ref_key UNIQUE (form_id, tx_ref);


--
-- Name: ledger_entries ledger_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ledger_entries
    ADD CONSTRAINT ledger_entries_pkey PRIMARY KEY (id);


--
-- Name: payout_accounts payout_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payout_accounts
    ADD CONSTRAINT payout_accounts_pkey PRIMARY KEY (id);


--
-- Name: refunds refunds_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_pkey PRIMARY KEY (id);


--
-- Name: registrations registrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.registrations
    ADD CONSTRAINT registrations_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_ref_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_ref_unique UNIQUE (ref);


--
-- Name: users unique_name_account_type; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_name_account_type UNIQUE (name, account_type);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: banks_nip_code_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX banks_nip_code_idx ON public.banks USING btree (nip_code);


--
-- Name: idx_banks_cbn_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_banks_cbn_code ON public.banks USING btree (cbn_code);


--
-- Name: idx_banks_country_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_banks_country_code ON public.banks USING btree (country);


--
-- Name: idx_brand_name_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_brand_name_trgm ON public.event_rankings USING gin (lower((brand_name)::text) public.gin_trgm_ops);


--
-- Name: idx_category_description_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_category_description_trgm ON public.event_rankings USING gin (lower(category_description) public.gin_trgm_ops);


--
-- Name: idx_category_name_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_category_name_trgm ON public.event_rankings USING gin (lower((category_name)::text) public.gin_trgm_ops);


--
-- Name: idx_description_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_description_trgm ON public.event_rankings USING gin (lower(description) public.gin_trgm_ops);


--
-- Name: idx_event_rankings_brand_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_brand_name ON public.event_rankings USING btree (lower((brand_name)::text));


--
-- Name: idx_event_rankings_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_category ON public.event_rankings USING btree (category_id);


--
-- Name: idx_event_rankings_date_range; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_date_range ON public.event_rankings USING btree (start_date, end_date);


--
-- Name: idx_event_rankings_end_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_end_date ON public.event_rankings USING btree (end_date);


--
-- Name: idx_event_rankings_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_event_rankings_id ON public.event_rankings USING btree (id);


--
-- Name: idx_event_rankings_price; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_price ON public.event_rankings USING btree (price);


--
-- Name: idx_event_rankings_rank; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_rank ON public.event_rankings USING btree (rank);


--
-- Name: idx_event_rankings_search_vector; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_search_vector ON public.event_rankings USING gin (search_vector);


--
-- Name: idx_event_rankings_start_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_start_date ON public.event_rankings USING btree (start_date);


--
-- Name: idx_event_rankings_weight; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_rankings_weight ON public.event_rankings USING btree (weight DESC);


--
-- Name: idx_events_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_created_by ON public.events USING btree (created_by);


--
-- Name: idx_events_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_id ON public.events USING btree (id);


--
-- Name: idx_events_slug; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_slug ON public.events USING btree (slug);


--
-- Name: idx_fee_tag_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_fee_tag_trgm ON public.event_rankings USING gin (lower(fee_tag) public.gin_trgm_ops);


--
-- Name: idx_forms_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_forms_created_by ON public.forms USING btree (created_by);


--
-- Name: idx_forms_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_forms_event_id ON public.forms USING btree (event_id);


--
-- Name: idx_forms_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_forms_id ON public.forms USING btree (id);


--
-- Name: idx_ledger_form_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ledger_form_id ON public.ledger_entries USING btree (form_id);


--
-- Name: idx_ledger_payout; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ledger_payout ON public.ledger_entries USING btree (payout_status, payout_date);


--
-- Name: idx_location_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_location_trgm ON public.event_rankings USING gin (lower(location) public.gin_trgm_ops);


--
-- Name: idx_mode_description_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_mode_description_trgm ON public.event_rankings USING gin (lower(mode_description) public.gin_trgm_ops);


--
-- Name: idx_mode_tag_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_mode_tag_trgm ON public.event_rankings USING gin (lower(mode_tag) public.gin_trgm_ops);


--
-- Name: idx_name_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_name_trgm ON public.event_rankings USING gin (lower((name)::text) public.gin_trgm_ops);


--
-- Name: idx_payout_accounts_bank_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_bank_id ON public.payout_accounts USING btree (bank_id);


--
-- Name: idx_payout_accounts_beneficiary_ref; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_payout_accounts_beneficiary_ref ON public.payout_accounts USING btree (beneficiary_ref);


--
-- Name: idx_payout_accounts_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_user_id ON public.payout_accounts USING btree (user_id);


--
-- Name: idx_registrations_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_registrations_id ON public.registrations USING btree (id);


--
-- Name: idx_transactions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transactions_id ON public.transactions USING btree (id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_id ON public.users USING btree (id);


--
-- Name: idx_users_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_name ON public.users USING btree (name);


--
-- Name: idx_users_phone; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_phone ON public.users USING btree (phone);


--
-- Name: events backup_deleted_events; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER backup_deleted_events BEFORE DELETE ON public.events FOR EACH ROW EXECUTE FUNCTION public.backup_deleted_record();


--
-- Name: events refresh_event_rankings_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER refresh_event_rankings_trigger AFTER INSERT OR DELETE OR UPDATE ON public.events FOR EACH STATEMENT EXECUTE FUNCTION public.refresh_event_rankings();


--
-- Name: events fk_events_created_by; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT fk_events_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: forms fk_forms_created_by; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forms
    ADD CONSTRAINT fk_forms_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: forms fk_forms_event_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forms
    ADD CONSTRAINT fk_forms_event_id FOREIGN KEY (event_id) REFERENCES public.events(id) ON DELETE SET NULL;


--
-- Name: ledger_entries ledger_entries_form_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ledger_entries
    ADD CONSTRAINT ledger_entries_form_id_fkey FOREIGN KEY (form_id) REFERENCES public.forms(id);


--
-- Name: payout_accounts payout_accounts_bank_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payout_accounts
    ADD CONSTRAINT payout_accounts_bank_id_fkey FOREIGN KEY (bank_id) REFERENCES public.banks(id);


--
-- Name: refunds refunds_registration_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_registration_id_fkey FOREIGN KEY (registration_id) REFERENCES public.registrations(id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: event_rankings; Type: MATERIALIZED VIEW DATA; Schema: public; Owner: postgres
--

REFRESH MATERIALIZED VIEW public.event_rankings;


--
-- PostgreSQL database dump complete
--

