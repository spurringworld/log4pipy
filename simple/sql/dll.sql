CREATE TABLE public.logs (
 message json NULL,
 ct timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
 id bigserial NOT NULL,
 CONSTRAINT logs_pk PRIMARY KEY (id)
);