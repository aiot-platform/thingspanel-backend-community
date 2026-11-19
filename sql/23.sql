-- ============================================================================
-- 23.sql 宠物看护（Pet Care）能力
--   1. 宠物业务表：档案 / 店铺 / 店员留言 / 动态时间线 / 相册 / 日报
--   2. 设备物模型模板：宠物店音视频看护设备（18 属性 / 7 事件 / 4 服务）
--      其中 4 个只读属性来自宠物项圈（穿戴设备），经边缘网关统一上报，
--      与音视频共用同一台设备、同一套物模型。
--   3. 前端菜单：宠物看护（宠物档案 / 店员留言 / 相册时间线）
-- 全部语句幂等，可重复执行。
-- 说明：device_templates 按租户隔离（tenant_id NOT NULL），物模型种子按执行时
--       已存在的租户逐个插入，主键使用 md5(前缀+tenant_id) 生成确定性 UUID。
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. 宠物业务表
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.pet_profiles (
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
COMMENT ON TABLE public.pet_profiles IS '宠物档案';

CREATE TABLE IF NOT EXISTS public.pet_shops (
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
COMMENT ON TABLE public.pet_shops IS '宠物店铺信息与营业时段';

CREATE TABLE IF NOT EXISTS public.pet_staff_notes (
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
COMMENT ON TABLE public.pet_staff_notes IS '店员留言';

CREATE TABLE IF NOT EXISTS public.pet_timelines (
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
COMMENT ON TABLE public.pet_timelines IS '宠物动态时间线';

CREATE TABLE IF NOT EXISTS public.pet_albums (
	id varchar(36) NOT NULL,
	tenant_id varchar(36) NOT NULL,
	device_id varchar(36) NOT NULL,
	url varchar(999) NULL, -- 图片地址
	created_at timestamptz(6) NOT NULL,
	updated_at timestamptz(6) NOT NULL,
	remark varchar(255) NULL,
	CONSTRAINT pet_albums_pkey PRIMARY KEY (id)
);
COMMENT ON TABLE public.pet_albums IS '宠物相册（抓拍）';

CREATE TABLE IF NOT EXISTS public.pet_daily_reports (
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
COMMENT ON TABLE public.pet_daily_reports IS '宠物日报';

CREATE INDEX IF NOT EXISTS idx_pet_staff_notes_device ON public.pet_staff_notes (tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_pet_timelines_device ON public.pet_timelines (tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_pet_albums_device ON public.pet_albums (tenant_id, device_id);

-- ---------------------------------------------------------------------------
-- 2. 设备物模型模板种子：宠物店音视频看护设备
--    模板主键    = md5('pet-tpl-' + tenant_id)::uuid
--    模型行主键  = md5('pet-tpl-' + <类型> + '-' + <标识符> + '-' + tenant_id)::uuid
-- ---------------------------------------------------------------------------

INSERT INTO public.device_templates (
	id, "name", author, "version", "description",
	tenant_id, created_at, updated_at, flag, "label"
)
SELECT (md5('pet-tpl-' || u.tenant_id)::uuid)::text,
	'宠物店音视频看护设备', 'ThingsPanel', '1.1.0', '18 属性 / 7 事件 / 4 服务',
	u.tenant_id, now(), now(), 1, 'pet'
FROM (
	SELECT DISTINCT tenant_id FROM public.users
	WHERE tenant_id IS NOT NULL AND tenant_id <> ''
) u
ON CONFLICT (id) DO NOTHING;

-- 只读属性（R）→ device_model_telemetry
INSERT INTO public.device_model_telemetry (
	id, device_template_id, data_name, data_identifier, read_write_flag,
	data_type, unit, description, additional_info, created_at, updated_at, tenant_id
)
SELECT (md5('pet-tpl-telemetry-' || m.data_identifier || '-' || u.tenant_id)::uuid)::text,
	(md5('pet-tpl-' || u.tenant_id)::uuid)::text,
	m.data_name, m.data_identifier, 'R', m.data_type, m.unit, m.description, m.additional_info::json,
	now(), now(), u.tenant_id
FROM (
	SELECT DISTINCT tenant_id FROM public.users
	WHERE tenant_id IS NOT NULL AND tenant_id <> ''
) u
CROSS JOIN (
	VALUES
	('音视频状态', 'av_status', 'Enum', '', '设备侧音视频链路状态，只读。0 空闲；1 推流中；2 对讲中；3 异常',
	 '[{"value_type":"Number","value":0,"description":"空闲"},{"value_type":"Number","value":1,"description":"推流中"},{"value_type":"Number","value":2,"description":"对讲中"},{"value_type":"Number","value":3,"description":"异常"}]'),
	('当前观看人数', 'viewer_count', 'Number', '人', '当前正在观看实时流的终端数，上限 4',
	 '{"min":0,"max":4,"step":1}'),
	('Wi-Fi 信号强度', 'wifi_rssi', 'Number', 'dBm', '设备侧 Wi-Fi 信号强度',
	 '{"min":-100,"max":0,"step":1}'),
	('存储卡状态', 'sd_status', 'Enum', '', '存储卡状态。0 未插卡；1 正常；2 异常；3 格式化中',
	 '[{"value_type":"Number","value":0,"description":"未插卡"},{"value_type":"Number","value":1,"description":"正常"},{"value_type":"Number","value":2,"description":"异常"},{"value_type":"Number","value":3,"description":"格式化中"}]'),
	('宠物状态', 'pet_activity', 'Enum', '', '设备端 AI 识别的宠物行为。0 未检测到宠物；1 休息；2 睡觉；3 玩耍；4 进食；5 饮水',
	 '[{"value_type":"Number","value":0,"description":"未检测到宠物"},{"value_type":"Number","value":1,"description":"休息"},{"value_type":"Number","value":2,"description":"睡觉"},{"value_type":"Number","value":3,"description":"玩耍"},{"value_type":"Number","value":4,"description":"进食"},{"value_type":"Number","value":5,"description":"饮水"}]'),
	('活跃度', 'activity_level', 'Number', '%', '最近 10 分钟活动量折算，0 为有效值',
	 '{"min":0,"max":100,"step":1}'),
	('舍内温度', 'env_temperature', 'Number', '℃', '宠物活动区域温度，一位小数',
	 '{"min":-20,"max":60,"step":0.1}'),
	('舍内湿度', 'env_humidity', 'Number', '%', '宠物活动区域相对湿度',
	 '{"min":0,"max":100,"step":1}'),
	('今日剩余零食次数', 'treat_left', 'Number', '次', '今日还可投喂的次数，由店铺设定上限',
	 '{"min":0,"max":10,"step":1}'),
	('项圈在线状态', 'collar_online', 'Boolean', '', '项圈与边缘网关的链路状态。0 离线；1 在线',
	 '[]'),
	('项圈电量', 'collar_battery', 'Number', '%', '项圈电池剩余电量',
	 '{"min":0,"max":100,"step":1}'),
	('宠物体温', 'pet_body_temperature', 'Number', '℃', '项圈贴体测得的宠物体温，一位小数',
	 '{"min":30,"max":45,"step":0.1}'),
	('宠物心率', 'pet_heart_rate', 'Number', 'bpm', '项圈光电容积脉搏波测得的宠物心率',
	 '{"min":0,"max":250,"step":1}')
) AS m(data_name, data_identifier, data_type, unit, description, additional_info)
ON CONFLICT (device_template_id, data_identifier) DO NOTHING;

-- 读写属性（RW）→ device_model_attributes
INSERT INTO public.device_model_attributes (
	id, device_template_id, data_name, data_identifier, read_write_flag,
	data_type, unit, description, additional_info, created_at, updated_at, tenant_id
)
SELECT (md5('pet-tpl-attribute-' || m.data_identifier || '-' || u.tenant_id)::uuid)::text,
	(md5('pet-tpl-' || u.tenant_id)::uuid)::text,
	m.data_name, m.data_identifier, 'RW', m.data_type, m.unit, m.description, m.additional_info::json,
	now(), now(), u.tenant_id
FROM (
	SELECT DISTINCT tenant_id FROM public.users
	WHERE tenant_id IS NOT NULL AND tenant_id <> ''
) u
CROSS JOIN (
	VALUES
	('画质', 'video_quality', 'Enum', '', '实时流编码画质档位，切换后设备重新协商码流',
	 '[{"value_type":"Number","value":1,"description":"流畅"},{"value_type":"Number","value":2,"description":"标清"},{"value_type":"Number","value":3,"description":"高清"},{"value_type":"Number","value":4,"description":"超清"}]'),
	('设备拾音开关', 'mic_switch', 'Boolean', '', '控制设备麦克风是否采集并上传声音',
	 '[]'),
	('设备扬声器音量', 'speaker_volume', 'Number', '%', '对讲时播放观看端声音的设备音量',
	 '{"min":0,"max":100,"step":10}'),
	('夜视模式', 'night_vision', 'Enum', '', '红外夜视策略。0 自动；1 始终开启；2 始终关闭',
	 '[{"value_type":"Number","value":0,"description":"自动"},{"value_type":"Number","value":1,"description":"始终开启"},{"value_type":"Number","value":2,"description":"始终关闭"}]'),
	('画面翻转', 'image_flip', 'Enum', '', '设备倒装时用于校正画面方向。0 正常；1 旋转180度',
	 '[{"value_type":"Number","value":0,"description":"正常"},{"value_type":"Number","value":1,"description":"旋转180度"}]')
) AS m(data_name, data_identifier, data_type, unit, description, additional_info)
ON CONFLICT (device_template_id, data_identifier) DO NOTHING;

-- 事件 → device_model_events（参数放在 params，read_write_flag 表示参数数据类型）
INSERT INTO public.device_model_events (
	id, device_template_id, data_name, data_identifier, params,
	description, additional_info, created_at, updated_at, tenant_id
)
SELECT (md5('pet-tpl-event-' || m.data_identifier || '-' || u.tenant_id)::uuid)::text,
	(md5('pet-tpl-' || u.tenant_id)::uuid)::text,
	m.data_name, m.data_identifier, m.params::json,
	m.description, '[]'::json, now(), now(), u.tenant_id
FROM (
	SELECT DISTINCT tenant_id FROM public.users
	WHERE tenant_id IS NOT NULL AND tenant_id <> ''
) u
CROSS JOIN (
	VALUES
	('推流异常', 'stream_error',
	 '[{"id":1,"data_name":"错误类型","data_identifier":"error_code","read_write_flag":"String","description":"0 编码器故障；1 摄像头无信号；2 网络拥塞"}]',
	 'ERROR 级别。设备推流失败时上报，av_status 同步为异常，恢复推流后解除'),
	('存储卡异常', 'sd_card_error',
	 '[{"id":1,"data_name":"存储卡状态","data_identifier":"sd_error_status","read_write_flag":"String","description":"0 未插卡；1 正常；2 异常；3 格式化中"}]',
	 'WARN 级别。存储卡读写异常时上报，不影响实时流'),
	('吠叫提醒', 'bark_alert',
	 '[{"id":1,"data_name":"声音强度","data_identifier":"volume_db","read_write_flag":"Number","description":"0~120 dB"}]',
	 'INFO 级别。识别到持续吠叫时上报，仅提醒'),
	('舍内环境异常', 'env_abnormal',
	 '[{"id":1,"data_name":"异常类型","data_identifier":"abnormal_type","read_write_flag":"String","description":"0 温度过高；1 温度过低；2 湿度异常"},{"id":2,"data_name":"当前值","data_identifier":"value","read_write_flag":"Number","description":"异常指标当前值"}]',
	 'WARN 级别。温湿度超出舒适范围时上报，回到范围内后解除'),
	('项圈电量低', 'collar_low_battery',
	 '[{"id":1,"data_name":"剩余电量","data_identifier":"battery","read_write_flag":"Number","description":"剩余电量百分比 0~100"}]',
	 'WARN 级别。项圈电量低于阈值时上报，提醒店员充电'),
	('项圈离线', 'collar_offline', '[]',
	 'WARN 级别。项圈与边缘网关断连超时上报，collar_online 同步为离线'),
	('宠物生理指标异常', 'pet_health_abnormal',
	 '[{"id":1,"data_name":"异常类型","data_identifier":"type","read_write_flag":"String","description":"0 体温过高；1 体温过低；2 心率过快；3 心率过慢"},{"id":2,"data_name":"当前值","data_identifier":"value","read_write_flag":"Number","description":"异常指标当前值"}]',
	 'WARN 级别。项圈测得体温或心率超出正常范围时上报，回到范围内后解除')
) AS m(data_name, data_identifier, params, description)
ON CONFLICT (device_template_id, data_identifier) DO NOTHING;

-- 服务 → device_model_commands（params 为入参列表；返回码写在 description）
INSERT INTO public.device_model_commands (
	id, device_template_id, data_name, data_identifier, params,
	description, additional_info, created_at, updated_at, tenant_id
)
SELECT (md5('pet-tpl-command-' || m.data_identifier || '-' || u.tenant_id)::uuid)::text,
	(md5('pet-tpl-' || u.tenant_id)::uuid)::text,
	m.data_name, m.data_identifier, m.params::json,
	m.description, '[]'::json, now(), now(), u.tenant_id
FROM (
	SELECT DISTINCT tenant_id FROM public.users
	WHERE tenant_id IS NOT NULL AND tenant_id <> ''
) u
CROSS JOIN (
	VALUES
	('开启实时流', 'start_live',
	 '[{"id":1,"data_name":"请求画质","data_identifier":"quality","param_type":"Enum","data_type":"number","description":"1 流畅；2 标清；3 高清；4 超清","enum_config":[{"value":1,"description":"流畅"},{"value":2,"description":"标清"},{"value":3,"description":"高清"},{"value":4,"description":"超清"}]}]',
	 '观看端点击实时流时调用。返回 0 已开始推流；1 设备忙；2 观看人数已满；3 摄像头无信号'),
	('停止实时流', 'stop_live', '[]',
	 '最后一个观看端退出时调用。返回 0 已停止；1 当前未推流'),
	('抓拍', 'snapshot', '[]',
	 '推流中截取当前一帧并保存。返回 0 抓拍成功；1 未在推流；2 存储失败'),
	('投喂零食', 'dispense_treat', '[]',
	 '远程投放一份零食，受 treat_left 限制。返回 0 已投喂；1 今日次数已用完；2 设备忙；3 余粮不足')
) AS m(data_name, data_identifier, params, description)
ON CONFLICT (device_template_id, data_identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. 前端菜单：宠物看护
-- ---------------------------------------------------------------------------

INSERT INTO public.sys_ui_elements (
	id, parent_id, element_code, element_type, orders, param1, param2, param3,
	authority, description, created_at, remark, multilingual, route_path
) VALUES
('7c9f0a10-1b21-4c31-9d41-000000000001', '0', 'pet', 1, 116, '/pet', 'mdi:paw', '0',
 '["TENANT_ADMIN","SYS_ADMIN"]'::json, '宠物看护', now(), '', 'route.pet', 'layout.base'),
('7c9f0a10-1b21-4c31-9d41-000000000002', '7c9f0a10-1b21-4c31-9d41-000000000001', 'pet_profile', 3, 1161, '/pet/profile', 'mdi:dog-side', '0',
 '["TENANT_ADMIN","SYS_ADMIN"]'::json, '宠物档案', now(), '', 'route.pet_profile', 'view.pet_profile'),
('7c9f0a10-1b21-4c31-9d41-000000000003', '7c9f0a10-1b21-4c31-9d41-000000000001', 'pet_staff-note', 3, 1162, '/pet/staff-note', 'mdi:message-text', '0',
 '["TENANT_ADMIN","SYS_ADMIN"]'::json, '店员留言', now(), '', 'route.pet_staff-note', 'view.pet_staff-note'),
('7c9f0a10-1b21-4c31-9d41-000000000004', '7c9f0a10-1b21-4c31-9d41-000000000001', 'pet_album', 3, 1163, '/pet/album', 'mdi:image-multiple', '0',
 '["TENANT_ADMIN","SYS_ADMIN"]'::json, '相册时间线', now(), '', 'route.pet_album', 'view.pet_album')
ON CONFLICT (id) DO NOTHING;
