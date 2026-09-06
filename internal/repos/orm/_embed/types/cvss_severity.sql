DO
$$
    BEGIN
        CREATE TYPE cvss_severity AS ENUM (
            'critical',
            'high',
            'medium',
            'low',
            'none'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$
