ALTER TABLE exercises
    ALTER COLUMN body_part        TYPE TEXT USING body_part::TEXT,
    ALTER COLUMN primary_muscle   TYPE TEXT USING primary_muscle::TEXT,
    ALTER COLUMN secondary_muscle TYPE TEXT USING secondary_muscle::TEXT;

DROP TYPE IF EXISTS muscle_enum;
DROP TYPE IF EXISTS body_part_enum;