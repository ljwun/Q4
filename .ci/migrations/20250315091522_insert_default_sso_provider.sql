-- Insert default value to "sso_providers" table
INSERT INTO "sso_providers" ("name", "created_at", "updated_at", "deleted_at") 
VALUES
    ("Internal", now(), now(), NULL),
    ("Google", now(), now(), NULL),
    ("Microsoft", now(), now(), NULL),
    ("GitHub", now(), now(), NULL);