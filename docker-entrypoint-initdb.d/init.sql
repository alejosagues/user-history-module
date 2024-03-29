CREATE TABLE "user" (
    id SERIAL PRIMARY KEY, name VARCHAR(64) NOT NULL
);

CREATE TABLE api_calls (
    id SERIAL PRIMARY KEY, endpoint VARCHAR(64) NOT NULL, user_id INT, call_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (user_id) REFERENCES "user" (id)
);

CREATE OR REPLACE FUNCTION generate_api_calls() RETURNS 
VOID AS 
$$
DECLARE
	user_record RECORD;
	 endpoint_names VARCHAR[] := ARRAY['/api/endpoint1' , '/api/endpoint2' , '/api/endpoint3' , '/api/endpoint4' , '/api/endpoint5'];
BEGIN
	FOR user_record IN SELECT id FROM "user" LOOP
	    FOR i IN 1..5 LOOP
	        FOR j IN 1..50 LOOP
	            INSERT INTO "api_calls" ( endpoint , user_id , call_timestamp ) VALUES ( endpoint_names[i] , user_record . id , CURRENT_TIMESTAMP ) ;
            END LOOP;
        END	LOOP;
    END	LOOP;
END;
$$
LANGUAGE plpgsql; 

CREATE OR REPLACE FUNCTION generate_users() RETURNS 
VOID AS 
$$
BEGIN
	FOR i IN 1..5000 LOOP
	    INSERT INTO "user" ( name ) VALUES ( CONCAT('username', i) ) ;
    END LOOP;
END;
$$
LANGUAGE plpgsql; 

SELECT generate_users();

SELECT generate_api_calls();