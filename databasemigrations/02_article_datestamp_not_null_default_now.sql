ALTER TABLE articles
  ALTER COLUMN datestamp SET NOT NULL;
ALTER TABLE articles
  ALTER COLUMN datestamp SET DEFAULT now();
