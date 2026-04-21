<p align="center">
  <img src="https://img.shields.io/github/license/libai07/routecheck?color=blueviolet">
</p>

## RouteCheck

中文回程路由检测工具。

## 功能

- 默认测试北京、上海、广州、枣庄、济宁、吉林三网
- 支持自定义 JSON 目标文件
- 自动识别三网顶级线路：电信 CN2 GIA、联通 9929、移动 CMIN2
- 一键安装 amd64 / arm64 Linux 二进制

## 安装

```sh
curl -sSf https://raw.githubusercontent.com/libai07/routecheck/main/install.sh | sh
```

安装脚本会自动处理 root / sudo。若系统支持 `setcap`，会自动授予 `cap_net_raw`，普通用户也可以运行。

## 使用

默认测试：

```sh
routecheck
```

测试自定义目标：

```sh
routecheck -targets ./targets.example.json
```

安装后直接使用自定义目标：

```sh
curl -sSf https://raw.githubusercontent.com/libai07/routecheck/main/install.sh | sh -s -- -targets ./targets.example.json
```

## 默认目标

- 北京电信、北京联通、北京移动
- 上海电信、上海联通、上海移动
- 广州电信、广州联通、广州移动
- 枣庄电信、枣庄联通、枣庄移动
- 济宁电信、济宁联通、济宁移动
- 吉林电信、吉林联通、吉林移动

## 自定义目标格式

```json
[
  { "name": "枣庄电信", "ip": "219.146.1.66" },
  { "name": "济宁联通", "ip": "202.102.154.3" },
  { "name": "吉林移动", "ip": "211.141.16.99" }
]
```

也支持对象格式：

```json
{
  "targets": [
    { "name": "枣庄电信", "ip": "219.146.1.66" },
    { "name": "济宁联通", "ip": "202.102.154.3" },
    { "name": "吉林移动", "ip": "211.141.16.99" }
  ]
}
```
