package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window

	cfg   *Config
	stats *Stats

	statusText binding.String
	statusIcon binding.String

	bytesIn    binding.String
	bytesOut   binding.String
	speedIn    binding.String
	speedOut   binding.String
	uptimeText binding.String
	assignedIP binding.String

	serverHost  *widget.Entry
	serverPort  *widget.Entry
	serverKey   *widget.Entry
	username    *widget.Entry
	password    *widget.Entry
	mfaCode     *widget.Entry
	saveCreds   *widget.Check
	autoConnect *widget.Check

	connectBtn    *widget.Button
	disconnectBtn *widget.Button

	tunnel *Tunnel

// SaaS integration
apiClient       *APIClient
serviceSelector *ServiceSelector
orgBadge        *widget.Label
services        []MyService
currentOrg      *MyOrganization
selectedService *MyService

	logs binding.StringList
}

func NewApp() *App {
	a := &App{
		fyneApp: app.NewWithID("com.defendcore.vpn"),
		stats:   NewStats(),

		statusText: binding.NewString(),
		statusIcon: binding.NewString(),

		bytesIn:    binding.NewString(),
		bytesOut:   binding.NewString(),
		speedIn:    binding.NewString(),
		speedOut:   binding.NewString(),
		uptimeText: binding.NewString(),
		assignedIP: binding.NewString(),

		logs: binding.NewStringList(),
	}
	fyne.Do(func() { a.statusText.Set("Disconnected") })
	fyne.Do(func() { a.statusIcon.Set("⚪") })
	a.bytesIn.Set("0 B")
	a.bytesOut.Set("0 B")
	a.speedIn.Set("0 B/s")
	a.speedOut.Set("0 B/s")
	a.uptimeText.Set("00:00:00")
	fyne.Do(func() { a.assignedIP.Set("—") })

a.tunnel = NewTunnel(a.stats)
a.serviceSelector = NewServiceSelector(a.onServiceSelected)
a.orgBadge = widget.NewLabel("")

return a
}

func (a *App) Run() {
	a.window = a.fyneApp.NewWindow("DefendCore VPN Client")
	a.window.Resize(fyne.NewSize(500, 700))

	cfg, err := LoadConfig()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.cfg = cfg

	a.buildUI()
	a.window.ShowAndRun()
}

func (a *App) buildUI() {
	statusLabel := widget.NewLabelWithData(a.statusText)
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	statusLabel.Alignment = fyne.TextAlignCenter

	statusIconLabel := widget.NewLabelWithData(a.statusIcon)
	statusIconLabel.Alignment = fyne.TextAlignCenter

	statusBox := container.NewVBox(
		statusIconLabel,
		statusLabel,
		widget.NewSeparator(),
	)

	a.connectBtn = widget.NewButtonWithIcon("Connect", theme.MediaPlayIcon(), a.onConnect)
	a.connectBtn.Importance = widget.HighImportance

	a.disconnectBtn = widget.NewButtonWithIcon("Disconnect", theme.MediaStopIcon(), a.onDisconnect)
	fyne.Do(func() { a.disconnectBtn.Disable() })

	btnBox := container.NewGridWithColumns(2, a.connectBtn, a.disconnectBtn)

	a.serverHost = widget.NewEntry()
	a.serverHost.SetText(a.cfg.ServerHost)
	a.serverHost.SetPlaceHolder("192.168.174.132")

	a.serverPort = widget.NewEntry()
	a.serverPort.SetText(fmt.Sprintf("%d", a.cfg.ServerPort))
	a.serverPort.SetPlaceHolder("51820")

	a.serverKey = widget.NewEntry()
	a.serverKey.SetText(a.cfg.ServerPubKey)
	a.serverKey.SetPlaceHolder("Server public key (base64)")

	a.username = widget.NewEntry()
	a.username.SetText(a.cfg.Username)
	a.username.SetPlaceHolder("admin@defendcore.local")

	a.password = widget.NewPasswordEntry()
	a.password.SetPlaceHolder("Password")

	a.mfaCode = widget.NewEntry()
	a.mfaCode.SetPlaceHolder("6-digit MFA code")
	a.mfaCode.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		if len(s) != 6 {
			return fmt.Errorf("must be 6 digits")
		}
		return nil
	}

	a.saveCreds = widget.NewCheck("Save credentials", nil)
	a.saveCreds.SetChecked(a.cfg.SavePassword)

	a.autoConnect = widget.NewCheck("Auto-connect on startup", nil)
	a.autoConnect.SetChecked(a.cfg.AutoConnect)

	configForm := widget.NewForm(
		widget.NewFormItem("Server Host", a.serverHost),
		widget.NewFormItem("Server Port", a.serverPort),
		widget.NewFormItem("Server Key", a.serverKey),
		widget.NewFormItem("Username", a.username),
		widget.NewFormItem("Password", a.password),
		widget.NewFormItem("MFA Code", a.mfaCode),
	)

	configCard := widget.NewCard("Configuration", "", container.NewVBox(
		configForm,
		a.saveCreds,
		a.autoConnect,
	))

	infoForm := widget.NewForm(
		widget.NewFormItem("Assigned IP", widget.NewLabelWithData(a.assignedIP)),
		widget.NewFormItem("Uptime", widget.NewLabelWithData(a.uptimeText)),
		widget.NewFormItem("↓ Received", widget.NewLabelWithData(a.bytesIn)),
		widget.NewFormItem("↑ Sent", widget.NewLabelWithData(a.bytesOut)),
		widget.NewFormItem("↓ Speed", widget.NewLabelWithData(a.speedIn)),
		widget.NewFormItem("↑ Speed", widget.NewLabelWithData(a.speedOut)),
	)

	infoCard := widget.NewCard("Connection Info", "", infoForm)

	logsList := widget.NewListWithData(
		a.logs,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		},
	)
	logsList.Resize(fyne.NewSize(480, 150))
	logsCard := widget.NewCard("Logs", "", logsList)

	serviceCard := widget.NewCard("Available VPN Services", "", container.NewVBox(
		a.orgBadge,
		widget.NewSeparator(),
		a.serviceSelector.Container(),
	))

	content := container.NewVBox(
		statusBox,
		btnBox,
		widget.NewSeparator(),
		configCard,
		serviceCard,
		infoCard,
		logsCard,
	)

	scroll := container.NewVScroll(content)
	a.window.SetContent(scroll)

	go a.statsTicker()
}

