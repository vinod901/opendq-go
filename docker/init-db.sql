-- Create marquez user and database
CREATE USER marquez WITH PASSWORD 'marquez';
CREATE DATABASE marquez OWNER marquez;

-- Create openfga database
CREATE DATABASE openfga;

-- Create keycloak database
CREATE DATABASE keycloak;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE marquez TO marquez;
GRANT ALL PRIVILEGES ON DATABASE openfga TO postgres;
GRANT ALL PRIVILEGES ON DATABASE keycloak TO postgres;
