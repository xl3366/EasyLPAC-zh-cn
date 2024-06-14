package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/makiuchi-d/gozxing"
	nativeDialog "github.com/sqweek/dialog"
	"golang.design/x/clipboard"
	"image/color"
	"strings"
)

var WMain fyne.Window
var spacer *canvas.Rectangle

func InitMainWindow() fyne.Window {
	w := App.NewWindow(fmt.Sprintf("EasyLPAC 速易卡(版本: %s)", Version))
	w.Resize(fyne.Size{
		Width:  850,
		Height: 545,
	})
	w.SetMaster()

	statusBar := container.NewGridWrap(fyne.Size{
		Width:  100,
		Height: DownloadButton.MinSize().Height,
	}, StatusLabel, StatusProcessBar)

	spacer = canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(1, 1))

	topToolBar := container.NewBorder(
		layout.NewSpacer(),
		nil,
		container.New(layout.NewHBoxLayout(), OpenLogButton, spacer, RefreshButton, spacer),
		FreeSpaceLabel,
		container.NewBorder(
			nil,
			nil,
			widget.NewLabel("读卡器:"),
			nil,
			container.NewHBox(container.NewGridWrap(fyne.Size{
				Width:  280,
				Height: ApduDriverSelect.MinSize().Height,
			}, ApduDriverSelect), ApduDriverRefreshButton)),
	)

	profileTabContent := container.NewBorder(
		topToolBar,
		container.NewBorder(
			nil,
			nil,
			nil,
			container.NewHBox(ProfileMaskCheck, DownloadButton,
				// spacer, DiscoveryButton,
				spacer, SetNicknameButton,
				spacer, SwitchStateButton,
				spacer, DeleteProfileButton),
			statusBar),
		nil,
		nil,
		ProfileList)
	ProfileTab = container.NewTabItem("配置文件", profileTabContent)

	notificationTabContent := container.NewBorder(
		topToolBar,
		container.NewBorder(
			nil,
			nil,
			nil,
			container.NewHBox(NotificationMaskCheck,
				spacer, ProcessNotificationButton,
				spacer, ProcessAllNotificationButton,
				spacer, BatchRemoveNotificationButton,
				spacer, RemoveNotificationButton),
			statusBar),
		nil,
		nil,
		NotificationList)
	NotificationTab = container.NewTabItem("通知", notificationTabContent)

	chipInfoTabContent := container.NewBorder(
		topToolBar,
		container.NewBorder(
			nil,
			nil,
			nil,
			nil,
			statusBar),
		nil,
		nil,
		container.NewBorder(
			container.NewVBox(
				container.NewHBox(
					EidLabel, CopyEidButton, layout.NewSpacer(), EUICCManufacturerLabel),
				container.NewHBox(
					DefaultDpAddressLabel, SetDefaultSmdpButton, layout.NewSpacer(), ViewCertInfoButton),
				container.NewHBox(
					RootDsAddressLabel, layout.NewSpacer(), CopyEuiccInfo2Button)),
			nil,
			nil,
			nil,
			container.NewScroll(EuiccInfo2Entry),
		))
	ChipInfoTab = container.NewTabItem("芯片信息", chipInfoTabContent)

	settingsTabContent := container.NewVBox(
		&widget.Label{Text: "LPAC 调试输出", TextStyle: fyne.TextStyle{Bold: true}},
		&widget.Check{
			Text:    "启用环境变量 LIBEUICC_DEBUG_HTTP",
			Checked: false,
			OnChanged: func(b bool) {
				ConfigInstance.DebugHTTP = b
			},
		},
		&widget.Check{
			Text:    "启用环境变量 LIBEUICC_DEBUG_APDU",
			Checked: false,
			OnChanged: func(b bool) {
				ConfigInstance.DebugAPDU = b
			},
		},
		&widget.Label{Text: "EasyLPAC 设置", TextStyle: fyne.TextStyle{Bold: true}},
		&widget.Check{
			Text:    "自动处理通知",
			Checked: true,
			OnChanged: func(b bool) {
				ConfigInstance.AutoMode = b
			},
		})
	SettingsTab = container.NewTabItem("设置", settingsTabContent)

	thankstoText := widget.NewRichTextFromMarkdown(`
# 鸣谢

[lpac](https://github.com/estkme-group/lpac) C-based eUICC LPA

[eUICC Manual](https://euicc-manual.osmocom.org) eUICC Developer Manual

[fyne](https://github.com/fyne-io/fyne) Material Design GUI toolkit`)

	aboutText := widget.NewRichTextFromMarkdown(`
# EasyLPAC

lpac GUI Frontend

[Github](https://github.com/creamlike1024/EasyLPAC) Repo `)

	suyikaText := widget.NewRichTextFromMarkdown(`
# 速易卡

[MFF2转(QFN8)4FF卡板](https://item.taobao.com/item.htm?id=730209105541) 淘宝

[ST33G(WLCSP11)转4FF卡板](https://item.taobao.com/item.htm?id=723325804913) 淘宝 `)

	aboutTabContent := container.NewBorder(
		nil,
		container.NewBorder(nil, nil,
			container.NewHBox(
				widget.NewLabel(fmt.Sprintf("版本: %s", Version)),
				LpacVersionLabel),
			widget.NewLabel(fmt.Sprintf("eUICC 数据: %s", EUICCDataVersion))),
		nil,
		nil,
		container.NewCenter(container.NewVBox(thankstoText, aboutText,suyikaText)))
	AboutTab = container.NewTabItem("关于", aboutTabContent)

	Tabs = container.NewAppTabs(ProfileTab, NotificationTab, ChipInfoTab, SettingsTab, AboutTab)

	w.SetContent(Tabs)

	return w
}