func (a *App) statsTicker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.stats.Tick()
		a.updateStats()
	}
}

func (a *App) updateStats() {
	in, out, spIn, spOut, uptime := a.stats.Snapshot()
	a.bytesIn.Set(formatBytes(in))
	a.bytesOut.Set(formatBytes(out))
	a.speedIn.Set(formatBytes(uint64(spIn)) + "/s")
	a.speedOut.Set(formatBytes(uint64(spOut)) + "/s")
	a.uptimeText.Set(formatDuration(uptime))
}

func (a *App) addLog(msg string) {
ts := time.Now().Format("15:04:05")
line := fmt.Sprintf("[%s] %s", ts, msg)

// Fyne v2.8 requires UI updates on the main thread.
fyne.Do(func() {
list, _ := a.logs.Get()
list = append(list, line)
if len(list) > 100 {
list = list[len(list)-100:]
}
a.logs.Set(list)
})
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (a *App) onConnect() {
	a.addLog("Connecting...")
	fyne.Do(func() { a.connectBtn.Disable() })
	fyne.Do(func() { a.statusIcon.Set("🟡") })
	fyne.Do(func() { a.statusText.Set("Connecting...") })
	fyne.Do(func() { a.connectBtn.SetText("Connecting...") })

	go a.connectAsync()
}

func (a *App) connectAsync() {
	defer func() {
		if r := recover(); r != nil {
			a.addLog(fmt.Sprintf("panic: %v", r))
			fyne.Do(func() { a.connectBtn.Enable() })
			fyne.Do(func() { a.connectBtn.SetText("Connect") })
		}
	}()

	// 1. Validate form
	host := a.serverHost.Text
	port := a.serverPort.Text
	email := a.username.Text
	password := a.password.Text
	mfaCode := a.mfaCode.Text

	if host == "" || port == "" || email == "" || password == "" {
		a.addLog("❌ Server host, port, email, password required")
		a.resetConnectButton()
		return
	}

	apiURL := fmt.Sprintf("http://%s:8080", host)

	// 2. Create API client
	api := NewAPIClient(apiURL)
	a.addLog(fmt.Sprintf("Logging in as %s...", email))

	// 3. Login
	a.addLog("[debug] calling api.Login...")
	a.addLog("[debug] calling api.Login...")
	loginResp, err := api.Login(email, password, mfaCode)
	if err != nil {
		a.addLog(fmt.Sprintf("❌ Login failed: %v", err))
		a.resetConnectButton()
		return
	}
	a.addLog(fmt.Sprintf("✅ Login successful (user: %s)", loginResp.User.Email))
a.apiClient = api

// Fetch user's organization + services
a.addLog("[debug] fetching organization...")
org, orgErr := api.GetMyOrganization()
if orgErr != nil {
a.addLog(fmt.Sprintf("⚠️  Org fetch failed: %v", orgErr))
} else if org != nil {
a.currentOrg = org
a.orgBadge.SetText(fmt.Sprintf("%s [%s]", org.Name, org.Plan))
a.addLog(fmt.Sprintf("Organization: %s (%s)", org.Name, org.Plan))
} else {
a.addLog("⚠️  No organization")
}

a.addLog("[debug] fetching services...")
services, err := api.ListMyServices()
if err != nil {
a.addLog(fmt.Sprintf("⚠️  Services fetch failed: %v", err))
} else if len(services) == 0 {
a.addLog("⚠️  No services assigned")
} else {
a.services = services
a.addLog(fmt.Sprintf("✅ Found %d VPN service(s)", len(services)))
for _, s := range services {
a.addLog(fmt.Sprintf("   %s %s", s.TypeIcon, s.Name))
}
a.serviceSelector.SetServices(services)
}
	a.addLog("[debug] calling api.ListDevices...")
	a.addLog("[debug] calling api.ListDevices...")

	// 4. List devices
	devices, err := api.ListDevices()
	if err != nil {
		a.addLog(fmt.Sprintf("❌ Failed to list devices: %v", err))
		a.resetConnectButton()
		return
	}
	if len(devices) == 0 {
		a.addLog("❌ No devices found. Create one in dashboard.")
		a.resetConnectButton()
		return
	}

	// 5. Use first active device (or specific device from config)
	var device *Device
	for i := range devices {
		if devices[i].Status == "active" {
			device = &devices[i]
			break
		}
	}
	if device == nil {
		device = &devices[0]
	}
	a.addLog(fmt.Sprintf("Using device: %s (%s)", device.Name, device.ID[:8]))

	// 6. Get device config
	deviceCfg, err := api.GetDeviceConfig(device.ID)
	if err != nil {
		a.addLog(fmt.Sprintf("❌ Failed to get config: %v", err))
		a.resetConnectButton()
		return
	}
	a.addLog(fmt.Sprintf("Got config: assigned_ip=%s", deviceCfg.AssignedIP))

	// 7. Save config
	a.cfg.ServerHost = host
	a.cfg.ServerPubKey = a.serverKey.Text
	a.cfg.Username = email
	a.cfg.UserID = loginResp.User.ID
	a.cfg.DeviceID = device.ID
	if a.saveCreds.Checked {
		a.cfg.SavePassword = true
	}
	if a.autoConnect.Checked {
		a.cfg.AutoConnect = true
	}
	_ = SaveConfig(a.cfg)

	// 8. TODO: Start VPN tunnel
	// For now, just mark as connected
	// 8. Start real VPN tunnel
	a.addLog("Starting VPN tunnel...")
	a.stats = NewStats()
	a.tunnel.stats = a.stats

	if err := a.tunnel.Start(
		host,
		deviceCfg.ServerPort,
		deviceCfg.ServerPubKey,
		loginResp.User.ID,
		device.ID,
	); err != nil {
		a.addLog(fmt.Sprintf("❌ Tunnel failed: %v", err))
		a.resetConnectButton()
		return
	}

	a.addLog("✅ VPN tunnel established")
	fyne.Do(func() { a.statusIcon.Set("🟢") })
	fyne.Do(func() { a.statusText.Set("Connected") })
	fyne.Do(func() { a.assignedIP.Set(deviceCfg.AssignedIP) })
	fyne.Do(func() { a.connectBtn.SetText("Connect") })
	fyne.Do(func() { a.disconnectBtn.Enable() })
}

func (a *App) resetConnectButton() {
	fyne.Do(func() { a.statusIcon.Set("🔴") })
	fyne.Do(func() { a.statusText.Set("Error") })
	fyne.Do(func() { a.connectBtn.Enable() })
	fyne.Do(func() { a.connectBtn.SetText("Connect") })
}

func (a *App) simulateTraffic() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			// Only simulate if connected
			in, _ := a.assignedIP.Get()
			if in == "—" {
				return
			}
			// Random-ish traffic
			a.stats.AddIn(1024 * 64)  // 64 KB/s in
			a.stats.AddOut(1024 * 16) // 16 KB/s out
		}
	}()
}

