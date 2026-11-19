# 宠物看护 · C 端 App 与 B 端管理系统 开发计划

涉及仓库：`thingspanel-backend-community`、`thingspanel-frontend-community`、`app`（pet / pet-wxmini 两分支），统一在 `1.2.1-dev` / 对应分支上开发。

---

## 一、角色与现状

| 端 | 使用方 | 现状 |
| --- | --- | --- |
| app（pet / pet-wxmini） | C 端宠物家长 | 面板 UI 完整（实时/动态/相册/更多）；但鉴权走后台 `x-token`，宠物档案、日报、留言、时间线、相册全部为本地 mock，未接入 C 端体系 |
| 管理系统（front + backend） | B 端：经销商 / 宠物店 / 宠物医院 | 已有行业方案管理页 + 宠物档案/店员留言/相册 3 页；设备 CRUD 复用 ThingsPanel 原生能力；缺门店/经销商层级、运营看板与告警闭环 |

关键断点：

1. 后端已有 C 端体系（`/api/v1/c/register|login|devices|devices/bind`，独立 JWT），但只提供设备绑定，**没有任何 C 端宠物业务接口**。
2. `pet/*` 系列接口全部按 `tenantID`（租户后台维度）过滤，C 端用户无法使用。
3. `POST /c/devices/bind` 只校验「设备存在 + 未重复绑定」，**任意注册用户可绑定任意设备**，无授权。
4. `pet/rtc/token` 挂在后台 JWT 下，C 端拿不到入会凭证。
5. 租户为扁平结构，缺「经销商 → 门店」层级。

---

## 二、已确认决策

1. **绑定方式**：门店出码（门店生成绑定码 / 二维码），家长扫码绑定。
2. **家长登录**：先做手机号 + 密码（复用现有 `/c/register`、`/c/login`），暂不做短信验证码。
3. **C 端信息架构**：按宠物组织 —— 首页为宠物列表，进入某只宠物后打开看护面板。
4. **租户层级**：走租户父子结构，支持「经销商 → 宠物医院 → 门店」三级，也支持单一角色（纯门店 / 纯医院 / 纯经销商）；行业方案可沿层级下放。

---

## 三、Phase 0：打通「家长 ↔ 门店」最小闭环

### M1 后端（thingspanel-backend-community）

- `sql/25.sql`（幂等）：
  - 新增 `device_bind_codes`：`id / tenant_id / device_id / code / expires_at / max_uses / used_count / status / created_by / created_at / updated_at`，`code` 唯一索引。
  - 租户表新增 `parent_id` 与索引（先核对 `sql/1.sql` 中租户表实际结构再定稿）。
- C 端接口（新增 `internal/service/c_pet.go`，复用现有 pet service，`tenantID` 由绑定关系推导，并校验「当前 C 端用户已绑定该设备」）：
  - `GET  /api/v1/c/pets` —— 按绑定关系返回宠物档案 + 设备 + 门店
  - `GET  /api/v1/c/pets/:device_id/profile`
  - `GET  /api/v1/c/pets/:device_id/daily-report`
  - `GET  /api/v1/c/pets/:device_id/staff-notes`
  - `GET  /api/v1/c/pets/:device_id/timeline`
  - `GET  /api/v1/c/pets/:device_id/album`
  - `POST /api/v1/c/pets/:device_id/rtc/token`
- `POST /api/v1/c/devices/bind` 改造：入参由裸 `device_id` 改为 `code`，校验有效期 / 可用次数 / 门店归属，绑定成功后累加 `used_count`。
- 方案分配支持沿租户层级下放（经销商为下属门店开通）。
- 验证：curl 串通 注册 → 门店生成码 → 扫码绑定 → 取宠物列表 → 取 RTC token。

### M2 B 端管理系统（thingspanel-frontend-community）

- 门店「绑定码管理」页：生成 / 复制二维码 / 失效 / 查看已用次数。
- 租户层级树展示 + 经销商给下属门店开通行业方案。
- supabase 与 men 菜单种子同步更新（`sys_ui_elements`）。

### M3 C 端 App（pet 与 pet-wxmini 两分支同步）

- 新增页面：登录 / 注册、宠物列表、扫码绑定。
- 服务层新增 C 端 API 封装（`access_token` + `serverAddress`），把 `pet-platform.*` 的 mock 切到 `/c/pets/*`。
- 面板入口由宠物列表带 `device_id` 进入；RTC 入会改走 C 端 token 接口。
- 兜底：微信开发者工具下支持手输绑定码（相机扫码需真机）。

---

## 四、Phase 1：B 端可运营 + C 端可感知

### B 端

- 组织层级落地：经销商 → 门店数据可见性、下属门店汇总。
- 门店工作台：设备在线率、观看情况、待办（项圈电量低 / 生理指标异常）、今日日报完成情况。
- 补齐后台页面：日报管理、时间线管理（后端接口已就绪、前端缺页）、C 端用户管理、告警 / 事件中心（7 个事件 + 级别 + 处置状态）。
- 医院场景扩展：笼位 / 住院记录、就诊与护理交接。

### C 端

- 告警触达：小程序订阅消息 / App 推送（后端已有 Aliyun SMS 与 SMTP，未与宠物事件联动）。

### 后端（共建）

- 日报生成链路：AI 行为事件 → 定时聚合 → `pet_daily_reports`（当前只有 `GET`，无写入方）。
- 事件处置状态机 + 推送服务。

验收：门店能看到本店告警并标记处置；异常发生后家长手机收到提醒。

---

## 五、技术前提与风险

1. **租户表结构**：需先核对 `sql/1.sql` 中租户表字段，决定 `parent_id` 是新增还是复用现有层级字段。
2. **一设备一宠物**：`pet_profiles.device_id` 为唯一键，当前「多宠物 = 多设备」；若后续要一设备多宠物需改唯一键。
3. **小程序扫码**：需相机权限 + 真机调试；音视频 `live-pusher/live-player` 同样不支持开发者工具。
4. **跨仓库文档**：本计划同时约束三个仓库，改动提交时按仓库分别提交。

---

## 六、交付顺序

`M1 后端` → `M2 B 端绑定码与层级` → `M3 C 端登录 / 宠物列表 / 绑定 / 面板接真实数据`，先做到「能绑、能看、能对讲」，再推进 Phase 1 运营能力。
