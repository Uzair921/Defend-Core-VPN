--
-- PostgreSQL database dump
--

\restrict SNSlPBjKFLRiuZb0yE2pjGHZp9FaZxZFvbt3fu6tLdS5eiythsYs2CAx7nXigrd

-- Dumped from database version 18.6 (Ubuntu 18.6-0ubuntu0.26.04.1)
-- Dumped by pg_dump version 18.6 (Ubuntu 18.6-0ubuntu0.26.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: access_policies; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.access_policies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    device_id uuid,
    type text NOT NULL,
    value text NOT NULL,
    action text NOT NULL,
    priority integer DEFAULT 100 NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    CONSTRAINT access_policies_action_check CHECK ((action = ANY (ARRAY['allow'::text, 'deny'::text]))),
    CONSTRAINT access_policies_type_check CHECK ((type = ANY (ARRAY['cidr'::text, 'domain'::text, 'url'::text, 'ip'::text, 'port'::text])))
);


ALTER TABLE public.access_policies OWNER TO defendcore_user;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    action text NOT NULL,
    target text,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.audit_logs OWNER TO defendcore_user;

--
-- Name: device_vpn_configs; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.device_vpn_configs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    device_id uuid NOT NULL,
    service_id uuid NOT NULL,
    assigned_ip text NOT NULL,
    custom_routes jsonb,
    custom_dns text[],
    is_default boolean DEFAULT false NOT NULL,
    last_connected_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.device_vpn_configs OWNER TO defendcore_user;

--
-- Name: devices; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.devices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    public_key text NOT NULL,
    platform text NOT NULL,
    last_seen timestamp with time zone,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.devices OWNER TO defendcore_user;

--
-- Name: dns_policies; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.dns_policies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    domain text NOT NULL,
    resolved_ip text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.dns_policies OWNER TO defendcore_user;

--
-- Name: invoices; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.invoices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    invoice_number text NOT NULL,
    amount_cents integer NOT NULL,
    currency text DEFAULT 'USD'::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    paid_at timestamp with time zone,
    due_at timestamp with time zone,
    items jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT invoices_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'paid'::text, 'overdue'::text, 'cancelled'::text])))
);


ALTER TABLE public.invoices OWNER TO defendcore_user;

--
-- Name: organization_subscriptions; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.organization_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    service_type text NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    max_users integer DEFAULT 10 NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    price_cents_per_month integer,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone,
    CONSTRAINT organization_subscriptions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'suspended'::text, 'cancelled'::text])))
);


ALTER TABLE public.organization_subscriptions OWNER TO defendcore_user;

--
-- Name: organization_users; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.organization_users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT organization_users_role_check CHECK ((role = ANY (ARRAY['admin'::text, 'user'::text])))
);


ALTER TABLE public.organization_users OWNER TO defendcore_user;

--
-- Name: organizations; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.organizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    email text NOT NULL,
    phone text,
    website text,
    status text DEFAULT 'active'::text NOT NULL,
    plan text DEFAULT 'basic'::text NOT NULL,
    max_users integer DEFAULT 10 NOT NULL,
    max_services integer DEFAULT 2 NOT NULL,
    bandwidth_limit_gb integer,
    billing_email text,
    subscription_start timestamp with time zone,
    subscription_end timestamp with time zone,
    monthly_price_cents integer,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT organizations_plan_check CHECK ((plan = ANY (ARRAY['trial'::text, 'basic'::text, 'pro'::text, 'enterprise'::text]))),
    CONSTRAINT organizations_status_check CHECK ((status = ANY (ARRAY['active'::text, 'suspended'::text, 'trial'::text, 'cancelled'::text])))
);


ALTER TABLE public.organizations OWNER TO defendcore_user;

--
-- Name: platform_audit; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.platform_audit (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    actor_id uuid,
    actor_role text,
    organization_id uuid,
    action text NOT NULL,
    target_type text,
    target_id uuid,
    details jsonb,
    ip_address text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.platform_audit OWNER TO defendcore_user;

--
-- Name: policy_group_members; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.policy_group_members (
    group_id uuid NOT NULL,
    user_id uuid NOT NULL
);


ALTER TABLE public.policy_group_members OWNER TO defendcore_user;

--
-- Name: policy_group_rules; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.policy_group_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    group_id uuid NOT NULL,
    type text NOT NULL,
    value text NOT NULL,
    action text NOT NULL,
    priority integer DEFAULT 100 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT policy_group_rules_action_check CHECK ((action = ANY (ARRAY['allow'::text, 'deny'::text]))),
    CONSTRAINT policy_group_rules_type_check CHECK ((type = ANY (ARRAY['cidr'::text, 'domain'::text, 'url'::text, 'ip'::text, 'port'::text])))
);


ALTER TABLE public.policy_group_rules OWNER TO defendcore_user;

--
-- Name: policy_groups; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.policy_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.policy_groups OWNER TO defendcore_user;

--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.refresh_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.refresh_tokens OWNER TO defendcore_user;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.schema_migrations (
    version text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO defendcore_user;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    device_id uuid,
    server_id uuid,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    bytes_in bigint DEFAULT 0 NOT NULL,
    bytes_out bigint DEFAULT 0 NOT NULL,
    client_ip text
);


ALTER TABLE public.sessions OWNER TO defendcore_user;

--
-- Name: user_vpn_services; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.user_vpn_services (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    service_id uuid NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    granted_by uuid,
    granted_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone,
    CONSTRAINT user_vpn_services_role_check CHECK ((role = ANY (ARRAY['user'::text, 'admin'::text])))
);


ALTER TABLE public.user_vpn_services OWNER TO defendcore_user;

--
-- Name: users; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_superadmin boolean DEFAULT false NOT NULL,
    organization_id uuid,
    invited_by uuid,
    invitation_accepted_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO defendcore_user;

--
-- Name: vpn_configs; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.vpn_configs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    device_id uuid NOT NULL,
    server_id uuid,
    assigned_ip text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.vpn_configs OWNER TO defendcore_user;

--
-- Name: vpn_servers; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.vpn_servers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    public_ip text NOT NULL,
    region text,
    status text DEFAULT 'offline'::text NOT NULL,
    version text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.vpn_servers OWNER TO defendcore_user;

--
-- Name: vpn_service_audit; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.vpn_service_audit (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    service_id uuid,
    user_id uuid,
    action text NOT NULL,
    details jsonb,
    ip_address text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.vpn_service_audit OWNER TO defendcore_user;

--
-- Name: vpn_service_types; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.vpn_service_types (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    description text,
    icon text,
    category text NOT NULL,
    supports_split_tunnel boolean DEFAULT false NOT NULL,
    supports_kill_switch boolean DEFAULT false NOT NULL,
    supports_mfa boolean DEFAULT true NOT NULL,
    supports_policies boolean DEFAULT true NOT NULL,
    supports_ztna boolean DEFAULT false NOT NULL,
    default_routes jsonb DEFAULT '["0.0.0.0/0"]'::jsonb NOT NULL,
    default_dns text[] DEFAULT ARRAY['1.1.1.1'::text, '8.8.8.8'::text],
    default_mtu integer DEFAULT 1420 NOT NULL,
    default_keepalive integer DEFAULT 25 NOT NULL,
    display_order integer DEFAULT 100 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.vpn_service_types OWNER TO defendcore_user;

--
-- Name: vpn_services; Type: TABLE; Schema: public; Owner: defendcore_user
--

CREATE TABLE public.vpn_services (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    service_type text NOT NULL,
    description text,
    server_id uuid,
    subnet text NOT NULL,
    server_ip text NOT NULL,
    dns_servers text[] DEFAULT ARRAY['10.8.0.1'::text],
    max_clients integer DEFAULT 100 NOT NULL,
    current_clients integer DEFAULT 0 NOT NULL,
    bandwidth_limit_mbps integer,
    custom_routes jsonb,
    custom_mtu integer,
    status text DEFAULT 'active'::text NOT NULL,
    owner_user_id uuid,
    is_public boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    organization_id uuid,
    CONSTRAINT vpn_services_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text, 'maintenance'::text])))
);


ALTER TABLE public.vpn_services OWNER TO defendcore_user;

