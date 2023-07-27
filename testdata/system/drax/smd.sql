--
-- PostgreSQL database dump
--

-- Dumped from database version 11.20 (Ubuntu 11.20-1.pgdg18.04+1)
-- Dumped by pg_dump version 14.8 (Ubuntu 14.8-1.pgdg18.04+1)

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
-- Name: metric_helpers; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA metric_helpers;


ALTER SCHEMA metric_helpers OWNER TO postgres;

--
-- Name: user_management; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA user_management;


ALTER SCHEMA user_management OWNER TO postgres;

--
-- Name: pg_stat_statements; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_stat_statements WITH SCHEMA public;


--
-- Name: EXTENSION pg_stat_statements; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_stat_statements IS 'track execution statistics of all SQL statements executed';


--
-- Name: pg_stat_kcache; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_stat_kcache WITH SCHEMA public;


--
-- Name: EXTENSION pg_stat_kcache; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_stat_kcache IS 'Kernel statistics gathering';


--
-- Name: set_user; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS set_user WITH SCHEMA public;


--
-- Name: EXTENSION set_user; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION set_user IS 'similar to SET ROLE but with added logging';


--
-- Name: group_namespace; Type: TYPE; Schema: public; Owner: hmsdsuser
--

CREATE TYPE public.group_namespace AS ENUM (
    'partition',
    'group'
);


ALTER TYPE public.group_namespace OWNER TO hmsdsuser;

--
-- Name: group_type; Type: TYPE; Schema: public; Owner: hmsdsuser
--

CREATE TYPE public.group_type AS ENUM (
    'partition',
    'exclusive',
    'shared'
);


ALTER TYPE public.group_type OWNER TO hmsdsuser;

--
-- Name: get_btree_bloat_approx(); Type: FUNCTION; Schema: metric_helpers; Owner: postgres
--

CREATE FUNCTION metric_helpers.get_btree_bloat_approx(OUT i_database name, OUT i_schema_name name, OUT i_table_name name, OUT i_index_name name, OUT i_real_size numeric, OUT i_extra_size numeric, OUT i_extra_ratio double precision, OUT i_fill_factor integer, OUT i_bloat_size double precision, OUT i_bloat_ratio double precision, OUT i_is_na boolean) RETURNS SETOF record
    LANGUAGE sql IMMUTABLE STRICT SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
SELECT current_database(), nspname AS schemaname, tblname, idxname, bs*(relpages)::bigint AS real_size,
  bs*(relpages-est_pages)::bigint AS extra_size,
  100 * (relpages-est_pages)::float / relpages AS extra_ratio,
  fillfactor,
  CASE WHEN relpages > est_pages_ff
    THEN bs*(relpages-est_pages_ff)
    ELSE 0
  END AS bloat_size,
  100 * (relpages-est_pages_ff)::float / relpages AS bloat_ratio,
  is_na
  -- , 100-(pst).avg_leaf_density AS pst_avg_bloat, est_pages, index_tuple_hdr_bm, maxalign, pagehdr, nulldatawidth, nulldatahdrwidth, reltuples, relpages -- (DEBUG INFO)
FROM (
  SELECT coalesce(1 +
         ceil(reltuples/floor((bs-pageopqdata-pagehdr)/(4+nulldatahdrwidth)::float)), 0 -- ItemIdData size + computed avg size of a tuple (nulldatahdrwidth)
      ) AS est_pages,
      coalesce(1 +
         ceil(reltuples/floor((bs-pageopqdata-pagehdr)*fillfactor/(100*(4+nulldatahdrwidth)::float))), 0
      ) AS est_pages_ff,
      bs, nspname, tblname, idxname, relpages, fillfactor, is_na
      -- , pgstatindex(idxoid) AS pst, index_tuple_hdr_bm, maxalign, pagehdr, nulldatawidth, nulldatahdrwidth, reltuples -- (DEBUG INFO)
  FROM (
      SELECT maxalign, bs, nspname, tblname, idxname, reltuples, relpages, idxoid, fillfactor,
            ( index_tuple_hdr_bm +
                maxalign - CASE -- Add padding to the index tuple header to align on MAXALIGN
                  WHEN index_tuple_hdr_bm%maxalign = 0 THEN maxalign
                  ELSE index_tuple_hdr_bm%maxalign
                END
              + nulldatawidth + maxalign - CASE -- Add padding to the data to align on MAXALIGN
                  WHEN nulldatawidth = 0 THEN 0
                  WHEN nulldatawidth::integer%maxalign = 0 THEN maxalign
                  ELSE nulldatawidth::integer%maxalign
                END
            )::numeric AS nulldatahdrwidth, pagehdr, pageopqdata, is_na
            -- , index_tuple_hdr_bm, nulldatawidth -- (DEBUG INFO)
      FROM (
          SELECT n.nspname, ct.relname AS tblname, i.idxname, i.reltuples, i.relpages,
              i.idxoid, i.fillfactor, current_setting('block_size')::numeric AS bs,
              CASE -- MAXALIGN: 4 on 32bits, 8 on 64bits (and mingw32 ?)
                WHEN version() ~ 'mingw32' OR version() ~ '64-bit|x86_64|ppc64|ia64|amd64' THEN 8
                ELSE 4
              END AS maxalign,
              /* per page header, fixed size: 20 for 7.X, 24 for others */
              24 AS pagehdr,
              /* per page btree opaque data */
              16 AS pageopqdata,
              /* per tuple header: add IndexAttributeBitMapData if some cols are null-able */
              CASE WHEN max(coalesce(s.stanullfrac,0)) = 0
                  THEN 2 -- IndexTupleData size
                  ELSE 2 + (( 32 + 8 - 1 ) / 8) -- IndexTupleData size + IndexAttributeBitMapData size ( max num filed per index + 8 - 1 /8)
              END AS index_tuple_hdr_bm,
              /* data len: we remove null values save space using it fractionnal part from stats */
              sum( (1-coalesce(s.stanullfrac, 0)) * coalesce(s.stawidth, 1024)) AS nulldatawidth,
              max( CASE WHEN a.atttypid = 'pg_catalog.name'::regtype THEN 1 ELSE 0 END ) > 0 AS is_na
          FROM (
              SELECT idxname, reltuples, relpages, tbloid, idxoid, fillfactor,
                  CASE WHEN indkey[i]=0 THEN idxoid ELSE tbloid END AS att_rel,
                  CASE WHEN indkey[i]=0 THEN i ELSE indkey[i] END AS att_pos
              FROM (
                  SELECT idxname, reltuples, relpages, tbloid, idxoid, fillfactor, indkey, generate_series(1,indnatts) AS i
                  FROM (
                      SELECT ci.relname AS idxname, ci.reltuples, ci.relpages, i.indrelid AS tbloid,
                          i.indexrelid AS idxoid,
                          coalesce(substring(
                              array_to_string(ci.reloptions, ' ')
                              from 'fillfactor=([0-9]+)')::smallint, 90) AS fillfactor,
                          i.indnatts,
                          string_to_array(textin(int2vectorout(i.indkey)),' ')::int[] AS indkey
                      FROM pg_index i
                      JOIN pg_class ci ON ci.oid=i.indexrelid
                      WHERE ci.relam=(SELECT oid FROM pg_am WHERE amname = 'btree')
                        AND ci.relpages > 0
                  ) AS idx_data
              ) AS idx_data_cross
          ) i
          JOIN pg_attribute a ON a.attrelid = i.att_rel
                             AND a.attnum = i.att_pos
          JOIN pg_statistic s ON s.starelid = i.att_rel
                             AND s.staattnum = i.att_pos
          JOIN pg_class ct ON ct.oid = i.tbloid
          JOIN pg_namespace n ON ct.relnamespace = n.oid
          GROUP BY 1,2,3,4,5,6,7,8,9,10
      ) AS rows_data_stats
  ) AS rows_hdr_pdg_stats
) AS relation_stats;
$$;


ALTER FUNCTION metric_helpers.get_btree_bloat_approx(OUT i_database name, OUT i_schema_name name, OUT i_table_name name, OUT i_index_name name, OUT i_real_size numeric, OUT i_extra_size numeric, OUT i_extra_ratio double precision, OUT i_fill_factor integer, OUT i_bloat_size double precision, OUT i_bloat_ratio double precision, OUT i_is_na boolean) OWNER TO postgres;

--
-- Name: get_table_bloat_approx(); Type: FUNCTION; Schema: metric_helpers; Owner: postgres
--

CREATE FUNCTION metric_helpers.get_table_bloat_approx(OUT t_database name, OUT t_schema_name name, OUT t_table_name name, OUT t_real_size numeric, OUT t_extra_size double precision, OUT t_extra_ratio double precision, OUT t_fill_factor integer, OUT t_bloat_size double precision, OUT t_bloat_ratio double precision, OUT t_is_na boolean) RETURNS SETOF record
    LANGUAGE sql IMMUTABLE STRICT SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
SELECT
  current_database(),
  schemaname,
  tblname,
  (bs*tblpages) AS real_size,
  ((tblpages-est_tblpages)*bs) AS extra_size,
  CASE WHEN tblpages - est_tblpages > 0
    THEN 100 * (tblpages - est_tblpages)/tblpages::float
    ELSE 0
  END AS extra_ratio,
  fillfactor,
  CASE WHEN tblpages - est_tblpages_ff > 0
    THEN (tblpages-est_tblpages_ff)*bs
    ELSE 0
  END AS bloat_size,
  CASE WHEN tblpages - est_tblpages_ff > 0
    THEN 100 * (tblpages - est_tblpages_ff)/tblpages::float
    ELSE 0
  END AS bloat_ratio,
  is_na
FROM (
  SELECT ceil( reltuples / ( (bs-page_hdr)/tpl_size ) ) + ceil( toasttuples / 4 ) AS est_tblpages,
    ceil( reltuples / ( (bs-page_hdr)*fillfactor/(tpl_size*100) ) ) + ceil( toasttuples / 4 ) AS est_tblpages_ff,
    tblpages, fillfactor, bs, tblid, schemaname, tblname, heappages, toastpages, is_na
    -- , tpl_hdr_size, tpl_data_size, pgstattuple(tblid) AS pst -- (DEBUG INFO)
  FROM (
    SELECT
      ( 4 + tpl_hdr_size + tpl_data_size + (2*ma)
        - CASE WHEN tpl_hdr_size%ma = 0 THEN ma ELSE tpl_hdr_size%ma END
        - CASE WHEN ceil(tpl_data_size)::int%ma = 0 THEN ma ELSE ceil(tpl_data_size)::int%ma END
      ) AS tpl_size, bs - page_hdr AS size_per_block, (heappages + toastpages) AS tblpages, heappages,
      toastpages, reltuples, toasttuples, bs, page_hdr, tblid, schemaname, tblname, fillfactor, is_na
      -- , tpl_hdr_size, tpl_data_size
    FROM (
      SELECT
        tbl.oid AS tblid, ns.nspname AS schemaname, tbl.relname AS tblname, tbl.reltuples,
        tbl.relpages AS heappages, coalesce(toast.relpages, 0) AS toastpages,
        coalesce(toast.reltuples, 0) AS toasttuples,
        coalesce(substring(
          array_to_string(tbl.reloptions, ' ')
          FROM 'fillfactor=([0-9]+)')::smallint, 100) AS fillfactor,
        current_setting('block_size')::numeric AS bs,
        CASE WHEN version()~'mingw32' OR version()~'64-bit|x86_64|ppc64|ia64|amd64' THEN 8 ELSE 4 END AS ma,
        24 AS page_hdr,
        23 + CASE WHEN MAX(coalesce(s.null_frac,0)) > 0 THEN ( 7 + count(s.attname) ) / 8 ELSE 0::int END
           + CASE WHEN bool_or(att.attname = 'oid' and att.attnum < 0) THEN 4 ELSE 0 END AS tpl_hdr_size,
        sum( (1-coalesce(s.null_frac, 0)) * coalesce(s.avg_width, 0) ) AS tpl_data_size,
        bool_or(att.atttypid = 'pg_catalog.name'::regtype)
          OR sum(CASE WHEN att.attnum > 0 THEN 1 ELSE 0 END) <> count(s.attname) AS is_na
      FROM pg_attribute AS att
        JOIN pg_class AS tbl ON att.attrelid = tbl.oid
        JOIN pg_namespace AS ns ON ns.oid = tbl.relnamespace
        LEFT JOIN pg_stats AS s ON s.schemaname=ns.nspname
          AND s.tablename = tbl.relname AND s.inherited=false AND s.attname=att.attname
        LEFT JOIN pg_class AS toast ON tbl.reltoastrelid = toast.oid
      WHERE NOT att.attisdropped
        AND tbl.relkind = 'r'
      GROUP BY 1,2,3,4,5,6,7,8,9,10
      ORDER BY 2,3
    ) AS s
  ) AS s2
) AS s3 WHERE schemaname NOT LIKE 'information_schema';
$$;


ALTER FUNCTION metric_helpers.get_table_bloat_approx(OUT t_database name, OUT t_schema_name name, OUT t_table_name name, OUT t_real_size numeric, OUT t_extra_size double precision, OUT t_extra_ratio double precision, OUT t_fill_factor integer, OUT t_bloat_size double precision, OUT t_bloat_ratio double precision, OUT t_is_na boolean) OWNER TO postgres;

--
-- Name: pg_stat_statements(boolean); Type: FUNCTION; Schema: metric_helpers; Owner: postgres
--

CREATE FUNCTION metric_helpers.pg_stat_statements(showtext boolean) RETURNS SETOF public.pg_stat_statements
    LANGUAGE sql IMMUTABLE STRICT SECURITY DEFINER
    AS $$
  SELECT * FROM public.pg_stat_statements(showtext);
$$;


ALTER FUNCTION metric_helpers.pg_stat_statements(showtext boolean) OWNER TO postgres;

--
-- Name: comp_ethernet_interfaces_update(); Type: FUNCTION; Schema: public; Owner: hmsdsuser
--

CREATE FUNCTION public.comp_ethernet_interfaces_update() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    comp_ethernet_interface RECORD;
BEGIN
    FOR comp_ethernet_interface IN SELECT id, json_build_array(json_build_object('IPAddress', ipaddr, 'Network', '')) as ip_addresses
                                   FROM comp_eth_interfaces
                                   WHERE ipaddr != ''
        LOOP
            UPDATE comp_eth_interfaces
            SET ip_addresses = comp_ethernet_interface.ip_addresses
            WHERE id = comp_ethernet_interface.id;
        END LOOP;
END;
$$;


ALTER FUNCTION public.comp_ethernet_interfaces_update() OWNER TO hmsdsuser;

--
-- Name: comp_lock_update_reservations(); Type: FUNCTION; Schema: public; Owner: hmsdsuser
--

CREATE FUNCTION public.comp_lock_update_reservations() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    lock_member RECORD;
BEGIN
    FOR lock_member IN SELECT
        component_lock_members.component_id AS "comp_id",
        component_lock_members.lock_id AS "lock_id",
        component_locks.created AS "created",
        component_locks.lifetime AS "lifetime"
    FROM component_lock_members LEFT JOIN component_locks ON component_lock_members.lock_id = component_locks.id LOOP
        INSERT INTO reservations (
            component_id, create_timestamp, expiration_timestamp, deputy_key, reservation_key, v1_lock_id)
        VALUES (
            lock_member.comp_id,
            lock_member.created,
            lock_member.created + (lock_member.lifetime || ' seconds')::interval,
            lock_member.comp_id || ':dk:' || lock_member.lock_id::text,
            lock_member.comp_id || ':rk:' || lock_member.lock_id::text,
            lock_member.lock_id);
    END LOOP;
END;
$$;


ALTER FUNCTION public.comp_lock_update_reservations() OWNER TO hmsdsuser;

--
-- Name: hwinv_by_loc_update_parents(); Type: FUNCTION; Schema: public; Owner: hmsdsuser
--

CREATE FUNCTION public.hwinv_by_loc_update_parents() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    node_id RECORD;
BEGIN
    FOR node_id IN SELECT id FROM hwinv_by_loc WHERE type = 'Node' LOOP
        UPDATE hwinv_by_loc SET parent_node = node_id.id WHERE id SIMILAR TO node_id.id||'([[:alpha:]][[:alnum:]]*)?';
    END LOOP;
    UPDATE hwinv_by_loc SET parent_node = id WHERE parent_node = '';
END;
$$;


ALTER FUNCTION public.hwinv_by_loc_update_parents() OWNER TO hmsdsuser;

--
-- Name: hwinv_hist_prune(); Type: FUNCTION; Schema: public; Owner: hmsdsuser
--

CREATE FUNCTION public.hwinv_hist_prune() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    comp_id RECORD;
    fru_event1 RECORD;
    fru_event2 RECORD;
BEGIN
    FOR comp_id IN SELECT distinct id FROM hwinv_hist LOOP
        SELECT * INTO fru_event1 FROM hwinv_hist WHERE id = comp_id.id ORDER BY timestamp ASC LIMIT 1;
        FOR fru_event2 IN SELECT * FROM hwinv_hist WHERE id = comp_id.id AND timestamp != fru_event1.timestamp ORDER BY timestamp ASC LOOP
            IF fru_event2.fru_id = fru_event1.fru_id THEN
                DELETE FROM hwinv_hist WHERE id = fru_event2.id AND fru_id = fru_event2.fru_id AND timestamp = fru_event2.timestamp;
            ELSE
                fru_event1 = fru_event2;
            END IF;
        END LOOP;
    END LOOP;
END;
$$;


ALTER FUNCTION public.hwinv_hist_prune() OWNER TO hmsdsuser;

