-- Adds treatment_validation support
ALTER TABLE settings ADD treatment_schema jsonb;
ALTER TABLE settings ADD validation_url varchar(255);
