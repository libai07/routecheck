<p align="center">
  <img src="https://img.shields.io/github/license/libai07/routecheck?color=blueviolet">
</p>

## RouteCheck

中文回程路由检测工具。

## 功能

- 默认测试北京、上海、天津、重庆四个直辖市三网目标
- 支持自定义 JSON 目标文件
- 自动识别三网顶级线路：电信 CN2 GIA、联通 9929、移动 CMIN2
- 只使用 NextTrace / nexttrace-tiny 探测 hop，RouteCheck 自己按 CIDR 规则判定线路
- 一键安装 amd64 / arm64 Linux 二进制

## 判定口径

RouteCheck 检测的是本机到国内目标 IP 的回程路径。顶级线路只按路径中可见的核心骨干特征段判定：

- 电信 CN2 GIA：`59.43.0.0/16`
- 联通 9929：`218.105.0.0/16`
- 移动 CMIN2：`223.120.128.0/17`

RouteCheck 只读取 NextTrace 探测到的 hop IP；线路结论由 RouteCheck 的固定 CIDR 规则判定，不做目标 ASN 校验，也不依赖 NextTrace 的 GeoIP/ASN 数据。

普通骨干辅助识别：

- 电信 163：`202.97.0.0/16`
- 联通 4837：`219.158.0.0/16`
- 移动 CMI：`223.120.0.0/17`
- 移动 CMNET：`221.183.0.0/16`

## 安装

```sh
curl -sSf https://raw.githubusercontent.com/libai07/routecheck/main/install.sh | sh
```

安装脚本会自动处理 root / sudo，把 `routecheck` 安装到 `/usr/local/bin/routecheck`。如果本机没有 `nexttrace` 或 `nexttrace-tiny`，会自动通过 [NextTrace 官方安装脚本](https://nxtrace.org/nt) 安装。

如果手动安装二进制，运行前需要确保 `nexttrace` 或 `nexttrace-tiny` 在 `PATH` 中可用。

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
- 天津电信、天津联通、天津移动
- 重庆电信、重庆联通、重庆移动

## 自定义目标格式

```json
[
  { "name": "北京电信", "ip": "219.141.140.10" },
  { "name": "上海联通", "ip": "210.22.70.3" },
  { "name": "天津移动", "ip": "211.137.160.50" },
  { "name": "重庆电信", "ip": "61.128.192.68" }
]
```

也支持对象格式：

```json
{
  "targets": [
    { "name": "北京电信", "ip": "219.141.140.10" },
    { "name": "上海联通", "ip": "210.22.70.3" },
    { "name": "天津移动", "ip": "211.137.160.50" },
    { "name": "重庆电信", "ip": "61.128.192.68" }
  ]
}
```
