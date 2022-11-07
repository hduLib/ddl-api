# ddl-api

## feature

- [x] 超星
- [ ] 省教育平台
- [ ] 中国大学mooc
- [ ] 智慧树

## api

GET `/dll/all`

### query

| key          | value  | description       |
|--------------|--------|-------------------|
| cx_account   |        |                   |
| cx_passwd    |        |                   |
| cx_loginType | cas/cx | 可选杭电cas登陆或者超星账号登陆 |

### resp

| key    | value         | description |
|--------|---------------|-------------|
| code   | 0/-1          | -1为寄        |
| data   | Array<ddl>    |             |
| msg    | string        | 错误          |
| errors | Array<string> | 错误列表        |

**ddl**

| key    | value               | description |
|--------|---------------------|-------------|
| Course | 课程名称                |             |
| Title  | 作业/考试标题             |             |
| Time   | ddl                 |             |
| Type   | 作业/考试               |             |
| From   | 超星/省平台/中国大学mooc/智慧树 |             |
