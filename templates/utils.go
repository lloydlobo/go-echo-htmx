package templates

import "context"

// Used by templates/components/title_templ.go
func GetPageTitle(ctx context.Context) string {
	title, ok := ctx.Value("pageTitle").(string)
	if !ok {
		return "Headcount | Home"
	}
	return title
}

func BoolToStrJS(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
