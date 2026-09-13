#!/usr/bin/env bash
# fanout 管理菜单
set -uo pipefail

WORK_DIR=/var/lib/fanout
SERVICE=fanout
BIN=/usr/local/bin/fanout
REPO="${REPO:-DongHua3/fanout-pro}"

# 样式与高对比终端配色
G='\033[1;32m'; R='\033[1;31m'; Y='\033[1;33m'; B='\033[1;36m'; M='\033[1;35m'; W='\033[1;37m'; D='\033[0;90m'; N='\033[0m'

need_root() {
  [[ $EUID -eq 0 ]] || { echo -e "${R}需要 root${N}"; exit 1; }
}

# ── init 系统抽象：systemd 与 OpenRC ────────────────────
if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
  INIT_SYS=systemd
  UNIT=/etc/systemd/system/${SERVICE}.service
else
  INIT_SYS=openrc
  UNIT=/etc/init.d/${SERVICE}
fi

svc_start()   { [[ $INIT_SYS == systemd ]] && systemctl start "$SERVICE"   || rc-service "$SERVICE" start; }
svc_stop()    { [[ $INIT_SYS == systemd ]] && systemctl stop "$SERVICE"    || rc-service "$SERVICE" stop; }
svc_restart() { [[ $INIT_SYS == systemd ]] && systemctl restart "$SERVICE" || rc-service "$SERVICE" restart; }
svc_reload()  { [[ $INIT_SYS == systemd ]] && systemctl daemon-reload || true; }
svc_enable()  { [[ $INIT_SYS == systemd ]] && systemctl enable "$SERVICE" >/dev/null 2>&1 || rc-update add "$SERVICE" default >/dev/null 2>&1; }
svc_disable() { [[ $INIT_SYS == systemd ]] && systemctl disable "$SERVICE" >/dev/null 2>&1 || rc-update del "$SERVICE" default >/dev/null 2>&1; }

svc_is_enabled() {
  if [[ $INIT_SYS == systemd ]]; then
    systemctl is-enabled --quiet "$SERVICE"
  else
    rc-update show default 2>/dev/null | grep -q "^ *${SERVICE} "
  fi
}

svc_enabled_text() {
  svc_is_enabled && echo enabled || echo disabled
}

svc_status_page() {
  if [[ $INIT_SYS == systemd ]]; then
    systemctl status "$SERVICE" --no-pager
  else
    rc-service "$SERVICE" status
  fi
}

svc_logs() {
  if [[ $INIT_SYS == systemd ]]; then
    journalctl -u "$SERVICE" -n "${1:-50}" --no-pager
  else
    tail -n "${1:-50}" /var/log/${SERVICE}.log 2>/dev/null || echo "  暂无日志"
  fi
}

svc_logs_follow() {
  if [[ $INIT_SYS == systemd ]]; then
    journalctl -u "$SERVICE" -f
  else
    tail -f /var/log/${SERVICE}.log
  fi
}

svc_state() {
  if [[ $INIT_SYS == systemd ]]; then
    systemctl is-active --quiet "$SERVICE" && echo running || echo stopped
  else
    rc-service "$SERVICE" status >/dev/null 2>&1 && echo running || echo stopped
  fi
}

# 端口以 settings.json 为准。老版本把 -web 写死在服务文件里，
# 两处各改各的会互相拽回旧值，所以这里只认工作目录下的配置。
web_port() {
  local p
  p=$(sed -n 's/.*"port"[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p' \
        "$WORK_DIR/settings.json" 2>/dev/null | head -1)
  [[ -n $p ]] && { echo "$p"; return; }
  # 兼容老安装：settings.json 还没生成时退回读服务文件
  grep -oE '\-web [0-9]+' "$UNIT" 2>/dev/null \
    | grep -oE '[0-9]+' | head -1 || echo 8899
}

public_ip() {
  curl -s --max-time 6 http://api.ipify.org 2>/dev/null || echo "<本机IP>"
}

pause() {
  echo
  echo -ne "  ${D}按回车键继续...${N}"
  read -r _
}

