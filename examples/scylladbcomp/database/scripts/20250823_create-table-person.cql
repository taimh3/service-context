CREATE TABLE IF NOT EXISTS auth.person (
    first_name TEXT,
    last_name TEXT,
    email LIST<TEXT>,
    PRIMARY KEY (first_name, last_name)
);
