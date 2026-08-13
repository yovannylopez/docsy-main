package web

import "time"

// AppLayoutData holds shared view data for authenticated app-layout pages.
type AppLayoutData struct {
	Title           string
	Subtitle        string
	UserName        string
	AvatarURL       string
	ActiveRoute     string
	SidebarExpanded bool
	AppVersion      string
	Year            int
	ThemeClass      string
	Storage         SidebarStorageData
	// AlertCount is the in-app notification badge (e.g. due documents).
	AlertCount int
}

// NewAppLayoutData builds layout defaults for app pages.
func NewAppLayoutData(title, subtitle, userName, activeRoute string) AppLayoutData {
	return AppLayoutData{
		Title:           title,
		Subtitle:        subtitle,
		UserName:        userName,
		AvatarURL:       "/static/assets/avatars/avatar-01.png",
		ActiveRoute:     activeRoute,
		SidebarExpanded: true,
		AppVersion:      "1.0.0",
		Year:            time.Now().Year(),
	}
}
