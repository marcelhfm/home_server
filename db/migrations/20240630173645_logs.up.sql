CREATE TABLE IF NOT EXISTS logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  datasource_id INT NOT NULL,
  message TEXT NOT NULL,
  CONSTRAINT fk_datsource FOREIGN KEY (datasource_id) REFERENCES datasources (id)
);



