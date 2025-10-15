package router

type CreateJarRequest struct {
	Name string `json:"name"`
}

type DeleteJarRequest struct {
	ID string `json:"id"`
}

type DeleteRequestRequest struct {
	ID string `json:"id"`
}