func (a *App) onDisconnect() {
	a.addLog("Disconnecting...")
	fyne.Do(func() { a.disconnectBtn.Disable() })

	// Stop tunnel
	if a.tunnel != nil {
		a.tunnel.Stop()
	}

	// Reset stats
	a.stats = NewStats()
	fyne.Do(func() { a.assignedIP.Set("—") })

	fyne.Do(func() { a.statusIcon.Set("⚪") })
	fyne.Do(func() { a.statusText.Set("Disconnected") })
	fyne.Do(func() { a.connectBtn.Enable() })
	fyne.Do(func() { a.connectBtn.SetText("Connect") })
	a.addLog("Disconnected")
}


// onServiceSelected is called when the user picks a service from the selector.
func (a *App) onServiceSelected(svc MyService) {
a.selectedService = &svc
a.addLog(fmt.Sprintf("Selected: %s %s", svc.TypeIcon, svc.Name))

// Auto-set server host from service
if svc.ServerIP != "" {
a.serverHost.SetText(svc.ServerIP)
}

// Start tunnel asynchronously
a.addLog("Starting tunnel...")
fyne.Do(func() { a.connectBtn.Disable() })
fyne.Do(func() { a.statusIcon.Set("🟡") })
fyne.Do(func() { a.statusText.Set("Connecting...") })

go a.startSelectedService()
}

