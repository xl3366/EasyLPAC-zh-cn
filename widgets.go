package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/publicsuffix"
	"slices"
	"strings"
	"time"
)

var StatusProcessBar *widget.ProgressBarInfinite
var StatusLabel *widget.Label
var SetNicknameButton *widget.Button
var DownloadButton *widget.Button
var DeleteProfileButton *widget.Button
var SwitchStateButton *widget.Button
var ProcessNotificationButton *widget.Button
var ProcessAllNotificationButton *widget.Button
var RemoveNotificationButton *widget.Button
var BatchRemoveNotificationButton *widget.Button

var ProfileList *widget.List
var NotificationList *widget.List

var FreeSpaceLabel *widget.Label
var OpenLogButton *widget.Button
var RefreshButton *widget.Button
var ProfileMaskCheck *widget.Check
var NotificationMaskCheck *widget.Check

var EidLabel *widget.Label
var DefaultDpAddressLabel *widget.Label
var RootDsAddressLabel *widget.Label
var EuiccInfo2Entry *ReadOnlyEntry
var CopyEidButton *widget.Button
var SetDefaultSmdpButton *widget.Button
var ViewCertInfoButton *widget.Button
var EUICCManufacturerLabel *widget.Label
var CopyEuiccInfo2Button *widget.Button

var ApduDriverSelect *widget.Select
var ApduDriverRefreshButton *widget.Button

var Tabs *container.AppTabs
var ProfileTab *container.TabItem
var NotificationTab *container.TabItem
var ChipInfoTab *container.TabItem
var SettingsTab *container.TabItem
var AboutTab *container.TabItem

var LpacVersionLabel *widget.Label

type ReadOnlyEntry struct{ widget.Entry }

func (entry *ReadOnlyEntry) TypedRune(_ rune)          {}
func (entry *ReadOnlyEntry) TypedKey(_ *fyne.KeyEvent) {}
func (entry *ReadOnlyEntry) TypedShortcut(shortcut fyne.Shortcut) {
	switch shortcut := shortcut.(type) {
	case *fyne.ShortcutCopy:
		entry.Entry.TypedShortcut(shortcut)
	}
}

func (entry *ReadOnlyEntry) TappedSecondary(ev *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	copyItem := fyne.NewMenuItem("复制", func() {
		c.SetContent(entry.SelectedText())
	})
	menu := fyne.NewMenu("", copyItem)
	widget.ShowPopUpMenuAtPosition(menu, fyne.CurrentApp().Driver().CanvasForObject(entry), ev.AbsolutePosition)
}

func NewReadOnlyEntry() *ReadOnlyEntry {
	entry := &ReadOnlyEntry{}
	entry.ExtendBaseWidget(entry) // 确保自定义的 widget 被正确地初始化
	entry.MultiLine = true        // 支持多行文本
	entry.TextStyle = fyne.TextStyle{Monospace: true}
	entry.Wrapping = fyne.TextWrapOff
	return entry
}

