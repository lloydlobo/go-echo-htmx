package handlers

type ContactService interface {
	CrudOps()
}

// func New(log *slog.Logger, cs ContactService)
func New(cs ContactService) *DefaultHandler {
	return &DefaultHandler{
		// Log: log,
		ContactService: cs,
	}

}

type DefaultHandler struct {
	// Log *slog.Logger
	ContactService ContactService
}
