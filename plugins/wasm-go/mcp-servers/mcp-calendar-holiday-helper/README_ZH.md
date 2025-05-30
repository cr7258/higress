# 中国黄历/假期助手

API认证需要的APP Code请在阿里云API市场申请: https://market.aliyun.com/apimarket/detail/cmapi00066017

## 什么是云市场API MCP服务

阿里云云市场是生态伙伴的交易服务平台，我们致力于为合作伙伴提供覆盖上云、商业化和售卖的全链路服务，帮助客户高效获取、部署和管理优质生态产品。云市场的API服务涵盖以下几个类目：应用开发、身份验证与金融、车辆交通与物流、企业服务、短信与运营商、AI应用与OCR、生活服务。
云市场API依托Higress提供MCP服务，您只需在云市场完成订阅并获取AppCode，通过Higress MCP Server进行配置，即可无缝集成云市场API服务。

## 如何在使用云市场API MCP服务

1. 进入API详情页，订阅该API。您可以优先使用免费试用。
2. 前往云市场用户控制台，使用阿里云账号登陆后查看已订阅API服务的AppCode，并配置到Higress MCP Server的配置中。注意：在阿里云市场订阅API服务后，您将获得AppCode。对于您订阅的所有API服务，此AppCode是相同的，您只需使用这一个AppCode即可访问所有已订阅的API服务。
3. 云市场用户控制台会实时展示已订阅的预付费API服务的可用额度，如您免费试用额度已用完，您可以选择重新订阅。

# MCP服务器配置文档

## 功能简介
`calendar-holiday-helper`服务器是一个专注于提供节假日相关信息以及黄历运势查询的服务平台。它支持多种API调用来获取包括但不限于节假日列表、具体日期的节假日详情、以及基于中国传统文化的黄历信息等数据。这些服务对于需要根据特定日期安排活动或希望了解某日吉凶情况的个人和组织非常有用。

## 工具简介

### 1. 节假日列表
- **描述**：此工具用于列出指定年份内的所有节假日。
- **应用场景**：适用于企业规划年度假期、旅游行业制定促销计划等场合。
- **参数说明**：
  - `year` (string)：需要查询的年份，默认查当年。非当年日期也返回当年节假日数据；来年的数据需等到当年12月份才能查询。

### 2. 节假日详情
- **描述**：该工具提供了一个具体的日期（默认为当天）下的节假日详细信息。
- **应用场景**：适合于个人或团队想要了解某一天是否为节假日及其具体名称时使用。
- **参数说明**：
  - `date` (string)：查询的日期，默认为当天。
  - `needDesc` (string)：是否需要返回当日公众日、国际日和我国传统节日的简介，值为1表示返回，默认不返回。

### 3. 黄历运势_新版_吉时
- **描述**：提供了基于中国传统历法的每日吉时查询服务。
- **应用场景**：对那些相信选择吉时进行重要决策的人群特别有用。
- **参数说明**：
  - `date` (string, 必填)：查询的日期，格式为yyyyMMdd。

### 4. 黄历运势_新版_吉神凶煞
- **描述**：展示了特定日期内影响运势的吉神与凶煞信息。
- **应用场景**：帮助用户避开不利因素并抓住有利时机。
- **参数说明**：
  - `date` (string, 必填)：查询的日期，格式为yyyyMMdd。

### 5. 黄历运势_新版_黄历
- **描述**：综合了农历、公历以及其他相关天文学信息的日历服务。
- **应用场景**：广泛应用于日常生活中的各种习俗活动安排。
- **参数说明**：
  - `date` (string, 必填)：查询的日期，格式为yyyyMMdd。

以上就是`calendar-holiday-helper`服务器提供的主要工具和服务概述。通过合理利用这些工具，用户能够更有效地管理时间，并根据需要调整自己的活动安排。
