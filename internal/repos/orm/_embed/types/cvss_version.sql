DO
$$
    BEGIN
        CREATE TYPE cvss_version AS ENUM (
            '2',
            '3.0',
            '3.1',
            '4.0'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$