show_info() {
  local state port bp pw ip dom cf sm la ver n state_badge autostart_badge
  state=$(svc_state); port=$(web_port)
  bp=$(cat "$WORK_DIR/basepath" 2>/dev/null || echo "")
  pw=$(cat "$WORK_DIR/password" 2>/dev/null || echo "-")
  ip=$(public_ip)
  ver=$("$BIN" -version 2>/dev/null || echo 'v2.0.8')

  dom=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  cf=$(sed -n 's/.*"cert_file"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  sm=$(sed -n 's/.*"ssl_mode"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  la=$(sed -n 's/.*"listen_addr"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)

  n=$(ls -d /var/run/netns/fo* 2>/dev/null | wc -l | tr -d ' ')

  if [[ $state == running ]]; then
    state_badge="${G}[运行中]${N}"
  else
    state_badge="${R}[已停止]${N}"
  fi

  if svc_is_enabled; then
    autostart_badge="${G}[已开启]${N}"
  else
    autostart_badge="${D}[已关闭]${N}"
  fi

  local path_part=""
  [[ -n "$bp" && "$bp" != "-" ]] && path_part="${bp#/}/"

  local panel_url="" proto_tag=""
  if [[ "$sm" == "caddy" ]] && [[ -n "$dom" ]]; then
    panel_url="https://${dom}/${path_part}"
    proto_tag="${G}[Caddy 443 反代]${N}"
  elif [[ -n "$cf" ]] && [[ "$sm" != "none" ]]; then
    local host_part="${dom:-$ip}"
    if [[ "$port" == "443" ]]; then
      panel_url="https://${host_part}/${path_part}"
    else
      panel_url="https://${host_part}:${port}/${path_part}"
    fi
    proto_tag="${G}[原生 HTTPS]${N}"
  elif [[ -n "$dom" ]]; then
    panel_url="http://${dom}:${port}/${path_part}"
    proto_tag="${Y}[纯域名 HTTP]${N}"
  else
    panel_url="http://${ip}:${port}/${path_part}"
    proto_tag="${D}[公网 IP 访问]${N}"
  fi

  echo
  echo -e "  ${B}┌─ [ 系统实时运行看板 ] ────────────────────────────────────────┐${N}"
  echo -e "  ${B}│${N}  服务状态: ${state_badge}          核心版本: ${B}${ver}${N}              ${B}│${N}"
  echo -e "  ${B}│${N}  开机自启: ${autostart_badge}              活跃隧道: ${G}${n} 个出口${N}           ${B}│${N}"
  echo -e "  ${B}├─ [ 管理面板访问凭据 ] ────────────────────────────────────────┤${N}"
  echo -e "  ${B}│${N}  管理地址: ${W}${panel_url}${N}"
  echo -e "  ${B}│${N}  协议模式: ${proto_tag}"
  if [[ "$sm" == "caddy" ]] && [[ -n "$dom" ]]; then
    echo -e "  ${B}│${N}  本地监听: ${D}http://127.0.0.1:${port}/${path_part}${N}"
  fi
  echo -e "  ${B}│${N}  访问口令: ${Y}${pw}${N}"
  echo -e "  ${B}└───────────────────────────────────────────────────────────────┘${N}"
}

list_tunnels() {
  local port bp pw ck
  port=$(web_port)
  bp=$(cat "$WORK_DIR/basepath" 2>/dev/null || echo "")
  pw=$(cat "$WORK_DIR/password" 2>/dev/null || echo "")
  ck=$(mktemp)

  curl -s --max-time 10 -c "$ck" -X POST -d "password=${pw}" \
    "http://127.0.0.1:${port}/${bp}/login" -o /dev/null
  echo
  curl -s --max-time 10 -b "$ck" "http://127.0.0.1:${port}/${bp}/api/tunnels" \
    > "$ck.json" 2>/dev/null
  rm -f "$ck"

  echo -e "  ${B}╔═══════════════════════════════════════════════════════════════╗${N}"
  echo -e "  ${B}║${N}                     ${W}当前运行中的隧道出口列表${N}                  ${B}║${N}"
  echo -e "  ${B}╚═══════════════════════════════════════════════════════════════╝${N}"
  echo

  if [[ ! -s "$ck.json" ]] || ! grep -q '"port"' "$ck.json" 2>/dev/null; then
    echo -e "  ${D}暂无运行中的隧道出口，请在 Web 管理面板中添加节点。${N}"
  else
    printf "  ${B}%-8s${N}  ${W}%-10s${N}  ${G}%-18s${N}  ${D}%s${N}\n" "本地端口" "运行状态" "出口公网 IP" "VPN Gate 节点"
    echo -e "  ${D}───────────────────────────────────────────────────────────────${N}"
    sed 's/{"slot"/\n{"slot"/g' "$ck.json" | while IFS= read -r line; do
      case "$line" in *'"slot"'*) ;; *) continue ;; esac
      p=$(echo "$line"  | sed -n 's/.*"port":\([0-9]*\).*/\1/p')
      st=$(echo "$line" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
      ip=$(echo "$line" | sed -n 's/.*"exit_ip":"\([^"]*\)".*/\1/p')
      hn=$(echo "$line" | sed -n 's/.*"hostname":"\([^"]*\)".*/\1/p')
      [[ -z $p ]] && continue
      local st_colored="${D}${st:--}${N}"
      [[ "$st" == "connected" || "$st" == "up" || "$st" == "running" ]] && st_colored="${G}● 活跃${N}"
      printf "  ${Y}%-8s${N}  %-19b  ${W}%-18s${N}  ${D}%s${N}\n" "$p" "$st_colored" "${ip:--}" "${hn:--}"
    done
  fi
  rm -f "$ck.json"
}

change_port() {
  local cur new
  cur=$(web_port)
  echo
  echo -e "  ${B}┌─ [ 修改管理面板监听端口 ] ───────────────────────────────────┐${N}"
  echo -e "  ${B}│${N}  当前端口: ${Y}${cur}${N}                                               ${B}│${N}"
  echo -e "  ${B}└──────────────────────────────────────────────────────────────┘${N}"
  read -rp "  请输入新端口 [1-65535] (回车保持不变): " new
  [[ -z $new ]] && { echo -e "  ${D}已取消修改${N}"; return; }
  if ! [[ $new =~ ^[0-9]+$ ]] || (( new < 1 || new > 65535 )); then
    echo -e "  ${R}[错误] 端口不合法，必须为 1-65535 的纯数字${N}"; return
  fi
  if ss -tln 2>/dev/null | grep -q ":${new} "; then
    echo -e "  ${R}[错误] 端口 ${new} 已被系统其它进程占用${N}"; return
  fi
  if [[ -f "$WORK_DIR/settings.json" ]]; then
    update_json_val "port" "$new"
  else
    printf '{\n  "port": %s,\n  "listen_addr": ""\n}\n' "$new" > "$WORK_DIR/settings.json"
    chmod 600 "$WORK_DIR/settings.json"
  fi
  sed -i "s/-web ${cur}/-web ${new}/" "$UNIT" 2>/dev/null
  svc_reload
  svc_restart
  echo -e "  ${G}[成功] 管理面板监听端口已成功改为: ${new} 并重启生效${N}"
}

reset_password() {
  local pw
  echo
  echo -e "  ${B}┌─ [ 修改管理面板访问口令 ] ───────────────────────────────────┐${N}"
  echo -e "  ${B}│${N}  改完后只影响新登录会话，当前已登录客户端无需重复输入        ${B}│${N}"
  echo -e "  ${B}└──────────────────────────────────────────────────────────────┘${N}"
  read -rp "  请输入新口令 (留空则随机生成无歧义密码): " pw
  if [[ -z $pw ]]; then
    pw=$(head -c 9 /dev/urandom | od -An -tx1 | tr -d ' \n')
  fi
  umask 077
  echo "$pw" > "$WORK_DIR/password"
  svc_restart
  echo -e "  ${G}[成功] 访问口令已成功重置为: ${Y}${pw}${N}"
}

reset_basepath() {
  local bp
  echo
  echo -e "  ${B}┌─ [ 修改管理面板访问路径前缀 ] ───────────────────────────────┐${N}"
  echo -e "  ${B}│${N}  隐藏在自定义路径后，能有效阻断端口全网扫描探针探测          ${B}│${N}"
  echo -e "  ${B}└──────────────────────────────────────────────────────────────┘${N}"
  read -rp "  请输入新访问路径 (例如 mypanel，留空则随机生成): " bp
  if [[ -z $bp ]]; then
    rm -f "$WORK_DIR/basepath"
    svc_restart
    sleep 2
    bp=$(cat "$WORK_DIR/basepath" 2>/dev/null)
  else
    bp=${bp#/}; bp=${bp%/}
    umask 077
    echo "$bp" > "$WORK_DIR/basepath"
    svc_restart
  fi
  echo -e "  ${G}[成功] 访问路径已成功更新为: ${B}/${bp}/${N}"
}

ipv6_state() {
  local a d
  a=$(sysctl -n net.ipv6.conf.all.disable_ipv6 2>/dev/null || echo 0)
  d=$(sysctl -n net.ipv6.conf.default.disable_ipv6 2>/dev/null || echo 0)
  [[ "$a" == 1 && "$d" == 1 ]] && echo disabled || echo enabled
}

toggle_ipv6() {
  local conf=/etc/sysctl.d/99-fanout-ipv6.conf
  echo
  if [[ $(ipv6_state) == disabled ]]; then
    read -rp "  当前已禁用 IPv6，要重新启用吗？[y/N]: " yes
    [[ ${yes,,} == y ]] || { echo "  已取消"; return; }
    rm -f "$conf"
    sysctl -qw net.ipv6.conf.all.disable_ipv6=0
    sysctl -qw net.ipv6.conf.default.disable_ipv6=0
    sysctl -qw net.ipv6.conf.lo.disable_ipv6=0
    echo -e "  ${G}已重新启用 IPv6${N}"
    return
  fi

  echo -e "  ${D}母机有全局 IPv6 时，没走隧道的流量可能从 IPv6 出去，暴露真实地址。${N}"
  read -rp "  确认禁用整机 IPv6？[y/N]: " yes
  [[ ${yes,,} == y ]] || { echo "  已取消"; return; }

  cat > "$conf" <<EOF
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
net.ipv6.conf.lo.disable_ipv6 = 1
EOF
  sysctl -qw net.ipv6.conf.all.disable_ipv6=1
  sysctl -qw net.ipv6.conf.default.disable_ipv6=1
  sysctl -qw net.ipv6.conf.lo.disable_ipv6=1
  svc_restart >/dev/null 2>&1
  echo -e "  ${G}已禁用 IPv6（重启后依然生效）${N}"
}

show_links() {
  echo
  echo -e "  ${B}┌─ [ 项目官方开源主页 ] ────────────────────────────────────────┐${N}"
  echo -e "  ${B}│${N}  GitHub 仓库: ${W}https://github.com/DongHua3/fanout-pro${N}        ${B}│${N}"
  echo -e "  ${B}│${N}  欢迎前往仓库 Star 点赞、提交反馈与查看最新版本发布日志         ${B}│${N}"
  echo -e "  ${B}└───────────────────────────────────────────────────────────────┘${N}"
}

# 老版本把 -web 写死在服务文件里，和 settings.json 互相拽回旧值。
# 更新时把端口搬进配置再从服务文件里摘掉，之后只认一处。
migrate_port_to_settings() {
  local unit_port
  unit_port=$(grep -oE '\-web [0-9]+' "$UNIT" 2>/dev/null | grep -oE '[0-9]+' | head -1)
  [[ -z $unit_port ]] && return

  if [[ ! -f "$WORK_DIR/settings.json" ]]; then
    printf '{\n  "port": %s,\n  "listen_addr": ""\n}\n' "$unit_port" > "$WORK_DIR/settings.json"
    chmod 600 "$WORK_DIR/settings.json"
  fi
  sed -i "s/-web ${unit_port} //" "$UNIT"
  svc_reload
  echo "  已把端口 ${unit_port} 迁移到 settings.json"
}

do_update() {
  local arch goarch tmp
  arch=$(uname -m)
  case "$arch" in
    x86_64) goarch=amd64 ;;
    aarch64|arm64) goarch=arm64 ;;
    *) echo -e "  ${R}不支持的架构 ${arch}${N}"; return ;;
  esac

  echo -e "\n  当前 $("$BIN" -version 2>/dev/null || echo '-')"
  tmp=$(mktemp -d)
  echo "  正在下载最新版..."
  if ! curl -fsSL "https://github.com/${REPO}/releases/latest/download/fanout-linux-${goarch}.tar.gz" \
       -o "$tmp/f.tar.gz"; then
    echo -e "  ${R}下载失败${N}"; rm -rf "$tmp"; return
  fi
  tar xzf "$tmp/f.tar.gz" -C "$tmp"
  svc_stop
  install -m 755 "$tmp/fanout" "$BIN"
  migrate_port_to_settings
  svc_start
  rm -rf "$tmp"
  echo -e "  ${G}已更新到 $("$BIN" -version 2>/dev/null)${N}"
}

