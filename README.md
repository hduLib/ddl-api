# ddl-api

## feature

- [x] 超星
- [x] 省教育平台
- [ ] 中国大学mooc
- [ ] 智慧树
- [x] 课表

## api

### ddl

GET `/dll/all`

#### query

| key           | value  | description       |
|---------------|--------|-------------------|
| cx_account    | string |                   |
| cx_passwd     | string |                   |
| cx_loginType  | cas/cx | 可选杭电cas登陆或者超星账号登陆 |
| zjooc_account | string |                   |
| zjooc_passwd  | string |                   |

#### resp

| key    | value         | description            |
|--------|---------------|------------------------|
| code   | 1/0/-1        | -1为寄，1为有错误，但是依旧返回了部分数据 |
| data   | Array<ddl>    |                        |
| msg    | string        | 错误或ok                  |
| errors | Array<string> | 错误列表                   |

**ddl**

| key    | value               | description |
|--------|---------------------|-------------|
| Course | 课程名称                | string      |
| Title  | 作业/考试标题             | string      |
| Time   | ddl                 | 时间戳（秒）      |
| Type   | 作业/测验/考试            | string      |
| From   | 超星/省平台/中国大学mooc/智慧树 | string      |

### 课表

GET `/courses/today`

> 仿写hdu-scriptable的课表api,提供无缝兼容,风格与ddl的风格不同

#### query

| key      | value  | description       |
|----------|--------|-------------------|
| username |        |                   |
| password |        |                   |

#### resp

| key  | value         | description |
|------|---------------|-------------|
| code | 0/-1          | -1为寄        |
| data | Array<course> |             |
| msg  | string        |             |

**course**

为了对hdu-scriptable的课表api保持兼容，几乎直接返回zjooc的课表api的payload，也有所缩减

