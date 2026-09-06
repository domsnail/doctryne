DO
$$
    BEGIN
        CREATE TYPE reference_type AS ENUM (
            'advisory',
            'article',
            'exploit',
            'fix',
            'government_resource',
            'media_coverage',
            'mitigation',
            'patch',
            'product',
            'release_notes',
            'vendor_advisory',
            'other'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$