func InitDownloadDialog() dialog.Dialog {
	smdpEntry := &widget.Entry{PlaceHolder: "留空使用默认SM-DP+"}
	matchIDEntry := &widget.Entry{PlaceHolder: "激活码(选填)"}
	confirmCodeEntry := &widget.Entry{PlaceHolder: "确认码(选填)"}
	imeiEntry := &widget.Entry{PlaceHolder: "终端IMEI(选填)"}

	formItems := []*widget.FormItem{
		{Text: "SM-DP+", Widget: smdpEntry},
		{Text: "激活码", Widget: matchIDEntry},
		{Text: "确认码", Widget: confirmCodeEntry},
		{Text: "IMEI", Widget: imeiEntry},
	}

	form := widget.NewForm(formItems...)
	var d dialog.Dialog
	showConfirmCodeNeededDialog := func() {
		dialog.ShowInformation("需要确认码",
			"此配置文件需要确认码才能下载.\n"+
				"请手动填写确认码.", WMain)
	}
	cancelButton := &widget.Button{
		Text: "取消",
		Icon: theme.CancelIcon(),
		OnTapped: func() {
			d.Hide()
		},
	}
	downloadButton := &widget.Button{
		Text:       "下载",
		Icon:       theme.ConfirmIcon(),
		Importance: widget.HighImportance,
		OnTapped: func() {
			d.Hide()
			pullConfig := PullInfo{
				SMDP:        strings.TrimSpace(smdpEntry.Text),
				MatchID:     strings.TrimSpace(matchIDEntry.Text),
				ConfirmCode: strings.TrimSpace(confirmCodeEntry.Text),
				IMEI:        strings.TrimSpace(imeiEntry.Text),
			}
			go func() {
				err := RefreshNotification()
				if err != nil {
					ShowLpacErrDialog(err)
					return
				}
				LpacProfileDownload(pullConfig)
			}()
		},
	}
	// 回调函数需要操作这两个 Button，预先声明
	var selectQRCodeButton *widget.Button
	var pasteFromClipboardButton *widget.Button
	disableButtons := func() {
		cancelButton.Disable()
		downloadButton.Disable()
		selectQRCodeButton.Disable()
		pasteFromClipboardButton.Disable()
	}
	enableButtons := func() {
		cancelButton.Enable()
		downloadButton.Enable()
		selectQRCodeButton.Enable()
		pasteFromClipboardButton.Enable()
	}

	selectQRCodeButton = &widget.Button{
		Text: "扫描图像文件",
		Icon: theme.FileImageIcon(),
		OnTapped: func() {
			go func() {
				disableButtons()
				defer enableButtons()
				fileBuilder := nativeDialog.File().Title("选择一个二维码图片文件")
				fileBuilder.Filters = []nativeDialog.FileFilter{
					{
						Desc:       "Image (*.png, *.jpg, *.jpeg)",
						Extensions: []string{"PNG", "JPG", "JPEG"},
					},
					{
						Desc:       "All files (*.*)",
						Extensions: []string{"*"},
					},
				}

				filename, err := fileBuilder.Load()
				if err != nil {
					if err.Error() != "Cancelled" {
						panic(err)
					}
				} else {
					result, err := ScanQRCodeImageFile(filename)
					if err != nil {
						dialog.ShowError(err, WMain)
					} else {
						pullInfo, confirmCodeNeeded, err2 := DecodeLpaActivationCode(result.String())
						if err2 != nil {
							dialog.ShowError(err2, WMain)
						} else {
							smdpEntry.SetText(pullInfo.SMDP)
							matchIDEntry.SetText(pullInfo.MatchID)
							if confirmCodeNeeded {
								go showConfirmCodeNeededDialog()
							}
						}
					}
				}
			}()
		},
	}
	pasteFromClipboardButton = &widget.Button{
		Text: "从剪贴板粘贴二维码或LPA:1激活码",
		Icon: theme.ContentPasteIcon(),
		OnTapped: func() {
			go func() {
				disableButtons()
				defer enableButtons()
				var err error
				var pullInfo PullInfo
				var confirmCodeNeeded bool
				var qrResult *gozxing.Result

				format, result, err := PasteFromClipboard()
				if err != nil {
					dialog.ShowError(err, WMain)
					return
				}
				switch format {
				case clipboard.FmtImage:
					qrResult, err = ScanQRCodeImageBytes(result)
					if err != nil {
						dialog.ShowError(err, WMain)
						return
					}
					pullInfo, confirmCodeNeeded, err = DecodeLpaActivationCode(qrResult.String())
				case clipboard.FmtText:
					pullInfo, confirmCodeNeeded, err = DecodeLpaActivationCode(CompleteActivationCode(string(result)))
				default:
					// Unreachable, should not be here.
					panic(nil)
				}
				if err != nil {
					dialog.ShowError(err, WMain)
					return
				}
				smdpEntry.SetText(pullInfo.SMDP)
				matchIDEntry.SetText(pullInfo.MatchID)
				if confirmCodeNeeded {
					go showConfirmCodeNeededDialog()
				}
			}()
		},
	}
	d = dialog.NewCustomWithoutButtons("下载", container.NewBorder(
		nil,
		container.NewVBox(spacer, container.NewCenter(selectQRCodeButton), spacer,
			container.NewCenter(pasteFromClipboardButton), spacer,
			container.NewCenter(container.NewHBox(cancelButton, spacer, downloadButton))),
		nil,
		nil,
		form), WMain)
	d.Resize(fyne.Size{
		Width:  520,
		Height: 380,
	})
	return d
}

