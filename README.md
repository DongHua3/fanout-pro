# fanout-pro (Fanout 增强版)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/DongHua3/fanout-pro)](https://github.com/DongHua3/fanout-pro/releases)

基于原版 fanout 深度重构与增强的 VPN Gate 出口网关与 3X-UI / Xray 智能分流路由系统。

把 VPN Gate 节点变成本地 SOCKS5 端口，支持 3X-UI 面板入站自由解绑/挂载、全节点质量大厅（家庭宽带/机房识别、IP 纯净度分、多维测速排序）与智能端口推荐。

---

## 🌟 增强版新增特性

1. **🌟 全节点大厅 (Node Explorer)**：
   - 界面顶栏直达「🌟 节点大厅」，可一览所有可用的 VPN Gate 节点。
   - **IP 质量画像**：结合 ippure.com / ping0.cc 理念，智能识别 `🏠 家庭宽带 (Residential)` 与 `🏢 机房 IDC`，计算纯净度与欺诈风险评分，展示运营商 ISP 与 ASN。
   - **多维自由排序**：支持按「质量评分降序」、「速度降序」、「Ping 延迟升序」自由排序与搜索。
   - **一键开通出口**：在大厅中看中任一优质节点，一键即可直开出口并挂载分流。
2. **🔄 3X-UI 自由绑定与一键解绑直连**：
   - **彻底修复解绑 Bug**：解决原版绑定出口后无法解除、无法退回直连的问题。
   - **单 443 核心节点场景完美支持**：支持只建一个 443 节点，日常保留母机直连（`freedom`），需要时自由挂载住宅/多国出口，不需要时一键点击「恢复直连」，瞬时清空 fanout 路由规则。
3. **🎯 常用端口推荐与冲突检测**：
   - 新建或修改入站端口时，智能展示常用推荐端口（443、8443、2053、2083、2096、80、8080 等）。
   - 实时比对本机与 3x-ui 已占用端口，已被占用的标红警示，空闲推荐的一键点击填入。
4. **⚡ 批量质量检测与 24h 本地持久化缓存**：
   - 内置批量 IP 质量评估引擎，自动持久化缓存 24 小时，避免重复消耗请求。
5. **📡 导出全部节点与客户端一键订阅源 (/sub)**：
   - 「导出链接」同时输出母机直连节点（如母机 443/2087 直连）与全部出口挂载节点，不再遗漏直连。
   - 新增标准 Base64 客户端订阅链接（`/sub?token=...`），可直接填入 V2RayN、Shadowrocket、Clash、Sing-box 等客户端，一键同步与自动更新。

---

## 快速安装与更新

需要 root 权限，Linux 系统（依赖 netns 与 openvpn）。

### 一键安装 / 更新

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/DongHua3/fanout-pro/main/install.sh)
```

### 已安装旧版如何更新到 Pro 增强版？

若您的机器上已安装了旧版 fanout，可通过以下任一方式无缝升级到增强版：

**方式一：一键更新命令**
```bash
REPO="DongHua3/fanout-pro" f update
```
或直接重新运行安装脚本（会自动保留您原有的端口、密码与配置）：
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/DongHua3/fanout-pro/main/install.sh)
```

**方式二：源码拉取与编译更新**
```bash
git clone https://github.com/DongHua3/fanout-pro.git
cd fanout-pro
go build -trimpath -ldflags "-s -w" -o /usr/local/bin/fanout .
systemctl restart fanout
```
升级完成后，浏览器按 `Ctrl + F5` 强制刷新管理界面，即可看到顶栏的 **「🌟 节点大厅」**、**核心 443 入站与直连卡片** 以及 **解绑直连** 按钮！

会自动下载对应架构的预编译二进制。也可以 clone 仓库后在源码目录运行同一个脚本，
那样会从源码编译（需要 Go 1.21+）。

依赖（openvpn / curl / openssl / iproute / iptables）会按发行版自动装，
apt、dnf、yum、pacman、apk、zypper 都认。没装 3x-ui 时还会顺带下载一份
Xray 到 `/var/lib/fanout/bin/`，装了则跳过，入站交给面板管。

服务用 systemd 或 OpenRC 都能装，装完自动开机自启。

**Alpine** 默认不带 bash，先装一下：

```bash
apk add bash curl
bash <(curl -fsSL https://raw.githubusercontent.com/DongHua3/fanout-pro/main/install.sh)
```

另外 fanout 要在 netns 里跑 openvpn，**宿主必须放开 `/dev/net/tun`**。
不少 LXC 小鸡没给这个权限，`ls /dev/net/tun` 不存在且 `mknod` 报
Operation not permitted 的话，这台机器用不了，跟发行版无关。

装完敲 `f` 打开管理菜单：

