CREATE TYPE body_part_enum AS ENUM (
    'upper_leg',
    'back',
    'chest',
    'shoulder',
    'upper_arm',
     'core'
);

CREATE TYPE muscle_enum AS ENUM (
    'pectoralis_major',
    'triceps_brachii',
    'biceps_brachii',
    'deltoids',
    'latissimus_dorsi',
    'rhomboids',
    'trapezius',
    'quadriceps',
    'hamstrings',
    'gluteus_maximus',
    'erector_spinae',
    'abdominals'
);


ALTER TABLE exercises
    ALTER COLUMN body_part        TYPE body_part_enum USING body_part::body_part_enum,
    ALTER COLUMN primary_muscle   TYPE muscle_enum USING primary_muscle::muscle_enum,
    ALTER COLUMN secondary_muscle TYPE muscle_enum USING secondary_muscle::muscle_enum;