--
-- Data for Name: access_policies; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.access_policies (id, user_id, device_id, type, value, action, priority, description, created_at, updated_at, created_by) FROM stdin;
73454884-0098-4d9a-bf53-d2673a2b17d8	54463435-f10f-45f0-b06f-6c33c0e6adf5	\N	cidr	10.0.0.0/24	allow	10	Dev network	2026-09-16 19:59:46.30377+05	2026-09-16 19:59:46.30377+05	54463435-f10f-45f0-b06f-6c33c0e6adf5
ac928e35-0142-466c-bd01-75035441a19d	54463435-f10f-45f0-b06f-6c33c0e6adf5	\N	cidr	10.0.2.0/24	deny	20	Finance (denied)	2026-09-16 19:59:53.398838+05	2026-09-16 19:59:53.398838+05	54463435-f10f-45f0-b06f-6c33c0e6adf5
14fcfa1a-e401-496a-a2d4-9d9580a569ff	54463435-f10f-45f0-b06f-6c33c0e6adf5	\N	domain	git.acme.com	allow	5	Git server	2026-09-16 20:00:03.319758+05	2026-09-16 20:00:03.319758+05	54463435-f10f-45f0-b06f-6c33c0e6adf5
2946221d-7d2d-4997-99a8-baa8db8d8b57	54463435-f10f-45f0-b06f-6c33c0e6adf5	\N	cidr	10.0.0.0/24	allow	10	Dev network	2026-09-16 20:00:13.58339+05	2026-09-16 20:00:13.58339+05	54463435-f10f-45f0-b06f-6c33c0e6adf5
33edbeb7-e7ba-4fdf-a652-38afd47c20cd	54463435-f10f-45f0-b06f-6c33c0e6adf5	\N	cidr	192.168.1.0/24	allow	50	Home network	2026-09-16 22:29:47.163255+05	2026-09-16 22:29:47.163255+05	54463435-f10f-45f0-b06f-6c33c0e6adf5
\.


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.audit_logs (id, user_id, action, target, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: device_vpn_configs; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.device_vpn_configs (id, device_id, service_id, assigned_ip, custom_routes, custom_dns, is_default, last_connected_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: devices; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.devices (id, user_id, name, public_key, platform, last_seen, status, created_at) FROM stdin;
584a1288-c328-4631-bcd7-44578f7bdfe0	54463435-f10f-45f0-b06f-6c33c0e6adf5	My Windows Laptop	RHplrl9crhGdZnGU48ipZQo+0gWBBS6/oh8PmmnndHk=	windows	\N	active	2026-09-15 21:48:54.461648+05
889f86c1-aab6-4563-854d-f263573164a8	54463435-f10f-45f0-b06f-6c33c0e6adf5	ubuntu	tO0cbNG2mkZQE/V1Yqx04ti1Xems81KegVMut7GfhXs=	linux	\N	active	2026-09-15 21:51:45.222675+05
d95b880d-7502-4457-b560-97f8144de746	54463435-f10f-45f0-b06f-6c33c0e6adf5	Windows VM	PqmwNPOzsWag5QK8mtDvu37DeiqlJCKaYiLUUdyuYmY=	windows	\N	active	2026-09-16 02:36:55.386518+05
\.


--
-- Data for Name: dns_policies; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.dns_policies (id, user_id, domain, resolved_ip, created_at) FROM stdin;
\.


--
-- Data for Name: invoices; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.invoices (id, organization_id, invoice_number, amount_cents, currency, status, period_start, period_end, paid_at, due_at, items, created_at) FROM stdin;
ce688f84-77b8-4483-8949-44ca7f835807	c2fa7ef0-d7e2-4505-8dba-2f386e279dbe	INV-2026-LEGACY-ce688f	29900	USD	pending	2026-09-01 05:00:00+05	2026-10-01 04:59:59+05	\N	2026-10-15 05:00:00+05	\N	2026-09-18 00:40:15.242279+05
f3aa1470-8434-466a-90b2-32f7db5b9991	c2fa7ef0-d7e2-4505-8dba-2f386e279dbe	INV-2026-A43807	29900	USD	paid	2026-09-15 05:00:00+05	2026-10-16 04:59:59+05	2026-09-18 00:56:08.186796+05	2026-10-30 05:00:00+05	\N	2026-09-18 00:56:02.291257+05
\.


--
-- Data for Name: organization_subscriptions; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.organization_subscriptions (id, organization_id, service_type, quantity, max_users, status, price_cents_per_month, started_at, expires_at) FROM stdin;
5df5cb5e-3ffb-4b7d-b82e-5ba570742da7	c2fa7ef0-d7e2-4505-8dba-2f386e279dbe	split_tunnel	1	100	active	19900	2026-09-18 00:34:51.542543+05	\N
f17c9c2e-f136-48f0-ae9e-4ac8d4c30db8	c2fa7ef0-d7e2-4505-8dba-2f386e279dbe	full_tunnel	1	50	active	9880	2026-09-18 00:35:53.649421+05	\N
\.


--
-- Data for Name: organization_users; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.organization_users (id, organization_id, user_id, role, status, joined_at) FROM stdin;
46066632-148e-4be9-8f05-a07640c839cb	c2fa7ef0-d7e2-4505-8dba-2f386e279dbe	54463435-f10f-45f0-b06f-6c33c0e6adf5	admin	active	2026-09-18 00:58:17.399398+05
\.


--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.organizations (id, name, slug, email, phone, website, status, plan, max_users, max_services, bandwidth_limit_gb, billing_email, subscription_start, subscription_end, monthly_price_cents, metadata, created_at, updated_at) FROM stdin;
c2fa7ef0-d7e2-4505-8dba-2f386e279dbe	ACME Corp	acme-corp-2ab727	admin@acme.com			active	pro	100	5	\N		\N	\N	29900	\N	2026-09-17 20:53:02.615689+05	2026-09-17 20:53:02.615689+05
\.


--
-- Data for Name: platform_audit; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.platform_audit (id, actor_id, actor_role, organization_id, action, target_type, target_id, details, ip_address, created_at) FROM stdin;
\.


--
-- Data for Name: policy_group_members; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.policy_group_members (group_id, user_id) FROM stdin;
\.


--
-- Data for Name: policy_group_rules; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.policy_group_rules (id, group_id, type, value, action, priority, created_at) FROM stdin;
\.


--
-- Data for Name: policy_groups; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.policy_groups (id, name, description, created_at) FROM stdin;
\.


--
-- Data for Name: refresh_tokens; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at) FROM stdin;
b33cee39-7de8-4a53-b7fa-c09a6f81e4b4	54463435-f10f-45f0-b06f-6c33c0e6adf5	bea55fe90d2b83c4d81cf8ca3a7b24bc303952af94a6fa0bc945427384af0cd9	2026-09-22 21:12:45.899238+05	t	2026-09-15 21:12:45.899502+05
4d079b2f-a6d4-46d9-b874-d8b1a9749e65	54463435-f10f-45f0-b06f-6c33c0e6adf5	428ff99c5c7c6143d65345c6073ee3cd875b8a96aa508e9ac2d421c2224efaf8	2026-09-22 21:12:13.915746+05	t	2026-09-15 21:12:13.916162+05
a4a7d4d4-dc49-4e64-8e53-e49e8563af55	54463435-f10f-45f0-b06f-6c33c0e6adf5	7eebf8b347e501a3572e4b3b0eadd4f2f9f8783b94faef1349969181c44cfbf5	2026-09-22 21:12:28.343337+05	t	2026-09-15 21:12:28.344821+05
44d73a69-8874-4ba6-a179-d40852febf93	54463435-f10f-45f0-b06f-6c33c0e6adf5	7a72e5c6961c05af850b1ead5509f2b50db141e77bb7108adb209b2047828b4f	2026-09-22 21:12:33.071419+05	t	2026-09-15 21:12:33.071568+05
83e8899a-8bcf-48ef-bd94-4dd15588a813	54463435-f10f-45f0-b06f-6c33c0e6adf5	b7697e6bc676fa81ea10dbe3e349c631e317bfac67295d42820225e0ca86ed81	2026-09-22 21:12:45.914372+05	t	2026-09-15 21:12:45.914485+05
3705a81d-cad0-452f-8ace-a426d604ae97	54463435-f10f-45f0-b06f-6c33c0e6adf5	ea80da7f7a92a57c248fe788ffa1290def1849c4ed0b4df31f80b31fcf7e53e3	2026-09-22 21:13:29.733041+05	t	2026-09-15 21:13:29.733214+05
47a6a560-4dec-4f51-af47-5b8c0b1f456a	54463435-f10f-45f0-b06f-6c33c0e6adf5	bba968382e23afe330e20cf18efb5d3cbe9c9acc1a858a40f8d2d9ec5d2e3227	2026-09-22 21:13:29.780377+05	t	2026-09-15 21:13:29.780471+05
02eed234-4851-4d48-88a2-a5e3e0918949	54463435-f10f-45f0-b06f-6c33c0e6adf5	be81fcc4c7fad05ab3c953415e2a1151a1ef7e8a35c34c88e077fb45db8a4b97	2026-09-22 21:17:57.564036+05	t	2026-09-15 21:17:57.564271+05
9786ef26-ea0c-4ac5-b06d-be54c2a09511	54463435-f10f-45f0-b06f-6c33c0e6adf5	d67525be20733d6d15009b016a2d803bad6cfadcdd213ec3ac5e7aac1c077e58	2026-09-22 21:17:57.606893+05	t	2026-09-15 21:17:57.606955+05
25895647-65db-4f27-bf62-4b3c58dcdee9	54463435-f10f-45f0-b06f-6c33c0e6adf5	f6b55cb23a0d58afc5a9227c77399560458800e145f7af85cfc2024a0f851c64	2026-09-22 21:18:11.805691+05	t	2026-09-15 21:18:11.80593+05
107b4072-5126-448a-a263-32b0fb187477	54463435-f10f-45f0-b06f-6c33c0e6adf5	238b588a82765ca54d08bfef911340139c2e0d1ea8ddcfd34eaebe251a2f3ba0	2026-09-22 21:18:11.848168+05	t	2026-09-15 21:18:11.848494+05
1d33abc8-80bd-4d70-96f8-53a81b8b560b	54463435-f10f-45f0-b06f-6c33c0e6adf5	d078dc8a5b8b38f6e55cbd1aab12dab9d959bd7fe6ed83221a249438e1057a1c	2026-09-22 21:41:28.361358+05	t	2026-09-15 21:41:28.362135+05
4b6696da-2700-4f27-8fd3-572e13a0b801	54463435-f10f-45f0-b06f-6c33c0e6adf5	1ff10855ed2668bec4451ee7c8a2bdb2a1225e5cab263329b8480b8f6a0bc88d	2026-09-22 22:30:46.615989+05	t	2026-09-15 22:30:46.617055+05
43f804a0-8f04-48b3-ac99-57ee2e9fae0a	54463435-f10f-45f0-b06f-6c33c0e6adf5	491831024c4a2bbc2b74f82456db01492c4db0ceb0b78f2f6e7279bf8238f7b9	2026-09-22 23:45:36.459618+05	t	2026-09-15 23:45:36.459829+05
0894e7a1-d9d4-423b-bc3a-d807b5d5ebd7	54463435-f10f-45f0-b06f-6c33c0e6adf5	98647bb28a60c23bdeb731e14f824d3019cdb90a6258d808a185aa99b290a16d	2026-09-23 00:15:17.175889+05	t	2026-09-16 00:15:17.176835+05
cef68a09-3b34-49b0-8b23-75b3ec3a542c	54463435-f10f-45f0-b06f-6c33c0e6adf5	ad08b57242e9ae35d9fb00bc13c02ad88ccace097d542176f80e122b784caea2	2026-09-23 00:30:58.085379+05	t	2026-09-16 00:30:58.08621+05
bde16948-ee16-445a-a123-ceebb91275cf	54463435-f10f-45f0-b06f-6c33c0e6adf5	8a89acea88b3fd3e7f60dcabdd648bc5f3cb22a988fad84c049c193d59d99a24	2026-09-23 00:46:57.082599+05	t	2026-09-16 00:46:57.082763+05
4b95f41a-1e73-4904-b1f4-04d7cca09c04	54463435-f10f-45f0-b06f-6c33c0e6adf5	33bc0ffe804f5101993692ec36966fd91617ae0d7889ac0d9cf5531a90647c00	2026-09-23 01:01:57.099708+05	t	2026-09-16 01:01:57.099916+05
a3cd3434-c2a8-4e43-99fd-c24780447a70	54463435-f10f-45f0-b06f-6c33c0e6adf5	d4a3a98175f248d38f571c291fe12104d66b02d3e5b1be0579799ec0e8d714e5	2026-09-23 02:09:36.777452+05	t	2026-09-16 02:09:36.777577+05
357712da-926d-4951-aef9-0807d92369e5	54463435-f10f-45f0-b06f-6c33c0e6adf5	626cd2ed8fd8972e1e59f7b6468df0a5ed642bd49d31bc351f978841b991e9cf	2026-09-23 02:24:49.06241+05	t	2026-09-16 02:24:49.062714+05
3088193f-f3a5-404e-8d4c-663c252677df	54463435-f10f-45f0-b06f-6c33c0e6adf5	fbd2a4a3e287e32315bf82f8b6010d0962a6703973944ba0d6675de9df9422b8	2026-09-23 10:35:46.412964+05	t	2026-09-16 10:35:46.4133+05
f1767b1d-9acd-40aa-be36-1a334c18893d	54463435-f10f-45f0-b06f-6c33c0e6adf5	d46f46da33c5422521e8843cfa6b349047893f11562516a0d0ac88cb91d07972	2026-09-22 21:36:54.425675+05	t	2026-09-15 21:36:54.426259+05
b699257f-6841-4a03-aeb9-2f142260011b	54463435-f10f-45f0-b06f-6c33c0e6adf5	d4cf6f76eeef1f4f40c5e273bb4010aa37b7d49791fcfc818d093d99a34d912d	2026-09-22 21:39:29.935447+05	t	2026-09-15 21:39:29.936143+05
183ea301-992a-4a59-981b-386733af435b	54463435-f10f-45f0-b06f-6c33c0e6adf5	7f46743a46fc34fa95b682fbea2bf3eac57cf921057fc60046ad8b1d80eb80e5	2026-09-22 21:45:22.905535+05	t	2026-09-15 21:45:22.906708+05
cde9ebe1-0322-4702-87ca-590385db4dc7	54463435-f10f-45f0-b06f-6c33c0e6adf5	2db18b972babfeab7cb4fe95ed8a67f533f76af31f337c3fe2d8b8f84d9ca18a	2026-09-22 21:46:24.18329+05	t	2026-09-15 21:46:24.184608+05
e3b43306-24e0-434b-8dcd-95b40ba648dc	54463435-f10f-45f0-b06f-6c33c0e6adf5	040e9e52033f2c258c1dcb9c77663885bd6bd113729665f2e12d777ef880d555	2026-09-22 21:47:16.401335+05	t	2026-09-15 21:47:16.401964+05
112b2c79-ede1-4ce4-9bac-6d6d334c23d7	54463435-f10f-45f0-b06f-6c33c0e6adf5	6bfe2ac762193b9f716592923ca223ad122838acb7ca924397f741e29515caf5	2026-09-22 21:48:54.447464+05	t	2026-09-15 21:48:54.448141+05
159f7f9a-746c-4606-b965-8dae22fda317	54463435-f10f-45f0-b06f-6c33c0e6adf5	3df590862011d18476919903f777382a6b033ad690fdd303033c910ee10a9caf	2026-09-23 00:03:54.750567+05	t	2026-09-16 00:03:54.751279+05
d4d21ebe-cf94-415e-835b-d1797bbb72a1	54463435-f10f-45f0-b06f-6c33c0e6adf5	6ed1a8ea7015a61f892393f18f6feab1677a4157ed96f0f7460c49abfe3b8133	2026-09-23 00:05:16.831502+05	t	2026-09-16 00:05:16.832362+05
1203287c-f46c-4079-aacf-e98e41822bbc	54463435-f10f-45f0-b06f-6c33c0e6adf5	94b57d5491c69a5868f263f11367ed163d7bca2dfbcaffc29394f827945d3d86	2026-09-23 02:36:55.368239+05	t	2026-09-16 02:36:55.369416+05
a66c6cb9-8fa5-46f2-ab43-bd33c49fb4dc	54463435-f10f-45f0-b06f-6c33c0e6adf5	cd7bf9e421025f190f7ca54e7cda78d86b4b78112462925c97bed2c5233b0f1e	2026-09-23 10:24:10.883712+05	t	2026-09-16 10:24:10.885339+05
d98da5c1-4ed9-4434-a0e6-fe0bc76f6fe4	54463435-f10f-45f0-b06f-6c33c0e6adf5	4d9456491be772d30427ced56b16c40df27ea4bd44112a38705b4629e0e067c7	2026-09-23 10:34:11.035607+05	t	2026-09-16 10:34:11.036599+05
50f28132-aa8d-4c3d-b226-70e4363b1c0d	54463435-f10f-45f0-b06f-6c33c0e6adf5	9886914337f0c39054bef916b165aad2fe179f750d260ae7cc07417837abb6d5	2026-09-23 10:35:51.041699+05	t	2026-09-16 10:35:51.042009+05
62ee1fd3-f00a-4f69-8eb8-e74414fd290f	54463435-f10f-45f0-b06f-6c33c0e6adf5	5ebf6fef232605250dc7ada4a667912ceadba9aa6efdb47677eea37fa27db09c	2026-09-23 19:59:41.020925+05	f	2026-09-16 19:59:41.0219+05
41db5c6f-222b-45e5-a2c0-22529750471f	54463435-f10f-45f0-b06f-6c33c0e6adf5	c1b2cb36704296f800a356e63b70a53b8e3ae5a1d11828f017ade09db7b6ea4c	2026-09-23 20:00:13.497451+05	f	2026-09-16 20:00:13.498275+05
ac2dbf06-74f8-4826-b464-5706ec8a22c3	54463435-f10f-45f0-b06f-6c33c0e6adf5	98514fcd27c4f1dfea19ab3d78f02e27440180c148543f6799f1f8242804749a	2026-09-23 20:03:36.273181+05	f	2026-09-16 20:03:36.274127+05
8156a6cd-b516-4ac5-a6b3-8a4f9fdf3dd3	54463435-f10f-45f0-b06f-6c33c0e6adf5	1e890265a70aa61575284cd86ead3309f8b8832a3329fffb298d65f2ac84440c	2026-09-23 22:27:27.071111+05	t	2026-09-16 22:27:27.071781+05
e1eb7625-118f-4c42-9cdb-e6cafb7159c3	54463435-f10f-45f0-b06f-6c33c0e6adf5	eb754b9985724835dc9d556ee2bf1759fdfe13cc2188b00282f3d94a18749275	2026-09-23 23:26:26.189476+05	f	2026-09-16 23:26:26.189815+05
0e156d22-c596-43ee-bc21-4ada59de1e86	54463435-f10f-45f0-b06f-6c33c0e6adf5	72d2d0f7e9cb7ca5586f3cc984d5a298d6b333422b0a2b122ac9ba9db4c51d4a	2026-09-23 23:53:39.421604+05	f	2026-09-16 23:53:39.423003+05
427cf0b7-cf71-423b-bb8f-b89ae9aa620e	54463435-f10f-45f0-b06f-6c33c0e6adf5	a99df39b305adb5336a9bfbbc8ae4302bc9905a83faf8d747b98c03af841264e	2026-09-24 00:11:10.485513+05	f	2026-09-17 00:11:10.486086+05
091d11a3-d939-4bf1-8539-2a6a089b9c4d	54463435-f10f-45f0-b06f-6c33c0e6adf5	c1e257835bf4efbef47f697427d94d3d35e8e3d497457ed7c6b8739c8e6bb00f	2026-09-24 00:11:23.260909+05	f	2026-09-17 00:11:23.261188+05
db6e6f7f-9923-4976-9f16-aab05d4adb5e	54463435-f10f-45f0-b06f-6c33c0e6adf5	32e743aa7c77ca3eda42335b607545c5d23ca9deb278c6a337c2401e8f2af4ba	2026-09-24 00:13:09.171429+05	f	2026-09-17 00:13:09.171681+05
4b778974-067a-4c7b-87bd-ee817e990bd6	54463435-f10f-45f0-b06f-6c33c0e6adf5	9a978f9c55a954ae4f7837eecb8eed349f85634e2d4527b919fc96336c3c1bea	2026-09-24 00:19:47.498327+05	f	2026-09-17 00:19:47.498611+05
b12aa37f-e266-49f2-8073-a3f2063ec0e1	54463435-f10f-45f0-b06f-6c33c0e6adf5	b78a3f9f9669c59be9104dd25924754cab7197cf152aede5045a7cf3d729c9e1	2026-09-24 00:28:02.442035+05	f	2026-09-17 00:28:02.443454+05
e430ca31-0b41-4cdd-9783-04941c5e7c9e	54463435-f10f-45f0-b06f-6c33c0e6adf5	e39e6495c4c0453792b2428296da97c59db63803b7be6f42063604c8c07710fb	2026-09-24 00:30:14.529291+05	f	2026-09-17 00:30:14.529636+05
78fff2bf-793d-4ce0-ad3a-74a1ab19f201	54463435-f10f-45f0-b06f-6c33c0e6adf5	1ebc787d4cd79e3c4ae9c6690738892bf1b6be8fc6134151704fa9591903544e	2026-09-24 13:03:02.483072+05	t	2026-09-17 13:03:02.483753+05
cede59e2-1abf-411e-8fe3-28b717eaaecf	54463435-f10f-45f0-b06f-6c33c0e6adf5	dbaa0493ad31686884f63f7f922b5926566b3b0f466d3d1c0cac70dd71369857	2026-09-24 19:45:32.233819+05	t	2026-09-17 19:45:32.234317+05
b294160f-744e-4526-a7f2-97ff82f80f6f	54463435-f10f-45f0-b06f-6c33c0e6adf5	b012c936a8feb662d389f4b14cf234be457c6c6b6f3cc385ecc8a30e14f8385d	2026-09-24 20:04:57.152627+05	f	2026-09-17 20:04:57.153271+05
652a9a00-1401-46a3-9bd3-7839a212e0ac	54463435-f10f-45f0-b06f-6c33c0e6adf5	5ba470fe29bd3778d02908d210339fdcc2ea4315a07f9a8e094e544945e70b84	2026-09-24 20:05:50.756422+05	f	2026-09-17 20:05:50.756885+05
5bc0e7a6-885b-48bf-bad5-1ab8395dde40	54463435-f10f-45f0-b06f-6c33c0e6adf5	5c5b0f978093bd2638b00b86c1eb591745932306185950cd114cdd8b4fe867ab	2026-09-24 20:01:11.605324+05	t	2026-09-17 20:01:11.605559+05
94184c0b-b6b7-4a38-a501-ae1f3dcbf612	54463435-f10f-45f0-b06f-6c33c0e6adf5	161f8a0812cd7bda86ddbb566c6873fb122cd3c5525cf2b050db403446ddd31a	2026-09-24 20:31:11.749205+05	t	2026-09-17 20:31:11.749288+05
ac0446ca-3a93-4be3-a7a9-c74ef6a0413d	54463435-f10f-45f0-b06f-6c33c0e6adf5	45af5c496c47c312cd423b1b08dfcc6b77791c5e1ecb7cb1f4c08d2d3e184ae5	2026-09-24 20:10:37.835275+05	t	2026-09-17 20:10:37.835741+05
35cedfb0-7297-467c-9719-c78654c383b0	54463435-f10f-45f0-b06f-6c33c0e6adf5	02a719db65a0470d6e27174beffd1679535967e3bd42dfaf5fa59b09cfb1bdef	2026-09-24 20:16:11.634459+05	t	2026-09-17 20:16:11.634795+05
a3454954-e42e-4567-b097-1c7b947266a3	54463435-f10f-45f0-b06f-6c33c0e6adf5	300029b02edc51436fda51b7b6d194ea43d65c370925eb493b34a1c32624198a	2026-09-24 20:53:02.559287+05	f	2026-09-17 20:53:02.560099+05
ba97cdeb-87ff-4cfe-a6a4-3fcf77f48a2b	54463435-f10f-45f0-b06f-6c33c0e6adf5	3a810ee66b2425995c296311e8821b763333bdb6d769d7cbf10d66fbfe796e97	2026-09-24 20:46:11.757409+05	t	2026-09-17 20:46:11.757645+05
1cdf04c6-13f4-448b-af05-039a9153d667	54463435-f10f-45f0-b06f-6c33c0e6adf5	74af89700e046ac13c7e0111a3a9fac9c31cdf75368292e8c19e799ebe8d0b51	2026-09-24 21:12:00.42907+05	f	2026-09-17 21:12:00.429239+05
47cc87e4-1195-4576-869b-b7a5a5beb042	54463435-f10f-45f0-b06f-6c33c0e6adf5	afb69a8a26660e335e76c7256fcdbd77252bafcc214fa9b17127fdff79831e2c	2026-09-24 21:12:27.854645+05	f	2026-09-17 21:12:27.854825+05
ed0cbbf4-18a4-4ba1-8396-6ab46c65bdec	54463435-f10f-45f0-b06f-6c33c0e6adf5	6192281483ab91404f14d662b4c57293bc8cfc0e922897dac7cacbc0ed312534	2026-09-24 21:01:11.613719+05	t	2026-09-17 21:01:11.614054+05
47b1b240-af19-4e34-a0c7-29bcb58fa68c	54463435-f10f-45f0-b06f-6c33c0e6adf5	d56d2fba4ba5559fa899b87e7f19ef84f3c7ae8d0d253cbfb4e6843efeaab33e	2026-09-24 21:18:21.044962+05	f	2026-09-17 21:18:21.045558+05
a583abb5-950a-4984-991d-c80166c1bf4e	54463435-f10f-45f0-b06f-6c33c0e6adf5	caef9d808a5122648eb3fee74326dffa896082ce4ec493bc3e66172afc009441	2026-09-24 20:57:31.443223+05	t	2026-09-17 20:57:31.446226+05
dfcfa397-64b0-4d7e-9594-54b0107bf29b	54463435-f10f-45f0-b06f-6c33c0e6adf5	0fa147b0654c0fdd74ada6f992a860f0d68282778211c08ce5e2e0911fd4ddb0	2026-09-24 21:18:11.737067+05	t	2026-09-17 21:18:11.737468+05
5844d751-6a76-4aa6-b1fe-7e49fdeae7f1	54463435-f10f-45f0-b06f-6c33c0e6adf5	3b4991d68e05b19cf83b0c744602bbb90ad42fa529d19e93e0cda858b3abe36b	2026-09-24 21:19:23.752207+05	t	2026-09-17 21:19:23.753038+05
5f139a0e-fe08-4664-b0aa-4d5a51d37a85	54463435-f10f-45f0-b06f-6c33c0e6adf5	658b354ae693d041c67f14da5d8b8bd58fb385e8e0abf67d4961a8c2c95cff86	2026-09-24 21:33:11.624233+05	t	2026-09-17 21:33:11.624355+05
e24a6fb8-cbb2-47e0-93cf-1c9a039338bc	54463435-f10f-45f0-b06f-6c33c0e6adf5	1a7cc0a21a6c819a959b1f782133ce059c8ef0ce2fca2bab203ddf4936b7e8c6	2026-09-24 21:48:11.719802+05	t	2026-09-17 21:48:11.72021+05
71273730-5150-4696-a776-196bf4ca1b82	54463435-f10f-45f0-b06f-6c33c0e6adf5	2de8c53665f99e6a38bdc995613e8c9b9a88b57d8d6d445e00fe2777e289c305	2026-09-24 22:03:11.788373+05	t	2026-09-17 22:03:11.788598+05
d98e7510-a804-469e-8089-2550a4c152a5	54463435-f10f-45f0-b06f-6c33c0e6adf5	84d122c52c5e2f9a26a962266065735409d6afdd74c145037741592fd5e5035f	2026-09-24 22:18:11.740074+05	t	2026-09-17 22:18:11.740227+05
32d5f5a2-8e7b-4089-906d-540eac565371	54463435-f10f-45f0-b06f-6c33c0e6adf5	4005b23b64a6ebdee146daed75767aacddf1efebe4ab4c049cc11708ea1fb915	2026-09-24 22:33:11.716827+05	t	2026-09-17 22:33:11.717388+05
6e8cff97-1627-441b-bdad-319666a6278c	54463435-f10f-45f0-b06f-6c33c0e6adf5	0813e1db65a867dbe47f252377e62eefce816998391ae8949949f32285c4e906	2026-09-24 23:27:01.299587+05	t	2026-09-17 23:27:01.299689+05
6e9bc2d3-cab2-491a-aa3a-478b17545a73	54463435-f10f-45f0-b06f-6c33c0e6adf5	27a75cbf727db05f3b14097cee3d742244db3ec30f21958df52c29006606e725	2026-09-24 23:42:11.429216+05	t	2026-09-17 23:42:11.429569+05
61ae2d2e-6101-4e5d-a66b-9c16897147a5	54463435-f10f-45f0-b06f-6c33c0e6adf5	34c009b800e00661208611dd5a7640948d80af009dcff5bd4141f6b1402626ea	2026-09-24 23:57:11.441802+05	t	2026-09-17 23:57:11.441917+05
6aded502-8737-46bc-8df7-3be7935a1acc	54463435-f10f-45f0-b06f-6c33c0e6adf5	cc446f1b2270ff20a7dc99a1f0f3c35d954f60029fc95e79418e97ed9a8228a1	2026-09-25 00:12:11.436779+05	f	2026-09-18 00:12:11.437309+05
a7bd3cd4-21c4-4e88-8ace-3d9e1b8c5602	54463435-f10f-45f0-b06f-6c33c0e6adf5	b0e5f2c0d3a154955ecbabf64d675062060f0643b5882f57a86baef1cf2f6987	2026-09-24 21:34:49.406434+05	t	2026-09-17 21:34:49.406528+05
e8233962-6b8d-4501-a18a-7cd18da82441	54463435-f10f-45f0-b06f-6c33c0e6adf5	885eea33e7aa38787bf8a1777da3c658577f1e22b817af2672d2c30c82e21b71	2026-09-25 00:28:27.728786+05	f	2026-09-18 00:28:27.729222+05
b5c94865-1aa1-490c-8a46-502b6bedca1b	54463435-f10f-45f0-b06f-6c33c0e6adf5	5886f9d3b7ff3594e296def35e53a05ff1340e86dd76036a33f1c68610cbafdd	2026-09-25 00:33:40.075544+05	t	2026-09-18 00:33:40.076076+05
39f3a2fe-3641-4955-b516-95c9c63adc85	54463435-f10f-45f0-b06f-6c33c0e6adf5	99d08bcd19f62b9ff62adb4471244c2d7ed9c371f26fc6f6432d3e4c54438475	2026-09-25 00:57:36.510955+05	f	2026-09-18 00:57:36.511532+05
077cc97f-4ab6-4540-ae9b-0bc9c1505f51	54463435-f10f-45f0-b06f-6c33c0e6adf5	1ec461bac3ff836c21770487687173de0c34fe1d40f64613d682a15bda6fc37c	2026-09-25 00:58:16.642322+05	f	2026-09-18 00:58:16.642791+05
e1e219eb-aec4-4a5d-a000-871e0ef620a1	54463435-f10f-45f0-b06f-6c33c0e6adf5	0abe07e8340c50017ce3e0dba60350da1ebcd3067cf8330e596566b2cd987498	2026-09-25 01:06:19.036022+05	f	2026-09-18 01:06:19.036267+05
e374eeb7-4db0-4c6e-a437-d4f500afb053	54463435-f10f-45f0-b06f-6c33c0e6adf5	9c79b344a87b7d67e76028e979023c3affa006bccc8ac4bf094253c976b1f99d	2026-09-25 01:16:41.38049+05	f	2026-09-18 01:16:41.380731+05
412d4e52-0fa3-4f2b-bbb5-3017cedd0a9d	54463435-f10f-45f0-b06f-6c33c0e6adf5	a4efac63cce1c1a6f312664043aa6d1730614dc1c317c65f8838bb7e1f2b13eb	2026-09-25 01:16:57.486201+05	f	2026-09-18 01:16:57.486555+05
e610f026-3d03-4ea3-91cf-03bf0b40f76f	54463435-f10f-45f0-b06f-6c33c0e6adf5	72a7e19336d9b98d75ea67d593f8e441bb4bd27de5e0cf57a839b5dc59b80ef6	2026-09-25 01:17:08.358429+05	f	2026-09-18 01:17:08.358597+05
56fe6523-4867-4edf-8287-9a74448db412	54463435-f10f-45f0-b06f-6c33c0e6adf5	a458b2885955af682de0d58afd8f9c8ba6310832c705379e671e9b1f65d9ad1f	2026-09-25 01:17:09.781189+05	f	2026-09-18 01:17:09.781412+05
f0f83541-5506-461d-8d4a-3952d17b9a35	54463435-f10f-45f0-b06f-6c33c0e6adf5	b53cf43130eb7d482908a581acfbcdf13a7cb5f41b19b9d29609a76c609288f3	2026-09-25 01:25:15.478602+05	f	2026-09-18 01:25:15.478894+05
e9adb80b-8b42-4f92-985a-1d44438c1af1	54463435-f10f-45f0-b06f-6c33c0e6adf5	c7f241c732975e94362dee2a287c4bdcb1a194c72a48187d392fe1c0b601e49f	2026-09-25 02:26:46.997153+05	f	2026-09-18 02:26:46.997467+05
1ce1ea25-a421-45f4-b321-0382633338a3	54463435-f10f-45f0-b06f-6c33c0e6adf5	f12062168b6e66e1f542fb5ac4511f4178f8ac2370208ea79bb66f8b5a14dd07	2026-09-25 02:26:51.279704+05	f	2026-09-18 02:26:51.279883+05
2d8bf5e1-e8bc-4b46-8027-70e3437885ce	54463435-f10f-45f0-b06f-6c33c0e6adf5	4247fb52e0d4fb5c2d1172218c63eb78fa903398a7c338c578e1a6ed443185a7	2026-09-25 02:29:47.395618+05	f	2026-09-18 02:29:47.396136+05
4135f4d4-f3fc-4f68-a061-52eb00ad246c	54463435-f10f-45f0-b06f-6c33c0e6adf5	644301ed20c3869b3c1beb822689a73acc763e5e20c666353c0db47e0fa352b0	2026-09-25 02:31:26.685529+05	f	2026-09-18 02:31:26.68623+05
51433e27-cdbb-4c88-b756-be5573c8b72e	54463435-f10f-45f0-b06f-6c33c0e6adf5	ca47a5086dbc3efac7bb4325c2aa29da8b6e5d0fce7d1ef393b7a4d20dc74f4b	2026-09-25 02:31:51.518863+05	f	2026-09-18 02:31:51.519063+05
821fb061-3e6b-4a8d-bfae-18bf65f29f01	54463435-f10f-45f0-b06f-6c33c0e6adf5	cfb5e1ab05d82f1d2237dda846a9f0f26322dfd7f2a5494238049983292a6426	2026-09-25 02:33:32.961877+05	f	2026-09-18 02:33:32.962525+05
a5eba2a1-b6ad-49c4-9c33-cfcf4a4f0dec	54463435-f10f-45f0-b06f-6c33c0e6adf5	18168ad7d326918ed5d1b4c689692c12926c1acca6971677de9c5605f3f3f1c6	2026-09-25 00:55:04.523385+05	t	2026-09-18 00:55:04.52422+05
538949fa-7284-4701-9a09-aebca75f3eef	54463435-f10f-45f0-b06f-6c33c0e6adf5	ad6f5e928069232e22c036cfd3045f376ec7d60016124250f64b41c7caa088cc	2026-09-25 02:37:11.768624+05	f	2026-09-18 02:37:11.769118+05
b97c4b02-e9df-4c99-a55c-903a1862a21e	54463435-f10f-45f0-b06f-6c33c0e6adf5	c79dfdef35690c5a5abfdc0681c84d3c149a59027224c89ae1aded3a75e23aea	2026-09-25 10:44:22.65806+05	f	2026-09-18 10:44:22.658453+05
6bc8ecba-051d-468d-9cae-3a58c77caaf7	54463435-f10f-45f0-b06f-6c33c0e6adf5	18422642ac53ba135524e075d1624556fd4cd65968366b550ea9f479a399890c	2026-09-25 10:46:32.82798+05	f	2026-09-18 10:46:32.828286+05
94f758b4-ec72-4a05-bb25-e69d4da194a7	54463435-f10f-45f0-b06f-6c33c0e6adf5	9ead561df5d4269e9ff7237050a29091d1c0b30254cec5f5ba9203d83141ac2b	2026-09-25 11:23:00.602304+05	t	2026-09-18 11:23:00.603549+05
e518b143-13eb-493c-8f7f-648f569aabeb	54463435-f10f-45f0-b06f-6c33c0e6adf5	c2938649d1bb30b7593edc9ab65d7f4af09246a3a234e4184c516c4f9687f821	2026-09-25 16:44:47.736274+05	t	2026-09-18 16:44:47.739686+05
0996acb4-8894-45d2-b606-7d277f8640bd	54463435-f10f-45f0-b06f-6c33c0e6adf5	309e58cc7a8ab844cf96fdd32177446ede50deed19fa7f4ea1bce3984724767a	2026-09-25 18:01:26.788553+05	f	2026-09-18 18:01:26.789746+05
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.schema_migrations (version, applied_at) FROM stdin;
001_initial	2026-09-15 21:11:21.973737+05
002_vpn_servers_last_seen	2026-09-16 00:04:33.666925+05
004_access_policies	2026-09-16 19:59:32.56892+05
006_vpn_services	2026-09-17 20:03:41.26244+05
007_multi_tenant	2026-09-17 20:48:39.861299+05
\.


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.sessions (id, user_id, device_id, server_id, started_at, ended_at, bytes_in, bytes_out, client_ip) FROM stdin;
bfb0fdbf-8360-4285-ac5d-21c848777ac5	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:03:46.057414+05	\N	21787	76	192.168.174.1
c62fa2ae-775f-48f9-bd2d-d45fdcd5455a	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:10:37.048731+05	\N	12316	304	192.168.174.1
89ab7c2a-759e-4302-be2e-9b5f6704d889	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:16:24.852885+05	\N	27754	76	192.168.174.1
64c24b4c-8ad5-4258-b092-c7318b6dfe55	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:12:13.455599+05	\N	11871	76	192.168.174.1
64394c37-dcad-4aff-a943-516ff72f3480	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:34:05.331564+05	\N	0	0	192.168.174.1
79a230d3-5411-421d-b4ee-afa3d0833e81	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 02:43:45.424804+05	\N	12835	1280	192.168.174.1
4a130b2c-eb24-4f1f-93aa-b5e783619d9e	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 02:50:14.419752+05	\N	24895	976	192.168.174.1
7896df38-53ae-4fde-8b3b-cedf908aa6b8	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:23:55.180634+05	2026-09-16 11:29:22.774797+05	0	0	192.168.174.1
a1c179dc-2c4c-4e44-91b2-ba2dd67abf31	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 02:55:44.734076+05	\N	11516	204	192.168.174.1
8037cb6e-43fb-439b-830c-23103ae12241	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:45:17.891533+05	\N	0	0	192.168.174.1
7247d908-c99c-4a83-b297-b98460e0c8b3	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:11:20.35497+05	2026-09-16 11:16:22.783887+05	0	0	192.168.174.1
727ed466-51b8-48de-a979-523a544f528f	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:10:21.473236+05	\N	0	0	192.168.174.1
978e2d68-1a92-49f2-997a-73ced61247c0	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:11:33.746528+05	\N	0	0	192.168.174.1
c4e5690a-5e29-43d4-8749-a9017b2d45d8	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 02:40:28.06568+05	2026-09-16 02:45:51.140228+05	0	304	192.168.174.1
048de06f-364c-4aa9-8f32-1613c83316e7	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:13:31.071155+05	\N	0	0	192.168.174.1
6d07b546-ebf8-466d-8ecf-f95d13d0da81	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 02:42:00.050577+05	2026-09-16 02:47:21.143962+05	0	304	192.168.174.1
2a9d3076-f901-4d9a-8779-bb605bfaa512	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:08:00.783552+05	\N	5254	0	192.168.174.1
7c536246-9754-46a5-81e7-2c928cfa3765	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:27:02.524186+05	2026-09-16 11:32:22.758698+05	0	0	192.168.174.1
da08d995-7db3-4571-937c-e11b0dc753cb	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 02:59:29.190339+05	\N	30492	828	192.168.174.1
9d2f6705-0f8d-4435-bf5d-48a240508d35	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 13:08:04.047272+05	2026-09-16 13:13:26.583502+05	0	0	192.168.174.1
af3526ff-32c7-4fdd-8414-ee1f2e771da4	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 03:10:43.071218+05	\N	8607	76	192.168.174.1
10b89156-54e6-4c6b-839e-4e76dde37ecb	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:15:33.359801+05	\N	0	0	192.168.174.1
dac4dbfa-7c61-4146-af91-88b4fd1c8575	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:32:07.607956+05	2026-09-16 12:37:17.752794+05	0	0	192.168.174.1
68b76f1e-bc22-4e2b-a562-89ce34001ab0	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:41:03.234445+05	2026-09-16 12:46:04.794096+05	0	0	192.168.174.1
b205ded1-f2a7-412a-a8c1-b92a37241f31	54463435-f10f-45f0-b06f-6c33c0e6adf5	889f86c1-aab6-4563-854d-f263573164a8	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 00:11:32.703444+05	\N	2604	620	192.168.174.129
c676ae5c-4165-4c9c-9dac-adf426f95d4b	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 13:11:46.091538+05	2026-09-16 13:16:56.60311+05	0	0	192.168.174.1
e5fbf0e0-122a-41e4-8612-f0a2db53e94a	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:35:43.005766+05	\N	26684	760	192.168.174.1
9be7033c-8f41-41b4-869a-3557a00f51d3	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 03:14:07.38661+05	\N	18349	1216	192.168.174.1
40265d9f-10b8-42a4-af57-0204a1c14599	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:31:55.365803+05	\N	0	0	192.168.174.1
80fe56c8-6b9a-4ab6-97d9-0e0cd0882c7e	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:41:24.264372+05	\N	6716	0	192.168.174.1
ef8af523-cedc-44de-8c84-4ad2e31448a9	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:42:48.22879+05	\N	0	0	192.168.174.1
da256f32-d59f-438c-8860-9c4fc5199061	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:43:54.544074+05	\N	0	0	192.168.174.1
074d1d7a-e6bc-4403-a587-f29517af7c31	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:19:20.242854+05	\N	42389	3268	192.168.174.1
58787f02-77e9-4c01-ad1c-695f377f13f4	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:37:05.821505+05	\N	0	0	192.168.174.1
b3846a99-5c40-4e87-9928-f9ec6b7df54c	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 11:28:50.960536+05	\N	0	0	192.168.174.1
952b0eb9-dc48-4d36-a35a-bbb141882c67	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 10:57:24.79117+05	\N	32886	2736	192.168.174.1
80d6cd9c-8a36-4848-b1f8-4cd089208080	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:38:56.383718+05	\N	0	0	192.168.174.1
6361c31f-11f1-4dbd-a97a-7dc42f0aec24	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:35:20.98662+05	\N	0	0	192.168.174.1
79ea170c-f91a-4c3d-8e08-7dcbdeb808db	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 13:22:31.682197+05	2026-09-16 13:27:56.596154+05	0	0	192.168.174.1
26992642-b9ef-4e4f-b784-5e462555b1fc	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 13:05:49.683642+05	2026-09-16 13:10:56.582717+05	0	0	192.168.174.1
9ee78e40-c30e-4f2a-8d72-3a3ef2434fe3	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:30:38.728507+05	\N	100621	1596	192.168.174.1
8767039e-908b-4cf4-a2ff-4cb048eb5d40	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 12:30:21.520773+05	2026-09-16 12:35:47.754008+05	0	0	192.168.174.1
d63a991c-4493-4126-b029-3382ba5c3fcf	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 13:13:53.555186+05	2026-09-16 13:18:56.604272+05	0	0	192.168.174.1
b76ab761-9b51-4ce0-b8ea-b4d708b309c8	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:17:48.248098+05	\N	0	0	192.168.174.1
7af61853-7b65-4b73-8428-551831676834	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 13:24:40.850503+05	2026-09-16 13:29:56.576187+05	0	0	192.168.174.1
d3460808-9691-47dd-b1bc-8723be1a54ec	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:47:06.03536+05	\N	10131	0	192.168.174.1
a6413608-cc9a-49ad-aee9-27d06edf6dc1	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:20:19.705504+05	\N	7661	0	192.168.174.1
4df386f6-9d9a-4ccf-ba3e-1bb0c0bf03f7	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:16:30.750038+05	\N	1733	0	192.168.174.1
4eabc7b4-2c84-42ee-b2cd-6476cb4c7e71	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:12:27.35159+05	\N	0	0	192.168.174.1
cc2bf728-70ce-4470-bccb-60a208cd667a	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 20:46:13.297665+05	2026-09-16 20:51:28.456305+05	0	228	192.168.174.1
ae6a5a7f-7d0d-4797-bd50-fcc73bbc8dc8	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:21:31.489183+05	\N	0	0	192.168.174.1
4f4d7e44-4606-4fa0-912d-9ef5e03a59c3	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:17:24.967937+05	\N	0	0	192.168.174.1
d9969c25-dd05-499c-8d06-190c468f875c	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:19:05.175662+05	\N	0	0	192.168.174.1
1140c077-1fb8-4e39-a3a9-3d6e58980bfe	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:40:59.550246+05	\N	0	0	192.168.174.1
7b9394b6-74af-4dc8-839b-bef0ff5d73c9	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:24:44.000788+05	\N	0	0	192.168.174.1
e74b2b8a-b5fe-4335-9ffa-9a1b98c366a3	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:42:34.20053+05	\N	0	0	192.168.174.1
98502dd1-b974-4dab-9d6b-230a2a45ff6c	54463435-f10f-45f0-b06f-6c33c0e6adf5	d95b880d-7502-4457-b560-97f8144de746	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 21:44:13.365347+05	\N	42568	3484	192.168.174.1
1bc4d155-bde0-40b2-b397-80a746163cc7	54463435-f10f-45f0-b06f-6c33c0e6adf5	889f86c1-aab6-4563-854d-f263573164a8	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 22:01:54.489054+05	\N	7616	200	192.168.174.129
6368667d-f1dd-4be9-aa3b-2aa635b9d619	54463435-f10f-45f0-b06f-6c33c0e6adf5	889f86c1-aab6-4563-854d-f263573164a8	98a95687-7ee8-4c11-9543-d09f7e322d2c	2026-09-16 22:19:27.945236+05	\N	3120	2000	192.168.174.129
\.


--
-- Data for Name: user_vpn_services; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.user_vpn_services (id, user_id, service_id, role, granted_by, granted_at, expires_at) FROM stdin;
42e2649e-f7c4-4731-9b59-61ead8b868a1	54463435-f10f-45f0-b06f-6c33c0e6adf5	a780a668-d0de-4fa0-b14d-51015766ed07	user	54463435-f10f-45f0-b06f-6c33c0e6adf5	2026-09-17 20:05:44.913117+05	\N
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.users (id, email, password_hash, role, status, created_at, updated_at, is_superadmin, organization_id, invited_by, invitation_accepted_at) FROM stdin;
54463435-f10f-45f0-b06f-6c33c0e6adf5	admin@defendcore.local	$2a$12$zVNLYCVVkjbpTUN1ZSp3ruymM2w9PJBesQ/enzUjQvsJqFMf5/BeK	user	active	2026-09-15 21:12:13.914403+05	2026-09-15 21:12:13.914403+05	t	\N	\N	\N
\.


--
-- Data for Name: vpn_configs; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.vpn_configs (id, user_id, device_id, server_id, assigned_ip, created_at) FROM stdin;
\.


--
-- Data for Name: vpn_servers; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.vpn_servers (id, name, public_ip, region, status, version, created_at, last_seen) FROM stdin;
98a95687-7ee8-4c11-9543-d09f7e322d2c	defendcore-lab-01	192.168.174.132	lab	online	0.1.0	2026-09-16 00:04:49.739497+05	2026-09-18 23:54:13.195354+05
\.


--
-- Data for Name: vpn_service_audit; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.vpn_service_audit (id, service_id, user_id, action, details, ip_address, created_at) FROM stdin;
73be2076-cf0b-4835-8b51-13e148384a89	a780a668-d0de-4fa0-b14d-51015766ed07	54463435-f10f-45f0-b06f-6c33c0e6adf5	service_created	{"name": "Corporate Split VPN", "type": "split_tunnel"}		2026-09-17 20:05:02.258577+05
a1df1978-f38e-47f3-b27d-ca782b481751	a780a668-d0de-4fa0-b14d-51015766ed07	54463435-f10f-45f0-b06f-6c33c0e6adf5	user_assigned	{"role": "user", "user_id": "54463435-f10f-45f0-b06f-6c33c0e6adf5"}		2026-09-17 20:05:44.916452+05
\.


--
-- Data for Name: vpn_service_types; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.vpn_service_types (id, code, name, description, icon, category, supports_split_tunnel, supports_kill_switch, supports_mfa, supports_policies, supports_ztna, default_routes, default_dns, default_mtu, default_keepalive, display_order, is_active, created_at, updated_at) FROM stdin;
80148217-5515-4f30-a81d-a787b3fd7585	full_tunnel	Full Tunnel VPN	Route all your internet traffic through secure VPN.	🔒	personal	f	t	t	t	f	["0.0.0.0/0"]	{1.1.1.1,8.8.8.8}	1420	25	10	t	2026-09-17 20:03:41.26244+05	2026-09-17 20:03:41.26244+05
18e93d0e-a865-437f-8a1f-f57bee511691	split_tunnel	Split Tunnel VPN	Only corporate resources through VPN.	⚡	corporate	t	f	t	t	f	["10.0.0.0/8", "192.168.0.0/16"]	{10.0.0.1,1.1.1.1}	1420	25	20	t	2026-09-17 20:03:41.26244+05	2026-09-17 20:03:41.26244+05
74f98079-4e4d-44e9-af62-61dc339978f8	remote_access	Remote Access VPN	Access your office network from anywhere.	👤	corporate	t	t	t	t	f	["10.0.0.0/8"]	{10.0.0.1}	1420	25	30	t	2026-09-17 20:03:41.26244+05	2026-09-17 20:03:41.26244+05
04968dfd-b4f2-4936-9460-a2a31e01135b	site_to_site	Site-to-Site VPN	Connect multiple office locations.	🏢	enterprise	f	f	t	t	f	["10.0.0.0/8"]	{10.0.0.1}	1420	25	40	t	2026-09-17 20:03:41.26244+05	2026-09-17 20:03:41.26244+05
401f1962-f77d-4ad2-9c0f-1c297ec451ed	zero_trust	Zero Trust VPN	Per-application access with verification.	🔐	enterprise	t	t	t	t	f	[]	{10.0.0.1}	1420	25	50	t	2026-09-17 20:03:41.26244+05	2026-09-17 20:03:41.26244+05
990c11a5-460f-4351-ad03-9e9f72d5828b	stealth	Stealth VPN	Obfuscated traffic bypasses firewalls.	🎭	personal	f	t	t	t	f	["0.0.0.0/0"]	{1.1.1.1}	1420	25	60	t	2026-09-17 20:03:41.26244+05	2026-09-17 20:03:41.26244+05
\.


--
-- Data for Name: vpn_services; Type: TABLE DATA; Schema: public; Owner: defendcore_user
--

COPY public.vpn_services (id, name, slug, service_type, description, server_id, subnet, server_ip, dns_servers, max_clients, current_clients, bandwidth_limit_mbps, custom_routes, custom_mtu, status, owner_user_id, is_public, created_at, updated_at, organization_id) FROM stdin;
a780a668-d0de-4fa0-b14d-51015766ed07	Corporate Split VPN	corporate-split-vpn-521ff4	split_tunnel	Dev team split tunnel	\N	10.9.0.0/24	10.9.0.1	{10.0.0.1,1.1.1.1}	50	0	\N	["10.0.0.0/8", "192.168.0.0/16"]	\N	active	54463435-f10f-45f0-b06f-6c33c0e6adf5	f	2026-09-17 20:05:02.254583+05	2026-09-17 20:05:02.254583+05	\N
\.


--
-- Name: access_policies access_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.access_policies
    ADD CONSTRAINT access_policies_pkey PRIMARY KEY (id);


--
-- Name: access_policies access_policies_user_id_device_id_type_value_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.access_policies
    ADD CONSTRAINT access_policies_user_id_device_id_type_value_key UNIQUE (user_id, device_id, type, value);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: device_vpn_configs device_vpn_configs_device_id_service_id_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.device_vpn_configs
    ADD CONSTRAINT device_vpn_configs_device_id_service_id_key UNIQUE (device_id, service_id);


--
-- Name: device_vpn_configs device_vpn_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.device_vpn_configs
    ADD CONSTRAINT device_vpn_configs_pkey PRIMARY KEY (id);


--
-- Name: devices devices_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_pkey PRIMARY KEY (id);


--
-- Name: dns_policies dns_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.dns_policies
    ADD CONSTRAINT dns_policies_pkey PRIMARY KEY (id);


--
-- Name: dns_policies dns_policies_user_id_domain_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.dns_policies
    ADD CONSTRAINT dns_policies_user_id_domain_key UNIQUE (user_id, domain);


--
-- Name: invoices invoices_invoice_number_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_invoice_number_key UNIQUE (invoice_number);


--
-- Name: invoices invoices_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);