func InitWidgets() {
	StatusProcessBar = widget.NewProgressBarInfinite()
	StatusProcessBar.Stop()
	StatusProcessBar.Hide()

	StatusLabel = widget.NewLabel("就绪.")

	DownloadButton = &widget.Button{Text: "下载",
		OnTapped: func() { go downloadButtonFunc() },
		Icon:     theme.DownloadIcon()}

	SetNicknameButton = &widget.Button{Text: "编辑别称",
		OnTapped: func() { go setNicknameButtonFunc() },
		Icon:     theme.DocumentCreateIcon()}

	DeleteProfileButton = &widget.Button{Text: "删除",
		OnTapped: func() { go deleteProfileButtonFunc() },
		Icon:     theme.DeleteIcon()}

	SwitchStateButton = &widget.Button{Text: "启用",
		OnTapped: func() { go switchStateButtonFunc() },
		Icon:     theme.ConfirmIcon()}

	ProfileList = initProfileList()
	NotificationList = initNotificationList()

	ProcessNotificationButton = &widget.Button{Text: "处理",
		OnTapped: func() { go processNotificationButtonFunc() },
		Icon:     theme.MediaPlayIcon()}

	ProcessAllNotificationButton = &widget.Button{Text: "处理全部",
		OnTapped: func() { go processAllNotificationButtonFunc() },
		Icon:     theme.MediaReplayIcon()}

	RemoveNotificationButton = &widget.Button{Text: "移除",
		OnTapped: func() { go removeNotificationButtonFunc() },
		Icon:     theme.ContentRemoveIcon()}

	BatchRemoveNotificationButton = &widget.Button{Text: "批量移除",
		OnTapped: func() { go batchRemoveNotificationButtonFunc() },
		Icon:     theme.DeleteIcon()}

	FreeSpaceLabel = widget.NewLabel("")

	OpenLogButton = &widget.Button{Text: "查看日志",
		OnTapped: func() { go OpenLog() },
		Icon:     theme.FolderOpenIcon()}

	RefreshButton = &widget.Button{Text: "刷新",
		OnTapped: func() { go Refresh() },
		Icon:     theme.ViewRefreshIcon()}

	ProfileMaskCheck = widget.NewCheck("掩码", func(b bool) {
		if b {
			ProfileMaskNeeded = true
			ProfileList.Refresh()
		} else {
			ProfileMaskNeeded = false
			ProfileList.Refresh()
		}
	})
	NotificationMaskCheck = widget.NewCheck("掩码", func(b bool) {
		if b {
			NotificationMaskNeeded = true
			NotificationList.Refresh()
		} else {
			NotificationMaskNeeded = false
			NotificationList.Refresh()
		}
	})

	EidLabel = widget.NewLabel("")
	DefaultDpAddressLabel = widget.NewLabel("")
	RootDsAddressLabel = widget.NewLabel("")
	EuiccInfo2Entry = NewReadOnlyEntry()
	EuiccInfo2Entry.Hide()
	CopyEidButton = &widget.Button{Text: "复制",
		OnTapped: func() { go copyEidButtonFunc() },
		Icon:     theme.ContentCopyIcon()}
	CopyEidButton.Hide()
	SetDefaultSmdpButton = &widget.Button{OnTapped: func() { go setDefaultSmdpButtonFunc() },
		Icon: theme.DocumentCreateIcon()}
	SetDefaultSmdpButton.Hide()
	ViewCertInfoButton = &widget.Button{Text: "证书颁发机构",
		OnTapped: func() { go viewCertInfoButtonFunc() },
		Icon:     theme.InfoIcon()}
	ViewCertInfoButton.Hide()
	EUICCManufacturerLabel = &widget.Label{}
	EUICCManufacturerLabel.Hide()
	CopyEuiccInfo2Button = &widget.Button{Text: "复制 eUICCInfo2",
		OnTapped: func() { go copyEuiccInfo2ButtonFunc() },
		Icon:     theme.ContentCopyIcon()}
	CopyEuiccInfo2Button.Hide()
	ApduDriverSelect = widget.NewSelect([]string{}, func(s string) { SetDriverIFID(s) })
	ApduDriverRefreshButton = &widget.Button{OnTapped: func() { go RefreshApduDriver() },
		Icon: theme.SearchReplaceIcon()}
	LpacVersionLabel = &widget.Label{}
}

func downloadButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded == true {
		ShowRefreshNeededDialog()
		return
	}
	InitDownloadDialog().Show()
}

func setNicknameButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	if SelectedProfile == Unselected {
		ShowSelectItemDialog()
		return
	}
	InitSetNicknameDialog().Show()
}

func deleteProfileButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	if SelectedProfile == Unselected {
		ShowSelectItemDialog()
		return
	}
	if Profiles[SelectedProfile].ProfileState == "enabled" {
		d := dialog.NewInformation("提示", "在删除之前，你应该禁用该配置文件.", WMain)
		d.Resize(fyne.Size{
			Width:  360,
			Height: 170,
		})
		d.Show()
		return
	}
	profileText := fmt.Sprint(
		"ICCID: ", Profiles[SelectedProfile].Iccid, "\n",
		"运营商: ", Profiles[SelectedProfile].ServiceProviderName, "\n",
	)
	if Profiles[SelectedProfile].ProfileNickname != nil {
		profileText += fmt.Sprint("别称: ", *Profiles[SelectedProfile].ProfileNickname, "\n")
	}
	dialog.ShowCustomConfirm("确认",
		"确认",
		"取消",
		container.NewVBox(container.NewCenter(widget.NewLabel("您确定要删除此配置文件吗?")),
			&widget.Label{Text: profileText}),
		func(b bool) {
			if b {
				go func() {
					if err := LpacProfileDelete(Profiles[SelectedProfile].Iccid); err != nil {
						ShowLpacErrDialog(err)
						Refresh()
					} else {
						notificationOrigin := Notifications
						Refresh()
						deleteNotification := findNewNotification(notificationOrigin, Notifications)
						if deleteNotification == nil {
							dialog.ShowError(errors.New("通知未找到"), WMain)
							return
						}
						if ConfigInstance.AutoMode {
							// 默认保留 delete 通知
							if err2 := LpacNotificationProcess(deleteNotification.SeqNumber, false); err2 != nil {
								dialog.ShowError(errors.New("已成功删除配置文件，但发送通知失败\n您应尝试手动发送删除通知"), WMain)
							} else {
								// Ask to remove delete notification
								// fixme 和手动操作通知模式重构
								var d *dialog.CustomDialog
								notNowButton := &widget.Button{
									Text: "现在不操作",
									Icon: theme.CancelIcon(),
									OnTapped: func() {
										d.Hide()
									},
								}
								removeButton := &widget.Button{
									Text: "移除",
									Icon: theme.DeleteIcon(),
									OnTapped: func() {
										go func() {
											d.Hide()
											if err3 := LpacNotificationRemove(deleteNotification.SeqNumber); err3 != nil {
												ShowLpacErrDialog(err3)
											}
											if err3 := RefreshNotification(); err3 != nil {
												ShowLpacErrDialog(err3)
												return
											}
											if err3 := RefreshChipInfo(); err3 != nil {
												ShowLpacErrDialog(err3)
												return
											}
										}()
									},
								}
								d = dialog.NewCustomWithoutButtons("移除通知",
									container.NewBorder(
										nil,
										container.NewCenter(container.NewHBox(notNowButton, spacer, removeButton)),
										nil,
										nil,
										container.NewVBox(
											&widget.Label{Text: "已成功删除配置文件并发送通知\n您现在要移除删除通知吗?",
												Alignment: fyne.TextAlignCenter},
											&widget.Label{Text: fmt.Sprintf("序号: %d\nICCID: %s\n操作: %s\n服务器: %s\n",
												deleteNotification.SeqNumber, deleteNotification.Iccid,
												deleteNotification.ProfileManagementOperation, deleteNotification.NotificationAddress)})),
									WMain)
								d.Show()
							}
						} else {
							dialog.ShowConfirm("删除成功",
								"配置文件已成功删除\n现在发送删除通知吗?\n",
								func(b bool) {
									if b {
										go processNotificationManually(deleteNotification.SeqNumber)
									}
								},
								WMain)
						}
					}
				}()
			}
		}, WMain)
}

func switchStateButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	if SelectedProfile == Unselected {
		ShowSelectItemDialog()
		return
	}
	if ProfileStateAllowDisable {
		if err := LpacProfileDisable(Profiles[SelectedProfile].Iccid); err != nil {
			ShowLpacErrDialog(err)
		}
	} else {
		if err := LpacProfileEnable(Profiles[SelectedProfile].Iccid); err != nil {
			ShowLpacErrDialog(err)
		}
	}
	if ConfigInstance.AutoMode {
		notificationsOrigin := Notifications
		Refresh()
		switchNotifications := findNewNotifications(notificationsOrigin, Notifications)
		// 考虑两种情况
		// 所有 Profile 禁用的情况下，启用 Profile 产生一个 enable 通知
		// 有一个 Profile 已启用，启用另外一个，产生一个 disable 和一个 enable 通知
		// 禁用 Profile，产生一个 disable 通知
		if switchNotifications == nil || len(switchNotifications) > 2 {
			dialog.ShowError(errors.New("未能找到通知"), WMain)
		} else {
			dialogText := "已成功启用配置文件\n"
			var hasError bool
			for _, notification := range switchNotifications {
				if err2 := LpacNotificationProcess(notification.SeqNumber, true); err2 != nil {
					hasError = true
					switch notification.ProfileManagementOperation {
					case "enable":
						dialogText += "启用通知处理失败\n"
						break
					case "disable":
						dialogText += "禁用通知处理失败\n"
						break
					}
				}
			}
			if hasError {
				dialog.ShowError(errors.New(dialogText), WMain)
			}
		}
	}
	Refresh()
	if ProfileStateAllowDisable {
		SwitchStateButton.SetText("启用")
		SwitchStateButton.SetIcon(theme.ConfirmIcon())
	}
}

func processNotificationButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	if SelectedNotification == Unselected {
		ShowSelectItemDialog()
		return
	}
	seq := Notifications[SelectedNotification].SeqNumber
	go processNotificationManually(seq)
}

func processAllNotificationButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	config := map[string]bool{
		"enable":  true,
		"disable": true,
		"install": true,
		"delete":  false,
	}
	enableCheck := &widget.Check{
		Text:    "enable",
		Checked: true,
		OnChanged: func(b bool) {
			config["enable"] = b
		},
	}
	disableCheck := &widget.Check{
		Text:    "disable",
		Checked: true,
		OnChanged: func(b bool) {
			config["disable"] = b
		},
	}
	installCheck := &widget.Check{
		Text:    "install",
		Checked: true,
		OnChanged: func(b bool) {
			config["install"] = b
		},
	}
	deleteCheck := &widget.Check{
		Text:    "delete",
		Checked: false,
		OnChanged: func(b bool) {
			config["delete"] = b
		},
	}
	dialog.ShowCustomConfirm("处理所有通知",
		"确定",
		"取消",
		container.NewVBox(
			&widget.Label{Text: "在处理后移除以下通知类型:"},
			enableCheck,
			disableCheck,
			installCheck,
			deleteCheck,
		),
		func(b bool) {
			if b {
				total := len(Notifications)
				var count int
				for _, notification := range Notifications {
					switch notification.ProfileManagementOperation {
					case "enable":
						if err := LpacNotificationProcess(notification.SeqNumber, config["enable"]); err != nil {
							count++
						}
					case "disable":
						if err := LpacNotificationProcess(notification.SeqNumber, config["disable"]); err != nil {
							count++
						}
					case "install":
						if err := LpacNotificationProcess(notification.SeqNumber, config["install"]); err != nil {
							count++
						}
					case "delete":
						if err := LpacNotificationProcess(notification.SeqNumber, config["delete"]); err != nil {
							count++
						}
					}
				}
				if err := RefreshNotification(); err != nil {
					ShowLpacErrDialog(err)
				}
				dialog.ShowCustom("操作完成",
					"确定",
					&widget.Label{Text: fmt.Sprintf("%d 个已处理\n%d 个成功\n%d 个失败", total, total-count, count)},
					WMain)
			}
		}, WMain)
}

func removeNotificationButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	if SelectedNotification == Unselected {
		ShowSelectItemDialog()
		return
	}
	dialog.ShowCustomConfirm("确认",
		"确认",
		"取消",
		&widget.Label{Text: "您确定要移除此通知吗?\n",
			Alignment: fyne.TextAlignCenter},
		func(b bool) {
			if b {
				if err := LpacNotificationRemove(Notifications[SelectedNotification].SeqNumber); err != nil {
					ShowLpacErrDialog(err)
				}

				if err := RefreshNotification(); err != nil {
					ShowLpacErrDialog(err)
					return
				}

				if err := RefreshChipInfo(); err != nil {
					ShowLpacErrDialog(err)
					return
				}
			}
		}, WMain)
}

func batchRemoveNotificationButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	config := map[string]bool{
		"enable":  true,
		"disable": true,
		"install": true,
		"delete":  false,
	}
	enableCheck := &widget.Check{
		Text:    "enable",
		Checked: true,
		OnChanged: func(b bool) {
			config["enable"] = b
		},
	}
	disableCheck := &widget.Check{
		Text:    "disable",
		Checked: true,
		OnChanged: func(b bool) {
			config["disable"] = b
		},
	}
	installCheck := &widget.Check{
		Text:    "Install",
		Checked: true,
		OnChanged: func(b bool) {
			config["install"] = b
		},
	}
	deleteCheck := &widget.Check{
		Text:    "delete",
		Checked: false,
		OnChanged: func(b bool) {
			config["delete"] = b
		},
	}
	dialog.ShowCustomConfirm("批量移除通知", "确认", "取消",
		container.NewVBox(
			&widget.Label{Text: "选择要移除的通知类型"},
			enableCheck,
			disableCheck,
			installCheck,
			deleteCheck),
		func(b bool) {
			if b {
				var failedCount int
				var total int
				for _, notification := range Notifications {
					switch notification.ProfileManagementOperation {
					case "enable":
						if err := LpacNotificationRemove(notification.SeqNumber); err != nil {
							failedCount++
						}
						total++
					case "disable":
						if err := LpacNotificationProcess(notification.SeqNumber, config["disable"]); err != nil {
							failedCount++
						}
						total++
					case "install":
						if err := LpacNotificationProcess(notification.SeqNumber, config["install"]); err != nil {
							failedCount++
						}
						total++
					case "delete":
						if err := LpacNotificationProcess(notification.SeqNumber, config["delete"]); err == nil {
							failedCount++
						}
						total++
					}
				}
				if err := RefreshNotification(); err != nil {
					ShowLpacErrDialog(err)
				}
				dialog.ShowCustom("操作完成",
					"确定",
					&widget.Label{Text: fmt.Sprintf("%d 个已处理\n%d 个成功\n%d 个失败", total, total-failedCount, failedCount)},
					WMain)
			}
		}, WMain)
}

func copyEidButtonFunc() {
	WMain.Clipboard().SetContent(ChipInfo.EidValue)
	CopyEidButton.SetText("已复制!")
	time.Sleep(2 * time.Second)
	CopyEidButton.SetText("复制")
}

func copyEuiccInfo2ButtonFunc() {
	WMain.Clipboard().SetContent(EuiccInfo2Entry.Text)
	CopyEuiccInfo2Button.SetText("已复制eUICCInfo2!")
	time.Sleep(2 * time.Second)
	CopyEuiccInfo2Button.SetText("复制eUICCInfo2")
}

func setDefaultSmdpButtonFunc() {
	if ConfigInstance.DriverIFID == "" {
		ShowSelectCardReaderDialog()
		return
	}
	if RefreshNeeded {
		ShowRefreshNeededDialog()
		return
	}
	InitSetDefaultSmdpDialog().Show()
}