func InitSetNicknameDialog() dialog.Dialog {
	entry := &widget.Entry{PlaceHolder: "留空删除别称"}
	form := []*widget.FormItem{
		{Text: "别称", Widget: entry},
	}
	d := dialog.NewForm("设置别称", "提交", "取消", form, func(b bool) {
		if b {
			if err := LpacProfileNickname(Profiles[SelectedProfile].Iccid, entry.Text); err != nil {
				ShowLpacErrDialog(err)
			}
			err := RefreshProfile()
			if err != nil {
				ShowLpacErrDialog(err)
			}
		}
	}, WMain)
	d.Resize(fyne.Size{
		Width:  400,
		Height: 180,
	})
	return d
}

func InitSetDefaultSmdpDialog() dialog.Dialog {
	entry := &widget.Entry{PlaceHolder: "留空删除默认的SM-DP+"}
	form := []*widget.FormItem{
		{Text: "默认SM-DP+", Widget: entry},
	}
	d := dialog.NewForm("设置默认SM-DP+", "提交", "取消", form, func(b bool) {
		if b {
			if err := LpacChipDefaultSmdp(entry.Text); err != nil {
				ShowLpacErrDialog(err)
			}
			err := RefreshChipInfo()
			if err != nil {
				ShowLpacErrDialog(err)
			}
		}
	}, WMain)
	d.Resize(fyne.Size{
		Width:  510,
		Height: 200,
	})
	return d
}

func ShowLpacErrDialog(err error) {
	go func() {
		l := &widget.Label{Text: fmt.Sprintf("%v", err)}
		content := container.NewVBox(
			container.NewCenter(container.NewHBox(
				widget.NewIcon(theme.ErrorIcon()),
				widget.NewLabel("lpac 错误"))),
			container.NewCenter(l),
			container.NewCenter(widget.NewLabel("请查看日志获取详细信息")))
		dialog.ShowCustom("错误", "确定", content, WMain)
	}()
}

func ShowSelectItemDialog() {
	go func() {
		d := dialog.NewInformation("信息", "请选择一个项目.", WMain)
		d.Resize(fyne.Size{
			Width:  220,
			Height: 160,
		})
		d.Show()
	}()
}

func ShowSelectCardReaderDialog() {
	go func() {
		dialog.ShowInformation("信息", "请选择一个读卡器.", WMain)
	}()
}

func ShowRefreshNeededDialog() {
	go func() {
		dialog.ShowInformation("信息", "请在继续之前进行刷新.\n", WMain)
	}()
}