do_uninstall() {
  local yes
  echo
  read -rp "  确认卸载？隧道和配置都会删除 [y/N]: " yes
  [[ ${yes,,} == y ]] || { echo "  已取消"; return; }

  svc_stop >/dev/null 2>&1
  svc_disable
  # 清掉残留的 netns 与 veth
  for ns in $(ip netns list 2>/dev/null | awk '{print $1}' | grep '^fo[0-9]'); do
    ip netns del "$ns" 2>/dev/null
  done
  for l in $(ip -o link show 2>/dev/null | awk -F': ' '{print $2}' | grep '^fov[0-9]'); do
    ip link del "$l" 2>/dev/null
  done
  rm -f "$UNIT" "$BIN" /usr/local/bin/f
  rm -rf "$WORK_DIR"
  svc_reload
  echo -e "  ${G}已卸载${N}"
  exit 0
}

update_json_val() {
  local key="$1" val="${2:-}"
  local f="$WORK_DIR/settings.json"
  [[ -f "$f" ]] || printf '{\n  "port": 8899,\n  "listen_addr": ""\n}\n' > "$f"
  local esc_val
  esc_val=$(echo "$val" | sed -e 's/[\\#&]/\\&/g')
  if grep -q "\"${key}\"[[:space:]]*:" "$f" 2>/dev/null; then
    if [[ -n "$val" && "$val" =~ ^[0-9]+$ ]]; then
      sed -i "s/\"${key}\"[[:space:]]*:[[:space:]]*[0-9]*/\"${key}\": ${val}/" "$f"
    else
      sed -i "s#\"${key}\"[[:space:]]*:[[:space:]]*\"[^\"]*\"#\"${key}\": \"${esc_val}\"#" "$f"
    fi
  else
    if [[ -n "$val" && "$val" =~ ^[0-9]+$ ]]; then
      sed -i "s#\"port\"#\"${key}\": ${val},\n  \"port\"#" "$f"
    else
      sed -i "s#\"port\"#\"${key}\": \"${esc_val}\",\n  \"port\"#" "$f"
    fi
  fi
  chmod 600 "$f"
}

