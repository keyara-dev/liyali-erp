-- Complete database cleanup - drops everything

-- Drop all triggers first
DO $$ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT n.nspname as schemaname, c.relname as tablename, t.tgname as triggername 
              FROM pg_trigger t 
              JOIN pg_class c ON t.tgrelid = c.oid 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND NOT t.tgisinternal) 
    LOOP
        EXECUTE 'DROP TRIGGER IF EXISTS ' || quote_ident(r.triggername) || ' ON ' || quote_ident(r.schemaname) || '.' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;

-- Drop all functions
DO $$ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT p.proname, oidvectortypes(p.proargtypes) as argtypes 
              FROM pg_proc p 
              JOIN pg_namespace n ON p.pronamespace = n.oid 
              WHERE n.nspname = 'public') 
    LOOP
        EXECUTE 'DROP FUNCTION IF EXISTS ' || quote_ident(r.proname) || '(' || r.argtypes || ') CASCADE';
    END LOOP;
END $$;

-- Drop all views
DO $$ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT c.relname as viewname 
              FROM pg_class c 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND c.relkind = 'v') 
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || quote_ident(r.viewname) || ' CASCADE';
    END LOOP;
END $$;

-- Drop all tables
DO $$ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT c.relname as tablename 
              FROM pg_class c 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND c.relkind = 'r') 
    LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;

-- Drop all sequences
DO $$ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT c.relname as sequencename 
              FROM pg_class c 
              JOIN pg_namespace n ON c.relnamespace = n.oid 
              WHERE n.nspname = 'public' AND c.relkind = 'S') 
    LOOP
        EXECUTE 'DROP SEQUENCE IF EXISTS ' || quote_ident(r.sequencename) || ' CASCADE';
    END LOOP;
END $$;

-- Drop all types
DO $$ 
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT t.typname 
              FROM pg_type t 
              JOIN pg_namespace n ON t.typnamespace = n.oid 
              WHERE n.nspname = 'public' AND t.typtype = 'e') 
    LOOP
        EXECUTE 'DROP TYPE IF EXISTS ' || quote_ident(r.typname) || ' CASCADE';
    END LOOP;
END $$;

SELECT 'Database completely cleaned' as status;