func viewCertInfoButtonFunc() {
	selectedCI := Unselected
	type ciWidgetEl struct {
		Country string
		Name    string
		KeyID   string
	}
	var ciWidgetEls []ciWidgetEl
	// ChipInfo 中 signing 和 verification 同时存在则有效
	for _, keyId := range ChipInfo.EUICCInfo2.EuiccCiPKIDListForSigning {
		if !slices.Contains(ChipInfo.EUICCInfo2.EuiccCiPKIDListForVerification, keyId) {
			continue
		}
		var element ciWidgetEl
		element.KeyID = keyId
		element.Name = "Unknown"
		if issuer := GetIssuer(keyId); issuer != nil {
			element.Country = issuer.Country
			element.Name = issuer.Name
		}
		ciWidgetEls = append(ciWidgetEls, element)
	}
	list := &widget.List{
		Length: func() int {
			return len(ciWidgetEls)
		},
		CreateItem: func() fyne.CanvasObject {
			return container.NewVBox(container.NewBorder(nil, nil,
				&widget.Label{}, &widget.Label{}),
				&widget.Label{})
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).SetText(ciWidgetEls[i].Name)
			o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(CountryCodeToEmoji(ciWidgetEls[i].Country))
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("KeyID: %s", ciWidgetEls[i].KeyID))
		},
		OnSelected: func(id widget.ListItemID) {
			selectedCI = id
		},
		OnUnselected: func(id widget.ListItemID) {
			selectedCI = Unselected
		},
	}
	certDataButtonFunc := func() {
		if selectedCI == Unselected {
			ShowSelectItemDialog()
		} else if issuer := GetIssuer(ciWidgetEls[selectedCI].KeyID); issuer == nil {
			dialog.ShowInformation("没有数据",
				"此证书的信息未包含在内.\n"+
					"如果您对此证书有任何信息,\n"+
					"请报告至 <euicc-dev-manual@septs.pw>\n"+
					"谢谢",
				WMain)
		} else {
			const CiUrl = "https://euicc-manual.septs.app/docs/pki/ci/files/"
			certificateURL := fmt.Sprint(CiUrl, issuer.KeyID, ".txt")
			if err := OpenProgram(certificateURL); err != nil {
				dialog.ShowError(err, WMain)
			}
		}
	}
	certDataButton := &widget.Button{
		Text:     "证书信息",
		OnTapped: certDataButtonFunc,
		Icon:     theme.InfoIcon(),
	}
	d := dialog.NewCustom("证书颁发机构", "确定",
		container.NewBorder(nil, container.NewCenter(certDataButton), nil, nil, list), WMain)
	d.Resize(fyne.Size{
		Width:  600,
		Height: 500,
	})
	d.Show()
}

func initProfileList() *widget.List {
	return &widget.List{
		Length: func() int {
			return len(Profiles)
		},
		CreateItem: func() fyne.CanvasObject {
			iccidLabel := &widget.Label{}
			nameLabel := &widget.Label{}
			stateLabel := &widget.Label{TextStyle: fyne.TextStyle{Bold: true}}
			enabledIcon := widget.NewIcon(theme.ConfirmIcon())
			profileIcon := widget.NewIcon(theme.FileImageIcon())
			providerLabel := &widget.Label{}
			return container.NewVBox(
				container.NewHBox(iccidLabel, layout.NewSpacer(), nameLabel),
				container.NewHBox(container.NewVBox(layout.NewSpacer(), stateLabel),
					enabledIcon, providerLabel, profileIcon, layout.NewSpacer()))
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			r1 := o.(*fyne.Container).Objects[0].(*fyne.Container)
			r2 := o.(*fyne.Container).Objects[1].(*fyne.Container)
			iccidLabel := r1.Objects[0].(*widget.Label)
			nameLabel := r1.Objects[2].(*widget.Label)
			stateLabel := r2.Objects[0].(*fyne.Container).Objects[1].(*widget.Label)
			enabledIcon := r2.Objects[1].(*widget.Icon)
			providerLabel := r2.Objects[2].(*widget.Label)
			profileIcon := r2.Objects[3].(*widget.Icon)

			iccid := Profiles[i].Iccid
			if ProfileMaskNeeded {
				iccid = Profiles[i].MaskedICCID()
			}
			iccidLabel.SetText(fmt.Sprintf("ICCID: %s", iccid))
			if Profiles[i].ProfileNickname != nil {
				nameLabel.SetText(*Profiles[i].ProfileNickname)
			} else {
				nameLabel.SetText(Profiles[i].ProfileName)
			}

			if Profiles[i].ProfileState == "enabled" {
				stateLabel.SetText("已启用")
				enabledIcon.Show()
			} else {
				stateLabel.SetText("已禁用")
				enabledIcon.Hide()
			}

			if Profiles[i].Icon != nil {
				profileIcon.SetResource(fyne.NewStaticResource(Profiles[i].Iccid, Profiles[i].Icon))
				profileIcon.Show()
			} else {
				profileIcon.Hide()
			}

			providerLabel.SetText("运营商: " + Profiles[i].ServiceProviderName)
		},
		OnSelected: func(id widget.ListItemID) {
			SelectedProfile = id
			if Profiles[SelectedProfile].ProfileState == "enabled" {
				ProfileStateAllowDisable = true
				SwitchStateButton.SetText("禁用")
				SwitchStateButton.SetIcon(theme.CancelIcon())
			} else {
				ProfileStateAllowDisable = false
				SwitchStateButton.SetText("启用")
				SwitchStateButton.SetIcon(theme.ConfirmIcon())
			}
		},
		OnUnselected: func(id widget.ListItemID) {
			SelectedProfile = Unselected
		}}
}

