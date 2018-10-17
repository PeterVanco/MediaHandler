package api

type Photo struct {
	Url  string `json:"url" binding:"required"`
	Keep bool   `json:"keep" binding:"required"`
}

type RootDefinition struct {
	Root string `json:"root"`
}