--
-- Name: create_application_user(text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.create_application_user(username text) RETURNS text
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
DECLARE
    pw text;
BEGIN
    SELECT user_management.random_password(20) INTO pw;
    EXECUTE format($$ CREATE USER %I WITH PASSWORD %L $$, username, pw);
    RETURN pw;
END
$_$;


ALTER FUNCTION user_management.create_application_user(username text) OWNER TO postgres;

--
-- Name: FUNCTION create_application_user(username text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.create_application_user(username text) IS 'Creates a user that can login, sets the password to a strong random one,
which is then returned';


--
-- Name: create_application_user_or_change_password(text, text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.create_application_user_or_change_password(username text, password text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
BEGIN
    PERFORM 1 FROM pg_roles WHERE rolname = username;

    IF FOUND
    THEN
        EXECUTE format($$ ALTER ROLE %I WITH PASSWORD %L $$, username, password);
    ELSE
        EXECUTE format($$ CREATE USER %I WITH PASSWORD %L $$, username, password);
    END IF;
END
$_$;


ALTER FUNCTION user_management.create_application_user_or_change_password(username text, password text) OWNER TO postgres;

--
-- Name: FUNCTION create_application_user_or_change_password(username text, password text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.create_application_user_or_change_password(username text, password text) IS 'USE THIS ONLY IN EMERGENCY!  The password will appear in the DB logs.
Creates a user that can login, sets the password to the one provided.
If the user already exists, sets its password.';


--
-- Name: create_role(text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.create_role(rolename text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
BEGIN
    -- set ADMIN to the admin user, so every member of admin can GRANT these roles to each other
    EXECUTE format($$ CREATE ROLE %I WITH ADMIN admin $$, rolename);
END;
$_$;


ALTER FUNCTION user_management.create_role(rolename text) OWNER TO postgres;

--
-- Name: FUNCTION create_role(rolename text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.create_role(rolename text) IS 'Creates a role that cannot log in, but can be used to set up fine-grained privileges';


--
-- Name: create_user(text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.create_user(username text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
BEGIN
    EXECUTE format($$ CREATE USER %I IN ROLE zalandos, admin $$, username);
    EXECUTE format($$ ALTER ROLE %I SET log_statement TO 'all' $$, username);
END;
$_$;


ALTER FUNCTION user_management.create_user(username text) OWNER TO postgres;

--
-- Name: FUNCTION create_user(username text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.create_user(username text) IS 'Creates a user that is supposed to be a human, to be authenticated without a password';


--
-- Name: drop_role(text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.drop_role(username text) RETURNS void
    LANGUAGE sql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
SELECT user_management.drop_user(username);
$$;


ALTER FUNCTION user_management.drop_role(username text) OWNER TO postgres;

--
-- Name: FUNCTION drop_role(username text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.drop_role(username text) IS 'Drop a human or application user.  Intended for cleanup (either after team changes or mistakes in role setup).
Roles (= users) that own database objects cannot be dropped.';


--
-- Name: drop_user(text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.drop_user(username text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
BEGIN
    EXECUTE format($$ DROP ROLE %I $$, username);
END
$_$;


ALTER FUNCTION user_management.drop_user(username text) OWNER TO postgres;

--
-- Name: FUNCTION drop_user(username text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.drop_user(username text) IS 'Drop a human or application user.  Intended for cleanup (either after team changes or mistakes in role setup).
Roles (= users) that own database objects cannot be dropped.';


--
-- Name: random_password(integer); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.random_password(length integer) RETURNS text
    LANGUAGE sql
    SET search_path TO 'pg_catalog'
    AS $$
WITH chars (c) AS (
    SELECT chr(33)
    UNION ALL
    SELECT chr(i) FROM generate_series (35, 38) AS t (i)
    UNION ALL
    SELECT chr(i) FROM generate_series (42, 90) AS t (i)
    UNION ALL
    SELECT chr(i) FROM generate_series (97, 122) AS t (i)
),
bricks (b) AS (
    -- build a pool of chars (the size will be the number of chars above times length)
    -- and shuffle it
    SELECT c FROM chars, generate_series(1, length) ORDER BY random()
)
SELECT substr(string_agg(b, ''), 1, length) FROM bricks;
$$;


ALTER FUNCTION user_management.random_password(length integer) OWNER TO postgres;

--
-- Name: revoke_admin(text); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.revoke_admin(username text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
BEGIN
    EXECUTE format($$ REVOKE admin FROM %I $$, username);
END
$_$;


ALTER FUNCTION user_management.revoke_admin(username text) OWNER TO postgres;

--
-- Name: FUNCTION revoke_admin(username text); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.revoke_admin(username text) IS 'Use this function to make a human user less privileged,
ie. when you want to grant someone read privileges only';


--
-- Name: terminate_backend(integer); Type: FUNCTION; Schema: user_management; Owner: postgres
--

CREATE FUNCTION user_management.terminate_backend(pid integer) RETURNS boolean
    LANGUAGE sql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
SELECT pg_terminate_backend(pid);
$$;


ALTER FUNCTION user_management.terminate_backend(pid integer) OWNER TO postgres;

--
-- Name: FUNCTION terminate_backend(pid integer); Type: COMMENT; Schema: user_management; Owner: postgres
--

COMMENT ON FUNCTION user_management.terminate_backend(pid integer) IS 'When there is a process causing harm, you can kill it using this function.  Get the pid from pg_stat_activity
(be careful to match the user name (usename) and the query, in order not to kill innocent kittens) and pass it to terminate_backend()';


--
-- Name: index_bloat; Type: VIEW; Schema: metric_helpers; Owner: postgres
--

CREATE VIEW metric_helpers.index_bloat AS
 SELECT get_btree_bloat_approx.i_database,
    get_btree_bloat_approx.i_schema_name,
    get_btree_bloat_approx.i_table_name,
    get_btree_bloat_approx.i_index_name,
    get_btree_bloat_approx.i_real_size,
    get_btree_bloat_approx.i_extra_size,
    get_btree_bloat_approx.i_extra_ratio,
    get_btree_bloat_approx.i_fill_factor,
    get_btree_bloat_approx.i_bloat_size,
    get_btree_bloat_approx.i_bloat_ratio,
    get_btree_bloat_approx.i_is_na
   FROM metric_helpers.get_btree_bloat_approx() get_btree_bloat_approx(i_database, i_schema_name, i_table_name, i_index_name, i_real_size, i_extra_size, i_extra_ratio, i_fill_factor, i_bloat_size, i_bloat_ratio, i_is_na);


ALTER TABLE metric_helpers.index_bloat OWNER TO postgres;

--
-- Name: pg_stat_statements; Type: VIEW; Schema: metric_helpers; Owner: postgres
--

CREATE VIEW metric_helpers.pg_stat_statements AS
 SELECT pg_stat_statements.userid,
    pg_stat_statements.dbid,
    pg_stat_statements.queryid,
    pg_stat_statements.query,
    pg_stat_statements.calls,
    pg_stat_statements.total_time,
    pg_stat_statements.min_time,
    pg_stat_statements.max_time,
    pg_stat_statements.mean_time,
    pg_stat_statements.stddev_time,
    pg_stat_statements.rows,
    pg_stat_statements.shared_blks_hit,
    pg_stat_statements.shared_blks_read,
    pg_stat_statements.shared_blks_dirtied,
    pg_stat_statements.shared_blks_written,
    pg_stat_statements.local_blks_hit,
    pg_stat_statements.local_blks_read,
    pg_stat_statements.local_blks_dirtied,
    pg_stat_statements.local_blks_written,
    pg_stat_statements.temp_blks_read,
    pg_stat_statements.temp_blks_written,
    pg_stat_statements.blk_read_time,
    pg_stat_statements.blk_write_time
   FROM metric_helpers.pg_stat_statements(true) pg_stat_statements(userid, dbid, queryid, query, calls, total_time, min_time, max_time, mean_time, stddev_time, rows, shared_blks_hit, shared_blks_read, shared_blks_dirtied, shared_blks_written, local_blks_hit, local_blks_read, local_blks_dirtied, local_blks_written, temp_blks_read, temp_blks_written, blk_read_time, blk_write_time);


ALTER TABLE metric_helpers.pg_stat_statements OWNER TO postgres;

--
-- Name: table_bloat; Type: VIEW; Schema: metric_helpers; Owner: postgres
--

CREATE VIEW metric_helpers.table_bloat AS
 SELECT get_table_bloat_approx.t_database,
    get_table_bloat_approx.t_schema_name,
    get_table_bloat_approx.t_table_name,
    get_table_bloat_approx.t_real_size,
    get_table_bloat_approx.t_extra_size,
    get_table_bloat_approx.t_extra_ratio,
    get_table_bloat_approx.t_fill_factor,
    get_table_bloat_approx.t_bloat_size,
    get_table_bloat_approx.t_bloat_ratio,
    get_table_bloat_approx.t_is_na
   FROM metric_helpers.get_table_bloat_approx() get_table_bloat_approx(t_database, t_schema_name, t_table_name, t_real_size, t_extra_size, t_extra_ratio, t_fill_factor, t_bloat_size, t_bloat_ratio, t_is_na);


ALTER TABLE metric_helpers.table_bloat OWNER TO postgres;

SET default_tablespace = '';

--
-- Name: comp_endpoints; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.comp_endpoints (
    id character varying(63) NOT NULL,
    type character varying(63) NOT NULL,
    domain character varying(192) NOT NULL,
    redfish_type character varying(63) NOT NULL,
    redfish_subtype character varying(63) NOT NULL,
    rf_endpoint_id character varying(63) NOT NULL,
    mac character varying(32),
    uuid character varying(64),
    odata_id character varying(512) NOT NULL,
    component_info json
);


ALTER TABLE public.comp_endpoints OWNER TO hmsdsuser;

--
-- Name: rf_endpoints; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.rf_endpoints (
    id character varying(63) NOT NULL,
    type character varying(63) NOT NULL,
    name text,
    hostname character varying(63),
    domain character varying(192),
    fqdn character varying(255),
    ip_info json DEFAULT '{}'::json,
    enabled boolean,
    uuid character varying(64),
    "user" character varying(128),
    password character varying(128),
    usessdp boolean,
    macrequired boolean,
    macaddr character varying(32),
    rediscoveronupdate boolean,
    templateid character varying(128),
    discovery_info json,
    ipaddr character varying(64) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.rf_endpoints OWNER TO hmsdsuser;

--
-- Name: comp_endpoints_info; Type: VIEW; Schema: public; Owner: hmsdsuser
--

CREATE VIEW public.comp_endpoints_info AS
 SELECT comp_endpoints.id,
    comp_endpoints.type,
    comp_endpoints.domain,
    comp_endpoints.redfish_type,
    comp_endpoints.redfish_subtype,
    comp_endpoints.mac,
    comp_endpoints.uuid,
    comp_endpoints.odata_id,
    comp_endpoints.rf_endpoint_id,
    rf_endpoints.fqdn AS rf_endpoint_fqdn,
    comp_endpoints.component_info,
    rf_endpoints."user" AS rf_endpoint_user,
    rf_endpoints.password AS rf_endpoint_password,
    rf_endpoints.enabled
   FROM (public.comp_endpoints
     LEFT JOIN public.rf_endpoints ON (((comp_endpoints.rf_endpoint_id)::text = (rf_endpoints.id)::text)));


ALTER TABLE public.comp_endpoints_info OWNER TO hmsdsuser;

--
-- Name: comp_eth_interfaces; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.comp_eth_interfaces (
    id character varying(32) NOT NULL,
    description text,
    macaddr character varying(32) NOT NULL,
    last_update timestamp with time zone,
    compid character varying(63) DEFAULT ''::character varying NOT NULL,
    comptype character varying(63) DEFAULT ''::character varying NOT NULL,
    ip_addresses json DEFAULT '[]'::json NOT NULL
);


ALTER TABLE public.comp_eth_interfaces OWNER TO hmsdsuser;

--
-- Name: component_group_members; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.component_group_members (
    component_id character varying(63) NOT NULL,
    group_id uuid NOT NULL,
    group_namespace character varying(255) NOT NULL,
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.component_group_members OWNER TO hmsdsuser;

--
-- Name: component_groups; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.component_groups (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description character varying(255) NOT NULL,
    tags character varying(255)[],
    annotations json DEFAULT '{}'::json,
    type public.group_type,
    namespace public.group_namespace,
    exclusive_group_identifier character varying(253)
);


ALTER TABLE public.component_groups OWNER TO hmsdsuser;

--
-- Name: components; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.components (
    id character varying(63) NOT NULL,
    type character varying(63) NOT NULL,
    state character varying(32) NOT NULL,
    admin character varying(32) DEFAULT ''::character varying NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    flag character varying(32) NOT NULL,
    role character varying(32) NOT NULL,
    nid bigint NOT NULL,
    subtype character varying(64) NOT NULL,
    nettype character varying(64) NOT NULL,
    arch character varying(64) NOT NULL,
    disposition character varying(64) DEFAULT ''::character varying NOT NULL,
    subrole character varying(32) DEFAULT ''::character varying NOT NULL,
    class character varying(32) DEFAULT ''::character varying NOT NULL,
    reservation_disabled boolean DEFAULT false NOT NULL,
    locked boolean DEFAULT false NOT NULL
);


ALTER TABLE public.components OWNER TO hmsdsuser;

--
-- Name: discovery_status; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.discovery_status (
    id integer NOT NULL,
    status character varying(128),
    last_update timestamp with time zone,
    details json
);


ALTER TABLE public.discovery_status OWNER TO hmsdsuser;

--
-- Name: hsn_interfaces; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.hsn_interfaces (
    nic character varying(32) NOT NULL,
    macaddr character varying(32) DEFAULT ''::character varying NOT NULL,
    hsn character varying(32) DEFAULT ''::character varying NOT NULL,
    node character varying(32) DEFAULT ''::character varying NOT NULL,
    ipaddr character varying(64) DEFAULT ''::character varying NOT NULL,
    last_update timestamp with time zone
);


ALTER TABLE public.hsn_interfaces OWNER TO hmsdsuser;

--
-- Name: hwinv_by_fru; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.hwinv_by_fru (
    fru_id character varying(255) NOT NULL,
    type character varying(63) NOT NULL,
    subtype character varying(63) NOT NULL,
    serial_number character varying(255) DEFAULT ''::character varying NOT NULL,
    part_number character varying(255) DEFAULT ''::character varying NOT NULL,
    manufacturer character varying(255) DEFAULT ''::character varying NOT NULL,
    fru_info json NOT NULL
);


ALTER TABLE public.hwinv_by_fru OWNER TO hmsdsuser;

--
-- Name: hwinv_by_loc; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.hwinv_by_loc (
    id character varying(63) NOT NULL,
    type character varying(63) NOT NULL,
    ordinal integer NOT NULL,
    status character varying(63) NOT NULL,
    parent character varying(63) DEFAULT ''::character varying NOT NULL,
    location_info json,
    fru_id character varying(255),
    parent_node character varying(63) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.hwinv_by_loc OWNER TO hmsdsuser;

--
-- Name: hwinv_by_loc_with_fru; Type: VIEW; Schema: public; Owner: hmsdsuser
--

CREATE VIEW public.hwinv_by_loc_with_fru AS
 SELECT hwinv_by_loc.id,
    hwinv_by_loc.type,
    hwinv_by_loc.ordinal,
    hwinv_by_loc.status,
    hwinv_by_loc.location_info,
    hwinv_by_loc.fru_id,
    hwinv_by_fru.type AS fru_type,
    hwinv_by_fru.subtype AS fru_subtype,
    hwinv_by_fru.fru_info
   FROM (public.hwinv_by_loc
     LEFT JOIN public.hwinv_by_fru ON (((hwinv_by_loc.fru_id)::text = (hwinv_by_fru.fru_id)::text)));


ALTER TABLE public.hwinv_by_loc_with_fru OWNER TO hmsdsuser;

--
-- Name: hwinv_by_loc_with_partition; Type: VIEW; Schema: public; Owner: hmsdsuser
--

CREATE VIEW public.hwinv_by_loc_with_partition AS
 SELECT hwinv_by_loc.id,
    hwinv_by_loc.type,
    hwinv_by_loc.ordinal,
    hwinv_by_loc.status,
    hwinv_by_loc.location_info,
    hwinv_by_loc.fru_id,
    hwinv_by_fru.type AS fru_type,
    hwinv_by_fru.subtype AS fru_subtype,
    hwinv_by_fru.fru_info,
    part_info.name AS partition
   FROM ((public.hwinv_by_loc
     LEFT JOIN public.hwinv_by_fru ON (((hwinv_by_loc.fru_id)::text = (hwinv_by_fru.fru_id)::text)))
     LEFT JOIN ( SELECT component_group_members.component_id AS id,
            component_groups.name
           FROM (public.component_group_members
             LEFT JOIN public.component_groups ON ((component_group_members.group_id = component_groups.id)))
          WHERE ((component_group_members.group_namespace)::text = '%%partition%%'::text)) part_info ON (((hwinv_by_loc.parent_node)::text = (part_info.id)::text)));


ALTER TABLE public.hwinv_by_loc_with_partition OWNER TO hmsdsuser;

--
-- Name: hwinv_hist; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.hwinv_hist (
    id character varying(63),
    fru_id character varying(255),
    event_type character varying(128),
    "timestamp" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.hwinv_hist OWNER TO hmsdsuser;

--
-- Name: job_state_rf_poll; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.job_state_rf_poll (
    comp_id character varying(63) NOT NULL,
    job_id uuid NOT NULL
);


ALTER TABLE public.job_state_rf_poll OWNER TO hmsdsuser;

--
-- Name: job_sync; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.job_sync (
    id uuid NOT NULL,
    type character varying(128),
    status character varying(128),
    last_update timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lifetime integer NOT NULL
);


ALTER TABLE public.job_sync OWNER TO hmsdsuser;

--
-- Name: node_nid_mapping; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.node_nid_mapping (
    id character varying(63) NOT NULL,
    nid bigint,
    role character varying(32) NOT NULL,
    name character varying(32) DEFAULT ''::character varying NOT NULL,
    node_info json,
    subrole character varying(32) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.node_nid_mapping OWNER TO hmsdsuser;

--
-- Name: power_mapping; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.power_mapping (
    id character varying(63) NOT NULL,
    powered_by character varying(63)[] NOT NULL
);


ALTER TABLE public.power_mapping OWNER TO hmsdsuser;

--
-- Name: reservations; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.reservations (
    component_id character varying(63) NOT NULL,
    create_timestamp timestamp with time zone NOT NULL,
    expiration_timestamp timestamp with time zone,
    deputy_key character varying,
    reservation_key character varying
);


ALTER TABLE public.reservations OWNER TO hmsdsuser;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO hmsdsuser;

--
-- Name: scn_subscriptions; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.scn_subscriptions (
    id integer NOT NULL,
    sub_url character varying(255) NOT NULL,
    subscription json DEFAULT '{}'::json
);


ALTER TABLE public.scn_subscriptions OWNER TO hmsdsuser;

--
-- Name: scn_subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: hmsdsuser
--

CREATE SEQUENCE public.scn_subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.scn_subscriptions_id_seq OWNER TO hmsdsuser;

--
-- Name: scn_subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hmsdsuser
--

ALTER SEQUENCE public.scn_subscriptions_id_seq OWNED BY public.scn_subscriptions.id;


--
-- Name: service_endpoints; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.service_endpoints (
    rf_endpoint_id character varying(63) NOT NULL,
    redfish_type character varying(63) NOT NULL,
    redfish_subtype character varying(63) NOT NULL,
    uuid character varying(64),
    odata_id character varying(512) NOT NULL,
    service_info json
);


ALTER TABLE public.service_endpoints OWNER TO hmsdsuser;

--
-- Name: service_endpoints_info; Type: VIEW; Schema: public; Owner: hmsdsuser
--

CREATE VIEW public.service_endpoints_info AS
 SELECT service_endpoints.rf_endpoint_id,
    service_endpoints.redfish_type,
    service_endpoints.redfish_subtype,
    service_endpoints.uuid,
    service_endpoints.odata_id,
    rf_endpoints.fqdn AS rf_endpoint_fqdn,
    service_endpoints.service_info
   FROM (public.service_endpoints
     LEFT JOIN public.rf_endpoints ON (((service_endpoints.rf_endpoint_id)::text = (rf_endpoints.id)::text)));


ALTER TABLE public.service_endpoints_info OWNER TO hmsdsuser;

--
-- Name: system; Type: TABLE; Schema: public; Owner: hmsdsuser
--

CREATE TABLE public.system (
    id integer NOT NULL,
    schema_version integer NOT NULL,
    system_info json
);


ALTER TABLE public.system OWNER TO hmsdsuser;

--
-- Name: scn_subscriptions id; Type: DEFAULT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.scn_subscriptions ALTER COLUMN id SET DEFAULT nextval('public.scn_subscriptions_id_seq'::regclass);


--
-- Data for Name: comp_endpoints; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.comp_endpoints (id, type, domain, redfish_type, redfish_subtype, rf_endpoint_id, mac, uuid, odata_id, component_info) FROM stdin;
x3000c0s19e3	NodeEnclosure		Chassis	RackMount	x3000c0s19b3			/redfish/v1/Chassis/RackMount	{"Name":"Computer System Chassis"}
x3000c0s29b0n0	Node		ComputerSystem	Physical	x3000c0s29b0		61df0000-9855-11ed-8000-b42e99a522e5	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["GracefulShutdown","On","ForceRestart","ForceOff"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"5","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/5","Description":"Ethernet Interface Lan5","MACAddress":"b4:2e:99:a5:22:e5"},{"RedfishId":"6","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/6","Description":"Ethernet Interface Lan6","MACAddress":"b4:2e:99:a5:22:e6"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"50:6b:4b:23:a7:c4"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:d9:9e:80"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b8:59:9f:d9:9e:81"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"50:6b:4b:28:50:5c"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
x3000c0s29b0	NodeBMC		Manager	BMC	x3000c0s29b0	da:d7:6d:7e:d2:41	b42e99a5-22e7-cd03-0010-debf00f6456e	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","FQDN":"AMIB42E99A522E7.hmn","Hostname":"AMIB42E99A522E7","InterfaceEnabled":true,"MACAddress":"b4:2e:99:a5:22:e7","PermanentMACAddress":"b4:2e:99:a5:22:e7"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","FQDN":"AMIB42E99A522E7.hmn","Hostname":"AMIB42E99A522E7","InterfaceEnabled":true,"MACAddress":"da:d7:6d:7e:d2:41","PermanentMACAddress":"da:d7:6d:7e:d2:41"}]}
x3000c0s7b0	NodeBMC		Manager	BMC	x3000c0s7b0	9a:18:69:f2:7c:c8	e005dd6e-debf-0010-e603-b42e99a52267	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","Hostname":"AMIB42E99A52267","InterfaceEnabled":true,"MACAddress":"b4:2e:99:a5:22:67","PermanentMACAddress":"b4:2e:99:a5:22:67"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","Hostname":"AMIB42E99A52267","InterfaceEnabled":true,"MACAddress":"9a:18:69:f2:7c:c8","PermanentMACAddress":"9a:18:69:f2:7c:c8"}]}
x3000c0s13e0	NodeEnclosure		Chassis	RackMount	x3000c0s13b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s13b0	NodeBMC		Manager	BMC	x3000c0s13b0	92:dd:c6:32:40:10	b42e99ab-2594-cb03-0010-debf40e8916d	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","FQDN":"AMIB42E99AB2594.hmn","Hostname":"AMIB42E99AB2594","InterfaceEnabled":true,"MACAddress":"b4:2e:99:ab:25:94","PermanentMACAddress":"b4:2e:99:ab:25:94"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","FQDN":"AMIB42E99AB2594.hmn","Hostname":"AMIB42E99AB2594","InterfaceEnabled":true,"MACAddress":"92:dd:c6:32:40:10","PermanentMACAddress":"92:dd:c6:32:40:10"}]}
x3000c0s9e0	NodeEnclosure		Chassis	RackMount	x3000c0s9b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s5e0	NodeEnclosure		Chassis	RackMount	x3000c0s5b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s17e0	NodeEnclosure		Chassis	RackMount	x3000c0s17b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s17b0n0	Node		ComputerSystem	Physical	x3000c0s17b0		61df0000-9855-11ed-8000-b42e993a25f8	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart","On","ForceOff","GracefulShutdown"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"b8:59:9f:c7:12:42"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:c7:12:43"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b4:2e:99:3a:25:f8"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"b4:2e:99:3a:25:f9"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
x3000c0s15e0	NodeEnclosure		Chassis	RackMount	x3000c0s15b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s7e0	NodeEnclosure		Chassis	RackMount	x3000c0s7b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s7b0n0	Node		ComputerSystem	Physical	x3000c0s7b0		61df0000-9855-11ed-8000-b42e99a52265	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["GracefulShutdown","ForceOff","On","ForceRestart"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"98:03:9b:aa:96:0c"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:d9:9d:30"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b8:59:9f:d9:9d:31"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"98:03:9b:aa:94:a8"},{"RedfishId":"5","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/5","Description":"Ethernet Interface Lan5","MACAddress":"b4:2e:99:a5:22:65"},{"RedfishId":"6","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/6","Description":"Ethernet Interface Lan6","MACAddress":"b4:2e:99:a5:22:66"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/Self"},{"@odata.id":"/redfish/v1/Chassis/Self"}]}]}
x3000c0w22	MgmtSwitch		Chassis	Drawer	x3000c0w22			/redfish/v1/Chassis/EthernetSwitch_0	{"Name":"DellEMC Networking S3048ON Chassis"}
x3000c0r24e0	HSNBoard		Chassis	Enclosure	x3000c0r24b0			/redfish/v1/Chassis/Enclosure	{"Name":"Enclosure","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":["ForceOff","GracefulRestart","ForceOn","ForceRestart","GracefulShutdown","On","Off","PowerCycle"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"}}}
x3000c0r24b0	RouterBMC		Manager	EnclosureManager	x3000c0r24b0			/redfish/v1/Managers/BMC	{"Name":"BMC","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart","StatefulReset"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Managers/BMC/Actions/Manager.Reset"},"Oem":{"#CrayProcess.Schedule":{"Name@Redfish.AllowableValues":["memtest","cpuburn"],"target":"/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"}}}}
x3000c0s19e1	NodeEnclosure		Chassis	RackMount	x3000c0s19b1			/redfish/v1/Chassis/RackMount	{"Name":"Computer System Chassis"}
x3000c0s15b0n0	Node		ComputerSystem	Physical	x3000c0s15b0		61df0000-9855-11ed-8000-b42e993a2618	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["GracefulShutdown","ForceOff","ForceRestart","On"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"b4:2e:99:3a:26:19"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"b8:59:9f:c7:11:22"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:c7:11:23"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b4:2e:99:3a:26:18"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/Self"},{"@odata.id":"/redfish/v1/Chassis/Self"}]}]}
x3000c0s15b0	NodeBMC		Manager	BMC	x3000c0s15b0	32:d4:86:66:c8:7b	808cde6e-debf-0010-e603-b42e993a261a	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","Hostname":"AMIB42E993A261A","InterfaceEnabled":true,"MACAddress":"b4:2e:99:3a:26:1a","PermanentMACAddress":"b4:2e:99:3a:26:1a"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","Hostname":"AMIB42E993A261A","InterfaceEnabled":true,"MACAddress":"32:d4:86:66:c8:7b","PermanentMACAddress":"32:d4:86:66:c8:7b"}]}
x3000c0s11e0	NodeEnclosure		Chassis	RackMount	x3000c0s11b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s29e0	NodeEnclosure		Chassis	RackMount	x3000c0s29b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s19b1n0	Node		ComputerSystem	Physical	x3000c0s19b1		b996c5f4-82a9-11e8-ab21-a4bf013ed1fa	/redfish/v1/Systems/QSBP82704191	{"Name":"S2600BPB","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["PushPowerButton","On","GracefulShutdown","ForceRestart","Nmi","ForceOn","ForceOff"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Systems/QSBP82704191/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/QSBP82704191/EthernetInterfaces/1","Description":"System NIC 1","MACAddress":"a4:bf:01:3e:d1:fa","PermanentMACAddress":"a4:bf:01:3e:d1:fa"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/QSBP82704191/EthernetInterfaces/2","Description":"System NIC 2","MACAddress":"a4:bf:01:3e:d1:fb","PermanentMACAddress":"a4:bf:01:3e:d1:fb"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/QSBP82704191/EthernetInterfaces/3","Description":"System NIC 3","MACAddress":"ff:ff:ff:ff:ff:ff","PermanentMACAddress":"ff:ff:ff:ff:ff:ff"}],"PowerURL":"/redfish/v1/Chassis/RackMount/Baseboard/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/RackMount/Baseboard/Power#/PowerControl/0","MemberId":"0","Name":"Server Power Control","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/QSBP82704191"},{"@odata.id":"/redfish/v1/Chassis/RackMount"}]}]}
x3000c0s19b1	NodeBMC		Manager	BMC	x3000c0s19b1	a4:bf:01:3e:d1:fe	6bec0d34-b372-85a6-0197-b3bb3ef31ac7	/redfish/v1/Managers/BMC	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/1","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013ED1FC","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:d1:fc"},{"RedfishId":"2","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/2","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013ED1FC","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:d1:fd"},{"RedfishId":"3","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/3","Description":"Network Interface on the Baseboard Management Controller","FQDN":"x3000c0s19b1","Hostname":"BMCA4BF013ED1FC","InterfaceEnabled":true,"MACAddress":"a4:bf:01:3e:d1:fe"}]}
x3000c0s19e2	NodeEnclosure		Chassis	RackMount	x3000c0s19b2			/redfish/v1/Chassis/RackMount	{"Name":"Computer System Chassis"}
x3000c0s19b2n0	Node		ComputerSystem	Physical	x3000c0s19b2		a4cd681d-8064-11e8-ab21-a4bf013ec02a	/redfish/v1/Systems/QSBP82703289	{"Name":"S2600BPB","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["PushPowerButton","On","GracefulShutdown","ForceRestart","Nmi","ForceOn","ForceOff"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Systems/QSBP82703289/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/QSBP82703289/EthernetInterfaces/1","Description":"System NIC 1","MACAddress":"a4:bf:01:3e:c0:2a","PermanentMACAddress":"a4:bf:01:3e:c0:2a"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/QSBP82703289/EthernetInterfaces/2","Description":"System NIC 2","MACAddress":"a4:bf:01:3e:c0:2b","PermanentMACAddress":"a4:bf:01:3e:c0:2b"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/QSBP82703289/EthernetInterfaces/3","Description":"System NIC 3","MACAddress":"ff:ff:ff:ff:ff:ff","PermanentMACAddress":"ff:ff:ff:ff:ff:ff"}],"PowerURL":"/redfish/v1/Chassis/RackMount/Baseboard/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/RackMount/Baseboard/Power#/PowerControl/0","MemberId":"0","Name":"Server Power Control","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/QSBP82703289"},{"@odata.id":"/redfish/v1/Chassis/RackMount"}]}]}
x3000c0s19b2	NodeBMC		Manager	BMC	x3000c0s19b2	a4:bf:01:3e:c0:2e	5b79f5f6-2a4d-729b-b018-aedd4ec66f35	/redfish/v1/Managers/BMC	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},"EthernetNICInfo":[{"RedfishId":"3","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/3","Description":"Network Interface on the Baseboard Management Controller","FQDN":"x3000c0s19b2","Hostname":"BMCA4BF013EC02C","InterfaceEnabled":true,"MACAddress":"a4:bf:01:3e:c0:2e"},{"RedfishId":"1","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/1","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013EC02C","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:c0:2c"},{"RedfishId":"2","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/2","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013EC02C","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:c0:2d"}]}
x3000c0s19b3n0	Node		ComputerSystem	Physical	x3000c0s19b3		3811a2bf-8140-11e8-ab21-a4bf013ecf66	/redfish/v1/Systems/QSBP82704059	{"Name":"S2600BPB","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["PushPowerButton","On","GracefulShutdown","ForceRestart","Nmi","ForceOn","ForceOff"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Systems/QSBP82704059/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/QSBP82704059/EthernetInterfaces/3","Description":"System NIC 3","MACAddress":"ff:ff:ff:ff:ff:ff","PermanentMACAddress":"ff:ff:ff:ff:ff:ff"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/QSBP82704059/EthernetInterfaces/1","Description":"System NIC 1","MACAddress":"a4:bf:01:3e:cf:66","PermanentMACAddress":"a4:bf:01:3e:cf:66"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/QSBP82704059/EthernetInterfaces/2","Description":"System NIC 2","MACAddress":"a4:bf:01:3e:cf:67","PermanentMACAddress":"a4:bf:01:3e:cf:67"}],"PowerURL":"/redfish/v1/Chassis/RackMount/Baseboard/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/RackMount/Baseboard/Power#/PowerControl/0","MemberId":"0","Name":"Server Power Control","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/QSBP82704059"},{"@odata.id":"/redfish/v1/Chassis/RackMount"}]}]}
x3000c0s19b3	NodeBMC		Manager	BMC	x3000c0s19b3	a4:bf:01:3e:cf:6a	4d9db2f7-2119-db35-b5c2-831629cbe59c	/redfish/v1/Managers/BMC	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/1","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013ECF68","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:cf:68"},{"RedfishId":"2","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/2","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013ECF68","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:cf:69"},{"RedfishId":"3","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/3","Description":"Network Interface on the Baseboard Management Controller","FQDN":"x3000c0s19b3","Hostname":"BMCA4BF013ECF68","InterfaceEnabled":true,"MACAddress":"a4:bf:01:3e:cf:6a"}]}
x3000c0s19e4	NodeEnclosure		Chassis	RackMount	x3000c0s19b4			/redfish/v1/Chassis/RackMount	{"Name":"Computer System Chassis"}
x3000c0s19b4n0	Node		ComputerSystem	Physical	x3000c0s19b4		c96d020d-8193-11e8-ab21-a4bf013ed290	/redfish/v1/Systems/QSBP82704221	{"Name":"S2600BPB","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["PushPowerButton","On","GracefulShutdown","ForceRestart","Nmi","ForceOn","ForceOff"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Systems/QSBP82704221/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/QSBP82704221/EthernetInterfaces/1","Description":"System NIC 1","MACAddress":"a4:bf:01:3e:d2:90","PermanentMACAddress":"a4:bf:01:3e:d2:90"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/QSBP82704221/EthernetInterfaces/2","Description":"System NIC 2","MACAddress":"a4:bf:01:3e:d2:91","PermanentMACAddress":"a4:bf:01:3e:d2:91"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/QSBP82704221/EthernetInterfaces/3","Description":"System NIC 3","MACAddress":"ff:ff:ff:ff:ff:ff","PermanentMACAddress":"ff:ff:ff:ff:ff:ff"}],"PowerURL":"/redfish/v1/Chassis/RackMount/Baseboard/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/RackMount/Baseboard/Power#/PowerControl/0","MemberId":"0","Name":"Server Power Control","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/QSBP82704221"},{"@odata.id":"/redfish/v1/Chassis/RackMount"}]}]}
x3000c0s19b4	NodeBMC		Manager	BMC	x3000c0s19b4	a4:bf:01:3e:d2:94	f7028065-6afb-7fed-9fad-c93bc5777bdd	/redfish/v1/Managers/BMC	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart"],"@Redfish.ActionInfo":"","target":"/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},"EthernetNICInfo":[{"RedfishId":"2","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/2","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013ED292","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:d2:93"},{"RedfishId":"3","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/3","Description":"Network Interface on the Baseboard Management Controller","FQDN":"x3000c0s19b4","Hostname":"BMCA4BF013ED292","InterfaceEnabled":true,"MACAddress":"a4:bf:01:3e:d2:94"},{"RedfishId":"1","@odata.id":"/redfish/v1/Managers/BMC/EthernetInterfaces/1","Description":"Network Interface on the Baseboard Management Controller","Hostname":"BMCA4BF013ED292","InterfaceEnabled":false,"MACAddress":"a4:bf:01:3e:d2:92"}]}
x3000c0s27e0	NodeEnclosure		Chassis	RackMount	x3000c0s27b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s27b0n0	Node		ComputerSystem	Physical	x3000c0s27b0		70518000-5ab2-11eb-8000-b42e99a52339	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["GracefulShutdown","On","ForceRestart","ForceOff"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"b4:2e:99:a5:23:39"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b4:2e:99:a5:23:3a"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":65535,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
x3000c0s27b0	NodeBMC		Manager	BMC	x3000c0s27b0	c6:7d:b3:54:6b:bd	407fdb6e-debf-0010-e603-b42e99a5233b	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","FQDN":"AMIB42E99A5233B.hmn","Hostname":"AMIB42E99A5233B","InterfaceEnabled":true,"MACAddress":"b4:2e:99:a5:23:3b","PermanentMACAddress":"b4:2e:99:a5:23:3b"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","FQDN":"AMIB42E99A5233B.hmn","Hostname":"AMIB42E99A5233B","InterfaceEnabled":true,"MACAddress":"c6:7d:b3:54:6b:bd","PermanentMACAddress":"c6:7d:b3:54:6b:bd"}]}
x3000c0s9b0n0	Node		ComputerSystem	Physical	x3000c0s9b0		61df0000-9855-11ed-8000-b42e993a2604	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["ForceRestart","GracefulShutdown","On","ForceOff"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:d9:9e:38"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b8:59:9f:d9:9e:39"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"98:03:9b:aa:95:e8"},{"RedfishId":"5","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/5","Description":"Ethernet Interface Lan5","MACAddress":"b4:2e:99:3a:26:04"},{"RedfishId":"6","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/6","Description":"Ethernet Interface Lan6","MACAddress":"b4:2e:99:3a:26:05"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"98:03:9b:aa:94:bc"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
x3000c0s9b0	NodeBMC		Manager	BMC	x3000c0s9b0	ea:b7:b5:d8:71:2b	407fdb6e-debf-0010-e603-b42e993a2606	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","Hostname":"AMIB42E993A2606","InterfaceEnabled":true,"MACAddress":"b4:2e:99:3a:26:06","PermanentMACAddress":"b4:2e:99:3a:26:06"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","Hostname":"AMIB42E993A2606","InterfaceEnabled":true,"MACAddress":"ea:b7:b5:d8:71:2b","PermanentMACAddress":"ea:b7:b5:d8:71:2b"}]}
x3000c0s11b0n0	Node		ComputerSystem	Physical	x3000c0s11b0		61df0000-9855-11ed-8000-e0d55e659162	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["ForceOff","On","GracefulShutdown","ForceRestart"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b8:59:9f:1d:d8:e3"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"98:03:9b:7f:bd:20"},{"RedfishId":"5","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/5","Description":"Ethernet Interface Lan5","MACAddress":"e0:d5:5e:65:91:62"},{"RedfishId":"6","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/6","Description":"Ethernet Interface Lan6","MACAddress":"e0:d5:5e:65:91:63"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"98:03:9b:7f:bd:1c"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:1d:d8:e2"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
x3000c0s11b0	NodeBMC		Manager	BMC	x3000c0s11b0	92:2f:1e:7e:bc:91	808cde6e-debf-0010-e703-e0d55e659164	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","FQDN":"AMIE0D55E659164.hmn","Hostname":"AMIE0D55E659164","InterfaceEnabled":true,"MACAddress":"e0:d5:5e:65:91:64","PermanentMACAddress":"e0:d5:5e:65:91:64"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","FQDN":"AMIE0D55E659164.hmn","Hostname":"AMIE0D55E659164","InterfaceEnabled":true,"MACAddress":"92:2f:1e:7e:bc:91","PermanentMACAddress":"92:2f:1e:7e:bc:91"}]}
x3000c0s5b0n0	Node		ComputerSystem	Physical	x3000c0s5b0		61df0000-9855-11ed-8000-b42e99a52285	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["ForceOff","On","GracefulShutdown","ForceRestart"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"b8:59:9f:c7:12:52"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:c7:12:53"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b4:2e:99:a5:22:85"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"b4:2e:99:a5:22:86"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/Self"},{"@odata.id":"/redfish/v1/Chassis/Self"}]}]}
x3000c0s5b0	NodeBMC		Manager	BMC	x3000c0s5b0	e6:cb:58:c7:aa:f1	e005dd6e-debf-0010-e603-b42e99a52287	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","Hostname":"AMIB42E99A52287","InterfaceEnabled":true,"MACAddress":"b4:2e:99:a5:22:87","PermanentMACAddress":"b4:2e:99:a5:22:87"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","Hostname":"AMIB42E99A52287","InterfaceEnabled":true,"MACAddress":"e6:cb:58:c7:aa:f1","PermanentMACAddress":"e6:cb:58:c7:aa:f1"}]}
x3000c0s3e0	NodeEnclosure		Chassis	RackMount	x3000c0s3b0			/redfish/v1/Chassis/Self	{"Name":"Computer System Chassis","Actions":{"#Chassis.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Chassis/Self/ResetActionInfo","target":"/redfish/v1/Chassis/Self/Actions/Chassis.Reset"}}}
x3000c0s3b0n0	Node		ComputerSystem	Physical	x3000c0s3b0		61df0000-9855-11ed-8000-b42e99a52271	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["ForceOff","GracefulShutdown","ForceRestart","On"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"b4:2e:99:a5:22:72"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"b8:59:9f:be:8f:2e"},{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"b8:59:9f:be:8f:2f"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b4:2e:99:a5:22:71"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
x3000c0s3b0	NodeBMC		Manager	BMC	x3000c0s3b0	42:a0:ee:51:42:ad	808cde6e-debf-0010-e603-b42e99a52273	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","Hostname":"AMIB42E99A52273","InterfaceEnabled":true,"MACAddress":"b4:2e:99:a5:22:73","PermanentMACAddress":"b4:2e:99:a5:22:73"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","Hostname":"AMIB42E99A52273","InterfaceEnabled":true,"MACAddress":"42:a0:ee:51:42:ad","PermanentMACAddress":"42:a0:ee:51:42:ad"}]}
x3000c0s17b0	NodeBMC		Manager	BMC	x3000c0s17b0	46:22:37:c9:0a:b2	e005dd6e-debf-0010-e603-b42e993a25fa	/redfish/v1/Managers/Self	{"Name":"Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":null,"@Redfish.ActionInfo":"/redfish/v1/Managers/Self/ResetActionInfo","target":"/redfish/v1/Managers/Self/Actions/Manager.Reset"},"Oem":{}},"EthernetNICInfo":[{"RedfishId":"bond0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/bond0","Description":"Ethernet Interface bond0","Hostname":"AMIB42E993A25FA","InterfaceEnabled":true,"MACAddress":"b4:2e:99:3a:25:fa","PermanentMACAddress":"b4:2e:99:3a:25:fa"},{"RedfishId":"usb0","@odata.id":"/redfish/v1/Managers/Self/EthernetInterfaces/usb0","Description":"Ethernet Interface usb0","Hostname":"AMIB42E993A25FA","InterfaceEnabled":true,"MACAddress":"46:22:37:c9:0a:b2","PermanentMACAddress":"46:22:37:c9:0a:b2"}]}
x3000c0s13b0n0	Node		ComputerSystem	Physical	x3000c0s13b0		61df0000-9855-11ed-8000-b42e99ab2592	/redfish/v1/Systems/Self	{"Name":"System","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["GracefulShutdown","ForceOff","ForceRestart","On"],"@Redfish.ActionInfo":"/redfish/v1/Systems/Self/ResetActionInfo","target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},"EthernetNICInfo":[{"RedfishId":"2","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/2","Description":"Ethernet Interface Lan2","MACAddress":"98:03:9b:b4:27:f7"},{"RedfishId":"3","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/3","Description":"Ethernet Interface Lan3","MACAddress":"b4:2e:99:ab:25:92"},{"RedfishId":"4","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/4","Description":"Ethernet Interface Lan4","MACAddress":"b4:2e:99:ab:25:93"},{"RedfishId":"1","@odata.id":"/redfish/v1/Systems/Self/EthernetInterfaces/1","Description":"Ethernet Interface Lan1","MACAddress":"98:03:9b:b4:27:f6"}],"PowerURL":"/redfish/v1/Chassis/Self/Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/Self/Power#/PowerControl/0","MemberId":"0","Name":"Chassis Power Control","PowerCapacityWatts":1600,"OEM":{},"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/Self"},{"@odata.id":"/redfish/v1/Systems/Self"}]}]}
\.