func initNotificationList() *widget.List {
	maskFQDNExceptPublicSuffix := func(fqdn string) string {
		suffix, _ := publicsuffix.PublicSuffix(fqdn)
		parts := strings.Split(fqdn, ".")
		suffixParts := strings.Split(suffix, ".")
		// 如果域名部分少于后缀部分，说明域名不合法或者是一个裸域名，直接返回掩码后的顶级域名
		if len(parts) <= len(suffixParts) {
			return strings.Repeat("x", len(parts[0])) + "." + suffix
		}
		// 掩盖除了后缀之外的所有部分
		for x := 0; x < len(parts)-len(suffixParts); x++ {
			parts[x] = strings.Repeat("x", len(parts[x]))
		}
		return strings.Join(parts, ".")
	}

	return &widget.List{
		Length: func() int {
			return len(Notifications)
		},
		CreateItem: func() fyne.CanvasObject {
			notificationAddressLabel := &widget.Label{}
			seqLabel := &widget.Label{}
			operationLabel := &widget.Label{TextStyle: fyne.TextStyle{Bold: true}}
			providerLabel := &widget.Label{}
			iccidLabel := &widget.Label{}
			providerIcon := widget.NewIcon(theme.FileImageIcon())
			return container.NewVBox(
				container.NewHBox(notificationAddressLabel, layout.NewSpacer(), seqLabel),
				container.NewHBox(container.NewVBox(layout.NewSpacer(), operationLabel), providerLabel, providerIcon, iccidLabel),
			)
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			notificationAddressLabel := o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label)
			seqLabel := o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[2].(*widget.Label)
			iccidLabel := o.(*fyne.Container).Objects[1].(*fyne.Container).Objects[3].(*widget.Label)
			operationLabel := o.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*widget.Label)
			providerLabel := o.(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*widget.Label)
			providerIcon := o.(*fyne.Container).Objects[1].(*fyne.Container).Objects[2].(*widget.Icon)

			iccid := Notifications[i].Iccid
			notificationAddress := Notifications[i].NotificationAddress
			if NotificationMaskNeeded {
				if iccid != "" {
					iccid = Notifications[i].MaskedICCID()
				}
				notificationAddress = maskFQDNExceptPublicSuffix(Notifications[i].NotificationAddress)
			}
			// ICCID
			if iccid == "" {
				iccid = "无 ICCID!"
			}
			iccidLabel.SetText(fmt.Sprint("(", iccid, ")"))
			// Notification Address
			notificationAddressLabel.SetText(notificationAddress)
			// Seq number
			seqLabel.SetText(fmt.Sprint("序号: ", Notifications[i].SeqNumber))
			// Operation
			operationLabel.
				SetText(Notifications[i].CapitalizedOperation())
			// Provider
			profile, err := findProfileByIccid(Notifications[i].Iccid)
			if err != nil {
				providerLabel.SetText("?已删除的配置文件")
				providerIcon.Hide()
			} else {
				name := profile.ServiceProviderName
				if profile.ProfileNickname != nil {
					name = *profile.ProfileNickname
				}
				providerLabel.SetText(name)
				if profile.Icon != nil {
					providerIcon.SetResource(fyne.NewStaticResource(profile.Iccid, profile.Icon))
					providerIcon.Show()
				} else {
					providerIcon.Hide()
				}
			}
		},
		OnSelected: func(id widget.ListItemID) {
			SelectedNotification = id
		},
		OnUnselected: func(id widget.ListItemID) {
			SelectedNotification = Unselected
		}}
}

