package main

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	healthInterval = 15 * time.Second
	healthFailures = 4 // 连续失败 4 次 (60 秒) 才判定彻底掉线，杜绝偶发网络抖动与第三方接口卡顿误杀
	healthTimeout  = 5 * time.Second
)

// WatchHealth 周期检查每条隧道的健康状况，确保节点离线时显示连接失败，并在节点重新上线后自动恢复连接。
// 绝不自动换节点，忠实保留用户选定的目标节点。
func (m *Manager) WatchHealth() {
	fails := map[int]int{}
	lastRetry := map[int]time.Time{}

	for range time.Tick(healthInterval) {
		for _, t := range m.Tunnels() {
			if t.Status == "failed" {
				// 处于失败状态的节点：每隔 30 秒自动探测是否已重新上线
				if time.Since(lastRetry[t.Slot]) >= 30*time.Second {
					lastRetry[t.Slot] = time.Now()
					go m.tryReconnectSameNode(t)
				}
				continue
			}

			if t.Status != "up" {
				continue
			}

			if m.tunnelHealthy(t) {
				fails[t.Slot] = 0
				continue
			}

			fails[t.Slot]++
			if fails[t.Slot] < healthFailures {
				log.Printf("隧道 %d (%s) 探测失败 %d 次 (连续失败 %d 次判定掉线)", t.Slot, t.Node.HostName, fails[t.Slot], healthFailures)
				continue
			}

			// 确认已掉线：直接标记连接失败，绝不自动换成其他节点！
			log.Printf("隧道 %d (%s) 已确认离线，标记连接失败，保留该节点等待恢复上线", t.Slot, t.Node.HostName)
			fails[t.Slot] = 0
			lastRetry[t.Slot] = time.Now()
			t.Status = "failed"
			t.Err = "节点已离线，等待恢复上线"
			t.ExitIP = ""
			t.teardownNetns()
			_ = m.saveState()
		}
	}
}

// tunnelHealthy 判断隧道是否还真的走在 VPN 上。
//
// 结合 Linux 内核路由表、真实物理 ICMP Ping 与多源 HTTP IP 对账，零误判。
func (m *Manager) tunnelHealthy(t *Tunnel) bool {
	// 1. Linux 内核路由级快速检查：确认默认路由是否指向 tun0 虚拟网卡
	outRoute, errRoute := exec.Command("ip", "netns", "exec", t.nsName(),
		"ip", "route", "get", "1.1.1.1").Output()
	if errRoute != nil || !strings.Contains(string(outRoute), "dev tun0") {
		return false
	}

	// 2. 真实 ICMP 往返 RTT 测试：极快 (0.02s)，通过隧道真实打通公网，且不依赖第三方外部 Web 服务
	if livePing := t.probeLiveLatency(); livePing > 0 {
		t.mu.Lock()
		t.Node.Ping = livePing
		t.mu.Unlock()
		return true
	}

	// 3. 若 ICMP 偶发丢包，多源 HTTP 出口 IP 校验作为后备对账
	endpoints := []string{
		"http://api.ipify.org",
		"http://icanhazip.com",
		"http://ifconfig.me/ip",
	}
	for _, ep := range endpoints {
		out, err := exec.Command("ip", "netns", "exec", t.nsName(),
			"curl", "-s", "--max-time", strconv.Itoa(int(healthTimeout.Seconds())), ep).Output()
		if err == nil {
			got := strings.TrimSpace(string(out))
			if got == t.ExitIP {
				return true
			}
		}
	}
	return false
}

// reconnect 就地把一条隧道换到别的节点上，保持槽位与端口不变，
// 这样已经分发出去的客户端配置仍然可用。
//
// oldHost 必须是本次重连前那条隧道真正绑着的节点名。调用方若已经
// 改过 t.Node（比如手动换节点），就要把改之前的名字传进来，
// 否则 rebind 找不到旧绑定，入站会掉成孤儿。
func (m *Manager) reconnect(t *Tunnel, oldHost string) {
	t.Status = "starting"
	t.Err = "正在换节点重连..."
	t.ExitIP = ""

	if t.ovpn != nil && t.ovpn.Process != nil {
		_ = t.ovpn.Process.Kill()
		t.ovpn = nil
	}
	t.teardownNetns()

	go func() {
		// 仅连接当前节点，绝不尝试候选切换
		m.bringUp(t, false)
		if t.Status != "up" {
			return
		}
		// 出站 tag 跟着节点名走，换了节点就要把原来指向它的入站重新绑过去，
		// 否则面板里的路由会指向一个已经不存在的出站。
		if t.Node.HostName != oldHost {
			if err := m.rebind(oldHost, t); err != nil {
				log.Printf("重连后同步 3x-ui 绑定失败: %v", err)
			}
			return
		}
		// 节点名没变也要重写一次出站：出口 IP 可能变了，
		// 而且上一轮换节点时留下的绑定需要重新指回来。
		if err := m.resync(t); err != nil {
			log.Printf("重连后重写 3x-ui 出站失败: %v", err)
		}
	}()
}
