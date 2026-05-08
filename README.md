# 秒杀系统设计与实现

## 1. 项目介绍

本项目是一个基于 Go-zero 的后端秒杀系统，实现了秒杀活动创建、用户参与秒杀、订单状态查询等核心功能。

秒杀系统的典型场景是：在短时间内，大量用户同时请求抢购少量商品或名额。  
本项目主要关注后端秒杀场景中的几个核心问题：

- 高并发请求下系统如何保持稳定
- 如何防止库存超卖
- 如何防止用户重复下单
- 如何避免大量请求直接打到 MySQL
- 如何保证订单数据最终落库

本项目可以没有前端，主要通过后端接口完成秒杀流程。

---

## 2. 技术栈

本项目计划使用以下技术：

- Go
- Go-zero
- MySQL
- Redis
- RabbitMQ

各组件作用如下：

| 技术 | 作用 |
|---|---|
| Go | 后端开发语言 |
| Go-zero | 后端微服务框架 |
| MySQL | 存储活动信息和订单信息 |
| Redis | 缓存库存、预减库存、防止重复下单 |
| RabbitMQ | 消息队列，用于削峰和异步创建订单 |

---

## 3. 核心功能

本项目计划实现以下功能：

### 3.1 创建秒杀活动

管理员创建秒杀活动，包括：

- 活动名称
- 活动库存
- 活动开始时间
- 活动结束时间

创建活动后，活动信息会保存到 MySQL，同时将活动库存初始化到 Redis 中。

### 3.2 用户参与秒杀

用户参与秒杀时，系统会依次进行以下判断：

1. 活动是否存在
2. 当前时间是否在活动时间范围内
3. 用户是否已经参与过该活动
4. Redis 中库存是否充足
5. 库存充足则进行 Redis 预减库存
6. 秒杀成功后将订单消息发送到 RabbitMQ
7. 返回“秒杀请求已受理”或“秒杀成功，订单处理中”

### 3.3 异步创建订单

秒杀请求通过 Redis 校验成功后，不直接写入 MySQL，而是将订单消息发送到 RabbitMQ。

后台消费者从 RabbitMQ 中读取消息，并异步创建订单，最终将订单信息写入 MySQL。

### 3.4 查询订单状态

用户可以通过用户 ID 和活动 ID 查询自己的秒杀结果。

订单状态可能包括：

- processing：处理中
- success：秒杀成功
- failed：秒杀失败

---

## 4. 系统整体流程

### 4.1 创建活动流程

```text
创建活动请求
        ↓
Go-zero API 服务接收请求
        ↓
活动信息写入 MySQL
        ↓
活动库存写入 Redis
        ↓
返回活动创建成功
````

### 4.2 秒杀请求流程

```text
用户发起秒杀请求
        ↓
判断活动是否存在
        ↓
判断活动是否在有效时间内
        ↓
判断用户是否重复下单
        ↓
Redis 预减库存
        ↓
库存不足，直接返回失败
        ↓
库存充足，发送订单消息到 RabbitMQ
        ↓
返回秒杀请求已受理
```

### 4.3 订单创建流程

```text
RabbitMQ 中存在订单消息
        ↓
订单消费者读取消息
        ↓
检查 MySQL 中是否已经存在订单
        ↓
不存在则创建订单
        ↓
