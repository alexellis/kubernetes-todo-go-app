CREATE TABLE todo (
  id              INT GENERATED ALWAYS AS IDENTITY,
  description     text NOT NULL,
  created_date    timestamp NOT NULL,
  completed_date  timestamp NULL
);