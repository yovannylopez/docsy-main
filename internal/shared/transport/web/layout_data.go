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
}

// NewAppLayoutData builds layout defaults for app pages.
func NewAppLayoutData(title, subtitle, userName, activeRoute string) AppLayoutData {
	return AppLayoutData{
		Title:           title,
		Subtitle:        subtitle,
		UserName:        userName,
		AvatarURL:       "/static/assets/images/avatar-default.svg",
		ActiveRoute:     activeRoute,
		SidebarExpanded: true,
		AppVersion:      "1.0.0",
		Year:            time.Now().Year(),
	}
}