set_domain() {
  local cur dom
  cur=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  if [[ -n "${1:-}" ]]; then
    if [[ "$1" == "clear" || "$1" == "--clear" || "$1" == "none" ]]; then
      dom=""
    else
      dom="$1"
    fi
  else
    echo
    echo -e "  当前绑定域名: ${B}${cur:-未绑定}${N}"
    read -rp "  输入新域名 (例如 panel.example.com，留空清除绑定): " dom
  fi
  dom=$(echo "${dom:-}" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
  update_json_val "domain" "${dom:-}"
  svc_restart
  if [[ -n "$dom" ]]; then
    echo -e "  ${G}域名已设置为: ${dom} 并重启服务${N}"
  else
    echo -e "  ${Y}已清除域名绑定并重启服务${N}"
  fi
}

bind_custom_ssl() {
  local c k dom cur_dom
  cur_dom=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  echo
  read -rp "  证书文件绝对路径 (.crt / .pem): " c
  c=$(echo "${c:-}" | xargs)
  if [[ ! -f "$c" ]]; then
    echo -e "  ${R}证书文件不存在: ${c}${N}"; return
  fi

  read -rp "  私钥文件绝对路径 (.key): " k
  k=$(echo "${k:-}" | xargs)
  if [[ ! -f "$k" ]]; then
    echo -e "  ${R}私钥文件不存在: ${k}${N}"; return
  fi

  if [[ -z "$cur_dom" ]]; then
    read -rp "  绑定域名 (例如 panel.example.com): " dom
    dom=$(echo "${dom:-}" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
    [[ -n "$dom" ]] && update_json_val "domain" "$dom"
  fi

  update_json_val "cert_file" "$c"
  update_json_val "key_file" "$k"
  update_json_val "ssl_mode" "custom"
  svc_restart
  echo -e "  ${G}自定义 SSL 证书已绑定并重启生效！${N}"
}

acme_standalone() {
  local dom email try_it
  dom=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  echo
  echo -e "${B}  一键申领 Let's Encrypt 证书 (Standalone 80 端口)${N}"
  if [[ -z "$dom" ]]; then
    read -rp "  输入解析到本机的域名: " dom
    dom=$(echo "${dom:-}" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
    [[ -z "$dom" ]] && { echo -e "  ${R}域名不能为空${N}"; return; }
    update_json_val "domain" "$dom"
  else
    read -rp "  域名 [回车沿用: ${dom}]: " new_dom
    if [[ -n "${new_dom:-}" ]]; then
      dom=$(echo "$new_dom" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
      update_json_val "domain" "$dom"
    fi
  fi

  if ss -tln 2>/dev/null | grep -q ":80 "; then
    echo -e "  ${Y}警告: 检测到 80 端口已被占用。Standalone 模式需要临时监听 80 端口。${N}"
    echo -e "  ${D}请先暂停占用 80 端口的服务，或改用 Cloudflare DNS 零端口模式。${N}"
    read -rp "  是否尝试继续？[y/N]: " try_it
    [[ ${try_it,,} == y ]] || { echo "  已取消"; return; }
  fi

  read -rp "  注册邮箱 (留空自动生成): " email
  email="${email:-admin@${dom}}"

  if ! command -v ~/.acme.sh/acme.sh >/dev/null 2>&1; then
    echo "  正在安装 acme.sh..."
    curl -fsSL https://get.acme.sh | sh -s email="$email" || {
      echo -e "  ${R}acme.sh 安装失败${N}"; return
    }
  fi

  mkdir -p "$WORK_DIR/ssl"
  chmod 700 "$WORK_DIR/ssl" 2>/dev/null || true
  ~/.acme.sh/acme.sh --set-default-ca --server letsencrypt >/dev/null 2>&1 || true
  if ! ~/.acme.sh/acme.sh --issue -d "$dom" --standalone --httpport 80; then
    echo -e "  ${R}证书签发失败，请确认域名已正确解析到本机 IP 且 80 端口未被阻断${N}"
    return
  fi

  ~/.acme.sh/acme.sh --install-cert -d "$dom" \
    --key-file "$WORK_DIR/ssl/privkey.pem" \
    --fullchain-file "$WORK_DIR/ssl/fullchain.pem" \
    --reloadcmd "systemctl restart fanout 2>/dev/null || rc-service fanout restart 2>/dev/null || true"

  update_json_val "cert_file" "$WORK_DIR/ssl/fullchain.pem"
  update_json_val "key_file" "$WORK_DIR/ssl/privkey.pem"
  update_json_val "ssl_mode" "acme_standalone"
  svc_restart
  echo -e "  ${G}ACME 证书申请并安装成功！已配置自动续期与热重载。${N}"
}

acme_cf_dns() {
  local dom token
  dom=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  echo
  echo -e "${B}  一键申领 Let's Encrypt 证书 (Cloudflare DNS API 零端口模式)${N}"
  echo -e "${D}  无需开放 80 端口，适合 80 端口被封禁或被其它服务占用的环境。${N}"
  if [[ -z "$dom" ]]; then
    read -rp "  输入托管在 Cloudflare 上的域名: " dom
    dom=$(echo "${dom:-}" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
    [[ -z "$dom" ]] && { echo -e "  ${R}域名不能为空${N}"; return; }
    update_json_val "domain" "$dom"
  else
    read -rp "  域名 [回车沿用: ${dom}]: " new_dom
    if [[ -n "${new_dom:-}" ]]; then
      dom=$(echo "$new_dom" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
      update_json_val "domain" "$dom"
    fi
  fi

  read -rp "  输入 Cloudflare API Token (需具备 DNS:Edit 权限): " token
  token=$(echo "${token:-}" | xargs)
  [[ -z "$token" ]] && { echo -e "  ${R}Token 不能为空${N}"; return; }

  if ! command -v ~/.acme.sh/acme.sh >/dev/null 2>&1; then
    echo "  正在安装 acme.sh..."
    curl -fsSL https://get.acme.sh | sh -s email="admin@${dom}" || {
      echo -e "  ${R}acme.sh 安装失败${N}"; return
    }
  fi

  mkdir -p "$WORK_DIR/ssl"
  chmod 700 "$WORK_DIR/ssl" 2>/dev/null || true
  export CF_Token="$token"
  ~/.acme.sh/acme.sh --set-default-ca --server letsencrypt >/dev/null 2>&1 || true
  if ! ~/.acme.sh/acme.sh --issue --dns dns_cf -d "$dom"; then
    echo -e "  ${R}证书签发失败，请检查 CF Token 权限与域名归属${N}"
    return
  fi

  ~/.acme.sh/acme.sh --install-cert -d "$dom" \
    --key-file "$WORK_DIR/ssl/privkey.pem" \
    --fullchain-file "$WORK_DIR/ssl/fullchain.pem" \
    --reloadcmd "systemctl restart fanout 2>/dev/null || rc-service fanout restart 2>/dev/null || true"

  update_json_val "cert_file" "$WORK_DIR/ssl/fullchain.pem"
  update_json_val "key_file" "$WORK_DIR/ssl/privkey.pem"
  update_json_val "ssl_mode" "acme_cf_dns"
  svc_restart
  echo -e "  ${G}Cloudflare DNS ACME 证书申请并安装成功！${N}"
}

inspect_ssl() {
  local cf kf
  cf=$(sed -n 's/.*"cert_file"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  kf=$(sed -n 's/.*"key_file"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
  echo
  if [[ -z "$cf" || ! -f "$cf" ]]; then
    echo -e "  ${Y}尚未配置或未找到证书文件${N}"
    return
  fi
  echo -e "  证书路径: ${cf}"
  echo -e "  私钥路径: ${kf:-未配置}"
  if command -v openssl >/dev/null 2>&1; then
    echo
    echo -e "  ${B}证书详情：${N}"
    openssl x509 -in "$cf" -noout -subject -issuer -dates 2>/dev/null || echo "  证书解析失败"
  fi
}

clear_ssl() {
  local yes
  echo
  read -rp "  确认清除 SSL 配置并恢复 HTTP 明文模式？[y/N]: " yes
  [[ ${yes,,} == y ]] || { echo "  已取消"; return; }
  update_json_val "cert_file" ""
  update_json_val "key_file" ""
  update_json_val "ssl_mode" "none"
  svc_restart
  echo -e "  ${G}SSL 配置已清除，面板恢复 HTTP 明文模式${N}"
}

setup_caddy() {
  local dom port force new_dom
  port=$(web_port)
  dom=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)

  echo
  echo -e "${B}  一键配置 Caddy 443 自动化反向代理${N}"
  echo -e "${D}  Caddy 自动申请与管理 Let's Encrypt 证书，实现免端口 (443) 纯净访问。${N}"
  echo

  # 443 端口与 Xray Reality 节点占用前置检测
  local p443
  p443=$(ss -tlnp 2>/dev/null | grep -E ':(443|https)\b' || true)
  if [[ -n "$p443" ]]; then
    echo -e "  ${R}[警告] 检测到 443 端口已被系统程序占用！${N}"
    echo -e "  ${D}${p443}${N}"
    if echo "$p443" | grep -qiE 'xray|3x-ui|xcl'; then
      echo -e "  ${Y}检测到 Xray / 3x-ui 节点正在使用 443 端口 (如 Reality 偷跑 443)。${N}"
      echo -e "  ${Y}Caddy 反代需要独占母机 443 端口，若继续部署将导致 Reality 节点冲突失效！${N}"
      echo -e "  ${G}建议：保留 443 给节点，使用「f ssl」配置面板原生 HTTPS (如 https://域名:${port}/)${N}"
    fi
    read -rp "  是否仍要强制继续配置 Caddy？[y/N]: " force
    [[ ${force,,} == y ]] || { echo "  已取消"; return; }
  fi

  if [[ -z "$dom" ]]; then
    read -rp "  输入解析到本 VPS 的域名 (例如 panel.example.com): " dom
    dom=$(echo "${dom:-}" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
    [[ -z "$dom" ]] && { echo -e "  ${R}域名不能为空${N}"; return; }
    update_json_val "domain" "$dom"
  else
    read -rp "  使用域名 [当前: ${dom}] (直接回车保持，或输入新域名): " new_dom
    if [[ -n "${new_dom:-}" ]]; then
      dom=$(echo "$new_dom" | sed -e 's|^https*://||' -e 's|/.*||' -e 's|:.*||' | tr '[:upper:]' '[:lower:]' | xargs)
      update_json_val "domain" "$dom"
    fi
  fi

  # 安装 Caddy
  if ! command -v caddy >/dev/null 2>&1; then
    echo "  正在安装 Caddy..."
    local mgr
    mgr=$(for m in apt-get dnf yum pacman apk zypper; do command -v "$m" >/dev/null && echo "$m" && break; done || true)
    case "${mgr:-}" in
      apt-get)
        apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl >/dev/null 2>&1 || true
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null || true
        curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null 2>&1 || true
        apt-get update -qq && apt-get install -y -qq caddy >/dev/null 2>&1
        ;;
      dnf|yum)
        $mgr install -y -q 'dnf-command(copr)' >/dev/null 2>&1 || true
        $mgr copr enable -y @caddy/caddy >/dev/null 2>&1 || true
        $mgr install -y -q caddy >/dev/null 2>&1
        ;;
      apk)
        apk add --no-cache caddy >/dev/null 2>&1
        ;;
      *)
        echo -e "  ${R}未识别的包管理器，请先手动安装 Caddy${N}"; return
        ;;
    esac
  fi

  if ! command -v caddy >/dev/null 2>&1; then
    echo -e "  ${R}Caddy 安装失败，请手动安装后重试${N}"; return
  fi

  # 写入 Caddyfile
  mkdir -p /etc/caddy
  cat > /etc/caddy/Caddyfile <<EOF
${dom} {
    encode gzip
    reverse_proxy 127.0.0.1:${port}
}
EOF

  # 将面板收敛至 127.0.0.1 本地监听，并设置 ssl_mode = caddy
  update_json_val "listen_addr" "127.0.0.1"
  update_json_val "ssl_mode" "caddy"
  update_json_val "cert_file" ""
  update_json_val "key_file" ""

  # 启动/重启 Caddy
  if command -v systemctl >/dev/null 2>&1; then
    systemctl enable caddy >/dev/null 2>&1 || true
    systemctl restart caddy
  elif command -v rc-service >/dev/null 2>&1; then
    rc-update add caddy default >/dev/null 2>&1 || true
    rc-service caddy restart
  fi

  svc_restart
  local bp
  bp=$(cat "$WORK_DIR/basepath" 2>/dev/null || echo "")
  echo
  echo -e "  ${G}[成功] Caddy 443 自动化反代已就绪！${N}"
  echo -e "  管理面板地址: ${B}https://${dom}/${bp}/${N} (免端口)"
  echo -e "  面板本地监听: ${D}http://127.0.0.1:${port}/${bp}/${N}"
}

ssl_menu() {
  while true; do
    clear
    local dom cf kf sm sub_ch
    dom=$(sed -n 's/.*"domain"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
    cf=$(sed -n 's/.*"cert_file"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
    kf=$(sed -n 's/.*"key_file"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)
    sm=$(sed -n 's/.*"ssl_mode"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$WORK_DIR/settings.json" 2>/dev/null || true)

    local sm_desc="${D}未配置 (HTTP 直连)${N}"
    if [[ -n "$cf" && "$sm" != "none" && "$sm" != "caddy" ]]; then
      if [[ "$sm" == "custom" ]]; then
        sm_desc="${G}[原生自定义 SSL 证书 (HTTPS)]${N}"
      elif [[ "$sm" == "acme_standalone" ]]; then
        sm_desc="${G}[ACME Standalone 自动证书 (HTTPS)]${N}"
      elif [[ "$sm" == "acme_cf_dns" ]]; then
        sm_desc="${G}[ACME Cloudflare DNS 证书 (HTTPS)]${N}"
      elif [[ "$sm" == "acme"* ]]; then
        sm_desc="${G}[ACME 自动证书 (HTTPS)]${N}"
      else
        sm_desc="${G}[原生 HTTPS 证书已生效]${N}"
      fi
    elif [[ "$sm" == "caddy" ]]; then
      sm_desc="${G}[Caddy 自动化反代 (443免端口)]${N}"
    elif [[ "$sm" == "none" ]] || [[ -n "$dom" && -z "$cf" ]]; then
      sm_desc="${Y}外部反代 / 纯域名 HTTP${N}"
    fi

    echo
    echo -e "  ${B}╔═══════════════════════════════════════════════════════════════╗${N}"
    echo -e "  ${B}║${N}       ${W}FANOUT PRO${N}  ${D}•  域名与 SSL (HTTPS) 安全管理${N}              ${B}║${N}"
    echo -e "  ${B}╚═══════════════════════════════════════════════════════════════╝${N}"
    echo
    echo -e "  ${B}┌─ [ 当前域名与证书状态 ] ──────────────────────────────────────┐${N}"
    echo -e "  ${B}│${N}  绑定域名: ${W}${dom:-未绑定}${N}"
    echo -e "  ${B}│${N}  当前模式: ${sm_desc}"
    if [[ -n "$cf" ]]; then
      echo -e "  ${B}│${N}  证书文件: ${D}${cf}${N}"
    fi
    if [[ -n "$kf" ]]; then
      echo -e "  ${B}│${N}  私钥文件: ${D}${kf}${N}"
    fi
    echo -e "  ${B}└───────────────────────────────────────────────────────────────┘${N}"
    echo
    echo -e "  ${B}── [ 域名与反向代理 ] ──────────────────────────────────────────${N}"
    echo -e "   ${Y} 1.${N} 设置 / 修改面板绑定域名 (用于反代或直接访问)"
    echo -e "   ${Y} 2.${N} 一键配置 Caddy 自动化反代 (独占443免端口，若已装3x-ui请勿用)"
    echo
    echo -e "  ${B}── [ SSL / HTTPS 证书申请与管理 ] ──────────────────────────────${N}"
    echo -e "   ${Y} 3.${N} ${G}绑定已有自定义证书路径 (.crt / .key，推荐复用 3x-ui 证书)${N}"
    echo -e "   ${Y} 4.${N} 一键申请 ACME 免费证书 (HTTP-80 Standalone 模式)"
    echo -e "   ${Y} 5.${N} ${G}一键申请 ACME 免费证书 (Cloudflare DNS 零端口，3x-ui 兼容)${N}"
    echo -e "   ${Y} 6.${N} 深度检测当前证书有效性与到期天数"
    echo -e "   ${Y} 7.${N} ${R}清除 SSL 配置 (安全降级为 HTTP 明文直连)${N}"
    echo
    echo -e "   ${R} 0.${N} 返回主菜单"
    echo -e "  ${B}────────────────────────────────────────────────────────────────${N}"
    echo -ne "  ${W}请输入选项 [0-7]:${N} "
    read -r sub_ch

    case "${sub_ch:-}" in
      1) set_domain; pause ;;
      2) setup_caddy; pause ;;
      3) bind_custom_ssl; pause ;;
      4) acme_standalone; pause ;;
      5) acme_cf_dns; pause ;;
      6) inspect_ssl; pause ;;
      7) clear_ssl; pause ;;
      0) return ;;
      *) ;;
    esac
  done
}