--
-- Data for Name: comp_eth_interfaces; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.comp_eth_interfaces (id, description, macaddr, last_update, compid, comptype, ip_addresses) FROM stdin;
0ee70fd98ce2	Ethernet Interface usb0	0e:e7:0f:d9:8c:e2	2023-07-17 07:50:10.869938+00	x3000c0s5b0	NodeBMC	[]
dee1eb777820	Ethernet Interface usb0	de:e1:eb:77:78:20	2023-07-17 07:50:10.941309+00	x3000c0s13b0	NodeBMC	[]
e0d55e659164	Ethernet Interface bond0	e0:d5:5e:65:91:64	2023-07-17 07:58:27.260353+00	x3000c0s11b0	NodeBMC	[]
a2b75b11b721	Ethernet Interface usb0	a2:b7:5b:11:b7:21	2023-07-17 07:50:11.707794+00	x3000c0s3b0	NodeBMC	[]
b42e99a522e5	Ethernet Interface Lan5	b4:2e:99:a5:22:e5	2023-07-17 10:08:14.27254+00	x3000c0s29b0n0	Node	[]
367a31dfe971	Ethernet Interface usb0	36:7a:31:df:e9:71	2023-07-17 07:50:11.790407+00	x3000c0s17b0	NodeBMC	[]
fe8315a3bbcf	Ethernet Interface usb0	fe:83:15:a3:bb:cf	2023-07-17 07:50:13.183452+00	x3000c0s7b0	NodeBMC	[]
a6685fdf950d	Ethernet Interface usb0	a6:68:5f:df:95:0d	2023-07-17 07:50:13.826083+00	x3000c0s15b0	NodeBMC	[]
b42e99a522e6	Ethernet Interface Lan6	b4:2e:99:a5:22:e6	2023-07-17 10:08:14.292085+00	x3000c0s29b0n0	Node	[]
506b4b23a7c4	Ethernet Interface Lan1	50:6b:4b:23:a7:c4	2023-07-17 10:08:14.152101+00	x3000c0s29b0n0	Node	[]
b8599fd99e81	Ethernet Interface Lan3	b8:59:9f:d9:9e:81	2023-07-17 10:08:14.428885+00	x3000c0s29b0n0	Node	[]
0040a6830845		00:40:a6:83:08:45	2023-07-17 07:57:05.017518+00	x3000c0r24b0	RouterBMC	[{"IPAddress":"10.254.1.23"}]
506b4b28505c	Ethernet Interface Lan4	50:6b:4b:28:50:5c	2023-07-17 10:08:14.168478+00	x3000c0s29b0n0	Node	[]
a4bf013ed294		a4:bf:01:3e:d2:94	2023-07-17 07:57:06.337829+00	x3000c0s19b4	NodeBMC	[{"IPAddress":"10.254.1.26"}]
b42e99a522e7	Ethernet Interface bond0	b4:2e:99:a5:22:e7	2023-07-17 07:50:32.088155+00	x3000c0s29b0	NodeBMC	[]
a4bf013ec02e		a4:bf:01:3e:c0:2e	2023-07-17 07:57:05.498879+00	x3000c0s19b2	NodeBMC	[{"IPAddress":"10.254.1.24"}]
ae246aae72e9	Ethernet Interface usb0	ae:24:6a:ae:72:e9	2023-07-17 07:58:27.260353+00	x3000c0s11b0	NodeBMC	[]
b8599fc71243	Ethernet Interface Lan2	b8:59:9f:c7:12:43	2023-07-17 10:08:15.201359+00	x3000c0s17b0n0	Node	[]
98039baa960c	Ethernet Interface Lan1	98:03:9b:aa:96:0c	2023-07-17 10:08:15.538836+00	x3000c0s7b0n0	Node	[]
a4bf013ecf6a		a4:bf:01:3e:cf:6a	2023-07-17 07:57:05.903494+00	x3000c0s19b3	NodeBMC	[{"IPAddress":"10.254.1.25"}]
b8599fd99d30	Bond0 - bond0.nmn0- kea	b8:59:9f:d9:9d:30	2023-07-17 10:10:02.812851+00	x3000c0s7b0n0	Node	[{"IPAddress":"10.252.1.7"},{"IPAddress":"10.102.3.22"},{"IPAddress":"10.254.1.10"},{"IPAddress":"10.1.1.5"}]
b8599fd99d31	Ethernet Interface Lan3	b8:59:9f:d9:9d:31	2023-07-17 10:08:15.716359+00	x3000c0s7b0n0	Node	[]
98039baa94a8	Ethernet Interface Lan4	98:03:9b:aa:94:a8	2023-07-17 10:08:15.554532+00	x3000c0s7b0n0	Node	[]
a4bf013ed1fe		a4:bf:01:3e:d1:fe	2023-07-17 07:57:06.706296+00	x3000c0s19b1	NodeBMC	[{"IPAddress":"10.254.1.27"}]
a4bf013ecf66	System NIC 1	a4:bf:01:3e:cf:66	2023-07-17 08:15:51.917565+00	x3000c0s19b3n0	Node	[]
b42e99a52272	Ethernet Interface Lan4	b4:2e:99:a5:22:72	2023-07-17 10:08:16.648041+00	x3000c0s3b0n0	Node	[]
b8599fc71253	Ethernet Interface Lan2	b8:59:9f:c7:12:53	2023-07-17 10:08:15.988347+00	x3000c0s5b0n0	Node	[]
b42e993a2619	Ethernet Interface Lan4	b4:2e:99:3a:26:19	2023-07-17 10:08:16.411855+00	x3000c0s15b0n0	Node	[]
b42e993a2604	Ethernet Interface Lan5	b4:2e:99:3a:26:04	2023-07-17 10:08:16.15421+00	x3000c0s9b0n0	Node	[]
b42e993a25f8	Ethernet Interface Lan3	b4:2e:99:3a:25:f8	2023-07-17 10:08:15.102806+00	x3000c0s17b0n0	Node	[]
98039bb427f7	Ethernet Interface Lan2	98:03:9b:b4:27:f7	2023-07-17 10:08:14.988057+00	x3000c0s13b0n0	Node	[]
b42e99ab2592	Ethernet Interface Lan3	b4:2e:99:ab:25:92	2023-07-17 10:08:14.888675+00	x3000c0s13b0n0	Node	[]
b42e99ab2593	Ethernet Interface Lan4	b4:2e:99:ab:25:93	2023-07-17 10:08:14.909949+00	x3000c0s13b0n0	Node	[]
98039bb427f6	Bond0 - bond0.nmn0- kea	98:03:9b:b4:27:f6	2023-07-17 10:10:02.888768+00	x3000c0s13b0n0	Node	[{"IPAddress":"10.252.1.11"},{"IPAddress":"10.1.1.9"},{"IPAddress":"10.102.3.26"},{"IPAddress":"10.254.1.18"}]
b8599f1dd8e3	Ethernet Interface Lan3	b8:59:9f:1d:d8:e3	2023-07-17 10:08:14.768917+00	x3000c0s11b0n0	Node	[]
b8599fbe8f2f	Ethernet Interface Lan2	b8:59:9f:be:8f:2f	2023-07-17 10:08:16.763978+00	x3000c0s3b0n0	Node	[]
b42e99a52271	Ethernet Interface Lan3	b4:2e:99:a5:22:71	2023-07-17 10:08:16.627731+00	x3000c0s3b0n0	Node	[]
b42e993a25f9	Ethernet Interface Lan4	b4:2e:99:3a:25:f9	2023-07-17 10:08:15.122583+00	x3000c0s17b0n0	Node	[]
b42e993a25fa	Ethernet Interface bond0	b4:2e:99:3a:25:fa	2023-07-17 07:50:11.790407+00	x3000c0s17b0	NodeBMC	[]
b42e99ab2594	Ethernet Interface bond0	b4:2e:99:ab:25:94	2023-07-17 07:50:10.941309+00	x3000c0s13b0	NodeBMC	[]
b42e99a52285	Ethernet Interface Lan3	b4:2e:99:a5:22:85	2023-07-17 10:08:15.873352+00	x3000c0s5b0n0	Node	[]
b42e99a52286	Ethernet Interface Lan4	b4:2e:99:a5:22:86	2023-07-17 10:08:15.893312+00	x3000c0s5b0n0	Node	[]
b8599fc71122	Bond0 - bond0.nmn0- kea	b8:59:9f:c7:11:22	2023-07-17 10:10:02.904897+00	x3000c0s15b0n0	Node	[{"IPAddress":"10.252.1.12"},{"IPAddress":"10.102.3.27"},{"IPAddress":"10.254.1.20"},{"IPAddress":"10.1.1.10"}]
98039b7fbd20	Ethernet Interface Lan4	98:03:9b:7f:bd:20	2023-07-17 10:08:14.567349+00	x3000c0s11b0n0	Node	[]
e0d55e659162	Ethernet Interface Lan5	e0:d5:5e:65:91:62	2023-07-17 10:08:14.640357+00	x3000c0s11b0n0	Node	[]
b42e99a52273	Ethernet Interface bond0	b4:2e:99:a5:22:73	2023-07-17 07:50:11.707794+00	x3000c0s3b0	NodeBMC	[]
b8599fc71123	Ethernet Interface Lan2	b8:59:9f:c7:11:23	2023-07-17 10:08:16.480973+00	x3000c0s15b0n0	Node	[]
b42e993a2618	Ethernet Interface Lan3	b4:2e:99:3a:26:18	2023-07-17 10:08:16.392702+00	x3000c0s15b0n0	Node	[]
9e388bad6121	Ethernet Interface usb0	9e:38:8b:ad:61:21	2023-07-17 07:50:32.088155+00	x3000c0s29b0	NodeBMC	[]
e0d55e659163	Ethernet Interface Lan6	e0:d5:5e:65:91:63	2023-07-17 10:08:14.658093+00	x3000c0s11b0n0	Node	[]
b42e993a261a	Ethernet Interface bond0	b4:2e:99:3a:26:1a	2023-07-17 07:50:13.826083+00	x3000c0s15b0	NodeBMC	[]
98039b7fbd1c	Ethernet Interface Lan1	98:03:9b:7f:bd:1c	2023-07-17 10:08:14.544262+00	x3000c0s11b0n0	Node	[]
b42e99a52265	Ethernet Interface Lan5	b4:2e:99:a5:22:65	2023-07-17 10:08:15.583584+00	x3000c0s7b0n0	Node	[]
b42e99a52266	Ethernet Interface Lan6	b4:2e:99:a5:22:66	2023-07-17 10:08:15.602483+00	x3000c0s7b0n0	Node	[]
b42e99a52267	Ethernet Interface bond0	b4:2e:99:a5:22:67	2023-07-17 07:50:13.183452+00	x3000c0s7b0	NodeBMC	[]
a4bf013ed1fb	System NIC 2	a4:bf:01:3e:d1:fb	2023-07-17 08:15:51.879539+00	x3000c0s19b1n0	Node	[]
a4bf013ed1fc	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:d1:fc	2023-07-17 08:15:51.879539+00	x3000c0s19b1	NodeBMC	[]
a4bf013ed1fd	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:d1:fd	2023-07-17 08:15:51.879539+00	x3000c0s19b1	NodeBMC	[]
a4bf013ec02a	System NIC 1	a4:bf:01:3e:c0:2a	2023-07-17 08:15:52.124232+00	x3000c0s19b2n0	Node	[]
a4bf013ec02b	System NIC 2	a4:bf:01:3e:c0:2b	2023-07-17 08:15:52.124232+00	x3000c0s19b2n0	Node	[]
a4bf013ec02c	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:c0:2c	2023-07-17 08:15:52.124232+00	x3000c0s19b2	NodeBMC	[]
a4bf013ec02d	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:c0:2d	2023-07-17 08:15:52.124232+00	x3000c0s19b2	NodeBMC	[]
ffffffffffff	System NIC 3	ff:ff:ff:ff:ff:ff	2023-07-17 08:15:51.879539+00	x3000c0s19b4n0	Node	[]
b42e99a5233b		b4:2e:99:a5:23:3b	2023-07-17 08:00:07.985452+00	x3000c0s27b0	NodeBMC	[{"IPAddress":"10.254.1.30"}]
a4bf013ed1fa	System NIC 1	a4:bf:01:3e:d1:fa	2023-07-17 09:51:59.045624+00	x3000c0s19b1n0	Node	[{"IPAddress":"10.252.1.22"}]
6230e3c24be6	CSI Handoff MAC	62:30:e3:c2:4b:e6	2023-07-17 10:08:14.377026+00	x3000c0s29b0n0	Node	[]
4a68a3c931ee	Ethernet Interface usb0	4a:68:a3:c9:31:ee	2023-07-17 07:50:14.878267+00	x3000c0s9b0	NodeBMC	[]
b42e993a2609	CSI Handoff MAC	b4:2e:99:3a:26:09	2023-07-17 10:08:15.332577+00	x3000c0s1b0n0	Node	[]
42a0ee5142ad	Ethernet Interface usb0	42:a0:ee:51:42:ad	2023-07-18 09:37:13.396207+00	x3000c0s3b0	NodeBMC	[]
9a1869f27cc8	Ethernet Interface usb0	9a:18:69:f2:7c:c8	2023-07-18 09:37:16.880437+00	x3000c0s7b0	NodeBMC	[]
7a90e7172b90	CSI Handoff MAC	7a:90:e7:17:2b:90	2023-07-17 10:08:15.659221+00	x3000c0s7b0n0	Node	[]
b8599fc712f2	Bond0 - bond0.nmn0- kea	b8:59:9f:c7:12:f2	2023-07-17 10:10:02.749154+00	x3000c0s1b0n0	Node	[{"IPAddress":"10.252.1.4"},{"IPAddress":"10.254.1.4"},{"IPAddress":"10.1.1.2"},{"IPAddress":"10.102.3.19"}]
b8599fd99e39	Ethernet Interface Lan3	b8:59:9f:d9:9e:39	2023-07-17 10:08:16.28192+00	x3000c0s9b0n0	Node	[]
98039baa95e8	Ethernet Interface Lan4	98:03:9b:aa:95:e8	2023-07-17 10:08:16.116595+00	x3000c0s9b0n0	Node	[]
b42e993a2605	Ethernet Interface Lan6	b4:2e:99:3a:26:05	2023-07-17 10:08:16.173599+00	x3000c0s9b0n0	Node	[]
98039baa94bc	Ethernet Interface Lan1	98:03:9b:aa:94:bc	2023-07-17 10:08:16.095862+00	x3000c0s9b0n0	Node	[]
b42e993a2606	Ethernet Interface bond0	b4:2e:99:3a:26:06	2023-07-17 07:50:14.878267+00	x3000c0s9b0	NodeBMC	[]
eab7b5d8712b	Ethernet Interface usb0	ea:b7:b5:d8:71:2b	2023-07-18 10:31:05.01076+00	x3000c0s9b0	NodeBMC	[]
a4bf013ecf67	System NIC 2	a4:bf:01:3e:cf:67	2023-07-17 08:15:51.917565+00	x3000c0s19b3n0	Node	[]
a4bf013ecf68	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:cf:68	2023-07-17 08:15:51.917565+00	x3000c0s19b3	NodeBMC	[]
a4bf013ecf69	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:cf:69	2023-07-17 08:15:51.917565+00	x3000c0s19b3	NodeBMC	[]
a4bf013ed290	System NIC 1	a4:bf:01:3e:d2:90	2023-07-17 08:15:52.441723+00	x3000c0s19b4n0	Node	[]
a4bf013ed291	System NIC 2	a4:bf:01:3e:d2:91	2023-07-17 08:15:52.441723+00	x3000c0s19b4n0	Node	[]
a4bf013ed293	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:d2:93	2023-07-17 08:15:52.441723+00	x3000c0s19b4	NodeBMC	[]
a4bf013ed292	Network Interface on the Baseboard Management Controller	a4:bf:01:3e:d2:92	2023-07-17 08:15:52.441723+00	x3000c0s19b4	NodeBMC	[]
b42e99a52339	Ethernet Interface Lan1	b4:2e:99:a5:23:39	2023-07-17 08:16:07.525938+00	x3000c0s27b0n0	Node	[]
b42e99a5233a	Ethernet Interface Lan2	b4:2e:99:a5:23:3a	2023-07-17 08:16:07.525938+00	x3000c0s27b0n0	Node	[]
c67db3546bbd	Ethernet Interface usb0	c6:7d:b3:54:6b:bd	2023-07-17 08:16:07.525938+00	x3000c0s27b0	NodeBMC	[]
220b3bc18341	CSI Handoff MAC	22:0b:3b:c1:83:41	2023-07-17 10:08:15.83732+00	x3000c0s5b0n0	Node	[]
f685e8353413	CSI Handoff MAC	f6:85:e8:35:34:13	2023-07-17 10:08:15.945121+00	x3000c0s5b0n0	Node	[]
96ed4cbf77c4	CSI Handoff MAC	96:ed:4c:bf:77:c4	2023-07-17 10:08:16.591166+00	x3000c0s3b0n0	Node	[]
4e3cfa1773a8	CSI Handoff MAC	4e:3c:fa:17:73:a8	2023-07-17 10:08:16.707911+00	x3000c0s3b0n0	Node	[]
b8599fc71252	Bond0 - bond0.nmn0- kea	b8:59:9f:c7:12:52	2023-07-17 10:10:02.790131+00	x3000c0s5b0n0	Node	[{"IPAddress":"10.252.1.6"},{"IPAddress":"10.102.3.21"},{"IPAddress":"10.254.1.8"},{"IPAddress":"10.1.1.4"}]
b42e99a52287	Ethernet Interface bond0	b4:2e:99:a5:22:87	2023-07-17 07:50:10.869938+00	x3000c0s5b0	NodeBMC	[]
e6cb58c7aaf1	Ethernet Interface usb0	e6:cb:58:c7:aa:f1	2023-07-18 09:37:13.399506+00	x3000c0s5b0	NodeBMC	[]
922f1e7ebc91	Ethernet Interface usb0	92:2f:1e:7e:bc:91	2023-07-18 10:32:36.539484+00	x3000c0s11b0	NodeBMC	[]
6a27c490fe7c	CSI Handoff MAC	6a:27:c4:90:fe:7c	2023-07-17 10:08:14.228807+00	x3000c0s29b0n0	Node	[]
624523172065	CSI Handoff MAC	62:45:23:17:20:65	2023-07-17 10:08:14.392779+00	x3000c0s29b0n0	Node	[]
3eed4ea57b93	CSI Handoff MAC	3e:ed:4e:a5:7b:93	2023-07-17 10:08:14.715335+00	x3000c0s11b0n0	Node	[]
c22945bce3dd	CSI Handoff MAC	c2:29:45:bc:e3:dd	2023-07-17 10:08:15.567787+00	x3000c0s7b0n0	Node	[]
f2f4274c7608	CSI Handoff MAC	f2:f4:27:4c:76:08	2023-07-17 10:08:15.67667+00	x3000c0s7b0n0	Node	[]
d6012911a7d9	CSI Handoff MAC	d6:01:29:11:a7:d9	2023-07-17 10:08:15.855465+00	x3000c0s5b0n0	Node	[]
3a3b033804d3	CSI Handoff MAC	3a:3b:03:38:04:d3	2023-07-17 10:08:15.960363+00	x3000c0s5b0n0	Node	[]
2e2b0a211d49	CSI Handoff MAC	2e:2b:0a:21:1d:49	2023-07-17 10:08:16.232414+00	x3000c0s9b0n0	Node	[]
462237c90ab2	Ethernet Interface usb0	46:22:37:c9:0a:b2	2023-07-18 09:37:30.196086+00	x3000c0s17b0	NodeBMC	[]
92ddc6324010	Ethernet Interface usb0	92:dd:c6:32:40:10	2023-07-18 09:37:39.987735+00	x3000c0s13b0	NodeBMC	[]
32d48666c87b	Ethernet Interface usb0	32:d4:86:66:c8:7b	2023-07-18 09:37:31.484297+00	x3000c0s15b0	NodeBMC	[]
dad76d7ed241	Ethernet Interface usb0	da:d7:6d:7e:d2:41	2023-07-18 11:01:59.973536+00	x3000c0s29b0	NodeBMC	[]
b8599fc71242	Bond0 - bond0.nmn0- kea	b8:59:9f:c7:12:42	2023-07-17 10:10:02.927461+00	x3000c0s17b0n0	Node	[{"IPAddress":"10.252.1.13"},{"IPAddress":"10.254.1.22"},{"IPAddress":"10.1.1.11"},{"IPAddress":"10.102.3.28"}]
36ef24b4b9cf	CSI Handoff MAC	36:ef:24:b4:b9:cf	2023-07-17 10:08:14.602896+00	x3000c0s11b0n0	Node	[]
5626dbbafae8	CSI Handoff MAC	56:26:db:ba:fa:e8	2023-07-17 10:08:14.732736+00	x3000c0s11b0n0	Node	[]
b42e993a2608	CSI Handoff MAC	b4:2e:99:3a:26:08	2023-07-17 10:08:15.351921+00	x3000c0s1b0n0	Node	[]
b8599fc712f3	CSI Handoff MAC	b8:59:9f:c7:12:f3	2023-07-17 10:08:15.432782+00	x3000c0s1b0n0	Node	[]
82c8ef683ce0	CSI Handoff MAC	82:c8:ef:68:3c:e0	2023-07-17 10:08:16.136474+00	x3000c0s9b0n0	Node	[]
4648f023f486	CSI Handoff MAC	46:48:f0:23:f4:86	2023-07-17 10:08:16.248384+00	x3000c0s9b0n0	Node	[]
b8599f1dd8e2	Bond0 - bond0.nmn0- kea	b8:59:9f:1d:d8:e2	2023-07-17 10:10:02.850892+00	x3000c0s11b0n0	Node	[{"IPAddress":"10.252.1.9"},{"IPAddress":"10.102.3.24"},{"IPAddress":"10.254.1.14"},{"IPAddress":"10.1.1.7"}]
ceb82023ca56	CSI Handoff MAC	ce:b8:20:23:ca:56	2023-07-17 10:08:16.609404+00	x3000c0s3b0n0	Node	[]
ee1dc394a27d	CSI Handoff MAC	ee:1d:c3:94:a2:7d	2023-07-17 10:08:16.723579+00	x3000c0s3b0n0	Node	[]
b8599fbe8f2e	Bond0 - bond0.nmn0- kea	b8:59:9f:be:8f:2e	2023-07-17 10:10:02.770124+00	x3000c0s3b0n0	Node	[{"IPAddress":"10.252.1.5"},{"IPAddress":"10.254.1.6"},{"IPAddress":"10.1.1.3"},{"IPAddress":"10.102.3.20"}]
b8599fd99e38	Bond0 - bond0.nmn0- kea	b8:59:9f:d9:9e:38	2023-07-17 10:10:02.831047+00	x3000c0s9b0n0	Node	[{"IPAddress":"10.252.1.8"},{"IPAddress":"10.254.1.12"},{"IPAddress":"10.1.1.6"},{"IPAddress":"10.102.3.23"}]
b8599fd99e80	Bond0 - bond0.nmn0- kea	b8:59:9f:d9:9e:80	2023-07-17 10:10:02.870474+00	x3000c0s29b0n0	Node	[{"IPAddress":"10.252.1.10"},{"IPAddress":"10.254.1.16"},{"IPAddress":"10.1.1.8"},{"IPAddress":"10.102.3.25"}]
\.


