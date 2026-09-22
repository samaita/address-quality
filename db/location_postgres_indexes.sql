-- Applied after COPY so index maintenance does not slow the snapshot import.
CREATE INDEX location_codes_normalized_idx
    ON location_codes (lowercase_normalized);

-- PostgreSQL-only similarity / fuzzy-search readiness.
CREATE INDEX location_codes_normalized_trgm_idx
    ON location_codes USING gin (lowercase_normalized gin_trgm_ops);