func processNotificationManually(seq int) {
	if err := LpacNotificationProcess(seq, false); err != nil {
		ShowLpacErrDialog(err)
		err2 := RefreshNotification()
		if err2 != nil {
			ShowLpacErrDialog(err2)
		}
	} else {
		var notification *Notification
		for _, n := range Notifications {
			if n.SeqNumber == seq {
				notification = n
				break
			}
		}
		if notification == nil {
			// 不应该出现
			dialog.ShowError(errors.New("未找到通知"), WMain)
			return
		}
		var d *dialog.CustomDialog
		notNowButton := &widget.Button{
			Text: "现在不操作",
			Icon: theme.CancelIcon(),
			OnTapped: func() {
				d.Hide()
			},
		}
		removeButton := &widget.Button{
			Text: "移除",
			Icon: theme.DeleteIcon(),
			OnTapped: func() {
				go func() {
					d.Hide()
					if err2 := LpacNotificationRemove(seq); err2 != nil {
						ShowLpacErrDialog(err2)
					}
					if err2 := RefreshNotification(); err2 != nil {
						ShowLpacErrDialog(err2)
						return
					}
					if err2 := RefreshChipInfo(); err2 != nil {
						ShowLpacErrDialog(err2)
						return
					}
				}()
			},
		}
		d = dialog.NewCustomWithoutButtons("移除通知",
			container.NewBorder(
				nil,
				container.NewCenter(container.NewHBox(notNowButton, spacer, removeButton)),
				nil,
				nil,
				container.NewVBox(
					&widget.Label{Text: "成功处理通知.\n你想立即移除此通知吗?",
						Alignment: fyne.TextAlignCenter},
					&widget.Label{Text: fmt.Sprintf("序号: %d\nICCID: %s\n操作: %s\n服务器: %s\n",
						notification.SeqNumber, notification.Iccid,
						notification.ProfileManagementOperation, notification.NotificationAddress)})),
			WMain)
		d.Show()
	}
}

func findNewNotification(origin, new []*Notification) *Notification {
	exists := make(map[int]bool)
	for _, notification := range origin {
		exists[notification.SeqNumber] = true
	}
	for _, notification := range new {
		if !exists[notification.SeqNumber] {
			return notification
		}
	}
	return nil
}

func findNewNotifications(origin, new []*Notification) []*Notification {
	exists := make(map[int]bool)
	var foundNotifications []*Notification
	for _, notification := range origin {
		exists[notification.SeqNumber] = true
	}
	for _, notification := range new {
		if !exists[notification.SeqNumber] {
			foundNotifications = append(foundNotifications, notification)
		}
	}
	return foundNotifications
}

func findProfileByIccid(iccid string) (*Profile, error) {
	for _, profile := range Profiles {
		if iccid == profile.Iccid {
			return profile, nil
		}
	}
	return nil, errors.New("未找到配置文件")
}