// startSelectedService establishes the tunnel for the currently selected service.
func (a *App) startSelectedService() {
defer func() {
if r := recover(); r != nil {
a.addLog(fmt.Sprintf("panic: %v", r))
a.resetConnectButton()
}
}()

if a.selectedService == nil {
a.addLog("❌ No service selected")
a.resetConnectButton()
return
}
if a.apiClient == nil {
a.addLog("❌ Not logged in")
a.resetConnectButton()
return
}

svc := *a.selectedService

// Fetch devices
devices, err := a.apiClient.ListDevices()
if err != nil {
a.addLog(fmt.Sprintf("❌ Failed to list devices: %v", err))
a.resetConnectButton()
return
}
if len(devices) == 0 {
a.addLog("❌ No devices found")
a.resetConnectButton()
return
}

device := &devices[0]
for i := range devices {
if devices[i].Status == "active" {
device = &devices[i]
break
}
}
a.addLog(fmt.Sprintf("Using device: %s (%s)", device.Name, device.ID[:8]))

// Fetch service config
serviceCfg, err := a.apiClient.GetServiceConfig(svc.ID)
if err != nil {
a.addLog(fmt.Sprintf("❌ Failed to get service config: %v", err))
a.resetConnectButton()
return
}
a.addLog(fmt.Sprintf("Got config: server=%s assigned_ip=%s", serviceCfg.ServerHost, serviceCfg.AssignedIP))

// Save config
a.cfg.ServerHost = serviceCfg.ServerHost
a.cfg.ServerPubKey = serviceCfg.ServerPublicKey
_ = SaveConfig(a.cfg)

// Resolve server public key (fallback to form if config empty)
serverKey := serviceCfg.ServerPublicKey
if serverKey == "" {
serverKey = a.serverKey.Text
a.addLog("⚠️  Using server key from form")
}
if serverKey == "" {
a.addLog("❌ Server public key is empty")
a.resetConnectButton()
return
}
a.addLog(fmt.Sprintf("Server key: %s...", serverKey[:16]))

// Start tunnel
a.addLog("Starting VPN tunnel...")
a.stats = NewStats()
a.tunnel.stats = a.stats

if err := a.tunnel.Start(
serviceCfg.ServerHost,
serviceCfg.ServerPort,
serverKey,
a.cfg.UserID,
device.ID,
); err != nil {
a.addLog(fmt.Sprintf("❌ Tunnel failed: %v", err))
a.resetConnectButton()
return
}

a.addLog("✅ VPN tunnel established")
fyne.Do(func() { a.statusIcon.Set("🟢") })
fyne.Do(func() { a.statusText.Set("Connected") })
fyne.Do(func() { a.assignedIP.Set(serviceCfg.AssignedIP) })
fyne.Do(func() { a.connectBtn.SetText("Connect") })
fyne.Do(func() { a.disconnectBtn.Enable() })
}
