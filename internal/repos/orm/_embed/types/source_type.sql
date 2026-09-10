DO
$$
    BEGIN
        CREATE TYPE source_type AS ENUM (
            'primary',
            'secondary'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$