--
-- Data for Name: component_group_members; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.component_group_members (component_id, group_id, group_namespace, joined_at) FROM stdin;
\.


--
-- Data for Name: component_groups; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.component_groups (id, name, description, tags, annotations, type, namespace, exclusive_group_identifier) FROM stdin;
\.


--
-- Data for Name: components; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.components (id, type, state, admin, enabled, flag, role, nid, subtype, nettype, arch, disposition, subrole, class, reservation_disabled, locked) FROM stdin;
x3000c0w22	MgmtSwitch	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s9e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s9b0n0	Node	Ready		t	OK	Management	100005		Sling	X86		Worker	River	f	t
x3000c0r24e0	HSNBoard	On		t	OK		-1		Sling	X86			River	f	f
x3000c0r24b0	RouterBMC	Ready		t	OK		-1		Sling	X86			River	f	f
x3000c0s19e1	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s19b1n0	Node	On		t	OK	Compute	1		Sling	X86			River	f	f
x3000c0s19b1	NodeBMC	Ready		t	OK		-1		Sling	X86			River	f	f
x3000c0s19e3	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s19b3n0	Node	On		t	OK	Compute	3		Sling	X86			River	f	f
x3000c0s19b3	NodeBMC	Ready		t	OK		-1		Sling	X86			River	f	f
x3000c0s19e2	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s19b2n0	Node	On		t	OK	Compute	2		Sling	X86			River	f	f
x3000c0s19b2	NodeBMC	Ready		t	OK		-1		Sling	X86			River	f	f
x3000c0s19e4	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s19b4n0	Node	On		t	OK	Compute	4		Sling	X86			River	f	f
x3000c0s19b4	NodeBMC	Ready		t	OK		-1		Sling	X86			River	f	f
x3000c0s27e0	NodeEnclosure	Off		t	OK		-1		Sling	X86			River	f	f
x3000c0s27b0n0	Node	Off		t	OK	Application	49169248		Sling	X86		UAN	River	f	f
x3000c0s27b0	NodeBMC	Ready		t	OK		-1		Sling	X86			River	f	f
x3000c0s1b0n0	Node	Populated		t	OK	Management	100001		Sling	X86		Master	River	f	t
x3000c0s9b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s11e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s11b0n0	Node	Ready		t	OK	Management	100006		Sling	X86		Worker	River	f	t
x3000c0s11b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s29e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s29b0n0	Node	Ready		t	OK	Management	100007		Sling	X86		Worker	River	f	t
x3000c0s29b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s13b0n0	Node	Ready		t	OK	Management	100008		Sling	X86		Storage	River	f	t
x3000c0s15b0n0	Node	Ready		t	OK	Management	100009		Sling	X86		Storage	River	f	t
x3000c0s17b0n0	Node	Ready		t	OK	Management	100010		Sling	X86		Storage	River	f	t
x3000c0s5e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s5b0n0	Node	Ready		t	OK	Management	100003		Sling	X86		Master	River	f	t
x3000c0s5b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s3e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s3b0n0	Node	Ready		t	OK	Management	100002		Sling	X86		Master	River	f	t
x3000c0s3b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s7e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s7b0n0	Node	Ready		t	OK	Management	100004		Sling	X86		Worker	River	f	t
x3000c0s7b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s17e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s17b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s13e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s13b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
x3000c0s15e0	NodeEnclosure	On		t	OK		-1		Sling	X86			River	f	f
x3000c0s15b0	NodeBMC	Ready		t	OK	Management	-1		Sling	X86			River	f	t
\.