--
-- Name: organization_subscriptions organization_subscriptions_organization_id_service_type_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_subscriptions
    ADD CONSTRAINT organization_subscriptions_organization_id_service_type_key UNIQUE (organization_id, service_type);


--
-- Name: organization_subscriptions organization_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_subscriptions
    ADD CONSTRAINT organization_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: organization_users organization_users_organization_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_users
    ADD CONSTRAINT organization_users_organization_id_user_id_key UNIQUE (organization_id, user_id);


--
-- Name: organization_users organization_users_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_users
    ADD CONSTRAINT organization_users_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_slug_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_slug_key UNIQUE (slug);


--
-- Name: platform_audit platform_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.platform_audit
    ADD CONSTRAINT platform_audit_pkey PRIMARY KEY (id);


--
-- Name: policy_group_members policy_group_members_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_group_members
    ADD CONSTRAINT policy_group_members_pkey PRIMARY KEY (group_id, user_id);


--
-- Name: policy_group_rules policy_group_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_group_rules
    ADD CONSTRAINT policy_group_rules_pkey PRIMARY KEY (id);


--
-- Name: policy_groups policy_groups_name_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_groups
    ADD CONSTRAINT policy_groups_name_key UNIQUE (name);


--
-- Name: policy_groups policy_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_groups
    ADD CONSTRAINT policy_groups_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: user_vpn_services user_vpn_services_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.user_vpn_services
    ADD CONSTRAINT user_vpn_services_pkey PRIMARY KEY (id);


