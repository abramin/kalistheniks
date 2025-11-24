INSERT INTO exercises (name, body_part, primary_muscle, secondary_muscle)
VALUES
  -- Barbell basics
  ('Back Squat',        'upper_leg',  'quadriceps',       'gluteus_maximus'),
  ('Deadlift',          'back',       'erector_spinae',   'gluteus_maximus'),
  ('Bench Press',       'chest',      'pectoralis_major', 'triceps_brachii'),
  ('Overhead Press',    'shoulder',   'deltoids',         'triceps_brachii'),
  ('Bent-over Row',     'back',       'latissimus_dorsi', 'biceps_brachii'),

  -- Bodyweight basics
  ('Push Up',           'chest',      'pectoralis_major', 'triceps_brachii'),
  ('Inverted Row',      'back',       'rhomboids',        'biceps_brachii'),
  ('Air Squat',         'upper_leg',  'quadriceps',       'gluteus_maximus'),
  ('Dip',               'upper_arm',  'triceps_brachii',  'pectoralis_major'),
  ('Plank',             'core',       'abdominals',       NULL);