--
-- Data for Name: discovery_status; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.discovery_status (id, status, last_update, details) FROM stdin;
0	NotStarted	2023-07-17 07:39:13.12565+00	{}
\.


--
-- Data for Name: hsn_interfaces; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.hsn_interfaces (nic, macaddr, hsn, node, ipaddr, last_update) FROM stdin;
\.


--
-- Data for Name: hwinv_by_fru; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.hwinv_by_fru (fru_id, type, subtype, serial_number, part_number, manufacturer, fru_info) FROM stdin;
Memory.SKHynix.HMA84GR7CJR4NXN.932F6A2D	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6A2D"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6953	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6953"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6924	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6924"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F692D	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F692D"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6945	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6945"}
NodeEnclosure.CrayInc.6NR272Z30MRYF110.GJK9N6612A0041	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-YF","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-YF-110","SerialNumber":"GJK9N6612A0041","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323ZP	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K323ZP","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K3226T	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K3226T","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102310803.GJK9N6612A0041	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-YF","Manufacturer":"Cray Inc.","PartNumber":"102310803","SerialNumber":"GJK9N6612A0041","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e99ab2592"}
Processor.AdvancedMicroDevicesInc.2609DBC0CDB401C	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7402 24-Core Processor","SerialNumber":"2609DBC0CDB401C","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":24,"TotalThreads":48,"Oem":null}
Memory.SKHynix.HMA84GR7CJR4NXN.3365E1B5	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365E1B5"}
Memory.SKHynix.HMA84GR7CJR4NXN.3365E1CB	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365E1CB"}
Memory.SKHynix.HMA82GR7CJR8NXN.434E6F61	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"434E6F61"}
Memory.SKHynix.HMA82GR7CJR8NXN.434E6ED3	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"434E6ED3"}
Memory.SKHynix.HMA82GR7CJR8NXN.434E6FCA	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"434E6FCA"}
FRUIDforx3000c0s13b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.SKHynix.HMA84GR7CJR4NXN.3365CE34	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365CE34"}
Memory.SKHynix.HMA84GR7CJR4NXN.3365E1D5	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365E1D5"}
Memory.SKHynix.HMA84GR7CJR4NXN.3365C810	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365C810"}
Memory.SKHynix.HMA84GR7CJR4NXN.3365E0A7	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365E0A7"}
Memory.SKHynix.HMA84GR7CJR4NXN.3365E155	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365E155"}
Processor.AdvancedMicroDevicesInc.2B48C8C1793C083	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7702 64-Core Processor","SerialNumber":"2B48C8C1793C083","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":64,"TotalThreads":128,"Oem":null}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD169	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD169"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD276	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD276"}
NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0032	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-100","SerialNumber":"GJG7N9412A0032","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324D9	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K324D9","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324EV	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K324EV","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102261800.GJG7N9412A0032	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102261800","SerialNumber":"GJG7N9412A0032","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e99a52265"}
Processor.AdvancedMicroDevicesInc.2B48D3481D1403B	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7702 64-Core Processor","SerialNumber":"2B48D3481D1403B","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":64,"TotalThreads":128,"Oem":null}
FRUIDforx3000c0s7b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.SKHynix.HMA82GR7CJR8NXN.434E6FD0	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"434E6FD0"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD148	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD148"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD267	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD267"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD14C	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD14C"}
NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0063	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-100","SerialNumber":"GJG7N9412A0063","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322TM	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K322TM","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K3245Z	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K3245Z","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102261803.GJG7N9412A0063	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102261803","SerialNumber":"GJG7N9412A0063","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e993a25f8"}
NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0041	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-100","SerialNumber":"GJG7N9412A0041","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K32433	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K32433","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323ZW	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K323ZW","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102261803.GJG7N9412A0041	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102261803","SerialNumber":"GJG7N9412A0041","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e993a2618"}
Processor.AdvancedMicroDevicesInc.2B494759233C015	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7402 24-Core Processor","SerialNumber":"2B494759233C015","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":24,"TotalThreads":48,"Oem":null}
FRUIDforx3000c0s15b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD199	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD199"}
NodeEnclosure.CrayInc.6NR272Z30MR00101.GJG9N2612A0007	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-101","SerialNumber":"GJG9N2612A0007","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322NJ	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K322NJ","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324C4	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K324C4","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102319800.GJG9N2612A0007	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102319800","SerialNumber":"GJG9N2612A0007","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e99a522e5"}
Processor.AdvancedMicroDevicesInc.2B48C8C1793C060	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7702 64-Core Processor","SerialNumber":"2B48C8C1793C060","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":64,"TotalThreads":128,"Oem":null}
Memory.Samsung.M393A4K40DB3CWE.037EA5AA	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037EA5AA"}
Memory.Samsung.M393A4K40DB3CWE.037FD96B	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037FD96B"}
Memory.Samsung.M393A4K40DB3CWE.037FDAC8	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037FDAC8"}
Memory.Samsung.M393A4K40DB3CWE.037FDA9F	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037FDA9F"}
FRUIDforx3000c0s29b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
MgmtSwitch.DELL.CN0J4T5KCES008770102	MgmtSwitch					{"AssetTag":"2C9ZZP2","ChassisType":"Drawer","Model":"S3048ON","Manufacturer":"DELL","PartNumber":"","SerialNumber":"CN0J4T5KCES008770102","SKU":""}
NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0064	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-100","SerialNumber":"GJG7N9412A0064","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324ET	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K324ET","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322MZ	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K322MZ","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102261800.GJG7N9412A0064	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102261800","SerialNumber":"GJG7N9412A0064","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e993a2604"}
Processor.AdvancedMicroDevicesInc.2B48D3481D1403D	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7702 64-Core Processor","SerialNumber":"2B48D3481D1403D","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":64,"TotalThreads":128,"Oem":null}
Memory.SKHynix.HMA84GR7CJR4NXN.932F69EA	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F69EA"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6A24	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6A24"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6A2F	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6A2F"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103I252FJ	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103I252FJ","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
HSNBoard.HPE.101878104A.BC19190015	HSNBoard					{"AssetTag":"","ChassisType":"Enclosure","Model":"HPE_Slingshot_1_Top-of-Rack_Switch","Manufacturer":"HPE","PartNumber":"101878104.A","SerialNumber":"BC19190015","SKU":""}
FRUIDforx3000c0r24b0	RouterBMC					{"ManagerType":"EnclosureManager","Model":"","Manufacturer":"","PartNumber":"","SerialNumber":""}
NodeEnclosure.IntelCorporation.QSBP82704191	NodeEnclosure					{"AssetTag":"","ChassisType":"RackMount","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"..................","SerialNumber":"QSBP82704191","SKU":""}
Node.IntelCorporation.102072300.QSBP82704191	Node					{"AssetTag":"....................","BiosVersion":"SE5C620.86B.02.01.0010.C0001.010620200716","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"102072300","SerialNumber":"QSBP82704191","SKU":"....................","SystemType":"Physical","UUID":"b996c5f4-82a9-11e8-ab21-a4bf013ed1fa"}
FRUIDforx3000c0s19b1n0p0	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
FRUIDforx3000c0s19b1n0p1	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA48	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA48"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA18	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA18"}
NodeEnclosure.IntelCorporation.QSBP82704059	NodeEnclosure					{"AssetTag":"","ChassisType":"RackMount","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"..................","SerialNumber":"QSBP82704059","SKU":""}
Memory.Hynix.HMA42GR7AFR4NUH.512DC9B2	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DC9B2"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA54	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA54"}
FRUIDforx3000c0s19b1	NodeBMC					{"ManagerType":"BMC","Model":"S2600BPB","Manufacturer":"","PartNumber":"","SerialNumber":""}
NodeEnclosure.IntelCorporation.QSBP82703289	NodeEnclosure					{"AssetTag":"","ChassisType":"RackMount","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"..................","SerialNumber":"QSBP82703289","SKU":""}
Node.IntelCorporation.102072300.QSBP82703289	Node					{"AssetTag":"....................","BiosVersion":"SE5C620.86B.02.01.0010.C0001.010620200716","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"102072300","SerialNumber":"QSBP82703289","SKU":"....................","SystemType":"Physical","UUID":"a4cd681d-8064-11e8-ab21-a4bf013ec02a"}
FRUIDforx3000c0s19b2n0p0	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
FRUIDforx3000c0s19b2n0p1	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA3A	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA3A"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA16	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA16"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA4F	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA4F"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA40	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA40"}
FRUIDforx3000c0s19b2	NodeBMC					{"ManagerType":"BMC","Model":"S2600BPB","Manufacturer":"","PartNumber":"","SerialNumber":""}
Node.IntelCorporation.102072300.QSBP82704059	Node					{"AssetTag":"....................","BiosVersion":"SE5C620.86B.02.01.0010.C0001.010620200716","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"102072300","SerialNumber":"QSBP82704059","SKU":"....................","SystemType":"Physical","UUID":"3811a2bf-8140-11e8-ab21-a4bf013ecf66"}
FRUIDforx3000c0s19b3n0p0	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
FRUIDforx3000c0s19b3n0p1	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA4E	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA4E"}
Memory.Hynix.HMA42GR7AFR4NUH.512DC82C	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DC82C"}
Memory.Hynix.HMA42GR7AFR4NUH.512DC968	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DC968"}
Memory.Hynix.HMA42GR7AFR4NUH.512DC9BF	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DC9BF"}
FRUIDforx3000c0s19b3	NodeBMC					{"ManagerType":"BMC","Model":"S2600BPB","Manufacturer":"","PartNumber":"","SerialNumber":""}
NodeEnclosure.IntelCorporation.QSBP82704221	NodeEnclosure					{"AssetTag":"","ChassisType":"RackMount","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"..................","SerialNumber":"QSBP82704221","SKU":""}
Node.IntelCorporation.102072300.QSBP82704221	Node					{"AssetTag":"....................","BiosVersion":"SE5C620.86B.02.01.0010.C0001.010620200716","Model":"S2600BPB","Manufacturer":"Intel Corporation","PartNumber":"102072300","SerialNumber":"QSBP82704221","SKU":"....................","SystemType":"Physical","UUID":"c96d020d-8193-11e8-ab21-a4bf013ed290"}
FRUIDforx3000c0s19b4n0p0	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
FRUIDforx3000c0s19b4n0p1	Processor					{"InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel Xeon processor","SerialNumber":"","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"0xb3","EffectiveModel":"0x3","IdentificationRegisters":"50-65-4","MicrocodeInfo":"","Step":"","VendorID":"Intel(R) Xeon(R) Silver 4108 CPU @ 1.80GHz"},"ProcessorType":"CPU","TotalCores":8,"TotalThreads":8,"Oem":{}}
Memory.Hynix.HMA42GR7AFR4NUH.512DC9FE	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DC9FE"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA09	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA09"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA4A	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA4A"}
Memory.Hynix.HMA42GR7AFR4NUH.512DCA35	Memory					{"BaseModuleType":"RDIMM","BusWidthBits":72,"CapacityMiB":16384,"DataWidthBits":64,"ErrorCorrection":"MultiBitECC","Manufacturer":"Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":2400,"PartNumber":"HMA42GR7AFR4N-UH    ","RankCount":2,"SerialNumber":"512DCA35"}
FRUIDforx3000c0s19b4	NodeBMC					{"ManagerType":"BMC","Model":"S2600BPB","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.SKHynix.HMA84GR7CJR4NXN.932F69E8	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F69E8"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324AQ	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K324AQ","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322LL	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K322LL","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.GIGABYTE.000000000001.GJG9N2612A0006	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C17","Model":"R272-Z30-00","Manufacturer":"GIGABYTE","PartNumber":"000000000001","SerialNumber":"GJG9N2612A0006","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"70518000-5ab2-11eb-8000-b42e99a52339"}
Processor.AdvancedMicroDevicesInc.2B494759233C0A6	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7402 24-Core Processor                ","SerialNumber":"2B494759233C0A6","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":24,"TotalThreads":48,"Oem":null}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD0A4	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD0A4"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD0AB	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD0AB"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DCFD8	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DCFD8"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD0B3	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD0B3"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DCF95	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DCF95"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DCF89	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DCF89"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD09F	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD09F"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD09E	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD09E"}
FRUIDforx3000c0s27b0n0g1k1	Drive					{"Manufacturer":"","SerialNumber":"S45PNA0MC45466      ","PartNumber":"","Model":"SAMSUNG MZ7LH480HAHQ-00005","SKU":"","CapacityBytes":503424483328,"Protocol":"","MediaType":"","RotationSpeedRPM":0,"BlockSizeBytes":0,"CapableSpeedGbs":0,"FailurePredicted":false,"EncryptionAbility":"","EncryptionStatus":"","NegotiatedSpeedGbs":0,"PredictedMediaLifeLeftPercent":0}
FRUIDforx3000c0s27b0n0g1k2	Drive					{"Manufacturer":"","SerialNumber":"S45PNA0MC45465      ","PartNumber":"","Model":"SAMSUNG MZ7LH480HAHQ-00005","SKU":"","CapacityBytes":503424483328,"Protocol":"","MediaType":"","RotationSpeedRPM":0,"BlockSizeBytes":0,"CapableSpeedGbs":0,"FailurePredicted":false,"EncryptionAbility":"","EncryptionStatus":"","NegotiatedSpeedGbs":0,"PredictedMediaLifeLeftPercent":0}
FRUIDforx3000c0s27b0n0g1k3	Drive					{"Manufacturer":"","SerialNumber":"S455NY0MB52363      ","PartNumber":"","Model":"SAMSUNG MZ7LH1T9HMLT-00005","SKU":"","CapacityBytes":2013667524608,"Protocol":"","MediaType":"","RotationSpeedRPM":0,"BlockSizeBytes":0,"CapableSpeedGbs":0,"FailurePredicted":false,"EncryptionAbility":"","EncryptionStatus":"","NegotiatedSpeedGbs":0,"PredictedMediaLifeLeftPercent":0}
FRUIDforx3000c0s27b0n0g1k4	Drive					{"Manufacturer":"","SerialNumber":"S455NY0MB51690      ","PartNumber":"","Model":"SAMSUNG MZ7LH1T9HMLT-00005","SKU":"","CapacityBytes":2013667524608,"Protocol":"","MediaType":"","RotationSpeedRPM":0,"BlockSizeBytes":0,"CapableSpeedGbs":0,"FailurePredicted":false,"EncryptionAbility":"","EncryptionStatus":"","NegotiatedSpeedGbs":0,"PredictedMediaLifeLeftPercent":0}
FRUIDforx3000c0s27b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
NodeEnclosure.GIGABYTE.01234567.01234567890123456789AB	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"GIGABYTE","PartNumber":"01234567","SerialNumber":"01234567890123456789AB","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103I252A9	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103I252A9","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.GIGABYTE.000000000001.01234567890123456789AB	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"GIGABYTE","PartNumber":"000000000001","SerialNumber":"01234567890123456789AB","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-e0d55e659162"}
Processor.AdvancedMicroDevicesInc.2B47E9FEEE9C013	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3200,"Model":"AMD Eng Sample: 100-000000053-04_32/20_N","SerialNumber":"2B47E9FEEE9C013","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":64,"TotalThreads":128,"Oem":null}
Memory.SKHynix.HMA84GR7CJR4NXN.3365E039	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"3365E039"}
FRUIDforx3000c0s11b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.SKHynix.HMA84GR7CJR4NXN.932F69E1	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F69E1"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6998	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6998"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6A30	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6A30"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6911	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6911"}
FRUIDforx3000c0s9b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0042	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-100","SerialNumber":"GJG7N9412A0042","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K32429	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K32429","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323YY	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K323YY","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102261700.GJG7N9412A0042	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102261700","SerialNumber":"GJG7N9412A0042","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e99a52285"}
Processor.AdvancedMicroDevicesInc.2B494759233C0A2	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7402 24-Core Processor","SerialNumber":"2B494759233C0A2","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":24,"TotalThreads":48,"Oem":null}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD26C	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD26C"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD146	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD146"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD14D	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD14D"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD1F3	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD1F3"}
FRUIDforx3000c0s5b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.Samsung.M393A4K40DB3CWE.037EA7F8	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037EA7F8"}
Memory.Samsung.M393A4K40DB3CWE.037FDA3D	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037FDA3D"}
Memory.Samsung.M393A4K40DB3CWE.037EA83B	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037EA83B"}
Memory.Samsung.M393A4K40DB3CWE.037EA7F2	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"Samsung","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"M393A4K40DB3-CWE    ","SerialNumber":"037EA7F2"}
NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0037	NodeEnclosure					{"AssetTag":"01234567890123456789AB","ChassisType":"RackMount","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"6NR272Z30MR-00-100","SerialNumber":"GJG7N9412A0037","SKU":"01234567890123456789AB"}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324G7	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K324G7","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322H5	NodeEnclosurePowerSupply					{"Manufacturer":"Liteon Power","SerialNumber":"6K9L10103K322H5","Model":"PS-2801-9L1","PartNumber":"","PowerCapacityWatts":800,"PowerInputWatts":969,"PowerOutputWatts":800,"PowerSupplyType":""}
Node.CrayInc.102261700.GJG7N9412A0037	Node					{"AssetTag":"Free form asset tag","BiosVersion":"C37","Model":"R272-Z30-00","Manufacturer":"Cray Inc.","PartNumber":"102261700","SerialNumber":"GJG7N9412A0037","SKU":"01234567890123456789AB","SystemType":"Physical","UUID":"61df0000-9855-11ed-8000-b42e99a52271"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD10F	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD10F"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD108	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD108"}
FRUIDforx3000c0s3b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
Memory.SKHynix.HMA84GR7CJR4NXN.932F697B	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F697B"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F693D	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F693D"}
Memory.SKHynix.HMA84GR7CJR4NXN.932F6A0A	Memory					{"BusWidthBits":48,"CapacityMiB":31249,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA84GR7CJR4N-XN    ","SerialNumber":"932F6A0A"}
Processor.AdvancedMicroDevicesInc.2B494759233C0A7	Processor					{"InstructionSet":"x86-64","Manufacturer":"Advanced Micro Devices, Inc.","MaxSpeedMHz":3350,"Model":"AMD EPYC 7402 24-Core Processor","SerialNumber":"2B494759233C0A7","PartNumber":"","ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"AMD Zen Processor Family","EffectiveModel":"0x31","IdentificationRegisters":"178bfbff00830f10","MicrocodeInfo":"","Step":"0x0","VendorID":"AuthenticAMD"},"ProcessorType":"CPU","TotalCores":24,"TotalThreads":48,"Oem":null}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD281	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD281"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD147	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD147"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD205	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD205"}
Memory.SKHynix.HMA82GR7CJR8NXN.533DD26E	Memory					{"BusWidthBits":48,"CapacityMiB":15625,"DataWidthBits":40,"ErrorCorrection":"MultiBitECC","Manufacturer":"SK Hynix","MemoryType":"DRAM","MemoryDeviceType":"DDR4","OperatingSpeedMhz":3200,"PartNumber":"HMA82GR7CJR8N-XN    ","SerialNumber":"533DD26E"}
FRUIDforx3000c0s17b0	NodeBMC					{"ManagerType":"BMC","Model":"410410600","Manufacturer":"","PartNumber":"","SerialNumber":""}
\.


--
-- Data for Name: hwinv_by_loc; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.hwinv_by_loc (id, type, ordinal, status, parent, location_info, fru_id, parent_node) FROM stdin;
x3000c0s5e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0042	x3000c0s5e0
x3000c0s5e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K32429	x3000c0s5e0t0
x3000c0s5e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323YY	x3000c0s5e0t1
x3000c0s5b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7402 24-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":61}}	Node.CrayInc.102261700.GJG7N9412A0042	x3000c0s5b0n0
x3000c0s5b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B494759233C0A2	x3000c0s5b0n0
x3000c0s5b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d5	Memory	5	Empty		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD26C	x3000c0s5b0n0
x3000c0s5b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD146	x3000c0s5b0n0
x3000c0s5b0n0d7	Memory	7	Empty		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d1	Memory	1	Empty		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d3	Memory	3	Empty		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD14D	x3000c0s5b0n0
x3000c0s5b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s5b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD1F3	x3000c0s5b0n0
x3000c0s5b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s5b0n0
x3000c0s13e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MRYF110.GJK9N6612A0041	x3000c0s13e0
x3000c0s9b0n0d1	Memory	1	Populated		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F69E1	x3000c0s9b0n0
x3000c0s9b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d5	Memory	5	Populated		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6998	x3000c0s9b0n0
x3000c0s9b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F69EA	x3000c0s9b0n0
x3000c0s9b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A24	x3000c0s9b0n0
x3000c0s9b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A30	x3000c0s9b0n0
x3000c0s9b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d3	Memory	3	Populated		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6911	x3000c0s9b0n0
x3000c0s9b0n0d7	Memory	7	Populated		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A2F	x3000c0s9b0n0
x3000c0s9b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:22:59+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s9b0	x3000c0s9b0
x3000c0s5b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:23:33+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s5b0	x3000c0s5b0
x3000c0s3b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD10F	x3000c0s3b0n0
x3000c0s3b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD108	x3000c0s3b0n0
x3000c0s3b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD276	x3000c0s3b0n0
x3000c0s3b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d3	Memory	3	Empty		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:23:29+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s3b0	x3000c0s3b0
x3000c0s13b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.434E6F61	x3000c0s13b0n0
x3000c0s13b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d1	Memory	1	Empty		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.434E6FD0	x3000c0s13b0n0
x3000c0s13b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:24:21+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s13b0	x3000c0s13b0
x3000c0s15e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0041	x3000c0s15e0
x3000c0s15e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323ZW	x3000c0s15e0t1
x3000c0s9e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0064	x3000c0s9e0
x3000c0s9e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324ET	x3000c0s9e0t0
x3000c0s9e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322MZ	x3000c0s9e0t1
x3000c0s9b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7702 64-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":244}}	Node.CrayInc.102261800.GJG7N9412A0064	x3000c0s9b0n0
x3000c0s9b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B48D3481D1403D	x3000c0s9b0n0
x3000c0s9b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F69E8	x3000c0s9b0n0
x3000c0s9b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s9b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s9b0n0
x3000c0s7e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0032	x3000c0s7e0
x3000c0s7e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324D9	x3000c0s7e0t0
x3000c0s7e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324EV	x3000c0s7e0t1
x3000c0s7b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7702 64-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":244}}	Node.CrayInc.102261800.GJG7N9412A0032	x3000c0s7b0n0
x3000c0s7b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B48D3481D1403B	x3000c0s7b0n0
x3000c0s7b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F693D	x3000c0s7b0n0
x3000c0s7b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A2D	x3000c0s7b0n0
x3000c0s7b0n0d1	Memory	1	Populated		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6953	x3000c0s7b0n0
x3000c0s7b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6945	x3000c0s7b0n0
x3000c0s7b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s17b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s13e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323ZP	x3000c0s13e0t1
x3000c0s13e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K3226T	x3000c0s13e0t0
x3000c0s13b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7402 24-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":61}}	Node.CrayInc.102310803.GJK9N6612A0041	x3000c0s13b0n0
x3000c0s13b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2609DBC0CDB401C	x3000c0s13b0n0
x3000c0s13b0n0d3	Memory	3	Empty		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.434E6ED3	x3000c0s13b0n0
x3000c0s13b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.434E6FCA	x3000c0s13b0n0
x3000c0s29b0n0d1	Memory	1	Populated		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037FDA9F	x3000c0s29b0n0
x3000c0s29b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037EA83B	x3000c0s29b0n0
x3000c0s29b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037EA7F2	x3000c0s29b0n0
x3000c0s29b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:52:53+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s29b0	x3000c0s29b0
x3000c0s7b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F697B	x3000c0s7b0n0
x3000c0s7b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0n0d7	Memory	7	Populated		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A0A	x3000c0s7b0n0
x3000c0s7b0n0d3	Memory	3	Populated		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F6924	x3000c0s7b0n0
x3000c0s7b0n0d5	Memory	5	Populated		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.932F692D	x3000c0s7b0n0
x3000c0s7b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s7b0n0
x3000c0s7b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:23:00+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s7b0	x3000c0s7b0
x3000c0s13b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d5	Memory	5	Empty		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s13b0n0d7	Memory	7	Empty		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s13b0n0
x3000c0s15e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K32433	x3000c0s15e0t0
x3000c0s15b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7402 24-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":61}}	Node.CrayInc.102261803.GJG7N9412A0041	x3000c0s15b0n0
x3000c0s15b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B494759233C015	x3000c0s15b0n0
x3000c0s15b0n0d5	Memory	5	Empty		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d7	Memory	7	Empty		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD267	x3000c0s15b0n0
x3000c0s15b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD14C	x3000c0s15b0n0
x3000c0s15b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d3	Memory	3	Empty		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD199	x3000c0s15b0n0
x3000c0s11e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.GIGABYTE.01234567.01234567890123456789AB	x3000c0s11e0
x3000c0s11e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.48"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103I252A9	x3000c0s11e0t0
x3000c0s29e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00101.GJG9N2612A0007	x3000c0s29e0
x3000c0s29e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322NJ	x3000c0s29e0t1
x3000c0s29e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324C4	x3000c0s29e0t0
x3000c0s29b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7702 64-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":244}}	Node.CrayInc.102319800.GJG9N2612A0007	x3000c0s29b0n0
x3000c0s29b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B48C8C1793C060	x3000c0s29b0n0
x3000c0s29b0n0d3	Memory	3	Populated		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037FD96B	x3000c0s29b0n0
x3000c0w22	MgmtSwitch	0	Populated		{"Id":"EthernetSwitch_0","Name":"DellEMC Networking S3048ON Chassis","Description":"","HostName":""}	MgmtSwitch.DELL.CN0J4T5KCES008770102	x3000c0w22
x3000c0r24e0	HSNBoard	0	Populated		{"Id":"Enclosure","Name":"Enclosure","Description":"HPE_Slingshot_1_Top-of-Rack_Switch","HostName":""}	HSNBoard.HPE.101878104A.BC19190015	x3000c0r24e0
x3000c0r24b0	RouterBMC	0	Populated		{"DateTime":"2023-07-17T08:15:44+00:00","DateTimeLocalOffset":"+00:00","Description":"Shasta Manager","FirmwareVersion":"","Id":"BMC","Name":"BMC"}	FRUIDforx3000c0r24b0	x3000c0r24b0
x3000c0s19e3	NodeEnclosure	0	Populated		{"Id":"RackMount","Name":"Computer System Chassis","Description":"System Chassis","HostName":""}	NodeEnclosure.IntelCorporation.QSBP82704059	x3000c0s19e3
x3000c0s19b3n0	Node	0	Populated		{"Id":"QSBP82704059","Name":"S2600BPB","Description":"Computer system providing compute resources","HostName":"","ProcessorSummary":{"Count":2,"Model":"Central Processor"},"MemorySummary":{"TotalSystemMemoryGiB":64}}	Node.IntelCorporation.102072300.QSBP82704059	x3000c0s19b3n0
x3000c0s19b3n0p0	Processor	0	Populated		{"Id":"CPU1","Name":"Processor 1","Description":"","Socket":"CPU 1"}	FRUIDforx3000c0s19b3n0p0	x3000c0s19b3n0
x3000c0s19b3n0p1	Processor	1	Populated		{"Id":"CPU2","Name":"Processor 2","Description":"","Socket":"CPU 2"}	FRUIDforx3000c0s19b3n0p1	x3000c0s19b3n0
x3000c0s19b3n0d3	Memory	3	Populated		{"Id":"Memory4","Name":"Memory 4","Description":"System Memory","MemoryLocation":{"Socket":12,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA4E	x3000c0s19b3n0
x3000c0s19b3n0d0	Memory	0	Populated		{"Id":"Memory1","Name":"Memory 1","Description":"System Memory","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DC82C	x3000c0s19b3n0
x3000c0s19b3n0d1	Memory	1	Populated		{"Id":"Memory2","Name":"Memory 2","Description":"System Memory","MemoryLocation":{"Socket":4,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DC968	x3000c0s19b3n0
x3000c0s19b3n0d2	Memory	2	Populated		{"Id":"Memory3","Name":"Memory 3","Description":"System Memory","MemoryLocation":{"Socket":8,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DC9BF	x3000c0s19b3n0
x3000c0s19b3n0g1k0	Drive	0	Empty		{"Id":"HDD1","Name":"HDD 0 Status","Description":""}	\N	x3000c0s19b3n0
x3000c0s19b3n0g1k1	Drive	1	Empty		{"Id":"HDD2","Name":"HDD 1 Status","Description":""}	\N	x3000c0s19b3n0
x3000c0s19b3n0g1k2	Drive	2	Empty		{"Id":"HDD3","Name":"HDD 2 Status","Description":""}	\N	x3000c0s19b3n0
x3000c0s19b3	NodeBMC	0	Populated		{"DateTime":"2023-07-17T08:10:07+00:00","DateTimeLocalOffset":"","Description":"Baseboard Management Controller","FirmwareVersion":"2.37.1f190479","Id":"BMC","Name":"Manager"}	FRUIDforx3000c0s19b3	x3000c0s19b3
x3000c0s19e1	NodeEnclosure	0	Populated		{"Id":"RackMount","Name":"Computer System Chassis","Description":"System Chassis","HostName":""}	NodeEnclosure.IntelCorporation.QSBP82704191	x3000c0s19e1
x3000c0s19b1n0	Node	0	Populated		{"Id":"QSBP82704191","Name":"S2600BPB","Description":"Computer system providing compute resources","HostName":"","ProcessorSummary":{"Count":2,"Model":"Central Processor"},"MemorySummary":{"TotalSystemMemoryGiB":64}}	Node.IntelCorporation.102072300.QSBP82704191	x3000c0s19b1n0
x3000c0s19b1n0p0	Processor	0	Populated		{"Id":"CPU1","Name":"Processor 1","Description":"","Socket":"CPU 1"}	FRUIDforx3000c0s19b1n0p0	x3000c0s19b1n0
x3000c0s19b1n0p1	Processor	1	Populated		{"Id":"CPU2","Name":"Processor 2","Description":"","Socket":"CPU 2"}	FRUIDforx3000c0s19b1n0p1	x3000c0s19b1n0
x3000c0s19b1n0d1	Memory	1	Populated		{"Id":"Memory2","Name":"Memory 2","Description":"System Memory","MemoryLocation":{"Socket":4,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA48	x3000c0s19b1n0
x3000c0s19b1n0d2	Memory	2	Populated		{"Id":"Memory3","Name":"Memory 3","Description":"System Memory","MemoryLocation":{"Socket":8,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA18	x3000c0s19b1n0
x3000c0s19b1n0d3	Memory	3	Populated		{"Id":"Memory4","Name":"Memory 4","Description":"System Memory","MemoryLocation":{"Socket":12,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DC9B2	x3000c0s19b1n0
x3000c0s19b1n0d0	Memory	0	Populated		{"Id":"Memory1","Name":"Memory 1","Description":"System Memory","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA54	x3000c0s19b1n0
x3000c0s19b1n0g1k1	Drive	1	Empty		{"Id":"HDD2","Name":"HDD 1 Status","Description":""}	\N	x3000c0s19b1n0
x3000c0s19b1n0g1k2	Drive	2	Empty		{"Id":"HDD3","Name":"HDD 2 Status","Description":""}	\N	x3000c0s19b1n0
x3000c0s19b1n0g1k0	Drive	0	Empty		{"Id":"HDD1","Name":"HDD 0 Status","Description":""}	\N	x3000c0s19b1n0
x3000c0s19b1	NodeBMC	0	Populated		{"DateTime":"2023-07-17T08:10:07+00:00","DateTimeLocalOffset":"","Description":"Baseboard Management Controller","FirmwareVersion":"2.37.1f190479","Id":"BMC","Name":"Manager"}	FRUIDforx3000c0s19b1	x3000c0s19b1
x3000c0s19e2	NodeEnclosure	0	Populated		{"Id":"RackMount","Name":"Computer System Chassis","Description":"System Chassis","HostName":""}	NodeEnclosure.IntelCorporation.QSBP82703289	x3000c0s19e2
x3000c0s19b2n0	Node	0	Populated		{"Id":"QSBP82703289","Name":"S2600BPB","Description":"Computer system providing compute resources","HostName":"","ProcessorSummary":{"Count":2,"Model":"Central Processor"},"MemorySummary":{"TotalSystemMemoryGiB":64}}	Node.IntelCorporation.102072300.QSBP82703289	x3000c0s19b2n0
x3000c0s19b2n0p0	Processor	0	Populated		{"Id":"CPU1","Name":"Processor 1","Description":"","Socket":"CPU 1"}	FRUIDforx3000c0s19b2n0p0	x3000c0s19b2n0
x3000c0s19b2n0p1	Processor	1	Populated		{"Id":"CPU2","Name":"Processor 2","Description":"","Socket":"CPU 2"}	FRUIDforx3000c0s19b2n0p1	x3000c0s19b2n0
x3000c0s19b2n0d0	Memory	0	Populated		{"Id":"Memory1","Name":"Memory 1","Description":"System Memory","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA3A	x3000c0s19b2n0
x3000c0s19b2n0d1	Memory	1	Populated		{"Id":"Memory2","Name":"Memory 2","Description":"System Memory","MemoryLocation":{"Socket":4,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA16	x3000c0s19b2n0
x3000c0s19b2n0d2	Memory	2	Populated		{"Id":"Memory3","Name":"Memory 3","Description":"System Memory","MemoryLocation":{"Socket":8,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA4F	x3000c0s19b2n0
x3000c0s19b2n0d3	Memory	3	Populated		{"Id":"Memory4","Name":"Memory 4","Description":"System Memory","MemoryLocation":{"Socket":12,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA40	x3000c0s19b2n0
x3000c0s19b2n0g1k0	Drive	0	Empty		{"Id":"HDD1","Name":"HDD 0 Status","Description":""}	\N	x3000c0s19b2n0
x3000c0s19b2n0g1k1	Drive	1	Empty		{"Id":"HDD2","Name":"HDD 1 Status","Description":""}	\N	x3000c0s19b2n0
x3000c0s19b2n0g1k2	Drive	2	Empty		{"Id":"HDD3","Name":"HDD 2 Status","Description":""}	\N	x3000c0s19b2n0
x3000c0s19b2	NodeBMC	0	Populated		{"DateTime":"2023-07-17T08:05:33+00:00","DateTimeLocalOffset":"","Description":"Baseboard Management Controller","FirmwareVersion":"2.37.1f190479","Id":"BMC","Name":"Manager"}	FRUIDforx3000c0s19b2	x3000c0s19b2
x3000c0s19e4	NodeEnclosure	0	Populated		{"Id":"RackMount","Name":"Computer System Chassis","Description":"System Chassis","HostName":""}	NodeEnclosure.IntelCorporation.QSBP82704221	x3000c0s19e4
x3000c0s19b4n0	Node	0	Populated		{"Id":"QSBP82704221","Name":"S2600BPB","Description":"Computer system providing compute resources","HostName":"","ProcessorSummary":{"Count":2,"Model":"Central Processor"},"MemorySummary":{"TotalSystemMemoryGiB":64}}	Node.IntelCorporation.102072300.QSBP82704221	x3000c0s19b4n0
x3000c0s19b4n0p0	Processor	0	Populated		{"Id":"CPU1","Name":"Processor 1","Description":"","Socket":"CPU 1"}	FRUIDforx3000c0s19b4n0p0	x3000c0s19b4n0
x3000c0s19b4n0p1	Processor	1	Populated		{"Id":"CPU2","Name":"Processor 2","Description":"","Socket":"CPU 2"}	FRUIDforx3000c0s19b4n0p1	x3000c0s19b4n0
x3000c0s19b4n0d0	Memory	0	Populated		{"Id":"Memory1","Name":"Memory 1","Description":"System Memory","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DC9FE	x3000c0s19b4n0
x3000c0s19b4n0d1	Memory	1	Populated		{"Id":"Memory2","Name":"Memory 2","Description":"System Memory","MemoryLocation":{"Socket":4,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA09	x3000c0s19b4n0
x3000c0s19b4n0d2	Memory	2	Populated		{"Id":"Memory3","Name":"Memory 3","Description":"System Memory","MemoryLocation":{"Socket":8,"MemoryController":0,"Channel":0,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA4A	x3000c0s19b4n0
x3000c0s19b4n0d3	Memory	3	Populated		{"Id":"Memory4","Name":"Memory 4","Description":"System Memory","MemoryLocation":{"Socket":12,"MemoryController":1,"Channel":3,"Slot":1}}	Memory.Hynix.HMA42GR7AFR4NUH.512DCA35	x3000c0s19b4n0
x3000c0s19b4n0g1k0	Drive	0	Empty		{"Id":"HDD1","Name":"HDD 0 Status","Description":""}	\N	x3000c0s19b4n0
x3000c0s19b4n0g1k1	Drive	1	Empty		{"Id":"HDD2","Name":"HDD 1 Status","Description":""}	\N	x3000c0s19b4n0
x3000c0s19b4n0g1k2	Drive	2	Empty		{"Id":"HDD3","Name":"HDD 2 Status","Description":""}	\N	x3000c0s19b4n0
x3000c0s19b4	NodeBMC	0	Populated		{"DateTime":"2023-07-17T07:59:52+00:00","DateTimeLocalOffset":"","Description":"Baseboard Management Controller","FirmwareVersion":"2.37.1f190479","Id":"BMC","Name":"Manager"}	FRUIDforx3000c0s19b4	x3000c0s19b4
x3000c0s27e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.GIGABYTE.01234567.01234567890123456789AB	x3000c0s27e0
x3000c0s27e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324AQ	x3000c0s27e0t0
x3000c0s27e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322LL	x3000c0s27e0t1
x3000c0s27b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7402 24-Core Processor                "},"MemorySummary":{"TotalSystemMemoryGiB":122}}	Node.GIGABYTE.000000000001.GJG9N2612A0006	x3000c0s27b0n0
x3000c0s27b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B494759233C0A6	x3000c0s27b0n0
x3000c0s27b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD0A4	x3000c0s27b0n0
x3000c0s27b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD0AB	x3000c0s27b0n0
x3000c0s27b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d1	Memory	1	Populated		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DCFD8	x3000c0s27b0n0
x3000c0s27b0n0d7	Memory	7	Populated		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD0B3	x3000c0s27b0n0
x3000c0s27b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DCF95	x3000c0s27b0n0
x3000c0s27b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DCF89	x3000c0s27b0n0
x3000c0s27b0n0d3	Memory	3	Populated		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD09F	x3000c0s27b0n0
x3000c0s27b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d5	Memory	5	Populated		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD09E	x3000c0s27b0n0
x3000c0s27b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s27b0n0
x3000c0s27b0n0g1k1	Drive	1	Populated		{"Id":"1","Name":"SAMSUNG MZ7LH480HAHQ-00005","Description":"This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation."}	FRUIDforx3000c0s27b0n0g1k1	x3000c0s27b0n0
x3000c0s27b0n0g1k2	Drive	2	Populated		{"Id":"2","Name":"SAMSUNG MZ7LH480HAHQ-00005","Description":"This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation."}	FRUIDforx3000c0s27b0n0g1k2	x3000c0s27b0n0
x3000c0s27b0n0g1k3	Drive	3	Populated		{"Id":"3","Name":"SAMSUNG MZ7LH1T9HMLT-00005","Description":"This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation."}	FRUIDforx3000c0s27b0n0g1k3	x3000c0s27b0n0
x3000c0s27b0n0g1k4	Drive	4	Populated		{"Id":"4","Name":"SAMSUNG MZ7LH1T9HMLT-00005","Description":"This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation."}	FRUIDforx3000c0s27b0n0g1k4	x3000c0s27b0n0
x3000c0s27b0	NodeBMC	0	Populated		{"DateTime":"2023-07-17T08:15:38+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.09","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s27b0	x3000c0s27b0
x3000c0s11e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103I252FJ	x3000c0s11e0t1
x3000c0s11b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD Eng Sample: 100-000000053-04_32/20_N"},"MemorySummary":{"TotalSystemMemoryGiB":244}}	Node.GIGABYTE.000000000001.01234567890123456789AB	x3000c0s11b0n0
x3000c0s11b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B47E9FEEE9C013	x3000c0s11b0n0
x3000c0s11b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365CE34	x3000c0s11b0n0
x3000c0s11b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365C810	x3000c0s11b0n0
x3000c0s11b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d3	Memory	3	Populated		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365E039	x3000c0s11b0n0
x3000c0s11b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365E1B5	x3000c0s11b0n0
x3000c0s11b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365E1CB	x3000c0s11b0n0
x3000c0s11b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d5	Memory	5	Populated		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365E1D5	x3000c0s11b0n0
x3000c0s11b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s11b0n0
x3000c0s11b0n0d1	Memory	1	Populated		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365E0A7	x3000c0s11b0n0
x3000c0s11b0n0d7	Memory	7	Populated		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA84GR7CJR4NXN.3365E155	x3000c0s11b0n0
x3000c0s11b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:23:40+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s11b0	x3000c0s11b0
x3000c0s3e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0037	x3000c0s3e0
x3000c0s3e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324G7	x3000c0s3e0t0
x3000c0s3e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322H5	x3000c0s3e0t1
x3000c0s3b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7702 64-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":61}}	Node.CrayInc.102261700.GJG7N9412A0037	x3000c0s3b0n0
x3000c0s3b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B48C8C1793C083	x3000c0s3b0n0
x3000c0s3b0n0d1	Memory	1	Empty		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d7	Memory	7	Empty		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD169	x3000c0s3b0n0
x3000c0s3b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d5	Memory	5	Empty		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s3b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s3b0n0
x3000c0s29b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037EA5AA	x3000c0s29b0n0
x3000c0s29b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037EA7F8	x3000c0s29b0n0
x3000c0s29b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s29b0n0d5	Memory	5	Populated		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037FDAC8	x3000c0s29b0n0
x3000c0s29b0n0d7	Memory	7	Populated		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.Samsung.M393A4K40DB3CWE.037FDA3D	x3000c0s29b0n0
x3000c0s29b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s29b0n0
x3000c0s17e0	NodeEnclosure	0	Populated		{"Id":"Self","Name":"Computer System Chassis","Description":"Chassis Self","HostName":""}	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0063	x3000c0s17e0
x3000c0s17e0t0	NodeEnclosurePowerSupply	0	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322TM	x3000c0s17e0t0
x3000c0s17e0t1	NodeEnclosurePowerSupply	1	Populated		{"Name":"PS-2801-9L1","FirmwareVersion":"48.46.48.51"}	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K3245Z	x3000c0s17e0t1
x3000c0s17b0n0	Node	0	Populated		{"Id":"Self","Name":"System","Description":"System Self","HostName":"","ProcessorSummary":{"Count":1,"Model":"AMD EPYC 7402 24-Core Processor"},"MemorySummary":{"TotalSystemMemoryGiB":61}}	Node.CrayInc.102261803.GJG7N9412A0063	x3000c0s17b0n0
x3000c0s17b0n0p0	Processor	0	Populated		{"Id":"1","Name":"Processor 1","Description":"Processor Instance 1","Socket":"P0"}	Processor.AdvancedMicroDevicesInc.2B494759233C0A7	x3000c0s17b0n0
x3000c0s17b0n0d6	Memory	6	Empty		{"Id":"15","Name":"Memory 15","Description":"Memory Instance 15","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d10	Memory	10	Populated		{"Id":"4","Name":"Memory 4","Description":"Memory Instance 4","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD147	x3000c0s17b0n0
x3000c0s17b0n0d5	Memory	5	Empty		{"Id":"14","Name":"Memory 14","Description":"Memory Instance 14","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d7	Memory	7	Empty		{"Id":"16","Name":"Memory 16","Description":"Memory Instance 16","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d11	Memory	11	Empty		{"Id":"5","Name":"Memory 5","Description":"Memory Instance 5","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d12	Memory	12	Populated		{"Id":"6","Name":"Memory 6","Description":"Memory Instance 6","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD281	x3000c0s17b0n0
x3000c0s17b0n0d3	Memory	3	Empty		{"Id":"12","Name":"Memory 12","Description":"Memory Instance 12","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d13	Memory	13	Empty		{"Id":"7","Name":"Memory 7","Description":"Memory Instance 7","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD205	x3000c0s17b0n0
x3000c0s17b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d2	Memory	2	Empty		{"Id":"11","Name":"Memory 11","Description":"Memory Instance 11","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d4	Memory	4	Empty		{"Id":"13","Name":"Memory 13","Description":"Memory Instance 13","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d8	Memory	8	Populated		{"Id":"2","Name":"Memory 2","Description":"Memory Instance 2","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD26E	x3000c0s17b0n0
x3000c0s17b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0n0d1	Memory	1	Empty		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s17b0n0
x3000c0s17b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:24:01+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s17b0	x3000c0s17b0
x3000c0s15b0n0d1	Memory	1	Empty		{"Id":"10","Name":"Memory 10","Description":"Memory Instance 10","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d9	Memory	9	Empty		{"Id":"3","Name":"Memory 3","Description":"Memory Instance 3","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d14	Memory	14	Populated		{"Id":"8","Name":"Memory 8","Description":"Memory Instance 8","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	Memory.SKHynix.HMA82GR7CJR8NXN.533DD148	x3000c0s15b0n0
x3000c0s15b0n0d0	Memory	0	Empty		{"Id":"1","Name":"Memory 1","Description":"Memory Instance 1","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0n0d15	Memory	15	Empty		{"Id":"9","Name":"Memory 9","Description":"Memory Instance 9","MemoryLocation":{"Socket":0,"MemoryController":0,"Channel":0,"Slot":0}}	\N	x3000c0s15b0n0
x3000c0s15b0	NodeBMC	0	Populated		{"DateTime":"2023-07-18T10:30:11+00:00","DateTimeLocalOffset":"+00:00","Description":"BMC","FirmwareVersion":"12.84.17","Id":"Self","Name":"Manager"}	FRUIDforx3000c0s15b0	x3000c0s15b0
\.


--
-- Data for Name: hwinv_hist; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.hwinv_hist (id, fru_id, event_type, "timestamp") FROM stdin;
x3000c0s5e0	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0042	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K32429	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323YY	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0n0	Node.CrayInc.102261700.GJG7N9412A0042	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0n0p0	Processor.AdvancedMicroDevicesInc.2B494759233C0A2	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0n0d14	Memory.SKHynix.HMA82GR7CJR8NXN.533DD146	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0n0d10	Memory.SKHynix.HMA82GR7CJR8NXN.533DD26C	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0n0d12	Memory.SKHynix.HMA82GR7CJR8NXN.533DD1F3	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0n0d8	Memory.SKHynix.HMA82GR7CJR8NXN.533DD14D	Detected	2023-07-17 07:50:11.022748+00
x3000c0s5b0	FRUIDforx3000c0s5b0	Detected	2023-07-17 07:50:11.022748+00
x3000c0s13e0	NodeEnclosure.CrayInc.6NR272Z30MRYF110.GJK9N6612A0041	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K3226T	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323ZP	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0n0	Node.CrayInc.102310803.GJK9N6612A0041	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0n0p0	Processor.AdvancedMicroDevicesInc.2609DBC0CDB401C	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0n0d12	Memory.SKHynix.HMA82GR7CJR8NXN.434E6FCA	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0n0d8	Memory.SKHynix.HMA82GR7CJR8NXN.434E6ED3	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0n0d14	Memory.SKHynix.HMA82GR7CJR8NXN.434E6F61	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0n0d10	Memory.SKHynix.HMA82GR7CJR8NXN.434E6FD0	Detected	2023-07-17 07:50:11.05024+00
x3000c0s13b0	FRUIDforx3000c0s13b0	Detected	2023-07-17 07:50:11.05024+00
x3000c0s3e0	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0037	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324G7	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322H5	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0n0	Node.CrayInc.102261700.GJG7N9412A0037	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0n0p0	Processor.AdvancedMicroDevicesInc.2B48C8C1793C083	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0n0d12	Memory.SKHynix.HMA82GR7CJR8NXN.533DD276	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0n0d8	Memory.SKHynix.HMA82GR7CJR8NXN.533DD10F	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0n0d10	Memory.SKHynix.HMA82GR7CJR8NXN.533DD108	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0n0d14	Memory.SKHynix.HMA82GR7CJR8NXN.533DD169	Detected	2023-07-17 07:50:11.785092+00
x3000c0s3b0	FRUIDforx3000c0s3b0	Detected	2023-07-17 07:50:11.785092+00
x3000c0s17e0	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0063	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322TM	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K3245Z	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0n0	Node.CrayInc.102261803.GJG7N9412A0063	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0n0p0	Processor.AdvancedMicroDevicesInc.2B494759233C0A7	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0n0d14	Memory.SKHynix.HMA82GR7CJR8NXN.533DD205	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0n0d8	Memory.SKHynix.HMA82GR7CJR8NXN.533DD26E	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0n0d12	Memory.SKHynix.HMA82GR7CJR8NXN.533DD281	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0n0d10	Memory.SKHynix.HMA82GR7CJR8NXN.533DD147	Detected	2023-07-17 07:50:11.863463+00
x3000c0s17b0	FRUIDforx3000c0s17b0	Detected	2023-07-17 07:50:11.863463+00
x3000c0s7e0	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0032	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324D9	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324EV	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0	Node.CrayInc.102261800.GJG7N9412A0032	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0p0	Processor.AdvancedMicroDevicesInc.2B48D3481D1403B	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d1	Memory.SKHynix.HMA84GR7CJR4NXN.932F6953	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d12	Memory.SKHynix.HMA84GR7CJR4NXN.932F6945	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d10	Memory.SKHynix.HMA84GR7CJR4NXN.932F693D	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d14	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A2D	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d5	Memory.SKHynix.HMA84GR7CJR4NXN.932F692D	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d8	Memory.SKHynix.HMA84GR7CJR4NXN.932F697B	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d3	Memory.SKHynix.HMA84GR7CJR4NXN.932F6924	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0n0d7	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A0A	Detected	2023-07-17 07:50:13.257337+00
x3000c0s7b0	FRUIDforx3000c0s7b0	Detected	2023-07-17 07:50:13.257337+00
x3000c0s15e0	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0041	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K323ZW	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K32433	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0n0	Node.CrayInc.102261803.GJG7N9412A0041	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0n0p0	Processor.AdvancedMicroDevicesInc.2B494759233C015	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0n0d8	Memory.SKHynix.HMA82GR7CJR8NXN.533DD267	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0n0d14	Memory.SKHynix.HMA82GR7CJR8NXN.533DD148	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0n0d10	Memory.SKHynix.HMA82GR7CJR8NXN.533DD199	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0n0d12	Memory.SKHynix.HMA82GR7CJR8NXN.533DD14C	Detected	2023-07-17 07:50:13.912296+00
x3000c0s15b0	FRUIDforx3000c0s15b0	Detected	2023-07-17 07:50:13.912296+00
x3000c0s9e0	NodeEnclosure.CrayInc.6NR272Z30MR00100.GJG7N9412A0064	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322MZ	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324ET	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0	Node.CrayInc.102261800.GJG7N9412A0064	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0p0	Processor.AdvancedMicroDevicesInc.2B48D3481D1403D	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d14	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A30	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d8	Memory.SKHynix.HMA84GR7CJR4NXN.932F69EA	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d12	Memory.SKHynix.HMA84GR7CJR4NXN.932F69E8	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d3	Memory.SKHynix.HMA84GR7CJR4NXN.932F6911	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d5	Memory.SKHynix.HMA84GR7CJR4NXN.932F6998	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d7	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A2F	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d1	Memory.SKHynix.HMA84GR7CJR4NXN.932F69E1	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0n0d10	Memory.SKHynix.HMA84GR7CJR4NXN.932F6A24	Detected	2023-07-17 07:50:14.95253+00
x3000c0s9b0	FRUIDforx3000c0s9b0	Detected	2023-07-17 07:50:14.95253+00
x3000c0s29e0	NodeEnclosure.CrayInc.6NR272Z30MR00101.GJG9N2612A0007	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324C4	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322NJ	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0	Node.CrayInc.102319800.GJG9N2612A0007	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0p0	Processor.AdvancedMicroDevicesInc.2B48C8C1793C060	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d3	Memory.Samsung.M393A4K40DB3CWE.037FD96B	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d12	Memory.Samsung.M393A4K40DB3CWE.037EA7F2	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d14	Memory.Samsung.M393A4K40DB3CWE.037EA83B	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d1	Memory.Samsung.M393A4K40DB3CWE.037FDA9F	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d8	Memory.Samsung.M393A4K40DB3CWE.037EA7F8	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d10	Memory.Samsung.M393A4K40DB3CWE.037EA5AA	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d7	Memory.Samsung.M393A4K40DB3CWE.037FDA3D	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0n0d5	Memory.Samsung.M393A4K40DB3CWE.037FDAC8	Detected	2023-07-17 07:50:32.165823+00
x3000c0s29b0	FRUIDforx3000c0s29b0	Detected	2023-07-17 07:50:32.165823+00
x3000c0w22	MgmtSwitch.DELL.CN0J4T5KCES008770102	Detected	2023-07-17 07:51:04.867487+00
x3000c0s11e0	NodeEnclosure.GIGABYTE.01234567.01234567890123456789AB	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103I252A9	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103I252FJ	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0	Node.GIGABYTE.000000000001.01234567890123456789AB	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0p0	Processor.AdvancedMicroDevicesInc.2B47E9FEEE9C013	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d1	Memory.SKHynix.HMA84GR7CJR4NXN.3365E0A7	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d3	Memory.SKHynix.HMA84GR7CJR4NXN.3365E039	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d8	Memory.SKHynix.HMA84GR7CJR4NXN.3365E1CB	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d14	Memory.SKHynix.HMA84GR7CJR4NXN.3365C810	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d5	Memory.SKHynix.HMA84GR7CJR4NXN.3365E1D5	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d12	Memory.SKHynix.HMA84GR7CJR4NXN.3365E1B5	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d10	Memory.SKHynix.HMA84GR7CJR4NXN.3365CE34	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0n0d7	Memory.SKHynix.HMA84GR7CJR4NXN.3365E155	Detected	2023-07-17 07:58:27.397221+00
x3000c0s11b0	FRUIDforx3000c0s11b0	Detected	2023-07-17 07:58:27.397221+00
x3000c0r24e0	HSNBoard.HPE.101878104A.BC19190015	Detected	2023-07-17 08:15:45.06076+00
x3000c0r24b0	FRUIDforx3000c0r24b0	Detected	2023-07-17 08:15:45.06076+00
x3000c0s19e1	NodeEnclosure.IntelCorporation.QSBP82704191	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0	Node.IntelCorporation.102072300.QSBP82704191	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0p0	FRUIDforx3000c0s19b1n0p0	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0p1	FRUIDforx3000c0s19b1n0p1	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0d1	Memory.Hynix.HMA42GR7AFR4NUH.512DCA48	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0d2	Memory.Hynix.HMA42GR7AFR4NUH.512DCA18	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0d3	Memory.Hynix.HMA42GR7AFR4NUH.512DC9B2	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1n0d0	Memory.Hynix.HMA42GR7AFR4NUH.512DCA54	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19b1	FRUIDforx3000c0s19b1	Detected	2023-07-17 08:15:51.973822+00
x3000c0s19e3	NodeEnclosure.IntelCorporation.QSBP82704059	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0	Node.IntelCorporation.102072300.QSBP82704059	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0p0	FRUIDforx3000c0s19b3n0p0	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0p1	FRUIDforx3000c0s19b3n0p1	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0d3	Memory.Hynix.HMA42GR7AFR4NUH.512DCA4E	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0d0	Memory.Hynix.HMA42GR7AFR4NUH.512DC82C	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0d1	Memory.Hynix.HMA42GR7AFR4NUH.512DC968	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3n0d2	Memory.Hynix.HMA42GR7AFR4NUH.512DC9BF	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19b3	FRUIDforx3000c0s19b3	Detected	2023-07-17 08:15:51.999715+00
x3000c0s19e2	NodeEnclosure.IntelCorporation.QSBP82703289	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0	Node.IntelCorporation.102072300.QSBP82703289	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0p0	FRUIDforx3000c0s19b2n0p0	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0p1	FRUIDforx3000c0s19b2n0p1	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0d0	Memory.Hynix.HMA42GR7AFR4NUH.512DCA3A	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0d1	Memory.Hynix.HMA42GR7AFR4NUH.512DCA16	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0d2	Memory.Hynix.HMA42GR7AFR4NUH.512DCA4F	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2n0d3	Memory.Hynix.HMA42GR7AFR4NUH.512DCA40	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19b2	FRUIDforx3000c0s19b2	Detected	2023-07-17 08:15:52.197818+00
x3000c0s19e4	NodeEnclosure.IntelCorporation.QSBP82704221	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0	Node.IntelCorporation.102072300.QSBP82704221	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0p0	FRUIDforx3000c0s19b4n0p0	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0p1	FRUIDforx3000c0s19b4n0p1	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0d0	Memory.Hynix.HMA42GR7AFR4NUH.512DC9FE	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0d1	Memory.Hynix.HMA42GR7AFR4NUH.512DCA09	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0d2	Memory.Hynix.HMA42GR7AFR4NUH.512DCA4A	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4n0d3	Memory.Hynix.HMA42GR7AFR4NUH.512DCA35	Detected	2023-07-17 08:15:52.525093+00
x3000c0s19b4	FRUIDforx3000c0s19b4	Detected	2023-07-17 08:15:52.525093+00
x3000c0s27e0	NodeEnclosure.GIGABYTE.01234567.01234567890123456789AB	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27e0t0	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K324AQ	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27e0t1	NodeEnclosurePowerSupply.LiteonPower.6K9L10103K322LL	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0	Node.GIGABYTE.000000000001.GJG9N2612A0006	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0p0	Processor.AdvancedMicroDevicesInc.2B494759233C0A6	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d12	Memory.SKHynix.HMA82GR7CJR8NXN.533DD0A4	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d14	Memory.SKHynix.HMA82GR7CJR8NXN.533DD0AB	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d1	Memory.SKHynix.HMA82GR7CJR8NXN.533DCFD8	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d7	Memory.SKHynix.HMA82GR7CJR8NXN.533DD0B3	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d8	Memory.SKHynix.HMA82GR7CJR8NXN.533DCF95	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d10	Memory.SKHynix.HMA82GR7CJR8NXN.533DCF89	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d3	Memory.SKHynix.HMA82GR7CJR8NXN.533DD09F	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0d5	Memory.SKHynix.HMA82GR7CJR8NXN.533DD09E	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0g1k1	FRUIDforx3000c0s27b0n0g1k1	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0g1k2	FRUIDforx3000c0s27b0n0g1k2	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0g1k3	FRUIDforx3000c0s27b0n0g1k3	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0n0g1k4	FRUIDforx3000c0s27b0n0g1k4	Detected	2023-07-17 08:16:07.619952+00
x3000c0s27b0	FRUIDforx3000c0s27b0	Detected	2023-07-17 08:16:07.619952+00
\.


--
-- Data for Name: job_state_rf_poll; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.job_state_rf_poll (comp_id, job_id) FROM stdin;
\.


--
-- Data for Name: job_sync; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.job_sync (id, type, status, last_update, lifetime) FROM stdin;
\.


--
-- Data for Name: node_nid_mapping; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.node_nid_mapping (id, nid, role, name, node_info, subrole) FROM stdin;
\.


--
-- Data for Name: power_mapping; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.power_mapping (id, powered_by) FROM stdin;
\.


--
-- Data for Name: reservations; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.reservations (component_id, create_timestamp, expiration_timestamp, deputy_key, reservation_key) FROM stdin;
\.


--
-- Data for Name: rf_endpoints; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.rf_endpoints (id, type, name, hostname, domain, fqdn, ip_info, enabled, uuid, "user", password, usessdp, macrequired, macaddr, rediscoveronupdate, templateid, discovery_info, ipaddr) FROM stdin;
x3000c0s5b0	NodeBMC				x3000c0s5b0	{}	t	e005dd6e-debf-0010-e603-b42e99a52287	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:30:35.967171Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0w22	MgmtSwitch		x3000c0w22-rts:8083		x3000c0w22-rts:8083	{}	t		testuser		f	f		t		{"LastDiscoveryAttempt":"2023-07-17T07:51:04.789183Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"2019.1"}	
x3000c0s3b0	NodeBMC				x3000c0s3b0	{}	t	808cde6e-debf-0010-e603-b42e99a52273	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:30:37.858831Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s29b0	NodeBMC				x3000c0s29b0	{}	t	b42e99a5-22e7-cd03-0010-debf00f6456e	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T11:01:59.889157Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0r24b0	RouterBMC		x3000c0r24b0		x3000c0r24b0	{}	t		root		f	f	0040a6830845	t		{"LastDiscoveryAttempt":"2023-07-17T08:15:44.969492Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.2.0"}	
x3000c0s7b0	NodeBMC				x3000c0s7b0	{}	t	e005dd6e-debf-0010-e603-b42e99a52267	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:30:37.953203Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s19b1	NodeBMC		x3000c0s19b1		x3000c0s19b1	{}	t	029649c5-9569-41a8-8858-ab6f513d83f1	root		f	f	a4bf013ed1fe	t		{"LastDiscoveryAttempt":"2023-07-17T08:15:51.821000Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s19b3	NodeBMC		x3000c0s19b3		x3000c0s19b3	{}	t	4970d119-e1b1-4b9f-8002-39504b1869c7	root		f	f	a4bf013ecf6a	t		{"LastDiscoveryAttempt":"2023-07-17T08:15:51.879791Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s19b2	NodeBMC		x3000c0s19b2		x3000c0s19b2	{}	t	6143d866-fd48-468e-957b-72dd317d14bc	root		f	f	a4bf013ec02e	t		{"LastDiscoveryAttempt":"2023-07-17T08:15:52.077303Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s19b4	NodeBMC		x3000c0s19b4		x3000c0s19b4	{}	t	14007c57-da40-4f23-ae14-066e5117733c	root		f	f	a4bf013ed294	t		{"LastDiscoveryAttempt":"2023-07-17T08:15:52.428301Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s27b0	NodeBMC		x3000c0s27b0		x3000c0s27b0	{}	t	407fdb6e-debf-0010-e603-b42e99a5233b	root		f	f	b42e99a5233b	t		{"LastDiscoveryAttempt":"2023-07-17T08:16:07.506403Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s1b0	NodeBMC				x3000c0s1b0	{}	t		root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:30:38.256442Z","LastDiscoveryStatus":"HTTPsGetFailed"}	
x3000c0s17b0	NodeBMC				x3000c0s17b0	{}	t	e005dd6e-debf-0010-e603-b42e993a25fa	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:30:54.530377Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s13b0	NodeBMC				x3000c0s13b0	{}	t	b42e99ab-2594-cb03-0010-debf40e8916d	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:30:56.672469Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s15b0	NodeBMC				x3000c0s15b0	{}	t	808cde6e-debf-0010-e603-b42e993a261a	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:31:00.143863Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s9b0	NodeBMC				x3000c0s9b0	{}	t	407fdb6e-debf-0010-e603-b42e993a2606	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:31:04.994820Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
x3000c0s11b0	NodeBMC				x3000c0s11b0	{}	t	808cde6e-debf-0010-e703-e0d55e659164	root		f	f		t		{"LastDiscoveryAttempt":"2023-07-18T10:32:36.524238Z","LastDiscoveryStatus":"DiscoverOK","RedfishVersion":"1.7.0"}	
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.schema_migrations (version, dirty) FROM stdin;
22	f
\.


--
-- Data for Name: scn_subscriptions; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.scn_subscriptions (id, sub_url, subscription) FROM stdin;
1	cray-hmnfd-7c5b475bcc-gg87w_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-7c5b475bcc-gg87w_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
2	cray-hmnfd-7c5b475bcc-xk9qp_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-7c5b475bcc-xk9qp_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
3	cray-hmnfd-7c5b475bcc-qzfdl_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-7c5b475bcc-qzfdl_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
4	cray-hmnfd-7c5b475bcc-xk9qp_2http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-7c5b475bcc-xk9qp_2","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
5	cray-hmnfd-7c5b475bcc-qzfdl_2http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-7c5b475bcc-qzfdl_2","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
6	cray-hmnfd-65ff69dff-wrnfm_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-wrnfm_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
7	cray-hmnfd-65ff69dff-n67xx_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-n67xx_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
8	cray-hmnfd-65ff69dff-s6b7w_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-s6b7w_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
9	cray-hmnfd-65ff69dff-wrnfm_2http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-wrnfm_2","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
10	cray-hmnfd-65ff69dff-n67xx_2http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-n67xx_2","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
39	cray-hmnfd-65ff69dff-jvxjb_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-jvxjb_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
40	cray-hmnfd-65ff69dff-8mp9n_4http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-8mp9n_4","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
41	cray-hmnfd-65ff69dff-8mp9n_5http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-8mp9n_5","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
42	cray-hmnfd-65ff69dff-zzg7n_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-zzg7n_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
43	cray-hmnfd-65ff69dff-jvxjb_2http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-jvxjb_2","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
76	cray-hmnfd-65ff69dff-8th7v_1http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-8th7v_1","Enabled":true,"Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
77	cray-hmnfd-65ff69dff-8th7v_2http://cray-hmnfd/hmi/v1/scn	{"Subscriber":"cray-hmnfd-65ff69dff-8th7v_2","Enabled":true,"States":["on","off","empty","unknown","populated"],"Url":"http://cray-hmnfd/hmi/v1/scn"}
\.


--
-- Data for Name: service_endpoints; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.service_endpoints (rf_endpoint_id, redfish_type, redfish_subtype, uuid, odata_id, service_info) FROM stdin;
x3000c0s29b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:52:37+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s29b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1675808873\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s5b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:23:20+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s5b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1675800152\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s13b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s3b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:23:17+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s3b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1675795381\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s7b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s7b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s7b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s7b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:22:46+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s7b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1664840069\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s15b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s15b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s15b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s17b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s17b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s17b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s13b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s13b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s11b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s11b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0w22	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService(AccountLockoutCounterResetAfter,AccountLockoutDuration,AccountLockoutThreshold,Accounts,MaxPasswordLength,MinPasswordLength,Roles,ServiceEnabled)","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"","Name":"","Description":"","Status":{"Health":""},"ServiceEnabled":true,"AuthFailureLoggingThreshold":0,"MinPasswordLength":8,"AccountLockoutThreshold":0,"AccountLockoutDuration":0,"AccountLockoutCounterResetAfter":0,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s29b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s29b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s29b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0r24b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_2_1.AccountService","Id":"AccountService","Name":"Account Service","Description":"BMC User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0r24b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_3.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0r24b0	EventService			/redfish/v1/EventService	{"@odata.context":"","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_0_5.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":["StatusChange","ResourceUpdated","ResourceAdded","ResourceRemoved","Alert"],"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":""}}}
x3000c0r24b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_0.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-17T08:15:44Z","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0r24b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"","@odata.etag":"W/\\"1550244732\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_2_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate","title":"SimpleUpdate"}}}
x3000c0s19b3	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"BMC User Accounts","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":0,"MinPasswordLength":6,"AccountLockoutThreshold":0,"AccountLockoutDuration":0,"AccountLockoutCounterResetAfter":0,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s19b3	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"SessionService","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":1800,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s19b3	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":""},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s19b3	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":""},"ServiceEnabled":true,"DateTime":"","CompletedTaskOverWritePolicy":"Manual","LifeCycleEventOnTaskStateChange":null,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s19b3	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","SoftwareInventory":{"@odata.id":"/redfish/v1/UpdateService/SoftwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"}}}
x3000c0s19b4	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"BMC User Accounts","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":0,"MinPasswordLength":6,"AccountLockoutThreshold":0,"AccountLockoutDuration":0,"AccountLockoutCounterResetAfter":0,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s19b4	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"SessionService","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":1800,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s19b4	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":""},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s19b4	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":""},"ServiceEnabled":true,"DateTime":"","CompletedTaskOverWritePolicy":"Manual","LifeCycleEventOnTaskStateChange":null,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s19b1	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"BMC User Accounts","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":0,"MinPasswordLength":6,"AccountLockoutThreshold":0,"AccountLockoutDuration":0,"AccountLockoutCounterResetAfter":0,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s19b1	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"SessionService","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":1800,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s19b1	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":""},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s19b1	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":""},"ServiceEnabled":true,"DateTime":"","CompletedTaskOverWritePolicy":"Manual","LifeCycleEventOnTaskStateChange":null,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s19b1	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","SoftwareInventory":{"@odata.id":"/redfish/v1/UpdateService/SoftwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"}}}
x3000c0s19b2	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"BMC User Accounts","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":0,"MinPasswordLength":6,"AccountLockoutThreshold":0,"AccountLockoutDuration":0,"AccountLockoutCounterResetAfter":0,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s19b2	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"SessionService","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":1800,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s19b2	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":""},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s19b2	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":""},"ServiceEnabled":true,"DateTime":"","CompletedTaskOverWritePolicy":"Manual","LifeCycleEventOnTaskStateChange":null,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s19b2	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","SoftwareInventory":{"@odata.id":"/redfish/v1/UpdateService/SoftwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"}}}
x3000c0s19b4	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","SoftwareInventory":{"@odata.id":"/redfish/v1/UpdateService/SoftwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"}}}
x3000c0s27b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s27b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s27b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s27b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-17T08:15:34+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s27b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1664905557\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s11b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s11b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:23:11+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s13b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:24:06+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s11b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1674060882\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s5b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s5b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s5b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s3b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s3b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s3b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s17b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:23:47+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s17b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1664840070\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s13b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1646239868\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s15b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:29:53+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s15b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1664840099\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
x3000c0s9b0	AccountService			/redfish/v1/AccountService	{"@odata.context":"/redfish/v1/$metadata#AccountService.AccountService","@odata.id":"/redfish/v1/AccountService","@odata.type":"#AccountService.v1_5_0.AccountService","Id":"AccountService","Name":"Account Service","Description":"Redfish User Accounts","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"AuthFailureLoggingThreshold":3,"MinPasswordLength":8,"AccountLockoutThreshold":5,"AccountLockoutDuration":30,"AccountLockoutCounterResetAfter":30,"Accounts":{"@odata.id":"/redfish/v1/AccountService/Accounts"},"Roles":{"@odata.id":"/redfish/v1/AccountService/Roles"}}
x3000c0s9b0	SessionService			/redfish/v1/SessionService	{"@odata.context":"/redfish/v1/$metadata#SessionService.SessionService","@odata.id":"/redfish/v1/SessionService","@odata.type":"#SessionService.v1_1_5.SessionService","Id":"SessionService","Name":"Session Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"SessionTimeout":30,"Sessions":{"@odata.id":"/redfish/v1/SessionService/Sessions"}}
x3000c0s9b0	EventService			/redfish/v1/EventService	{"@odata.context":"/redfish/v1/$metadata#EventService.EventService","@odata.id":"/redfish/v1/EventService","@odata.type":"#EventService.v1_3_0.EventService","Id":"EventService","Name":"Event Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DeliveryRetryAttempts":3,"DeliveryRetryIntervalInSeconds":0,"EventTypesForSubscription":null,"EventTypesForSubscription@odata.count":0,"Subscriptions":{"@odata.id":"/redfish/v1/EventService/Subscriptions"},"Actions":{"#EventService.SubmitTestEvent":{"EventType@Redfish.AllowableValues":null,"target":"/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"}}}
x3000c0s9b0	TaskService			/redfish/v1/TaskService	{"@odata.context":"/redfish/v1/$metadata#TaskService.TaskService","@odata.id":"/redfish/v1/TaskService","@odata.type":"#TaskService.v1_1_3.TaskService","Id":"TaskService","Name":"Task Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"DateTime":"2023-07-18T10:22:39+00:00","CompletedTaskOverWritePolicy":"Oldest","LifeCycleEventOnTaskStateChange":true,"Tasks":{"@odata.id":"/redfish/v1/TaskService/Tasks"}}
x3000c0s9b0	UpdateService			/redfish/v1/UpdateService	{"@odata.context":"/redfish/v1/$metadata#UpdateService.UpdateService","@odata.etag":"W/\\"1664840076\\"","@odata.id":"/redfish/v1/UpdateService","@odata.type":"#UpdateService.v1_5_0.UpdateService","Id":"UpdateService","Name":"Update Service","Status":{"Health":"OK","State":"Enabled"},"ServiceEnabled":true,"FirmwareInventory":{"@odata.id":"/redfish/v1/UpdateService/FirmwareInventory"},"Actions":{"#UpdateService.SimpleUpdate":{"target":"/redfish/v1/UpdateService/Actions/SimpleUpdate"}}}
\.


--
-- Data for Name: system; Type: TABLE DATA; Schema: public; Owner: hmsdsuser
--

COPY public.system (id, schema_version, system_info) FROM stdin;
0	20	{}
\.


--
-- Name: scn_subscriptions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: hmsdsuser
--

SELECT pg_catalog.setval('public.scn_subscriptions_id_seq', 77, true);


--
-- Name: comp_endpoints comp_endpoints_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.comp_endpoints
    ADD CONSTRAINT comp_endpoints_pkey PRIMARY KEY (id);


--
-- Name: comp_eth_interfaces comp_eth_interfaces_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.comp_eth_interfaces
    ADD CONSTRAINT comp_eth_interfaces_pkey PRIMARY KEY (id);


--
-- Name: component_group_members component_group_members_component_id_group_namespace_key; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.component_group_members
    ADD CONSTRAINT component_group_members_component_id_group_namespace_key UNIQUE (component_id, group_namespace);


--
-- Name: component_group_members component_group_members_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.component_group_members
    ADD CONSTRAINT component_group_members_pkey PRIMARY KEY (component_id, group_id);


--
-- Name: component_groups component_groups_name_namespace_key; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.component_groups
    ADD CONSTRAINT component_groups_name_namespace_key UNIQUE (name, namespace);


--
-- Name: component_groups component_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.component_groups
    ADD CONSTRAINT component_groups_pkey PRIMARY KEY (id);


--
-- Name: components components_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.components
    ADD CONSTRAINT components_pkey PRIMARY KEY (id);


--
-- Name: discovery_status discovery_status_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.discovery_status
    ADD CONSTRAINT discovery_status_pkey PRIMARY KEY (id);


--
-- Name: hsn_interfaces hsn_interfaces_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.hsn_interfaces
    ADD CONSTRAINT hsn_interfaces_pkey PRIMARY KEY (nic);


--
-- Name: hwinv_by_fru hwinv_by_fru_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.hwinv_by_fru
    ADD CONSTRAINT hwinv_by_fru_pkey PRIMARY KEY (fru_id);


--
-- Name: hwinv_by_loc hwinv_by_loc_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.hwinv_by_loc
    ADD CONSTRAINT hwinv_by_loc_pkey PRIMARY KEY (id);


--
-- Name: job_state_rf_poll job_state_rf_poll_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.job_state_rf_poll
    ADD CONSTRAINT job_state_rf_poll_pkey PRIMARY KEY (comp_id);


--
-- Name: job_sync job_sync_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.job_sync
    ADD CONSTRAINT job_sync_pkey PRIMARY KEY (id);


--
-- Name: reservations locks_component_id_pk; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT locks_component_id_pk PRIMARY KEY (component_id);


--
-- Name: node_nid_mapping node_nid_mapping_nid_key; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.node_nid_mapping
    ADD CONSTRAINT node_nid_mapping_nid_key UNIQUE (nid);


--
-- Name: node_nid_mapping node_nid_mapping_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.node_nid_mapping
    ADD CONSTRAINT node_nid_mapping_pkey PRIMARY KEY (id);


--
-- Name: power_mapping power_mapping_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.power_mapping
    ADD CONSTRAINT power_mapping_pkey PRIMARY KEY (id);


--
-- Name: rf_endpoints rf_endpoints_fqdn_key; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.rf_endpoints
    ADD CONSTRAINT rf_endpoints_fqdn_key UNIQUE (fqdn);


--
-- Name: rf_endpoints rf_endpoints_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.rf_endpoints
    ADD CONSTRAINT rf_endpoints_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: scn_subscriptions scn_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.scn_subscriptions
    ADD CONSTRAINT scn_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: scn_subscriptions scn_subscriptions_sub_url_key; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.scn_subscriptions
    ADD CONSTRAINT scn_subscriptions_sub_url_key UNIQUE (sub_url);


--
-- Name: service_endpoints service_endpoints_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.service_endpoints
    ADD CONSTRAINT service_endpoints_pkey PRIMARY KEY (rf_endpoint_id, redfish_type);


--
-- Name: system system_pkey; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.system
    ADD CONSTRAINT system_pkey PRIMARY KEY (id);


--
-- Name: system system_schema_version_key; Type: CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.system
    ADD CONSTRAINT system_schema_version_key UNIQUE (schema_version);


--
-- Name: components_role_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX components_role_idx ON public.components USING btree (role);


--
-- Name: components_role_subrole_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX components_role_subrole_idx ON public.components USING btree (role, subrole);


--
-- Name: components_subrole_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX components_subrole_idx ON public.components USING btree (subrole);


--
-- Name: hwinvhist_event_type_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX hwinvhist_event_type_idx ON public.hwinv_hist USING btree (event_type);


--
-- Name: hwinvhist_fru_id_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX hwinvhist_fru_id_idx ON public.hwinv_hist USING btree (fru_id);


--
-- Name: hwinvhist_id_fruid_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX hwinvhist_id_fruid_idx ON public.hwinv_hist USING btree (id, fru_id);


--
-- Name: hwinvhist_id_fruid_ts_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX hwinvhist_id_fruid_ts_idx ON public.hwinv_hist USING btree (id, fru_id, "timestamp");


--
-- Name: hwinvhist_id_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX hwinvhist_id_idx ON public.hwinv_hist USING btree (id);


--
-- Name: hwinvhist_timestamp_idx; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX hwinvhist_timestamp_idx ON public.hwinv_hist USING btree ("timestamp");


--
-- Name: locks_create_timestamp_index; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX locks_create_timestamp_index ON public.reservations USING btree (create_timestamp);


--
-- Name: locks_deputy_key_index; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX locks_deputy_key_index ON public.reservations USING btree (deputy_key);


--
-- Name: locks_expiration_timestamp_index; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX locks_expiration_timestamp_index ON public.reservations USING btree (expiration_timestamp);


--
-- Name: locks_reservation_key_index; Type: INDEX; Schema: public; Owner: hmsdsuser
--

CREATE INDEX locks_reservation_key_index ON public.reservations USING btree (reservation_key);


--
-- Name: comp_endpoints comp_endpoints_rf_endpoint_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.comp_endpoints
    ADD CONSTRAINT comp_endpoints_rf_endpoint_id_fkey FOREIGN KEY (rf_endpoint_id) REFERENCES public.rf_endpoints(id) ON DELETE CASCADE;


--
-- Name: component_group_members component_group_members_component_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.component_group_members
    ADD CONSTRAINT component_group_members_component_id_fkey FOREIGN KEY (component_id) REFERENCES public.components(id) ON DELETE CASCADE;


--
-- Name: component_group_members component_group_members_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.component_group_members
    ADD CONSTRAINT component_group_members_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.component_groups(id) ON DELETE CASCADE;


--
-- Name: hwinv_by_loc hwinv_by_loc_fru_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.hwinv_by_loc
    ADD CONSTRAINT hwinv_by_loc_fru_id_fkey FOREIGN KEY (fru_id) REFERENCES public.hwinv_by_fru(fru_id);


--
-- Name: job_state_rf_poll job_state_rf_poll_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.job_state_rf_poll
    ADD CONSTRAINT job_state_rf_poll_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.job_sync(id) ON DELETE CASCADE;


--
-- Name: reservations locks_hardware_component_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT locks_hardware_component_id_fk FOREIGN KEY (component_id) REFERENCES public.components(id) ON DELETE CASCADE;


--
-- Name: service_endpoints service_endpoints_rf_endpoint_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hmsdsuser
--

ALTER TABLE ONLY public.service_endpoints
    ADD CONSTRAINT service_endpoints_rf_endpoint_id_fkey FOREIGN KEY (rf_endpoint_id) REFERENCES public.rf_endpoints(id) ON DELETE CASCADE;


--
-- Name: SCHEMA metric_helpers; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA metric_helpers TO admin;
GRANT USAGE ON SCHEMA metric_helpers TO robot_zmon;


--
-- Name: SCHEMA user_management; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA user_management TO admin;


--
-- Name: FUNCTION get_btree_bloat_approx(OUT i_database name, OUT i_schema_name name, OUT i_table_name name, OUT i_index_name name, OUT i_real_size numeric, OUT i_extra_size numeric, OUT i_extra_ratio double precision, OUT i_fill_factor integer, OUT i_bloat_size double precision, OUT i_bloat_ratio double precision, OUT i_is_na boolean); Type: ACL; Schema: metric_helpers; Owner: postgres
--

REVOKE ALL ON FUNCTION metric_helpers.get_btree_bloat_approx(OUT i_database name, OUT i_schema_name name, OUT i_table_name name, OUT i_index_name name, OUT i_real_size numeric, OUT i_extra_size numeric, OUT i_extra_ratio double precision, OUT i_fill_factor integer, OUT i_bloat_size double precision, OUT i_bloat_ratio double precision, OUT i_is_na boolean) FROM PUBLIC;
GRANT ALL ON FUNCTION metric_helpers.get_btree_bloat_approx(OUT i_database name, OUT i_schema_name name, OUT i_table_name name, OUT i_index_name name, OUT i_real_size numeric, OUT i_extra_size numeric, OUT i_extra_ratio double precision, OUT i_fill_factor integer, OUT i_bloat_size double precision, OUT i_bloat_ratio double precision, OUT i_is_na boolean) TO admin;
GRANT ALL ON FUNCTION metric_helpers.get_btree_bloat_approx(OUT i_database name, OUT i_schema_name name, OUT i_table_name name, OUT i_index_name name, OUT i_real_size numeric, OUT i_extra_size numeric, OUT i_extra_ratio double precision, OUT i_fill_factor integer, OUT i_bloat_size double precision, OUT i_bloat_ratio double precision, OUT i_is_na boolean) TO robot_zmon;


--
-- Name: FUNCTION get_table_bloat_approx(OUT t_database name, OUT t_schema_name name, OUT t_table_name name, OUT t_real_size numeric, OUT t_extra_size double precision, OUT t_extra_ratio double precision, OUT t_fill_factor integer, OUT t_bloat_size double precision, OUT t_bloat_ratio double precision, OUT t_is_na boolean); Type: ACL; Schema: metric_helpers; Owner: postgres
--

REVOKE ALL ON FUNCTION metric_helpers.get_table_bloat_approx(OUT t_database name, OUT t_schema_name name, OUT t_table_name name, OUT t_real_size numeric, OUT t_extra_size double precision, OUT t_extra_ratio double precision, OUT t_fill_factor integer, OUT t_bloat_size double precision, OUT t_bloat_ratio double precision, OUT t_is_na boolean) FROM PUBLIC;
GRANT ALL ON FUNCTION metric_helpers.get_table_bloat_approx(OUT t_database name, OUT t_schema_name name, OUT t_table_name name, OUT t_real_size numeric, OUT t_extra_size double precision, OUT t_extra_ratio double precision, OUT t_fill_factor integer, OUT t_bloat_size double precision, OUT t_bloat_ratio double precision, OUT t_is_na boolean) TO admin;
GRANT ALL ON FUNCTION metric_helpers.get_table_bloat_approx(OUT t_database name, OUT t_schema_name name, OUT t_table_name name, OUT t_real_size numeric, OUT t_extra_size double precision, OUT t_extra_ratio double precision, OUT t_fill_factor integer, OUT t_bloat_size double precision, OUT t_bloat_ratio double precision, OUT t_is_na boolean) TO robot_zmon;


--
-- Name: FUNCTION pg_stat_statements(showtext boolean); Type: ACL; Schema: metric_helpers; Owner: postgres
--

REVOKE ALL ON FUNCTION metric_helpers.pg_stat_statements(showtext boolean) FROM PUBLIC;
GRANT ALL ON FUNCTION metric_helpers.pg_stat_statements(showtext boolean) TO admin;
GRANT ALL ON FUNCTION metric_helpers.pg_stat_statements(showtext boolean) TO robot_zmon;


--
-- Name: FUNCTION pg_switch_wal(); Type: ACL; Schema: pg_catalog; Owner: postgres
--

GRANT ALL ON FUNCTION pg_catalog.pg_switch_wal() TO admin;


--
-- Name: FUNCTION pg_stat_statements_reset(); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.pg_stat_statements_reset() TO admin;


--
-- Name: FUNCTION set_user(text); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.set_user(text) TO admin;


--
-- Name: FUNCTION create_application_user(username text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.create_application_user(username text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.create_application_user(username text) TO admin;


--
-- Name: FUNCTION create_application_user_or_change_password(username text, password text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.create_application_user_or_change_password(username text, password text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.create_application_user_or_change_password(username text, password text) TO admin;


--
-- Name: FUNCTION create_role(rolename text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.create_role(rolename text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.create_role(rolename text) TO admin;


--
-- Name: FUNCTION create_user(username text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.create_user(username text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.create_user(username text) TO admin;


--
-- Name: FUNCTION drop_role(username text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.drop_role(username text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.drop_role(username text) TO admin;


--
-- Name: FUNCTION drop_user(username text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.drop_user(username text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.drop_user(username text) TO admin;


--
-- Name: FUNCTION revoke_admin(username text); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.revoke_admin(username text) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.revoke_admin(username text) TO admin;


--
-- Name: FUNCTION terminate_backend(pid integer); Type: ACL; Schema: user_management; Owner: postgres
--

REVOKE ALL ON FUNCTION user_management.terminate_backend(pid integer) FROM PUBLIC;
GRANT ALL ON FUNCTION user_management.terminate_backend(pid integer) TO admin;


--
-- Name: TABLE index_bloat; Type: ACL; Schema: metric_helpers; Owner: postgres
--

GRANT SELECT ON TABLE metric_helpers.index_bloat TO admin;
GRANT SELECT ON TABLE metric_helpers.index_bloat TO robot_zmon;


--
-- Name: TABLE pg_stat_statements; Type: ACL; Schema: metric_helpers; Owner: postgres
--

GRANT SELECT ON TABLE metric_helpers.pg_stat_statements TO admin;
GRANT SELECT ON TABLE metric_helpers.pg_stat_statements TO robot_zmon;


--
-- Name: TABLE table_bloat; Type: ACL; Schema: metric_helpers; Owner: postgres
--

GRANT SELECT ON TABLE metric_helpers.table_bloat TO admin;
GRANT SELECT ON TABLE metric_helpers.table_bloat TO robot_zmon;


--
-- Name: TABLE pg_stat_activity; Type: ACL; Schema: pg_catalog; Owner: postgres
--

GRANT SELECT ON TABLE pg_catalog.pg_stat_activity TO admin;


--
-- PostgreSQL database dump complete
--

