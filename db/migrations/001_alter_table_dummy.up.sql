DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' 
        AND table_name = 'dummy'
        AND column_name = 'updated'
    ) THEN
        ALTER TABLE dummy RENAME COLUMN UPDATED TO updated_at;
    END IF;
END $$;