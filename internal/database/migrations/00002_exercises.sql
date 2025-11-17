-- +goose Up
CREATE TABLE exercises (
    id                BIGSERIAL PRIMARY KEY,
    -- NULL for the built-in library shared by everyone.
    user_id           BIGINT REFERENCES users (id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    primary_muscle    TEXT NOT NULL,
    secondary_muscles TEXT[] NOT NULL DEFAULT '{}',
    equipment         TEXT NOT NULL,
    archived_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX exercises_owner_name_key ON exercises (coalesce(user_id, 0), lower(name));

INSERT INTO exercises (name, primary_muscle, secondary_muscles, equipment) VALUES
    ('Back Squat',               'quads',      '{glutes,hamstrings,core}', 'barbell'),
    ('Front Squat',              'quads',      '{glutes,core}',            'barbell'),
    ('Deadlift',                 'back',       '{hamstrings,glutes,forearms}', 'barbell'),
    ('Romanian Deadlift',        'hamstrings', '{glutes,back}',            'barbell'),
    ('Sumo Deadlift',            'glutes',     '{quads,hamstrings,back}',  'barbell'),
    ('Bench Press',              'chest',      '{triceps,shoulders}',      'barbell'),
    ('Incline Bench Press',      'chest',      '{shoulders,triceps}',      'barbell'),
    ('Close-Grip Bench Press',   'triceps',    '{chest,shoulders}',        'barbell'),
    ('Overhead Press',           'shoulders',  '{triceps,core}',           'barbell'),
    ('Barbell Row',              'back',       '{biceps,forearms}',        'barbell'),
    ('Pendlay Row',              'back',       '{biceps}',                 'barbell'),
    ('Power Clean',              'back',       '{quads,glutes,shoulders}', 'barbell'),
    ('Hip Thrust',               'glutes',     '{hamstrings}',             'barbell'),
    ('Barbell Curl',             'biceps',     '{forearms}',               'barbell'),
    ('Pull-Up',                  'back',       '{biceps}',                 'bodyweight'),
    ('Chin-Up',                  'back',       '{biceps}',                 'bodyweight'),
    ('Dip',                      'triceps',    '{chest,shoulders}',        'bodyweight'),
    ('Push-Up',                  'chest',      '{triceps,shoulders}',      'bodyweight'),
    ('Hanging Leg Raise',        'core',       '{}',                       'bodyweight'),
    ('Plank',                    'core',       '{}',                       'bodyweight'),
    ('Dumbbell Bench Press',     'chest',      '{triceps,shoulders}',      'dumbbell'),
    ('Incline Dumbbell Press',   'chest',      '{shoulders,triceps}',      'dumbbell'),
    ('Dumbbell Shoulder Press',  'shoulders',  '{triceps}',                'dumbbell'),
    ('Lateral Raise',            'shoulders',  '{}',                       'dumbbell'),
    ('One-Arm Dumbbell Row',     'back',       '{biceps}',                 'dumbbell'),
    ('Dumbbell Curl',            'biceps',     '{forearms}',               'dumbbell'),
    ('Hammer Curl',              'biceps',     '{forearms}',               'dumbbell'),
    ('Bulgarian Split Squat',    'quads',      '{glutes}',                 'dumbbell'),
    ('Goblet Squat',             'quads',      '{glutes,core}',            'dumbbell'),
    ('Walking Lunge',            'quads',      '{glutes,hamstrings}',      'dumbbell'),
    ('Leg Press',                'quads',      '{glutes}',                 'machine'),
    ('Leg Extension',            'quads',      '{}',                       'machine'),
    ('Lying Leg Curl',           'hamstrings', '{}',                       'machine'),
    ('Seated Leg Curl',          'hamstrings', '{}',                       'machine'),
    ('Standing Calf Raise',      'calves',     '{}',                       'machine'),
    ('Seated Calf Raise',        'calves',     '{}',                       'machine'),
    ('Chest Fly Machine',        'chest',      '{shoulders}',              'machine'),
    ('Lat Pulldown',             'back',       '{biceps}',                 'cable'),
    ('Seated Cable Row',         'back',       '{biceps}',                 'cable'),
    ('Face Pull',                'shoulders',  '{back}',                   'cable'),
    ('Triceps Pushdown',         'triceps',    '{}',                       'cable'),
    ('Overhead Cable Extension', 'triceps',    '{}',                       'cable'),
    ('Cable Crunch',             'core',       '{}',                       'cable'),
    ('Kettlebell Swing',         'glutes',     '{hamstrings,back}',        'kettlebell');

-- +goose Down
DROP TABLE exercises;
