DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' 
        AND table_name = 'dummy'
        AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE dummy RENAME COLUMN updated_at TO updated;
    END IF;
END $$;