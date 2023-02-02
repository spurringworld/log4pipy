CREATE TABLE public.trafficlogs (
	id serial4 NOT NULL,
	req_time int8 NULL,
	res_time int8 NULL,
	req_size int4 NULL,
	res_size int4 NULL,
	end_time int8 NULL,
	remote_addr varchar(255) NULL,
	local_addr varchar(255) NULL,
	remote_port int4 NULL,
	local_port int4 NULL,
	res_status int4 NULL,
	service_name varchar(255) NULL,
	mesh_name varchar(255) NULL,
	cluster_name varchar(255) NULL,
	message jsonb NULL,
	req_path varchar(255) NULL,
	req_method varchar(255) NULL,
	bond_type varchar(255) NULL,
	CONSTRAINT trafficlogs_pkey PRIMARY KEY (id)
);