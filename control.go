package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"math"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

const StatusProcess = 1
const StatusReady = 0
const Unselected = -1

var SelectedProfile = Unselected
var SelectedNotification = Unselected

var RefreshNeeded = true
var ProfileMaskNeeded bool
var NotificationMaskNeeded bool
var ProfileStateAllowDisable bool

var StatusChan = make(chan int)
var LockButtonChan = make(chan bool)

func RefreshProfile() error {
	var err error
	Profiles, err = LpacProfileList()
	if err != nil {
		return err
	}
	// 刷新 List
	ProfileList.Refresh()
	ProfileList.UnselectAll()
	SwitchStateButton.SetText("启用")
	SwitchStateButton.SetIcon(theme.ConfirmIcon())
	return nil
}

func RefreshNotification() error {
	var err error
	Notifications, err = LpacNotificationList()
	if err != nil {
		return err
	}
	sort.Slice(Notifications, func(i, j int) bool {
		return Notifications[i].SeqNumber < Notifications[j].SeqNumber
	})
	// 刷新 List
	NotificationList.Refresh()
	NotificationList.UnselectAll()
	return nil
}

func RefreshChipInfo() error {
	var err error
	ChipInfo, err = LpacChipInfo()
	if err != nil {
		return err
	}
	if ChipInfo == nil {
		return nil
	}

	convertToString := func(value interface{}) string {
		if value == nil {
			return "<未设置>"
		}
		if str, ok := value.(string); ok {
			return str
		}
		return "<未设置>"
	}

	EidLabel.SetText(fmt.Sprintf("EID: %s", ChipInfo.EidValue))
	DefaultDpAddressLabel.SetText(fmt.Sprintf("默认SM-DP+地址:  %s", convertToString(ChipInfo.EuiccConfiguredAddresses.DefaultDpAddress)))
	RootDsAddressLabel.SetText(fmt.Sprintf("根SM-DS地址:  %s", convertToString(ChipInfo.EuiccConfiguredAddresses.RootDsAddress)))
	// eUICC Manufacturer Label
	if eum := GetEUM(ChipInfo.EidValue); eum != nil {
		manufacturer := fmt.Sprint(eum.Manufacturer, " ", CountryCodeToEmoji(eum.Country))
		if productName := eum.ProductName(ChipInfo.EidValue); productName != "" {
			manufacturer = fmt.Sprint(productName, " (", manufacturer, ")")
		}
		EUICCManufacturerLabel.SetText("制造商: " + manufacturer)
	} else {
		EUICCManufacturerLabel.SetText("制造商: Unknown")
	}
	// EUICCInfo2 entry
	bytes, err := json.MarshalIndent(ChipInfo.EUICCInfo2, "", "  ")
	if err != nil {
		ShowLpacErrDialog(fmt.Errorf("芯片信息：解码EUICCInfo2失败\n%s", err))
	}
	EuiccInfo2Entry.SetText(string(bytes))
	// 计算剩余空间
	freeSpace := float64(ChipInfo.EUICCInfo2.ExtCardResource.FreeNonVolatileMemory) / 1024
	FreeSpaceLabel.SetText(fmt.Sprintf("剩余空间: %.2f KB", math.Round(freeSpace*100)/100))

	CopyEidButton.Show()
	SetDefaultSmdpButton.Show()
	EuiccInfo2Entry.Show()
	ViewCertInfoButton.Show()
	EUICCManufacturerLabel.Show()
	CopyEuiccInfo2Button.Show()
	return nil
}

func RefreshApduDriver() {
	var err error
	ApduDrivers, err = LpacDriverApduList()
	if err != nil {
		ShowLpacErrDialog(err)
	}
	var options []string
	for _, d := range ApduDrivers {
		// exclude YubiKey and CanoKey
		if strings.Contains(d.Name, "canokeys.org") || strings.Contains(d.Name, "YubiKey") {
			continue
		}
		options = append(options, d.Name)
	}
	ApduDriverSelect.SetOptions(options)
	ApduDriverSelect.ClearSelected()
	ConfigInstance.DriverIFID = ""
	ApduDriverSelect.Refresh()
}

func OpenLog() {
	if err := OpenProgram(ConfigInstance.LogDir); err != nil {
		d := dialog.NewError(err, WMain)
		d.Show()
	}
}

func OpenProgram(name string) error {
	var launcher string
	switch runtime.GOOS {
	case "windows":
		launcher = "explorer"
	case "darwin":
		launcher = "open"
	case "linux":
		launcher = "xdg-open"
	}
	if launcher == "" {
		return fmt.Errorf("不支持的平台，无法打开")
	}
	return exec.Command(launcher, name).Start()
}

func Refresh() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	err := RefreshProfile()
	if err != nil {
		ShowLpacErrDialog(err)
		return
	}
	err = RefreshNotification()
	if err != nil {
		ShowLpacErrDialog(err)
		return
	}
	err = RefreshChipInfo()
	if err != nil {
		ShowLpacErrDialog(err)
		return
	}
	RefreshNeeded = false
}

func UpdateStatusBarListener() {
	for {
		status := <-StatusChan
		switch status {
		case StatusProcess:
			StatusLabel.SetText("处理中...")
			StatusProcessBar.Start()
			StatusProcessBar.Show()
			continue
		case StatusReady:
			StatusLabel.SetText("就绪.")
			StatusProcessBar.Stop()
			StatusProcessBar.Hide()
			continue
		}
	}
}

func LockButtonListener() {
	buttons := []*widget.Button{
		RefreshButton, DownloadButton, SetNicknameButton, SwitchStateButton, DeleteProfileButton,
		ProcessNotificationButton, ProcessAllNotificationButton, RemoveNotificationButton, BatchRemoveNotificationButton,
		SetDefaultSmdpButton, ApduDriverRefreshButton,
	}
	checks := []*widget.Check{
		ProfileMaskCheck, NotificationMaskCheck,
	}
	for {
		lock := <-LockButtonChan
		if lock {
			for _, button := range buttons {
				button.Disable()
			}
			for _, check := range checks {
				check.Disable()
			}
			ApduDriverSelect.Disable()
		} else {
			for _, button := range buttons {
				button.Enable()
			}
			for _, check := range checks {
				check.Enable()
			}
			ApduDriverSelect.Enable()
		}
	}
}

func SetDriverIFID(name string) {
	for _, d := range ApduDrivers {
		if name == d.Name {
			// 未选择过读卡器
			if ConfigInstance.DriverIFID == "" {
				ConfigInstance.DriverIFID = d.Env
			} else {
				// 选择过读卡器，要求刷新
				ConfigInstance.DriverIFID = d.Env
				RefreshNeeded = true
			}
		}
	}
}
