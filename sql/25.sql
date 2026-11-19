-- ============================================================================
-- 25.sql 门店绑定码 + 租户层级（Phase 0 / M1）
--   1. public.tenants：租户层级主表。ThingsPanel 原生没有 tenants 表，租户以
--      users 表中 authority='TENANT_ADMIN' 且 tenant_id 非空的记录代表；本表作为
--      叠加层，补充 parent_id（经销商→宠物医院→门店）等层级信息，并从既有
--      TENANT_ADMIN 用户回填，保证老数据可用。
--   2. public.device_bind_codes：门店出码（绑定码/二维码），家长凭码绑定设备。
-- 全部语句幂等，可重复执行。
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. 租户层级表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.tenants (
	id varchar(36) NOT NULL, -- 租户 id（与 users.tenant_id 一致）
	parent_id varchar(36) NULL, -- 上级租户 id，顶层为 NULL
	"name" varchar(255) NULL, -- 租户名称
	role varchar(32) NULL, -- 角色：dealer-经销商 hospital-宠物医院 store-门店
	status int2 NOT NULL DEFAULT 1, -- 1-启用 0-停用
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT tenants_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE public.tenants IS '租户层级（经销商/宠物医院/门店）';
CREATE INDEX IF NOT EXISTS idx_tenants_parent ON public.tenants (parent_id);

-- 从既有 TENANT_ADMIN 用户回填租户记录（老数据补层级，幂等）
INSERT INTO public.tenants (id, parent_id, "name", role, status, created_at, updated_at, remark)
SELECT u.tenant_id, NULL, COALESCE(NULLIF(u."name", ''), u.email), 'tenant', 1, now(), now(), '由 TENANT_ADMIN 用户自动回填'
FROM public.users u
WHERE u.authority = 'TENANT_ADMIN' AND u.tenant_id IS NOT NULL AND u.tenant_id <> ''
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. 设备绑定码表（门店出码）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.device_bind_codes (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL, -- 出码门店（租户）id
	device_id varchar(36) NOT NULL, -- 目标设备 id
	code varchar(64) NOT NULL, -- 绑定码（对外唯一）
	expires_at timestamptz(6) NULL, -- 过期时间，NULL 表示不过期
	max_uses int4 NOT NULL DEFAULT 1, -- 最大可用次数
	used_count int4 NOT NULL DEFAULT 0, -- 已使用次数
	status int2 NOT NULL DEFAULT 1, -- 1-有效 0-失效
	created_by varchar(36) NULL, -- 出码人（后台用户 id）
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	CONSTRAINT device_bind_codes_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE public.device_bind_codes IS '设备绑定码（门店出码，家长扫码绑定）';
CREATE UNIQUE INDEX IF NOT EXISTS device_bind_codes_code_un ON public.device_bind_codes (code);
CREATE INDEX IF NOT EXISTS idx_device_bind_codes_tenant ON public.device_bind_codes (tenant_id);
CREATE INDEX IF NOT EXISTS idx_device_bind_codes_device ON public.device_bind_codes (device_id);
