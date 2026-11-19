-- ============================================================================
-- 24.sql 行业方案（Industry Solution）架构
--   1. 每个行业方案使用独立 PostgreSQL schema（sol_<code>），实现方案数据隔离；
--   2. 把 23.sql 建在 public 下的宠物业务表迁入 sol_pet_medical；
--   3. public 下新增平台级表：方案字典 / 租户分配 / C 端用户 / C 端用户设备绑定；
--   4. 种子：5 个行业方案字典 + 前端「行业方案」菜单。
-- 全部语句幂等，可重复执行。
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. 方案 schema（每个行业方案一个独占 schema）
-- ---------------------------------------------------------------------------
CREATE SCHEMA IF NOT EXISTS sol_pet_medical;
CREATE SCHEMA IF NOT EXISTS sol_yuezi;
CREATE SCHEMA IF NOT EXISTS sol_tuoyu;
CREATE SCHEMA IF NOT EXISTS sol_xumu;
CREATE SCHEMA IF NOT EXISTS sol_gongye;

-- ---------------------------------------------------------------------------
-- 2. 把 23.sql 残留的 public.pet_* 表迁入 sol_pet_medical
--    幂等守卫：仅当 public 存在且目标 schema 内不存在时才迁移。
-- ---------------------------------------------------------------------------
DO $$
DECLARE
	t text;
BEGIN
	FOREACH t IN ARRAY ARRAY[
		'pet_profiles', 'pet_shops', 'pet_staff_notes',
		'pet_timelines', 'pet_albums', 'pet_daily_reports'
	]
	LOOP
		IF EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = t
		) AND NOT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'sol_pet_medical' AND table_name = t
		) THEN
			EXECUTE format('ALTER TABLE public.%I SET SCHEMA sol_pet_medical', t);
		END IF;
	END LOOP;
END $$;

-- ---------------------------------------------------------------------------
-- 3. 宠物医疗方案业务表（sol_pet_medical schema 内，幂等）
--    schema 与 23.sql 内容保持一致，仅改变所属 schema。
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sol_pet_medical.pet_profiles (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL, -- 关联设备
	pet_name varchar(99) NULL, -- 宠物名字
	breed varchar(99) NULL, -- 品种
	age varchar(50) NULL, -- 年龄
	avatar varchar(999) NULL, -- 头像地址
	boarding_days int4 NULL DEFAULT 0, -- 寄养天数
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_profiles_pkey PRIMARY KEY (id),
	CONSTRAINT pet_profiles_device_un UNIQUE (device_id)
);
COMMENT ON TABLE sol_pet_medical.pet_profiles IS '宠物档案';

CREATE TABLE IF NOT EXISTS sol_pet_medical.pet_shops (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	"name" varchar(255) NULL, -- 店铺名称
	open_time varchar(10) NULL, -- 营业开始时间 HH:mm
	close_time varchar(10) NULL, -- 营业结束时间 HH:mm
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_shops_pkey PRIMARY KEY (id),
	CONSTRAINT pet_shops_device_un UNIQUE (device_id)
);
COMMENT ON TABLE sol_pet_medical.pet_shops IS '宠物店铺信息与营业时段';

