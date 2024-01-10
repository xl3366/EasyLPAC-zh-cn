package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var StatusProcessBar *widget.ProgressBarInfinite
var StatusLabel *widget.Label
var DownloadButton *widget.Button
var DeleteButton *widget.Button
var EnableButton *widget.Button
var ProfileList *widget.List
var NotificationList *widget.List

var ProfileListTitle *widget.Label
var NotificationListTitle *widget.Label

var FreeSpaceLabel *widget.Label
var OpenLogButton *widget.Button
var RefreshProfileButton *widget.Button
var RefreshNotificationButton *widget.Button
var ProcessNotificationButton *widget.Button
var RemoveNotificationButton *widget.Button

func InitWidgets() {
	StatusProcessBar = widget.NewProgressBarInfinite()
	StatusProcessBar.Stop()
	StatusProcessBar.Hide()
	StatusLabel = widget.NewLabel("Ready.")

	DownloadButton = widget.NewButton("Download", func() {
		downloadWindow := InitDownloadWindow()
		downloadWindow.Show()
	})
	DownloadButton.SetIcon(theme.DownloadIcon())
	DeleteButton = widget.NewButton("Delete", func() {
		if SelectedProfile < 0 || SelectedProfile >= len(Profiles) {
			SelectItemDialog()
			return
		}
		dialogText := fmt.Sprintf("Are you sure you want to delete this profile?\n\n%s\t\t%s",
			Profiles[SelectedProfile].Iccid,
			Profiles[SelectedProfile].ServiceProviderName)
		if Profiles[SelectedProfile].ProfileNickname != nil {
			dialogText += fmt.Sprintf("\t\t%s\n\n", Profiles[SelectedProfile].ProfileNickname)
		} else {
			dialogText += "\n\n"
		}
		d := dialog.NewConfirm("Confirm",
			dialogText,
			func(b bool) {
				if b {
					if err := LpacProfileDelete(Profiles[SelectedProfile].Iccid); err != nil {
						ErrDialog(err)
					}
					RefreshProfile()
					// SelectedProfile = -1
				} else {
					return
				}
			}, WMain)
		d.Show()
	})
	EnableButton = widget.NewButton("Enable", func() {
		if SelectedProfile < 0 || SelectedProfile >= len(Profiles) {
			SelectItemDialog()
			return
		}
		if err := LpacProfileEnable(Profiles[SelectedProfile].Iccid); err != nil {
			ErrDialog(err)
		}
		RefreshProfile()
	})

	ProfileList = widget.NewList(
		func() int {
			return len(Profiles)
		},
		func() fyne.CanvasObject {
			return widget.NewRichText()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			text := fmt.Sprintf("%s\t\t%s\t\t\t%s",
				Profiles[i].Iccid,
				Profiles[i].ProfileState,
				Profiles[i].ServiceProviderName)
			if Profiles[i].ProfileNickname != nil {
				text += fmt.Sprintf("\t\t\t%s", Profiles[i].ProfileNickname)
			}
			if Profiles[i].ProfileState == "enabled" {
				text = "**" + text + "**"
			}
			o.(*widget.RichText).ParseMarkdown(text)
		})
	ProfileList.OnSelected = func(id widget.ListItemID) {
		SelectedProfile = id
	}
	ProfileListTitle = widget.NewLabel(fmt.Sprintf("%19s\t\t\t%s\t\t\t%s\t\t\t\t%s", "ICCID", "Profile State", "Provider", "Nickname"))

	NotificationList = widget.NewList(
		func() int {
			return len(Notifications)
		},
		func() fyne.CanvasObject {
			return widget.NewRichText()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			text := fmt.Sprintf("%-4d%27s\t\t%s",
				Notifications[i].SeqNumber,
				Notifications[i].Iccid,
				Notifications[i].ProfileManagementOperation)
			if Notifications[i].ProfileManagementOperation == "install" {
				text += fmt.Sprintf("\t\t\t\t%s", Notifications[i].NotificationAddress)
			} else {
				text += fmt.Sprintf("\t\t\t%s", Notifications[i].NotificationAddress)
			}
			o.(*widget.RichText).ParseMarkdown(text)
		})
	NotificationList.OnSelected = func(id widget.ListItemID) {
		SelectedNotification = id
	}
	NotificationListTitle = widget.NewLabel(fmt.Sprintf("%s\t%19s\t\t\t\t%s\t\t\t\t%s", "Seq", "ICCID", "Operation", "Server"))

	ProcessNotificationButton = widget.NewButton("Process", func() {
		if SelectedNotification < 0 || SelectedNotification >= len(Notifications) {
			SelectItemDialog()
			return
		}
		seq := Notifications[SelectedNotification].SeqNumber
		if err := LpacNotificationProcess(seq); err != nil {
			ErrDialog(err)
		} else {
			dialogText := fmt.Sprintf("Successfully processed notification.\nDo you want to remove this notification now?\n\n%d\t\t%s\t\t%s\t\t%s\n\n",
				Notifications[SelectedNotification].SeqNumber,
				Notifications[SelectedNotification].Iccid,
				Notifications[SelectedNotification].ProfileManagementOperation,
				Notifications[SelectedNotification].NotificationAddress)
			d := dialog.NewConfirm("Remove Notification",
				dialogText,
				func(b bool) {
					if b {
						if err := LpacNotificationRemove(seq); err != nil {
							ErrDialog(err)
						}
						RefreshNotification()
					} else {
						RefreshNotification()
						return
					}
				}, WMain)
			d.Show()
		}
	})

	RemoveNotificationButton = widget.NewButton("Remove", func() {
		if SelectedNotification < 0 || SelectedNotification >= len(Notifications) {
			SelectItemDialog()
			return
		}
		if err := LpacNotificationRemove(Notifications[SelectedNotification].SeqNumber); err != nil {
			ErrDialog(err)
		}
		RefreshNotification()
	})

	FreeSpaceLabel = widget.NewLabel("")

	OpenLogButton = widget.NewButton("Open Log", OpenLog)
	OpenLogButton.SetIcon(theme.FolderOpenIcon())

	RefreshProfileButton = widget.NewButton("Refresh", func() {
		RefreshProfile()
	})
	RefreshProfileButton.SetIcon(theme.ViewRefreshIcon())

	RefreshNotificationButton = widget.NewButton("Refresh", func() {
		RefreshNotification()
	})
	RefreshNotificationButton.SetIcon(theme.ViewRefreshIcon())
}