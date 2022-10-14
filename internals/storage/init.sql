DROP TABLE IF EXISTS notes;
CREATE TABLE notes (
    id varchar PRIMARY KEY,
    full_name varchar NOT NULL,
    company varchar,
    phone varchar NOT NULL CHECK (phone SIMILAR TO '[+]?[0-9]{1,15}'),
    mail varchar NOT NULL,
    birth_date varchar CHECK (birth_date SIMILAR TO '[0-9]{1,2}.[0-9]{1,2}.[0-9]{4}'), -- symbol . is not a metachar for SIMILAR TO syntax
    image_id varchar CHECK (image_id SIMILAR TO '[0-9a-f]{32}')
);