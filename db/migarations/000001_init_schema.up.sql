CREATE TABLE IF NOT EXISTS datasources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS timeseries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  datasource_id UUID NOT NULL,
  metric JSONB,
  timestamp TIMESTAMPTZ NOT NULL,
  CONSTRAINT fk_datsource FOREIGN KEY (datasource_id) REFERENCES datasources (id)
);


-- SELECT create_hypertable('timeseries', by_range('timestamp'));