![管理菜单](https://images.joeyblog.net/2026/7/26/fanout-7-menu.png)

装完会打印管理界面地址、访问路径和口令：

```
管理界面  http://<你的IP>:8899/gwPuWHvaNr/
访问口令  f81120ac328d11c11b
```

路径和口令都是随机生成的，分别存在 `/var/lib/fanout/basepath` 和
`/var/lib/fanout/password`。路径不对一律返回 404，扫端口的看不到这里跑着什么。

## 使用

界面以**出口**为单位：一行就是一条隧道加上挂在它上面的节点链接。

点「新建出口」，选地区和数量，再选一个已有节点作模板，提交后 fanout 会并行
拉起隧道、为每个出口复制一份节点链接并绑好，进度按目标逐条回报。原来要手点
五步跨两栏的事，现在一次点击十几秒完成。

![新建出口](https://images.joeyblog.net/2026/7/27/fanout-wizard.png)

每行右侧两个按钮：换一个节点（出口 IP 变、端口不变，已分发的客户端配置不用改），
或者停掉这个出口。

点节点名进详情，可以改端口、备注、启停，管理客户端，以及改绑到别的出口：

![节点详情](https://images.joeyblog.net/2026/7/27/fanout-detail.png)

一个入站可以挂多套客户端凭据，分发给不同的人；每套都能单独重置，
重置后旧链接立即失效。

「导出链接」一次性拿到所有节点链接：

![导出链接](https://images.joeyblog.net/2026/7/27/fanout-export.png)

### 节点链接从哪来

同机装了 3x-ui 就直接接管面板里的入站，面板端口、路径、API token 全自动探测，
开了 SSL 也能识别。没装 3x-ui 时 fanout 自己跑一个 Xray，界面上多一个「新建节点」
按钮，可以选协议（VLESS / VMess / Trojan）、传输（TCP / WebSocket / gRPC /
HTTPUpgrade / XHTTP）和安全层（无 / TLS / REALITY）。

![新建节点](https://images.joeyblog.net/2026/7/27/fanout-newnode.png)

REALITY 的密钥对和 shortId 自动生成；TLS 不填证书就生成自签的，分享链接会带上
证书指纹让客户端固定信任。也可以填自己的证书路径。

接管 3x-ui 和自建这两种模式下，改端口、启停、加删客户端、绑定出口的操作完全一致，
用起来没有区别。

装了 [xray-cf-lite](https://github.com/byJoey/xray-cf-lite) 的机器会自动接管它生成的
三个节点。这个模式下节点归 xray-cf-lite 管，fanout 只负责给每个节点指定走哪条出口，
所以界面上不提供新建、删除和改节点的入口——想改端口或 UUID 去 xray-cf-lite 那边改。
两边共用同一份 Xray 配置，fanout 只往里加自己前缀的出站和分流规则，互不覆盖。

后端在设置面板里可以随时切换，本机没装的会置灰并说明原因；也可以用
`-panel 3x-ui` / `-panel native` / `-panel xray-cf-lite` 启动参数固定。
界面里选过之后会记住，重启仍然生效。

## 运维

装完后敲 `f` 打开管理菜单：启停、看日志、查隧道、改端口/口令/访问路径、更新、卸载。

```
  状态      运行中
  版本      fanout v0.1.1
  开机自启  enabled

  管理地址  http://1.2.3.4:8899/gwPuWHvaNr/
  访问口令  f81120ac328d11c11b

   1) 启动          2) 停止
   3) 重启          4) 查看日志
   5) 隧道列表      6) 连接信息
   7) 改端口        8) 改口令
   9) 改访问路径   10) 开机自启开关
  11) 更新         12) 卸载
```

也可以直接带参数用：

```bash
f info       # 连接信息
f list       # 隧道列表
f restart    # 重启
f log        # 跟踪日志
f update     # 更新到最新版
f uninstall  # 卸载
```

隧道状态存在 `/var/lib/fanout/state.json`，重启后自动恢复，端口保持不变。

健康检查每 10 秒跑一次，比对出口 IP 是否还是建立隧道时那个——openvpn 挂掉后
netns 仍能经母机 NAT 出网，只看通不通会漏判。连续两次不符就自动换节点重连，
槽位和端口不变，原先指向它的节点链接会自动改绑过去。

## 已知限制

- 只转发 TCP。SOCKS5 收到域名时在本机解析，隧道内不跑 UDP/DNS。
- VPN Gate 是志愿者节点，有相当比例已下线或满员（`AUTH_FAILED`）。
  启动时连不上会自动顺着同地区候选往下试，最多 6 个。
- 管理界面只有随机路径 + 口令登录，没有 HTTPS。放公网建议前面套一层反代。

## 许可

[MIT](LICENSE)。

节点来自 [VPN Gate](https://www.vpngate.net/)（筑波大学的学术实验项目），
本工具只是调用其公开的节点列表并用官方 openvpn 客户端连接，不修改也不代理其服务。
使用时请遵守 VPN Gate 的条款和你所在地的法律。

## 交流

- 交流群：<https://t.me/+ft-zI76oovgwNmRh>
- 视频教程：<https://youtube.com/@joeyblog>
- 博客：<https://joeyblog.net>

用着有问题、或者想要什么功能，去群里说或提 issue。