--
-- Name: user_vpn_services user_vpn_services_user_id_service_id_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.user_vpn_services
    ADD CONSTRAINT user_vpn_services_user_id_service_id_key UNIQUE (user_id, service_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: vpn_configs vpn_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_configs
    ADD CONSTRAINT vpn_configs_pkey PRIMARY KEY (id);


--
-- Name: vpn_servers vpn_servers_name_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_servers
    ADD CONSTRAINT vpn_servers_name_key UNIQUE (name);


--
-- Name: vpn_servers vpn_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_servers
    ADD CONSTRAINT vpn_servers_pkey PRIMARY KEY (id);


--
-- Name: vpn_service_audit vpn_service_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_service_audit
    ADD CONSTRAINT vpn_service_audit_pkey PRIMARY KEY (id);


--
-- Name: vpn_service_types vpn_service_types_code_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_service_types
    ADD CONSTRAINT vpn_service_types_code_key UNIQUE (code);


--
-- Name: vpn_service_types vpn_service_types_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_service_types
    ADD CONSTRAINT vpn_service_types_pkey PRIMARY KEY (id);


--
-- Name: vpn_services vpn_services_pkey; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_services
    ADD CONSTRAINT vpn_services_pkey PRIMARY KEY (id);


--
-- Name: vpn_services vpn_services_slug_key; Type: CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_services
    ADD CONSTRAINT vpn_services_slug_key UNIQUE (slug);


--
-- Name: idx_access_policies_device; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_access_policies_device ON public.access_policies USING btree (device_id);


--
-- Name: idx_access_policies_type; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_access_policies_type ON public.access_policies USING btree (type);


--
-- Name: idx_access_policies_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_access_policies_user ON public.access_policies USING btree (user_id);


--
-- Name: idx_audit_logs_created; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_audit_logs_created ON public.audit_logs USING btree (created_at DESC);


--
-- Name: idx_audit_logs_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_audit_logs_user ON public.audit_logs USING btree (user_id);


--
-- Name: idx_device_vpn_configs_device; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_device_vpn_configs_device ON public.device_vpn_configs USING btree (device_id);


--
-- Name: idx_device_vpn_configs_service; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_device_vpn_configs_service ON public.device_vpn_configs USING btree (service_id);


--
-- Name: idx_devices_user_id; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_devices_user_id ON public.devices USING btree (user_id);


--
-- Name: idx_dns_policies_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_dns_policies_user ON public.dns_policies USING btree (user_id);


--
-- Name: idx_invoices_org; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_invoices_org ON public.invoices USING btree (organization_id);


--
-- Name: idx_invoices_status; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_invoices_status ON public.invoices USING btree (status);


--
-- Name: idx_org_subs_org; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_org_subs_org ON public.organization_subscriptions USING btree (organization_id);


--
-- Name: idx_org_users_org; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_org_users_org ON public.organization_users USING btree (organization_id);


--
-- Name: idx_org_users_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_org_users_user ON public.organization_users USING btree (user_id);


--
-- Name: idx_organizations_slug; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_organizations_slug ON public.organizations USING btree (slug);


--
-- Name: idx_organizations_status; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_organizations_status ON public.organizations USING btree (status);


--
-- Name: idx_platform_audit_actor; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_platform_audit_actor ON public.platform_audit USING btree (actor_id);


--
-- Name: idx_platform_audit_org; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_platform_audit_org ON public.platform_audit USING btree (organization_id);


--
-- Name: idx_refresh_tokens_hash; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_refresh_tokens_hash ON public.refresh_tokens USING btree (token_hash);


--
-- Name: idx_refresh_tokens_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_refresh_tokens_user ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_sessions_active; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_sessions_active ON public.sessions USING btree (user_id, started_at DESC) WHERE (ended_at IS NULL);


--
-- Name: idx_sessions_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_sessions_user ON public.sessions USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_org; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_users_org ON public.users USING btree (organization_id);


--
-- Name: idx_users_superadmin; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_users_superadmin ON public.users USING btree (is_superadmin);


--
-- Name: idx_vpn_configs_device; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_vpn_configs_device ON public.vpn_configs USING btree (device_id);


--
-- Name: idx_vpn_service_audit_service; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_vpn_service_audit_service ON public.vpn_service_audit USING btree (service_id);


--
-- Name: idx_vpn_service_audit_user; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_vpn_service_audit_user ON public.vpn_service_audit USING btree (user_id);


--
-- Name: idx_vpn_services_org; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_vpn_services_org ON public.vpn_services USING btree (organization_id);


--
-- Name: idx_vpn_services_status; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_vpn_services_status ON public.vpn_services USING btree (status);


--
-- Name: idx_vpn_services_type; Type: INDEX; Schema: public; Owner: defendcore_user
--

CREATE INDEX idx_vpn_services_type ON public.vpn_services USING btree (service_type);


--
-- Name: access_policies access_policies_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.access_policies
    ADD CONSTRAINT access_policies_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: access_policies access_policies_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.access_policies
    ADD CONSTRAINT access_policies_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(id) ON DELETE CASCADE;


--
-- Name: access_policies access_policies_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.access_policies
    ADD CONSTRAINT access_policies_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: audit_logs audit_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: device_vpn_configs device_vpn_configs_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.device_vpn_configs
    ADD CONSTRAINT device_vpn_configs_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(id) ON DELETE CASCADE;


--
-- Name: device_vpn_configs device_vpn_configs_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.device_vpn_configs
    ADD CONSTRAINT device_vpn_configs_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.vpn_services(id) ON DELETE CASCADE;


--
-- Name: devices devices_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: dns_policies dns_policies_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.dns_policies
    ADD CONSTRAINT dns_policies_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: invoices invoices_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_subscriptions organization_subscriptions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_subscriptions
    ADD CONSTRAINT organization_subscriptions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_subscriptions organization_subscriptions_service_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_subscriptions
    ADD CONSTRAINT organization_subscriptions_service_type_fkey FOREIGN KEY (service_type) REFERENCES public.vpn_service_types(code);


--
-- Name: organization_users organization_users_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_users
    ADD CONSTRAINT organization_users_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_users organization_users_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.organization_users
    ADD CONSTRAINT organization_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: platform_audit platform_audit_actor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.platform_audit
    ADD CONSTRAINT platform_audit_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: platform_audit platform_audit_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.platform_audit
    ADD CONSTRAINT platform_audit_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: policy_group_members policy_group_members_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_group_members
    ADD CONSTRAINT policy_group_members_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.policy_groups(id) ON DELETE CASCADE;


--
-- Name: policy_group_members policy_group_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_group_members
    ADD CONSTRAINT policy_group_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: policy_group_rules policy_group_rules_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.policy_group_rules
    ADD CONSTRAINT policy_group_rules_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.policy_groups(id) ON DELETE CASCADE;


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(id) ON DELETE SET NULL;


--
-- Name: sessions sessions_server_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_server_id_fkey FOREIGN KEY (server_id) REFERENCES public.vpn_servers(id);


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_vpn_services user_vpn_services_granted_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.user_vpn_services
    ADD CONSTRAINT user_vpn_services_granted_by_fkey FOREIGN KEY (granted_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_vpn_services user_vpn_services_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.user_vpn_services
    ADD CONSTRAINT user_vpn_services_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.vpn_services(id) ON DELETE CASCADE;


--
-- Name: user_vpn_services user_vpn_services_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.user_vpn_services
    ADD CONSTRAINT user_vpn_services_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_invited_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id);


--
-- Name: users users_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: vpn_configs vpn_configs_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_configs
    ADD CONSTRAINT vpn_configs_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(id) ON DELETE CASCADE;


--
-- Name: vpn_configs vpn_configs_server_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_configs
    ADD CONSTRAINT vpn_configs_server_id_fkey FOREIGN KEY (server_id) REFERENCES public.vpn_servers(id);


--
-- Name: vpn_configs vpn_configs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_configs
    ADD CONSTRAINT vpn_configs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: vpn_service_audit vpn_service_audit_service_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_service_audit
    ADD CONSTRAINT vpn_service_audit_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.vpn_services(id) ON DELETE CASCADE;


--
-- Name: vpn_service_audit vpn_service_audit_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_service_audit
    ADD CONSTRAINT vpn_service_audit_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: vpn_services vpn_services_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_services
    ADD CONSTRAINT vpn_services_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: vpn_services vpn_services_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_services
    ADD CONSTRAINT vpn_services_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: vpn_services vpn_services_server_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_services
    ADD CONSTRAINT vpn_services_server_id_fkey FOREIGN KEY (server_id) REFERENCES public.vpn_servers(id) ON DELETE SET NULL;


--
-- Name: vpn_services vpn_services_service_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: defendcore_user
--

ALTER TABLE ONLY public.vpn_services
    ADD CONSTRAINT vpn_services_service_type_fkey FOREIGN KEY (service_type) REFERENCES public.vpn_service_types(code);


--
-- PostgreSQL database dump complete
--

\unrestrict SNSlPBjKFLRiuZb0yE2pjGHZp9FaZxZFvbt3fu6tLdS5eiythsYs2CAx7nXigrd