CREATE TABLE IF NOT EXISTS sol_pet_medical.pet_staff_notes (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	content text NULL, -- 留言内容
	staff_name varchar(99) NULL, -- 店员名字
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_staff_notes_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE sol_pet_medical.pet_staff_notes IS '店员留言';

CREATE TABLE IF NOT EXISTS sol_pet_medical.pet_timelines (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	event_type varchar(20) NULL, -- eat/play/sleep/care/treat
	title varchar(255) NULL, -- 动态文案
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_timelines_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE sol_pet_medical.pet_timelines IS '宠物动态时间线';

CREATE TABLE IF NOT EXISTS sol_pet_medical.pet_albums (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	url varchar(999) NULL, -- 图片地址
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_albums_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE sol_pet_medical.pet_albums IS '宠物相册（抓拍）';

CREATE TABLE IF NOT EXISTS sol_pet_medical.pet_daily_reports (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	report_date date NOT NULL, -- 日报日期
	eat_count int4 NULL DEFAULT 0, -- 进食次数
	sleep_minutes int4 NULL DEFAULT 0, -- 睡眠时长（分钟）
	play_minutes int4 NULL DEFAULT 0, -- 玩耍时长（分钟）
	drink_count int4 NULL DEFAULT 0, -- 饮水次数
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_daily_reports_pkey PRIMARY KEY (id),
	CONSTRAINT pet_daily_reports_device_date_un UNIQUE (device_id, report_date)
);
COMMENT ON TABLE sol_pet_medical.pet_daily_reports IS '宠物日报';

CREATE INDEX IF NOT EXISTS idx_pet_staff_notes_device ON sol_pet_medical.pet_staff_notes (tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_pet_timelines_device ON sol_pet_medical.pet_timelines (tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_pet_albums_device ON sol_pet_medical.pet_albums (tenant_id, device_id);

-- ---------------------------------------------------------------------------
-- 4. 平台级表（public schema）
-- ---------------------------------------------------------------------------

-- 4.1 行业方案字典
CREATE TABLE IF NOT EXISTS public.industry_solutions (
	id varchar(36) NOT NULL,
	code varchar(64) NOT NULL, -- 方案编码，对外稳定标识
	name varchar(99) NOT NULL, -- 方案名称
	description varchar(500) NULL,
	schema_name varchar(64) NOT NULL, -- 方案业务 schema（^sol_[a-z0-9_]+$）
	icon varchar(255) NULL,
	status int2 NOT NULL DEFAULT 1, -- 1-启用 0-停用
	sort int4 NOT NULL DEFAULT 0,
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT industry_solutions_pkey PRIMARY KEY (id),
	CONSTRAINT industry_solutions_code_un UNIQUE (code)
);
COMMENT ON TABLE public.industry_solutions IS '行业方案字典';

-- 4.2 方案↔租户 分配关系
CREATE TABLE IF NOT EXISTS public.tenant_solutions (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	solution_code varchar(64) NOT NULL,
	enabled int2 NOT NULL DEFAULT 1, -- 1-启用 0-停用
	config text NULL, -- 方案在该租户下的扩展配置
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	CONSTRAINT tenant_solutions_pkey PRIMARY KEY (id),
	CONSTRAINT tenant_solutions_tenant_code_un UNIQUE (tenant_id, solution_code)
);
COMMENT ON TABLE public.tenant_solutions IS '行业方案与租户的分配关系';

-- 4.3 C 端用户（独立于后台 users 表）
CREATE TABLE IF NOT EXISTS public.c_end_users (
	id varchar(36) NOT NULL,
	phone varchar(20) NULL,
	open_id varchar(128) NULL, -- 小程序 openid
	nickname varchar(99) NULL,
	avatar varchar(999) NULL,
	password_hash varchar(255) NULL,
	default_solution_code varchar(64) NULL, -- 默认所属方案
	status int2 NOT NULL DEFAULT 1, -- 1-正常 0-禁用
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT c_end_users_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE public.c_end_users IS 'C 端用户';
CREATE UNIQUE INDEX IF NOT EXISTS c_end_users_phone_un ON public.c_end_users (phone) WHERE phone IS NOT NULL;

-- 4.4 C 端用户↔设备 绑定关系（多对多，一台设备可被多人共享）
CREATE TABLE IF NOT EXISTS public.c_end_user_devices (
	id varchar(36) NOT NULL,
	c_user_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	solution_code varchar(64) NOT NULL,
	status int2 NOT NULL DEFAULT 1, -- 1-有效 0-解绑
	bound_at timestamptz(6) NOT NULL,
	created_at timestamptz(6) NOT NULL,
	CONSTRAINT c_end_user_devices_pkey PRIMARY KEY (id),
	CONSTRAINT c_end_user_devices_un UNIQUE (c_user_id, device_id)
);
COMMENT ON TABLE public.c_end_user_devices IS 'C 端用户与设备绑定关系';
CREATE INDEX IF NOT EXISTS idx_c_end_user_devices_device ON public.c_end_user_devices (device_id);

-- ---------------------------------------------------------------------------
-- 5. 种子数据
-- ---------------------------------------------------------------------------

-- 5.1 行业方案字典（5 个方案）
INSERT INTO public.industry_solutions (
	id, code, name, description, schema_name, icon, status, sort, created_at, updated_at, remark
) VALUES
('a1b2c3d4-0001-4000-8000-000000000001', 'pet_medical', '宠物医疗', '宠物寄养看护与实时音视频巡检', 'sol_pet_medical', 'mdi:paw', 1, 1, now(), now(), ''),
('a1b2c3d4-0001-4000-8000-000000000002', 'yuezi',       '月子中心', '母婴护理与新生儿看护',           'sol_yuezi',       'mdi:baby-face-outline', 1, 2, now(), now(), ''),
('a1b2c3d4-0001-4000-8000-000000000003', 'tuoyu',       '托育',     '婴幼儿托育机构安全管理',         'sol_tuoyu',       'mdi:human-child', 1, 3, now(), now(), ''),
('a1b2c3d4-0001-4000-8000-000000000004', 'xumu',        '畜牧行业', '养殖场牲畜环境与健康监测',       'sol_xumu',        'mdi:cow', 1, 4, now(), now(), ''),
('a1b2c3d4-0001-4000-8000-000000000005', 'gongye',      '工业采集', '工业设备数据采集与运行监控',     'sol_gongye',      'mdi:factory', 1, 5, now(), now(), '')
ON CONFLICT DO NOTHING;

-- 5.2 前端菜单：行业方案（仅平台超级管理员可见）
INSERT INTO public.sys_ui_elements (
	id, parent_id, element_code, element_type, orders, param1, param2, param3,
	authority, description, created_at, remark, multilingual, route_path
) VALUES
('7c9f0a20-2c32-4d42-9e52-000000000001', '0', 'industry-solution', 1, 118, '/industry-solution', 'mdi:view-dashboard-outline', '0',
 '["SYS_ADMIN"]'::json, '行业方案', now(), '', 'route.industry-solution', 'layout.base'),
('7c9f0a20-2c32-4d42-9e52-000000000002', '7c9f0a20-2c32-4d42-9e52-000000000001', 'industry-solution_manage', 3, 1181, '/industry-solution/manage', 'mdi:shape-plus-outline', '0',
 '["SYS_ADMIN"]'::json, '方案管理', now(), '', 'route.industry-solution_manage', 'view.industry-solution_manage')
ON CONFLICT (id) DO NOTHING;