订单数据写入 MySQL
```

---

## 5. 系统设计思路

### 5.1 使用 Redis 承接高并发请求

秒杀场景下，如果每一个请求都直接访问 MySQL，会导致数据库压力过大。

因此，本项目使用 Redis 存储活动库存，并在 Redis 中进行库存预减。
Redis 读写速度快，适合处理秒杀场景中的高频请求。

### 5.2 使用 Redis 防止库存超卖

库存扣减不直接操作 MySQL，而是在 Redis 中完成。

秒杀请求到达后，系统会先判断 Redis 中库存是否充足。
如果库存大于 0，则扣减库存；如果库存不足，则直接返回秒杀失败。

后续可以使用 Redis Lua 脚本，将“判断库存、判断重复下单、扣减库存、记录用户参与状态”合并为一个原子操作，从而进一步保证并发安全。

### 5.3 使用 Redis 防止重复下单

系统会在 Redis 中记录用户是否已经参与过某个活动。

Redis Key 设计示例：

```text
seckill:user:{activityId}:{userId}
```

如果该 Key 已经存在，说明用户已经参与过该活动，系统会直接拒绝重复请求。

### 5.4 使用 RabbitMQ 削峰

秒杀成功后，如果直接写入 MySQL，大量订单请求可能会瞬间压垮数据库。

因此，本项目使用 RabbitMQ 作为消息队列。
秒杀请求在 Redis 预减库存成功后，会先将订单消息发送到 RabbitMQ，由后台消费者异步创建订单。

这样可以将瞬时高并发请求转化为队列中的消息，由消费者按照一定速度逐步处理，减轻 MySQL 压力。

### 5.5 使用 MySQL 持久化订单数据

MySQL 负责保存最终的活动数据和订单数据。

Redis 主要用于高并发场景下的快速判断和库存预扣减，MySQL 用于保证数据最终可查询、可追溯。

---

## 6. 数据库设计

### 6.1 活动表 activity

| 字段名        | 类型       | 说明     |
| ---------- | -------- | ------ |
| id         | bigint   | 活动 ID  |
| name       | varchar  | 活动名称   |
| stock      | int      | 活动库存   |
| start_time | datetime | 活动开始时间 |
| end_time   | datetime | 活动结束时间 |
| created_at | datetime | 创建时间   |
| updated_at | datetime | 更新时间   |

### 6.2 订单表 seckill_order

| 字段名         | 类型       | 说明    |
| ----------- | -------- | ----- |
| id          | bigint   | 订单 ID |
| user_id     | bigint   | 用户 ID |
| activity_id | bigint   | 活动 ID |
| status      | varchar  | 订单状态  |
| created_at  | datetime | 创建时间  |
| updated_at  | datetime | 更新时间  |

订单表中可以设置联合唯一索引：

```text
user_id + activity_id
```

用于防止同一用户对同一活动重复创建订单。

---

## 7. Redis Key 设计

### 7.1 活动库存 Key

```text
seckill:stock:{activityId}
```

示例：

```text
seckill:stock:1 = 100
```

表示活动 ID 为 1 的活动剩余库存为 100。

### 7.2 用户参与记录 Key

```text
seckill:user:{activityId}:{userId}
```

示例：

```text
seckill:user:1:1001 = 1
```

表示用户 1001 已经参与过活动 1。

---

## 8. RabbitMQ 消息设计

秒杀成功后，发送到 RabbitMQ 的消息格式如下：

```json
{
  "user_id": 1001,
  "activity_id": 1
}
```

消费者接收到消息后，根据 user_id 和 activity_id 创建订单。

---

## 9. 接口设计

### 9.1 创建活动接口

```text
POST /activity/create
```

请求参数示例：

```json
{
  "name": "五一秒杀活动",
  "stock": 100,
  "start_time": "2026-05-08 10:00:00",
  "end_time": "2026-05-08 12:00:00"
}
```

返回示例：

```json
{
  "code": 0,
  "message": "活动创建成功",
  "data": {
    "activity_id": 1
  }
}
```

### 9.2 参与秒杀接口

```text
POST /seckill
```

请求参数示例：

```json
{
  "user_id": 1001,
  "activity_id": 1
}
```

返回示例：

```json
{
  "code": 0,
  "message": "秒杀请求已受理，订单处理中"
}
```

失败返回示例：

```json
{
  "code": 1,
  "message": "库存不足"
}
```

### 9.3 查询订单状态接口

```text
GET /order/status?user_id=1001&activity_id=1
```

返回示例：

```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "status": "success"
  }
}
```

---

## 10. 项目目录规划

```text
seckill-system/
├── README.md
├── go.mod
├── api/
│   └── seckill.api
├── internal/
│   ├── config/
│   ├── handler/
│   ├── logic/
│   ├── svc/
│   └── types/
├── model/
├── mq/
├── scripts/
│   └── init.sql
└── docs/
    └── design.md
```

目录说明：

| 目录               | 说明                          |
| ---------------- | --------------------------- |
| api              | Go-zero 接口定义文件              |
| internal/handler | 接收 HTTP 请求                  |
| internal/logic   | 核心业务逻辑                      |
| internal/svc     | 服务上下文，管理 MySQL、Redis、MQ 等依赖 |
| model            | 数据库模型                       |
| mq               | RabbitMQ 生产者和消费者            |
| scripts          | 数据库初始化 SQL                  |
| docs             | 项目设计文档                      |

---

## 11. 后续计划

本项目后续计划按照以下步骤实现：

1. 搭建 Go-zero 项目结构
2. 编写 MySQL 建表 SQL
3. 实现活动创建接口
4. 创建活动时初始化 Redis 库存
5. 实现秒杀接口
6. 使用 Redis 判断重复下单和预减库存
7. 接入 RabbitMQ
8. 实现订单消费者
9. 实现订单状态查询接口
10. 编写测试和运行说明

---

## 12. 当前进度

* [ ] 确定技术栈
* [ ] 初始化 GitHub 仓库
* [ ] 搭建 Go-zero 项目
* [ ] 编写数据库表
* [ ] 实现活动创建接口
* [ ] 接入 Redis
* [ ] 实现秒杀接口
* [ ] 接入 RabbitMQ
* [ ] 实现订单消费者
* [ ] 实现订单状态查询接口
* [ ] 完善 README 文档