menu() {
  while true; do
    clear
    echo
    echo -e "  ${B}╔═══════════════════════════════════════════════════════════════╗${N}"
    echo -e "  ${B}║${N}       ${W}FANOUT PRO${N}  ${D}•  VPN Gate 住宅多出口扇出网关管理${N}         ${B}║${N}"
    echo -e "  ${B}╚═══════════════════════════════════════════════════════════════╝${N}"
    show_info
    echo
    echo -e "  ${B}── [ 服务控制 ] ────────────────────────────────────────────────${N}"
    echo -e "   ${Y} 1.${N} 启动服务                    ${Y} 2.${N} 停止服务"
    echo -e "   ${Y} 3.${N} 重启服务                    ${Y} 4.${N} 查看实时运行日志"
    echo
    echo -e "  ${B}── [ 节点与隧道 ] ──────────────────────────────────────────────${N}"
    echo -e "   ${Y} 5.${N} 活跃隧道列表                ${Y} 6.${N} 详细连接与质量信息"
    echo
    echo -e "  ${B}── [ 面板与安全配置 ] ──────────────────────────────────────────${N}"
    echo -e "   ${Y} 7.${N} 修改面板监听端口            ${Y} 8.${N} 修改管理访问口令"
    echo -e "   ${Y} 9.${N} 修改路径前缀 (防扫描探测)   ${Y}10.${N} ${G}域名与 SSL (HTTPS) 设置${N}"
    echo
    echo -e "  ${B}── [ 系统运维 ] ────────────────────────────────────────────────${N}"
    echo -e "   ${Y}11.${N} 开机自启开关                ${Y}12.${N} 检查更新 / 升级版本"
    echo -e "   ${Y}13.${N} 卸载 Fanout                 ${Y}14.${N} 项目开源主页"
    echo
    echo -e "   ${R} 0.${N} 退出管理脚本"
    echo -e "  ${B}────────────────────────────────────────────────────────────────${N}"
    echo -ne "  ${W}请输入选项 [0-14]:${N} "
    read -r choice

    case "$choice" in
      1) svc_start   && echo -e "\n  ${G}[成功] 服务已启动${N}"; pause ;;
      2) svc_stop    && echo -e "\n  ${Y}[提示] 服务已停止${N}"; pause ;;
      3) svc_restart && echo -e "\n  ${G}[成功] 服务已重启${N}"; pause ;;
      4) echo; svc_logs 40; pause ;;
      5) list_tunnels; pause ;;
      6) show_info; pause ;;
      7) change_port; pause ;;
      8) reset_password; pause ;;
      9) reset_basepath; pause ;;
      10) ssl_menu ;;
      11)
        if svc_is_enabled; then
          svc_disable
          echo -e "\n  ${Y}[提示] 已关闭开机自启${N}"
        else
          svc_enable
          echo -e "\n  ${G}[成功] 已开启开机自启${N}"
        fi
        pause ;;
      12) do_update; pause ;;
      13) do_uninstall; pause ;;
      14) show_links; pause ;;
      0) exit 0 ;;
      *) ;;
    esac
  done
}

need_root

# 带参数时当普通命令用，不进菜单
case "${1:-}" in
  start)     svc_start ;;
  stop)      svc_stop ;;
  restart)   svc_restart ;;
  status)    svc_status_page ;;
  log)       svc_logs_follow ;;
  info)      show_info ;;
  list)      list_tunnels ;;
  domain)    shift; set_domain "$@" ;;
  ssl)       shift; ssl_menu "$@" ;;
  caddy)     shift; setup_caddy "$@" ;;
  update)    do_update ;;
  uninstall) do_uninstall ;;
  "")        menu ;;
  *)
    echo "用法: f [start|stop|restart|status|log|info|list|domain|ssl|caddy|update|uninstall]"
    echo "不带参数进入交互菜单"
    ;;
esac
